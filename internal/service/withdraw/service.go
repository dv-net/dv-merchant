package withdraw

import (
	"context"
	"errors"

	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dv-net/dv-merchant/internal/cache/settings"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/currency"
	"github.com/dv-net/dv-merchant/internal/service/exrate"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transactions"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transfers"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallet_addresses"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_withdrawal_from_processing_wallets"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_withdrawal_wallet_addresses"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/dv-net/dv-merchant/internal/util"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type IWithdrawServiceRunner interface {
	Run(context.Context, []models.Blockchain)
}

type IWithdrawService interface {
	IWithdrawServiceRunner
	ITransferService
	IWithdrawalService
	GetPrefetchWithdrawalAddress(ctx context.Context, user *models.User) ([]*models.PrefetchWithdrawAddressInfo, error)
}

const (
	CodeStatusNotEnoughResources   = 3000
	CodeStatusAddressIsTaken       = 3001
	CodeStatusBlockchainIsDisabled = 4000
	CodeStatusAddressEmptyBalance  = 4001
)

type service struct {
	transfersInProcess blockchainsInProcess
	storage            storage.IStorage
	logger             logger.Logger
	processing         processing.IProcessingTransfer
	processingWallet   processing.IProcessingWallet
	converter          currconv.ICurrencyConvertor
	currencyService    currency.ICurrency
	exRateService      exrate.IExRateSource
	settings           setting.ISettingService
}

var _ IWithdrawService = (*service)(nil)

func New(
	storage storage.IStorage,
	logger logger.Logger,
	processing processing.IProcessingTransfer,
	processingWallet processing.IProcessingWallet,
	converter currconv.ICurrencyConvertor,
	currencyService currency.ICurrency,
	exRateService exrate.IExRateSource,
	settingsSrv setting.ISettingService,
) IWithdrawService {
	return &service{
		transfersInProcess: blockchainsInProcess{
			mu:               sync.RWMutex{},
			blockchainsInUse: make(map[models.Blockchain]*atomic.Bool),
		},
		storage:          storage,
		logger:           logger,
		processing:       processing,
		processingWallet: processingWallet,
		converter:        converter,
		currencyService:  currencyService,
		exRateService:    exRateService,
		settings:         settingsSrv,
	}
}

type blockchainsInProcess struct {
	mu               sync.RWMutex
	blockchainsInUse map[models.Blockchain]*atomic.Bool
}

func (ts *blockchainsInProcess) get(b models.Blockchain) *atomic.Bool {
	ts.mu.RLock()
	sem, ok := ts.blockchainsInUse[b]
	ts.mu.RUnlock()

	if !ok {
		sem = &atomic.Bool{}
		ts.mu.Lock()
		ts.blockchainsInUse[b] = sem
		ts.mu.Unlock()
	}

	return sem
}

func (s *service) Run(ctx context.Context, blockchains []models.Blockchain) {
	interval := 2 * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			// Handle processing withdrawals with high priority
			s.processWithdrawalsFromProcessing(ctx)

			// Handle transfers with low amount
			s.processMultiTransfers(ctx)

			for _, blockchain := range blockchains {
				go s.processWithdrawalTransfers(ctx, blockchain)
			}
		}
	}
}

func (s *service) currencyRate(ctx context.Context, rateSource string, cur *models.Currency) (decimal.Decimal, error) {
	if cur.IsStablecoin {
		return decimal.NewFromInt(1), nil
	}
	rate, err := s.exRateService.GetCurrencyRate(ctx, rateSource, cur.Code, models.CurrencyCodeUSD)
	if err != nil {
		return decimal.Zero, fmt.Errorf("laod rate: %w", err)
	}

	decRate, err := decimal.NewFromString(rate)
	if err != nil {
		return decimal.Zero, fmt.Errorf("laod rate: %w", err)
	}

	return decRate, nil
}

func (s *service) processWithdrawalsFromProcessing(ctx context.Context) {
	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		processingWithdrawals, err := s.storage.WithdrawalsFromProcessing(repos.WithTx(tx)).GetQueuedWithdrawalsWithCurrencyAndUser(ctx)
		if err != nil {
			s.logger.Errorw("failed to fetch processing withdrawals", "error", err)
			return err
		}

		for _, withdrawal := range processingWithdrawals {
			if _, err = s.processProcessingWithdrawal(ctx, *withdrawal, tx); err != nil {
				s.logger.Errorw(
					"failed to process processing withdrawal",
					"error", err,
					"id", withdrawal.WithdrawalFromProcessingWallet.ID.String(),
				)

				continue
			}
		}

		return nil
	})
	if err != nil {
		s.logger.Errorw("processing withdrawals", "error", err)
	}
}

func (s *service) processWithdrawalTransfers(ctx context.Context, blockchain models.Blockchain) {
	locker := s.transfersInProcess.get(blockchain)
	if !locker.CompareAndSwap(false, true) {
		s.logger.Debugw("transfer iteration skipped by lock", "blockchain", blockchain)
		return
	}
	defer locker.Store(false)

	wallets, err := s.storage.WithdrawalWallets().GetWalletsForWithdrawal(ctx, blockchain)
	if err != nil {
		s.logger.Errorw("failed to get wallets for withdrawal", "error", err)
		return
	}

	for _, wallet := range wallets {
		transfer, err := s.processTransferByWallet(ctx, *wallet)
		if err != nil {
			if !isIgnoredLogError(err) {
				s.logger.Errorw("failed to process wallet", "error", err)
			}
			continue
		}

		if transfer != nil {
			s.logger.Infoln(
				"transfer initialized",
				"from", transfer.FromAddresses,
				"to", transfer.ToAddresses,
			)
		}
	}
}

func (s *service) processProcessingWithdrawal(
	ctx context.Context,
	row repo_withdrawal_from_processing_wallets.GetQueuedWithdrawalsWithCurrencyAndUserRow,
	tx pgx.Tx,
) (*models.Transfer, error) {
	if row.Currency.Blockchain == nil {
		return nil, ErrFiatCurrencyIsNotSupported
	}

	withdrawalSetting, err := s.settings.GetModelSetting(ctx, setting.WithdrawFromProcessing, settings.IModelSetting(&row.User))
	if err != nil || withdrawalSetting.Value == setting.FlagValueDisabled {
		return nil, ErrWithdrawalsFromProcessingDisabled
	}

	usdAmount, err := s.converter.Convert(
		ctx, currconv.ConvertDTO{
			Source:     row.User.RateSource.String(),
			From:       row.Currency.Code,
			To:         models.CurrencyCodeUSD,
			Amount:     row.WithdrawalFromProcessingWallet.Amount.String(),
			StableCoin: false,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("currency conversion: %w", err)
	}

	dto := TransferDto{
		ID:            uuid.New(),
		UserID:        row.User.ID,
		OwnerID:       row.User.ProcessingOwnerID.UUID,
		Kind:          models.TransferKindFromProcessing,
		FromAddresses: []string{row.WithdrawalFromProcessingWallet.AddressFrom},
		ToAddress:     row.WithdrawalFromProcessingWallet.AddressTo,
		Contract:      row.Currency.ContractAddress.String,
		Amount:        row.WithdrawalFromProcessingWallet.Amount,
		AmountUsd:     usdAmount,
		CurrencyID:    row.Currency.ID,
		Blockchain:    *row.Currency.Blockchain,
	}

	transfer, err := s.initializeTransfer(ctx, dto, &row.User, tx)
	if err != nil && !errors.Is(err, ErrTransfersDisabled) && !errors.Is(err, pgx.ErrNoRows) {
		return transfer, err
	}
	if transfer != nil {
		updateWithErr := s.storage.WithdrawalsFromProcessing(repos.WithTx(tx)).UpdateTransferID(
			ctx,
			repo_withdrawal_from_processing_wallets.UpdateTransferIDParams{
				ID:         row.WithdrawalFromProcessingWallet.ID,
				TransferID: transfer.ID,
				AmountUsd:  usdAmount,
			},
		)
		if updateWithErr != nil {
			return nil, updateWithErr
		}
	}

	return transfer, err
}

func (s *service) processTransferByWallet(ctx context.Context, wallet models.WithdrawalWallet) (*models.Transfer, error) {
	user, err := s.storage.Users().GetByID(ctx, wallet.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user by id %q: %w", wallet.UserID, err)
	}

	if err = s.checkTransferRequirementsByUser(ctx, user); err != nil {
		return nil, err
	}

	dto, err := s.prepareTransferDto(ctx, wallet, user, models.TransferKindFromAddress)
	if err != nil {
		return nil, fmt.Errorf("preapre transfer: %w", err)
	}

	s.logger.Infoln(
		"send processing request to transfer",
		"from", dto.FromAddresses,
		"to", dto.ToAddress,
		"amount", dto.Amount.String(),
		"amount_usd", dto.AmountUsd.String(),
	)

	return s.initializeTransfer(ctx, dto, user, nil)
}

func (s *service) prepareTransferDto(
	ctx context.Context,
	wallet models.WithdrawalWallet,
	user *models.User,
	kind models.TransferKind,
) (TransferDto, error) {
	rates, err := s.exRateService.LoadRatesList(ctx, user.RateSource.String())
	if err != nil {
		return TransferDto{}, fmt.Errorf("laod rates: %w", err)
	}

	params := repo_wallet_addresses.GetAddressForWithdrawalParams{
		UserID:       user.ID,
		Amount:       wallet.WithdrawalMinBalance.Decimal,
		CurrencyID:   wallet.CurrencyID,
		CurrencyIds:  rates.CurrencyIDs,
		CurrencyRate: rates.Rate,
		Blockchain:   wallet.Blockchain,
	}
	if wallet.WithdrawalMinBalanceUsd.Valid {
		params.MinUsd = wallet.WithdrawalMinBalanceUsd.Decimal
	}

	targetHotWallet, err := s.storage.WalletAddresses().GetAddressForWithdrawal(ctx, params)
	if err != nil {
		return TransferDto{}, fmt.Errorf("get hot wallet user:%q %w", wallet.UserID, err)
	}

	withdrawalAddrList, err := s.storage.WithdrawalWalletAddresses().GetAddressesList(ctx, wallet.ID)
	if err != nil || len(withdrawalAddrList) == 0 {
		return TransferDto{}, ErrWithdrawalAddressListEmpty
	}

	return TransferDto{
		ID:            uuid.New(),
		UserID:        user.ID,
		OwnerID:       user.ProcessingOwnerID.UUID,
		Kind:          kind,
		FromAddresses: []string{targetHotWallet.Address},
		ToAddress:     s.prepareWithdrawalAddressByUsed(ctx, targetHotWallet.Address, withdrawalAddrList),
		Contract:      targetHotWallet.ContractAddress.String,
		Amount:        targetHotWallet.Amount,
		AmountUsd:     targetHotWallet.AmountUsd,
		CurrencyID:    targetHotWallet.CurrencyID,
		Blockchain:    targetHotWallet.Blockchain,
	}, nil
}

func (s *service) initializeTransfer(
	ctx context.Context,
	dto TransferDto,
	user *models.User,
	tx pgx.Tx,
) (*models.Transfer, error) {
	transferStatus := models.TransferStatusNew
	params := processing.FundsWithdrawalParams{
		OwnerID:         dto.OwnerID,
		RequestID:       dto.ID,
		Blockchain:      dto.Blockchain,
		FromAddress:     dto.FromAddresses,
		ToAddress:       []string{dto.ToAddress},
		ContractAddress: dto.Contract,
		WholeAmount:     true,
	}

	if dto.Kind == models.TransferKindFromProcessing {
		params.Amount = dto.Amount.String()
		params.WholeAmount = false
	}
	// TODO make kind optional for any blockchain
	if dto.Blockchain.KindWithdrawalRequired() {
		tronTransferType, err := s.settings.GetModelSetting(ctx, setting.TransferType, setting.IModelSetting(user))
		if err != nil || tronTransferType == nil {
			params.Kind = util.Pointer(string(setting.TransferByBurnTRX))
		} else {
			params.Kind = util.Pointer(tronTransferType.Value)
		}
	}

	var errMessage *string

	_, processingErr := s.processing.FundsWithdrawal(ctx, params)

	if processingErr != nil {
		if connect.CodeOf(processingErr) == connect.CodeUnavailable {
			return nil, ErrProcessingUnavailable
		}
		rpcCode, ok := processing.ErrorRPCCode(processingErr)
		if connect.CodeOf(processingErr) == connect.CodeDeadlineExceeded ||
			(ok && (rpcCode >= CodeStatusNotEnoughResources && rpcCode <= CodeStatusBlockchainIsDisabled)) {
			s.logger.Debugln(
				"processing code status",
				"grpc_code", connect.CodeOf(processingErr).String(),
				"rpc_code", rpcCode,
			)
			if rpcCode == CodeStatusBlockchainIsDisabled {
				return nil, fmt.Errorf("blockchain [%s] is disabled: %w", params.Blockchain, ErrProcessingExplorerUnavailable)
			}

			return nil, processingErr
		}

		transferStatus = models.TransferStatusFailed
		errMessage = util.Pointer(processingErr.Error())
	}

	transfer, err := s.storage.Transfers(repos.WithTx(tx)).Create(ctx, repo_transfers.CreateParams{
		ID:            dto.ID,
		UserID:        dto.UserID,
		Kind:          dto.Kind,
		CurrencyID:    dto.CurrencyID,
		Status:        transferStatus,
		Stage:         models.ResolveTransferStageByStatus(transferStatus),
		FromAddresses: dto.FromAddresses,
		ToAddresses:   []string{dto.ToAddress},
		Amount:        dto.Amount,
		AmountUsd:     dto.AmountUsd,
		Blockchain:    dto.Blockchain,
		Message:       errMessage,
	})
	if err != nil {
		return nil, fmt.Errorf("transfer creation: %w", err)
	}

	return transfer, nil
}

func (s *service) prepareWithdrawalAddressByUsed(
	ctx context.Context,
	preparedAddressFrom string,
	withdrawalAddrList []string,
) string {
	addr, err := s.storage.Transactions().GetExistingWithdrawalAddress(
		ctx,
		repo_transactions.GetExistingWithdrawalAddressParams{
			FromAddr:          preparedAddressFrom,
			WithdrawAddresses: withdrawalAddrList,
		},
	)

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return tools.RandomSliceElement[string](withdrawalAddrList)
	}

	return addr
}

func (s *service) processMultiTransfers(ctx context.Context) {
	wallets, err := s.storage.WithdrawalWallets().GetForMultiWithdrawal(ctx, uuid.NullUUID{})
	if err != nil {
		s.logger.Errorw("failed to get withdrawal wallets", "error", err)
		return
	}

	for _, wallet := range wallets {
		if err = s.checkTransferRequirementsByUser(ctx, &wallet.User); err != nil {
			if !errors.Is(err, ErrTransfersDisabled) && !errors.Is(err, pgx.ErrNoRows) && !errors.Is(err, ErrWithdrawalsFromProcessingDisabled) {
				s.logger.Errorw("failed to process wallet", "error", err)
			}

			continue
		}

		rates, err := s.exRateService.LoadRatesList(ctx, wallet.User.RateSource.String())
		if err != nil {
			s.logger.Errorw("failed to load rates list", "error", err)
			continue
		}

		var preparedWallet string
		preparedWallet, err = s.prepareWalletToByMode(ctx, &wallet.User, wallet.MultiWithdrawalRule, wallet.Currency, wallet.Addresses)
		if err != nil {
			s.logger.Debugw("failed to prepare wallet", "error", err)
			continue
		}

		hotWalletData, err := s.storage.WalletAddresses().GetAddressForMultiWithdrawal(
			ctx,
			repo_wallet_addresses.GetAddressForMultiWithdrawalParams{
				CurrencyIds:  rates.CurrencyIDs,
				CurrencyRate: rates.Rate,
				UserID:       wallet.User.ID,
				Currency:     wallet.Currency.ID,
				Blockchain:   *wallet.Currency.Blockchain,
				MinUsd:       wallet.WithdrawalWallet.WithdrawalMinBalanceUsd,
				MinAmount:    wallet.WithdrawalWallet.WithdrawalMinBalance,
			},
		)
		if err != nil {
			s.logger.Debugw("failed to get wallet addresses", "error", err)
			continue
		}

		dto := TransferDto{
			ID:            uuid.New(),
			UserID:        wallet.User.ID,
			OwnerID:       wallet.User.ProcessingOwnerID.UUID,
			Kind:          models.TransferKindFromAddress,
			FromAddresses: hotWalletData.Addresses,
			ToAddress:     preparedWallet,
			Contract:      wallet.Currency.ContractAddress.String,
			Amount:        hotWalletData.TotalAmount,
			AmountUsd:     hotWalletData.AmountUsd,
			CurrencyID:    wallet.Currency.ID,
			Blockchain:    *wallet.Currency.Blockchain,
		}

		transfer, err := s.initializeTransfer(ctx, dto, &wallet.User, nil)
		if err != nil {
			s.logger.Errorw("failed to initialize transfer", "error", err)
			continue
		}

		if transfer != nil {
			s.logger.Infow(
				"multi transfer initialized",
				"from", transfer.FromAddresses,
				"to", transfer.ToAddresses,
			)
		}
	}
}

func (s *service) prepareWalletToByMode(
	ctx context.Context,
	u *models.User,
	rule models.MultiWithdrawalRule,
	curr models.Currency,
	withdrawalAddresses []string,
) (string, error) {
	switch rule.Mode {
	case models.MultiWithdrawalModeManual:
		if !rule.ManualAddress.Valid {
			return "", errors.New("manual address for multiple withdrawal is not set")
		}

		exists, err := s.storage.WithdrawalWalletAddresses().CheckAddressExists(
			ctx,
			repo_withdrawal_wallet_addresses.CheckAddressExistsParams{
				WithdrawalWalletID: rule.WithdrawalWalletID,
				Address:            rule.ManualAddress.String,
			},
		)

		if !exists || err != nil {
			return "", fmt.Errorf("failed to find manual withdrawal wallet in approved addresses")
		}

		return rule.ManualAddress.String, nil
	case models.MultiWithdrawalModeProcessing:
		res, err := s.processingWallet.GetOwnerProcessingWallet(ctx, processing.GetOwnerProcessingWalletsParams{
			OwnerID:    u.ProcessingOwnerID.UUID,
			Blockchain: curr.Blockchain,
		})
		if err != nil {
			return "", err
		}

		return res.Address, nil
	case models.MultiWithdrawalModeRandom:
		if len(withdrawalAddresses) == 0 {
			return "", errors.New("withdrawal addresses list is empty")
		}
		return tools.RandomSliceElement(withdrawalAddresses), nil
	default:
		return "", fmt.Errorf("mode '%s' is not supported", rule.Mode)
	}
}

func (s *service) checkTransferRequirementsByUser(ctx context.Context, u *models.User) error {
	if !u.ProcessingOwnerID.Valid { // TODO check user permissions for withdraw
		return errors.New("user has no processing owner")
	}

	transfersStatusFlag, err := s.settings.GetModelSetting(ctx, setting.TransfersStatus, setting.IModelSetting(u))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("fetch owner settings: %w", err)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrTransfersDisabled
	}

	// transfers disabled by user setting
	if transfersStatusFlag != nil && transfersStatusFlag.Value != setting.FlagValueEnabled {
		s.logger.Debugw("transfers disabled", "user_id ", u.ID.String())
		return ErrTransfersDisabled
	}

	return nil
}
