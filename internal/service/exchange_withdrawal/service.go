package exchange_withdrawal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/exchange_request"
	"github.com/dv-net/dv-merchant/internal/dto"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/exchange_manager"
	"github.com/dv-net/dv-merchant/internal/service/exchange_rules"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_chains"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_withdrawal_history"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_withdrawal_settings"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_exchanges"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/internal/util"
	binancemodels "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/models"
	bitgetmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/models"
	bybitmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/bybit/models"
	gateio "github.com/dv-net/dv-merchant/pkg/exchange_client/gate"
	htxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/models"
	kucoinmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/models"
	okxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/models"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/dv-net/dv-processing/pkg/avalidator"

	"github.com/go-mods/excel"
	"github.com/gocarina/gocsv"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

type IExchangeWithdrawalService interface {
	RunWithdrawalQueue(ctx context.Context)
	RunWithdrawalUpdater(ctx context.Context)
	CreateWithdrawalSetting(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug, d *dto.CreateWithdrawalSettingDTO) (*models.ExchangeWithdrawalSetting, error)
	UpdateWithdrawalSetting(ctx context.Context, userID uuid.UUID, settingID uuid.UUID, isEnabled bool) (*models.ExchangeWithdrawalSetting, error)
	DeleteWithdrawalSetting(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug, settingID uuid.UUID) error
	GetWithdrawalSettings(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) ([]*models.ExchangeWithdrawalSetting, error)
	GetWithdrawalSetting(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug, settingID uuid.UUID) (*models.ExchangeWithdrawalSetting, error)
	GetWithdrawalHistory(ctx context.Context, userID uuid.UUID, request *exchange_request.GetWithdrawalsRequest) (*storecmn.FindResponseWithFullPagination[*models.ExchangeWithdrawalHistoryDTO], error)
	DownloadWithdrawalHistory(ctx context.Context, userID uuid.UUID, request *exchange_request.GetWithdrawalsExportedRequest) (*bytes.Buffer, error)
	GetWithdrawalByID(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug, recordID uuid.UUID) (*models.ExchangeWithdrawalHistoryDTO, error)
}

type Service struct {
	logger      logger.Logger
	st          storage.IStorage
	exManager   exchange_manager.IExchangeManager
	currConvSvc currconv.ICurrencyConvertor
	exRulesSvc  exchange_rules.IExchangeRules
	settingSvc  setting.ISettingService
}

func (s *Service) DeleteWithdrawalSetting(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug, settingID uuid.UUID) error {
	err := repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		exchangeID, err := s.st.Exchanges().GetExchangeBySlug(ctx, slug)
		if err != nil {
			return fmt.Errorf("fetch exchange: %w", err)
		}
		err = s.st.ExchangeWithdrawalSettings(repos.WithTx(tx)).Delete(ctx, repo_exchange_withdrawal_settings.DeleteParams{
			UserID:     userID,
			ExchangeID: exchangeID.ID,
			ID:         settingID,
		})
		if err != nil {
			return fmt.Errorf("delete withdrawal setting: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) GetWithdrawalSetting(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug, settingID uuid.UUID) (*models.ExchangeWithdrawalSetting, error) {
	exchangeID, err := s.st.Exchanges().GetExchangeBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	settings, err := s.st.ExchangeWithdrawalSettings().GetByID(ctx, repo_exchange_withdrawal_settings.GetByIDParams{
		UserID:     userID,
		ExchangeID: exchangeID.ID,
		ID:         settingID,
	})
	if err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *Service) GetWithdrawalSettings(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug) ([]*models.ExchangeWithdrawalSetting, error) {
	exchangeID, err := s.st.Exchanges().GetExchangeBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	settings, err := s.st.ExchangeWithdrawalSettings().GetAllByUser(ctx, repo_exchange_withdrawal_settings.GetAllByUserParams{
		UserID:     userID,
		ExchangeID: exchangeID.ID,
	})
	if err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *Service) GetWithdrawalByID(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug, recordID uuid.UUID) (*models.ExchangeWithdrawalHistoryDTO, error) {
	userExchange, err := s.st.Exchanges().GetExchangeBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("get user exchange: %w", err)
	}

	res, err := s.st.ExchangeWithdrawalHistory().GetByUserAndOrderID(ctx, repo_exchange_withdrawal_history.GetByUserAndOrderIDParams{
		UserID:     userID,
		ID:         recordID,
		ExchangeID: userExchange.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("get withdrawal by user and order id: %w", err)
	}
	return &models.ExchangeWithdrawalHistoryDTO{
		ID:              res.ID,
		UserID:          res.UserID,
		ExchangeID:      res.ExchangeID,
		ExchangeOrderID: res.ExchangeOrderID,
		Address:         res.Address,
		Currency:        res.Currency,
		Chain:           res.Chain,
		Status:          res.Status,
		Txid:            res.Txid,
		CreatedAt:       res.CreatedAt,
		UpdatedAt:       res.UpdatedAt,
		Slug:            res.Slug,
		NativeAmount:    res.NativeAmount,
		FiatAmount:      res.FiatAmount,
	}, nil
}

func (s *Service) GetWithdrawalHistory(ctx context.Context, userID uuid.UUID, request *exchange_request.GetWithdrawalsRequest) (*storecmn.FindResponseWithFullPagination[*models.ExchangeWithdrawalHistoryDTO], error) {
	var exchangeID uuid.NullUUID
	if request.Slug != nil {
		exID, err := s.st.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlug(*request.Slug))
		if err != nil {
			return nil, fmt.Errorf("fetch exchange: %w", err)
		}
		exchangeID = uuid.NullUUID{
			UUID:  exID.ID,
			Valid: true,
		}
	}

	commonParams := storecmn.NewCommonFindParams()
	if request.PageSize != nil {
		commonParams.PageSize = request.PageSize
	}
	if request.Page != nil {
		commonParams.Page = request.Page
	}

	params := repo_exchange_withdrawal_history.GetWithdrawalHistoryByUserAndExchangeIDParams{
		UserID:           userID,
		ExchangeID:       exchangeID,
		CommonFindParams: *commonParams,
	}

	if request.CurrencyID != nil {
		params.Currency = *request.CurrencyID
	}

	if request.DateFrom != nil {
		dateFrom, err := util.ParseDate(*request.DateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from format: %w", err)
		}
		params.DateFrom = dateFrom
	}

	if request.DateTo != nil {
		dateTo, err := util.ParseDate(*request.DateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to format: %w", err)
		}
		params.DateTo = dateTo
	}

	return s.st.ExchangeWithdrawalHistory().GetAllByUserAndExchangeID(ctx, params)
}

func (s *Service) DownloadWithdrawalHistory(ctx context.Context, userID uuid.UUID, request *exchange_request.GetWithdrawalsExportedRequest) (*bytes.Buffer, error) {
	var exchangeID uuid.NullUUID
	if request.Slug != nil {
		exID, err := s.st.Exchanges().GetExchangeBySlug(ctx, models.ExchangeSlug(*request.Slug))
		if err != nil {
			return nil, fmt.Errorf("fetch exchange: %w", err)
		}
		exchangeID = uuid.NullUUID{
			UUID:  exID.ID,
			Valid: true,
		}
	}

	params := repo_exchange_withdrawal_history.GetWithdrawalExportHistoryByUserAndExchangeIDParams{
		UserID:     userID,
		ExchangeID: exchangeID,
	}

	if request.CurrencyID != nil {
		params.Currency = *request.CurrencyID
	}

	if request.DateFrom != nil {
		dateFrom, err := time.Parse(time.DateTime, *request.DateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from format: %w", err)
		}
		params.DateFrom = &dateFrom
	}

	if request.DateTo != nil {
		dateTo, err := time.Parse(time.DateTime, *request.DateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to format: %w", err)
		}
		params.DateTo = &dateTo
	}

	res, err := s.st.ExchangeWithdrawalHistory().GetAllByUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get withdrawal history: %w", err)
	}

	history, err := s.prepareWithdrawalHistoryForExport(ctx, res)
	if err != nil {
		return nil, fmt.Errorf("prepare withdrawal history for export: %w", err)
	}
	txBuffer := new(bytes.Buffer)
	switch request.Format {
	case "csv":
		if err := gocsv.Marshal(history, txBuffer); err != nil {
			return nil, fmt.Errorf("marshal withdrawal history: %w", err)
		}
	case "xlsx":
		excelFile := excelize.NewFile()
		if err := excelFile.SetSheetName(excelFile.GetSheetName(excelFile.GetActiveSheetIndex()), "Withdrawal history"); err != nil {
			return nil, fmt.Errorf("set active sheet name: %w", err)
		}
		defer func() { _ = excelFile.Close() }()
		excelWriter, err := excel.NewWriter(excelFile)
		if err != nil {
			return nil, fmt.Errorf("create excel writer: %w", err)
		}
		if err := excelWriter.SetActiveSheetName("Withdrawal history"); err != nil {
			return nil, fmt.Errorf("set active sheet name: %w", err)
		}
		if err := excelWriter.Marshal(&history); err != nil {
			return nil, fmt.Errorf("marshal withdrawal history: %w", err)
		}
		if _, err := excelWriter.File.WriteTo(txBuffer); err != nil {
			return nil, fmt.Errorf("write to buffer: %w", err)
		}
	}
	return txBuffer, nil
}

func (s *Service) prepareWithdrawalHistoryForExport(ctx context.Context, history []*models.ExchangeWithdrawalHistoryDTO) ([]*UserWithdrawalHistoryModel, error) {
	result := make([]*UserWithdrawalHistoryModel, 0, len(history))
	for _, item := range history {
		userExchange, err := s.st.Exchanges().GetByID(ctx, item.ExchangeID)
		if err != nil {
			return nil, fmt.Errorf("get user exchange by id: %w", err)
		}
		historyModel := &UserWithdrawalHistoryModel{
			ExchangeName: userExchange.Name,
			ExchangeID:   item.ExchangeID.String(),
			ExchangeSlug: item.Slug.String(),
			Address:      item.Address,
			Currency:     item.Currency,
			Chain:        item.Chain,
			Status:       item.Status.String(),
			Txid:         item.Txid.String,
			CreatedAt:    item.CreatedAt.Time,
		}
		if item.NativeAmount.Valid {
			historyModel.NativeAmount = item.NativeAmount.Decimal.String()
		}
		if item.FiatAmount.Valid {
			historyModel.FiatAmount = item.FiatAmount.Decimal.String()
		}
		if item.ExchangeOrderID.Valid {
			historyModel.ExchangeOrderID = item.ExchangeOrderID.String
		}
		result = append(result, historyModel)
	}
	return result, nil
}

func NewService(
	logger logger.Logger,
	st storage.IStorage,
	exManager exchange_manager.IExchangeManager,
	currConvSvc currconv.ICurrencyConvertor,
	exRulesSvc exchange_rules.IExchangeRules,
	settingSvc setting.ISettingService,
) IExchangeWithdrawalService {
	return &Service{
		logger:      logger,
		st:          st,
		exManager:   exManager,
		currConvSvc: currConvSvc,
		exRulesSvc:  exRulesSvc,
		settingSvc:  settingSvc,
	}
}

func (s *Service) CreateWithdrawalSetting(ctx context.Context, userID uuid.UUID, slug models.ExchangeSlug, d *dto.CreateWithdrawalSettingDTO) (*models.ExchangeWithdrawalSetting, error) {
	var withdrawalSetting *models.ExchangeWithdrawalSetting

	_, err := s.st.ExchangeChains().GetCurrencyIDBySlugAndChain(ctx, repo_exchange_chains.GetCurrencyIDBySlugAndChainParams{
		Slug:  slug,
		Chain: d.Chain,
	})
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("chain %s not found", d.Chain)
	}

	wdRule, err := s.exRulesSvc.GetWithdrawalRule(ctx, slug, userID.String(), d.CurrencyID)
	if err != nil {
		return nil, fmt.Errorf("get withdrawal rule: %w", err)
	}

	if ok, err := s.checkMinWithdrawalRequirement(wdRule, d); !ok {
		return nil, fmt.Errorf("min withdrawal requirement not met: %w", err)
	}

	err = repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		exID, err := s.st.Exchanges(repos.WithTx(tx)).GetExchangeBySlug(ctx, slug)
		if err != nil {
			return fmt.Errorf("fetch exchange: %w", err)
		}

		enabledCurrencies, err := s.st.ExchangeChains(repos.WithTx(tx)).GetEnabledCurrencies(ctx, slug)
		if err != nil {
			return fmt.Errorf("fetch enabled currencies: %w", err)
		}
		if !slices.ContainsFunc(enabledCurrencies, func(c *repo_exchange_chains.GetEnabledCurrenciesRow) bool {
			return c.ID.String == d.CurrencyID
		}) {
			return fmt.Errorf("currency not enabled: %s", d.CurrencyID)
		}

		existingSetting, err := s.st.ExchangeWithdrawalSettings(repos.WithTx(tx)).GetExisting(ctx, repo_exchange_withdrawal_settings.GetExistingParams{
			UserID:     userID,
			ExchangeID: exID.ID,
			Currency:   d.CurrencyID,
			Chain:      d.Chain,
		})
		if err == nil {
			return fmt.Errorf("user %s already has %s withdrawalSetting setup", existingSetting.UserID, existingSetting.Currency)
		}
		if err != nil && errors.Is(err, pgx.ErrNoRows) {
			newSetting, err := s.createWithdrawalSetting(ctx, exID.ID, userID, d, repos.WithTx(tx))
			if err != nil {
				if errors.Is(err, ErrInvalidAddress) {
					return err
				}
				return fmt.Errorf("create withdrawal withdrawalSetting: %w", err)
			}
			withdrawalSetting = newSetting
			return nil
		}
		if err != nil {
			return fmt.Errorf("fetch latest withdrawalSetting: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return withdrawalSetting, nil
}

func (s *Service) createWithdrawalSetting(ctx context.Context, exchangeID uuid.UUID, userID uuid.UUID, d *dto.CreateWithdrawalSettingDTO, opts ...repos.Option) (*models.ExchangeWithdrawalSetting, error) {
	minAmt, err := decimal.NewFromString(d.MinAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to convert min amount: %w", err)
	}
	createParams := repo_exchange_withdrawal_settings.CreateParams{
		UserID:     userID,
		ExchangeID: exchangeID,
		Currency:   d.CurrencyID,
		Chain:      d.Chain,
		Address:    d.Address,
		MinAmount:  minAmt,
	}
	cur, err := s.st.Currencies().GetByID(ctx, d.CurrencyID)
	if err != nil {
		return nil, fmt.Errorf("get currency by id: %w", err)
	}
	if !avalidator.ValidateAddressByBlockchain(d.Address, cur.Blockchain.String()) {
		return nil, ErrInvalidAddress
	}
	record, err := s.st.ExchangeWithdrawalSettings(opts...).Create(ctx, createParams)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (s *Service) updateWithdrawalSettingAddress(ctx context.Context, settingID uuid.UUID, userID uuid.UUID, address string, opts ...repos.Option) error { //nolint:unused
	err := s.st.ExchangeWithdrawalSettings(opts...).ChangeAddress(ctx, repo_exchange_withdrawal_settings.ChangeAddressParams{
		ID:      settingID,
		UserID:  userID,
		Address: address,
	})
	if err != nil {
		return fmt.Errorf("failed to update withdrawal setting address: %w", err)
	}
	return nil
}

func (s *Service) updateWithdrawalSettingMinAmount(ctx context.Context, settingID uuid.UUID, userID uuid.UUID, minAmount decimal.Decimal, opts ...repos.Option) error { //nolint:unused
	err := s.st.ExchangeWithdrawalSettings(opts...).ChangeMinAmount(ctx, repo_exchange_withdrawal_settings.ChangeMinAmountParams{
		ID:        settingID,
		UserID:    userID,
		MinAmount: minAmount,
	})
	if err != nil {
		return fmt.Errorf("failed to update withdrawal setting minAmount: %w", err)
	}
	return nil
}

func (s *Service) checkMinWithdrawalRequirement(rule *models.WithdrawalRulesDTO, d *dto.CreateWithdrawalSettingDTO) (bool, error) {
	minWdAmount, err := decimal.NewFromString(rule.MinWithdrawAmount)
	if err != nil {
		return false, fmt.Errorf("parse min withdrawal amount: %w", err)
	}
	minAmount, err := decimal.NewFromString(d.MinAmount)
	if err != nil {
		return false, fmt.Errorf("parse request withdrawal amount: %w", err)
	}

	if minAmount.LessThan(minWdAmount) {
		return false, fmt.Errorf("min withdrawal amount is %s", minWdAmount)
	}
	return true, nil
}

func (s *Service) createWithdrawalHistoryRecord(ctx context.Context, userID uuid.UUID, exID uuid.UUID, address string, currency string, chain string, connHash string, opts ...repos.Option) (*uuid.UUID, error) {
	params := repo_exchange_withdrawal_history.CreateParams{
		UserID:     userID,
		ExchangeID: exID,
		Address:    address,
		Status:     models.WithdrawalHistoryStatusNew,
		Chain:      chain,
		Currency:   currency,
		ExchangeConnectionHash: pgtype.Text{
			String: connHash,
			Valid:  true,
		},
	}

	wdHistoryID, err := s.st.ExchangeWithdrawalHistory(opts...).Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create withdrawal order: %w", err)
	}

	return &wdHistoryID.ID, nil
}

func (s *Service) processWithdrawals(ctx context.Context, userID uuid.UUID, settings []*models.ExchangeWithdrawalSetting) { //nolint:gocognit,funlen,gocyclo
	for _, setting := range settings {
		wdState, err := s.getWithdrawalState(ctx, userID, setting.ExchangeID)
		if err != nil {
			s.logger.Error("failed to fetch exchange withdrawal state", err, "userID", userID)
			continue
		}

		if *wdState == models.ExchangeWithdrawalStateDisabled {
			s.logger.Debug("skipping withdrawals", "userID", userID)
			continue
		}

		if !setting.IsEnabled {
			s.logger.Debug("skipped disabled withdrawal", "userID", userID, "withdrawalID", setting.ID)
			continue
		}

		userExchange, err := s.st.Exchanges().GetByID(ctx, setting.ExchangeID)
		if err != nil {
			s.logger.Error("failed to get user exchange by id", err)
			return
		}
		userExchangeClient, err := s.exManager.GetDriver(ctx, userExchange.Slug, userID)
		if err != nil {
			s.logger.Error("failed to get user exchange client", err)
			return
		}

		err = repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
			lastOrder, err := s.st.ExchangeWithdrawalHistory(repos.WithTx(tx)).GetLast(ctx, userID, userExchange.ID)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			if lastOrder != nil {
				return ErrWithdrawalPending
			}
			recordID, err := s.createWithdrawalHistoryRecord(ctx, userID, userExchange.ID, setting.Address, setting.Currency, setting.Chain, userExchangeClient.GetConnectionHash(), repos.WithTx(tx))
			if err != nil {
				return err
			}
			ticker, err := s.st.ExchangeChains().GetTickerByCurrencyID(ctx, repo_exchange_chains.GetTickerByCurrencyIDParams{
				Slug:       userExchange.Slug,
				CurrencyID: setting.Currency,
			})
			if err != nil {
				return err
			}
			wdRule, err := s.exRulesSvc.GetWithdrawalRule(ctx, userExchange.Slug, setting.UserID.String(), setting.Currency)
			if err != nil {
				return err
			}
			wdPrecision, err := strconv.Atoi(wdRule.WithdrawPrecision)
			if err != nil {
				return err
			}
			wdOrderParams := &models.CreateWithdrawalOrderParams{
				Address:             setting.Address,
				Currency:            setting.Currency,
				Chain:               setting.Chain,
				WithdrawalPrecision: wdPrecision,
				RecordID:            recordID,
			}
			if wdRule.Fee != "" {
				feeAmount, err := decimal.NewFromString(wdRule.Fee)
				if err != nil {
					return err
				}
				wdOrderParams.Fee = feeAmount
			}
			tokenBalance, err := userExchangeClient.GetCurrencyBalance(ctx, ticker)
			if err != nil {
				return err
			}
			if tokenBalance.LessThan(setting.MinAmount) {
				return ErrThresholdNotMet
			}
			wdOrderParams.NativeAmount = *tokenBalance
			// wdOrderParams.NativeAmount = wdOrderParams.NativeAmount.Sub(wdOrderParams.Fee) // fixme

			minWdAmt, err := decimal.NewFromString(wdRule.MinWithdrawAmount)
			if err != nil {
				return err
			}
			minWdAmt = minWdAmt.Add(wdOrderParams.Fee)
			wdOrderParams.MinWithdrawal = minWdAmt

			if wdOrderParams.NativeAmount.IsNegative() {
				return fmt.Errorf("withdrawal order negative amount (minimum amount with fee)")
			}
			if wdOrderParams.NativeAmount.LessThan(minWdAmt) {
				return fmt.Errorf("withdrawal order amount is less then minimum withdrawal amount with fee")
			}
			orderAmountFiat, err := s.currConvSvc.Convert(ctx, currconv.ConvertDTO{
				Source:     userExchange.Slug.String(),
				From:       ticker,
				To:         "USDT",
				Amount:     wdOrderParams.NativeAmount.Sub(wdOrderParams.Fee).String(),
				StableCoin: false,
			})
			if err != nil {
				return err
			}
			wdOrderParams.FiatAmount = orderAmountFiat

			updateParams := repo_exchange_withdrawal_history.UpdateParams{
				ID: *recordID,
				ExchangeConnectionHash: pgtype.Text{
					Valid:  true,
					String: userExchangeClient.GetConnectionHash(),
				},
			}

			s.logger.Info("creating withdrawal order",
				"userID", userID,
				"recordID", recordID.String(),
				"exchange", userExchange.Slug.String(),
				"currency", setting.Currency,
				"totalBalance", tokenBalance.String(),
				"withdrawalAmount", wdOrderParams.NativeAmount.String(),
				"fiatWithdrawalAmount", wdOrderParams.FiatAmount.String(),
				"withdrawalFee", wdOrderParams.Fee.String(),
				"exchangeConnectionHash", userExchangeClient.GetConnectionHash(),
			)

			var orderData *models.ExchangeWithdrawalDTO

			orderData, err = userExchangeClient.CreateWithdrawalOrder(ctx, wdOrderParams)
			if err != nil {
				if strings.Contains(err.Error(), "withdrawal balance locked") {
					s.logger.Info("insufficient balance due to withdrawal block count lock",
						"userID", userID,
						"recordID", recordID.String(),
						"exchange", userExchange.Slug.String(),
						"currency", setting.Currency,
						"totalBalance", tokenBalance.String(),
						"withdrawalAmount", wdOrderParams.NativeAmount.String(),
						"fiatWithdrawalAmount", wdOrderParams.FiatAmount.String(),
						"withdrawalFee", wdOrderParams.Fee.String(),
					)
					return ErrWithdrawalBalanceLocked
				}
				if strings.Contains(err.Error(), "withdrawal threshold not met") {
					s.logger.Info("insufficient balance due to withdrawal minimum",
						"userID", userID,
						"recordID", recordID.String(),
						"exchange", userExchange.Slug.String(),
						"currency", setting.Currency,
						"totalBalance", tokenBalance.String(),
						"withdrawalAmount", wdOrderParams.NativeAmount.String(),
						"fiatWithdrawalAmount", wdOrderParams.FiatAmount.String(),
						"withdrawalFee", wdOrderParams.Fee.String(),
					)
					return ErrThresholdNotMet
				}
				if strings.Contains(err.Error(), "temporary disabled due to user security action") {
					s.logger.Info("cannot withdrawal due to user recent security actions",
						"userID", userID,
						"recordID", recordID.String(),
						"exchange", userExchange.Slug.String(),
						"currency", setting.Currency,
						"totalBalance", tokenBalance.String(),
						"withdrawalAmount", wdOrderParams.NativeAmount.String(),
						"fiatWithdrawalAmount", wdOrderParams.FiatAmount.String(),
						"withdrawalFee", wdOrderParams.Fee.String(),
					)
					return ErrSoftLockByUserSecurityAction
				}
				if strings.Contains(err.Error(), "withdrawal address not whitelisted") {
					s.logger.Info("cannot withdrawal due to address not whitelisted",
						"userID", userID,
						"recordID", recordID.String(),
						"exchange", userExchange.Slug.String(),
						"currency", setting.Currency,
						"totalBalance", tokenBalance.String(),
						"withdrawalAmount", wdOrderParams.NativeAmount.String(),
						"fiatWithdrawalAmount", wdOrderParams.FiatAmount.String(),
						"withdrawalFee", wdOrderParams.Fee.String(),
					)
					return ErrWithdrawalAddessNotWhitelisted
				}
				s.logger.Warn("failed to create withdrawal order",
					"error", err.Error(),
					"userID", userID,
					"recordID", recordID.String(),
					"exchange", userExchange.Slug.String(),
				)
				updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusFailed.String()}
				updateParams.FailReason = pgtype.Text{Valid: true, String: err.Error()}

				return s.updateExchangeWithdrawal(ctx, userID, updateParams, tx)
			}

			// InternalOrderID is the order id generate by us, since this is the only way to identify the order by API
			// w/o querying the deposit-withdrawal history
			updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusInProgress.String()}
			if orderData.ExternalOrderID != "" {
				updateParams.ExchangeOrderID = pgtype.Text{Valid: true, String: orderData.ExternalOrderID}
			}
			if orderData.InternalOrderID != "" {
				updateParams.ExchangeOrderID = pgtype.Text{Valid: true, String: orderData.InternalOrderID}
			}
			updateParams.NativeAmount = decimal.NullDecimal{Valid: true, Decimal: wdOrderParams.NativeAmount}
			updateParams.FiatAmount = decimal.NullDecimal{Valid: true, Decimal: wdOrderParams.FiatAmount}
			if orderData.RetryReason != "" {
				updateParams.FailReason = pgtype.Text{Valid: true, String: orderData.RetryReason}
			}

			return s.updateExchangeWithdrawal(ctx, userID, updateParams, tx)
		})
		if err != nil {
			if errors.Is(err, ErrWithdrawalPending) ||
				errors.Is(err, ErrWithdrawalBalanceLocked) ||
				errors.Is(err, ErrThresholdNotMet) ||
				errors.Is(err, ErrSoftLockByUserSecurityAction) ||
				errors.Is(err, ErrWithdrawalAddessNotWhitelisted) {
				continue
			}
			if errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.ETIMEDOUT) {
				continue
			}
			s.logger.Error("failed to process withdrawal", err)
		}
	}
}

func (s *Service) handleWithdrawal(ctx context.Context, slug models.ExchangeSlug, order *models.ExchangeWithdrawalHistory, client exchange_manager.IExchangeClient, tx pgx.Tx) error {
	request := &models.GetWithdrawalByIDParams{}
	if order.ExchangeOrderID.Valid {
		request.ClientOrderID = &order.ExchangeOrderID.String
	}
	transferRecord, err := client.GetWithdrawalByID(ctx, request)
	if err != nil {
		return fmt.Errorf("fetch exchange transfer history: %w", err)
	}
	switch slug {
	case models.ExchangeSlugHtx:
		return s.handleHtxWithdrawal(ctx, transferRecord, order, tx)
	case models.ExchangeSlugOkx:
		return s.handleOkxWithdrawal(ctx, transferRecord, order, tx)
	case models.ExchangeSlugBinance:
		return s.handleBinanceWithdrawal(ctx, transferRecord, order, tx)
	case models.ExchangeSlugBitget:
		return s.handleBitgetWithdrawal(ctx, transferRecord, order, tx)
	case models.ExchangeSlugKucoin:
		return s.handleKucoinWithdrawal(ctx, transferRecord, order, tx)
	case models.ExchangeSlugGateio:
		return s.handleGateioWithdrawal(ctx, transferRecord, order, tx)
	case models.ExchangeSlugBybit:
		return s.handleBybitWithdrawal(ctx, transferRecord, order, tx)
	default:
		return fmt.Errorf("exchange %s not supported", slug.String())
	}
}

func (s *Service) handleGateioWithdrawal(ctx context.Context, record *models.WithdrawalStatusDTO, order *models.ExchangeWithdrawalHistory, tx pgx.Tx) error {
	updateParams := repo_exchange_withdrawal_history.UpdateParams{
		ID: order.ID,
	}

	switch record.Status {
	case gateio.WithdrawalStatusCancel.String(), gateio.WithdrawalStatusFail.String(), gateio.WithdrawalStatusInvalid.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusFailed.String()}
	case gateio.WithdrawalStatusRequest.String(), gateio.WithdrawalStatusManual.String(), gateio.WithdrawalStatusExtpend.String(),
		gateio.WithdrawalStatusProces.String(), gateio.WithdrawalStatusPend.String(), gateio.WithdrawalStatusDmove.String(), gateio.WithdrawalStatusReview.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusInProgress.String()}
	case gateio.WithdrawalStatusDone.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusCompleted.String()}
		updateParams.Txid = pgtype.Text{Valid: true, String: record.TxHash}
		updateParams.NativeAmount = decimal.NullDecimal{Valid: true, Decimal: record.NativeAmount}
	case gateio.WithdrawalStatusVerify.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusRecovery.String()}
		updateParams.FailReason = pgtype.Text{Valid: true, String: "manual verification required"}
	default:
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusFailed.String()}
	}

	return s.updateExchangeWithdrawal(ctx, order.UserID, updateParams, tx)
}

func (s *Service) handleKucoinWithdrawal(ctx context.Context, record *models.WithdrawalStatusDTO, order *models.ExchangeWithdrawalHistory, tx pgx.Tx) error {
	updateParams := repo_exchange_withdrawal_history.UpdateParams{
		ID: order.ID,
	}

	switch record.Status {
	case kucoinmodels.WithdrawalStatusFailure.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusFailed.String()}
	case kucoinmodels.WithdrawalStatusWalletProcessing.String(),
		kucoinmodels.WithdrawalStatusReview.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusInProgress.String()}
	case kucoinmodels.WithdrawalStatusSuccess.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusCompleted.String()}
		updateParams.Txid = pgtype.Text{Valid: true, String: record.TxHash}
		updateParams.NativeAmount = decimal.NullDecimal{Valid: true, Decimal: record.NativeAmount}
	default:
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusFailed.String()}
	}

	return s.updateExchangeWithdrawal(ctx, order.UserID, updateParams, tx)
}

func (s *Service) handleHtxWithdrawal(ctx context.Context, record *models.WithdrawalStatusDTO, order *models.ExchangeWithdrawalHistory, tx pgx.Tx) error {
	updateParams := repo_exchange_withdrawal_history.UpdateParams{
		ID: order.ID,
	}

	switch record.Status {
	case htxmodels.TransferStatePreTransfer.String(), htxmodels.TransferStatePass.String(), htxmodels.TransferStateSubmitted.String():
		return nil
	case htxmodels.TransferStateWalletTransfer.String(), htxmodels.TransferStateConfirmed.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusCompleted.String()}
		updateParams.Txid = pgtype.Text{Valid: true, String: record.TxHash}
		updateParams.NativeAmount = decimal.NullDecimal{Valid: true, Decimal: record.NativeAmount}
	default:
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusFailed.String()}
	}

	return s.updateExchangeWithdrawal(ctx, order.UserID, updateParams, tx)
}

func (s *Service) handleOkxWithdrawal(ctx context.Context, record *models.WithdrawalStatusDTO, order *models.ExchangeWithdrawalHistory, tx pgx.Tx) error {
	updateParams := repo_exchange_withdrawal_history.UpdateParams{
		ID: order.ID,
	}

	switch record.Status {
	case okxmodels.WithdrawalStatusWaitingWithdrawal.String(), okxmodels.WithdrawalStatusWaitingTransfer.String(), okxmodels.WithdrawalStatusBroadcasting.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusInProgress.String()}
	case okxmodels.WithdrawalStatusPendingValidation.String(),
		okxmodels.WithdrawalStatusDelayed.String(),
		okxmodels.WithdrawalStatusCanceling.String(),
		okxmodels.WithdrawalStatusCanceled.String(),
		okxmodels.WithdrawalStatusFailed.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusFailed.String()}
	case okxmodels.WithdrawalStatusSuccess.String(),
		okxmodels.WithdrawalStatusApproved.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusCompleted.String()}
		updateParams.Txid = pgtype.Text{Valid: true, String: record.TxHash}
	default:
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusRecovery.String()}
	}

	return s.updateExchangeWithdrawal(ctx, order.UserID, updateParams, tx)
}

func (s *Service) handleBinanceWithdrawal(ctx context.Context, record *models.WithdrawalStatusDTO, order *models.ExchangeWithdrawalHistory, tx pgx.Tx) error {
	updateParams := repo_exchange_withdrawal_history.UpdateParams{
		ID: order.ID,
	}

	switch record.Status {
	case binancemodels.WithdrawalStatusEmailSent.String(),
		binancemodels.WithdrawalStatusProcessing.String(),
		binancemodels.WithdrawalStatusAwaitingApproval.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusInProgress.String()}
	case binancemodels.WithdrawalStatusRejected.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusFailed.String()}
	case binancemodels.WithdrawalStatusCompleted.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusCompleted.String()}
		updateParams.Txid = pgtype.Text{Valid: true, String: record.TxHash}
	default:
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusRecovery.String()}
	}

	return s.updateExchangeWithdrawal(ctx, order.UserID, updateParams, tx)
}

func (s *Service) handleBitgetWithdrawal(ctx context.Context, record *models.WithdrawalStatusDTO, order *models.ExchangeWithdrawalHistory, tx pgx.Tx) error {
	updateParams := repo_exchange_withdrawal_history.UpdateParams{
		ID: order.ID,
	}

	switch record.Status {
	case bitgetmodels.WithdrawalStatusPending.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusInProgress.String()}
	case bitgetmodels.WithdrawalStatusFailed.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusFailed.String()}
	case bitgetmodels.WithdrawalStatusSuccess.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusCompleted.String()}
		updateParams.Txid = pgtype.Text{Valid: true, String: record.TxHash}
	default:
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusRecovery.String()}
	}

	return s.updateExchangeWithdrawal(ctx, order.UserID, updateParams, tx)
}

func (s *Service) handleBybitWithdrawal(ctx context.Context, record *models.WithdrawalStatusDTO, order *models.ExchangeWithdrawalHistory, tx pgx.Tx) error {
	updateParams := repo_exchange_withdrawal_history.UpdateParams{
		ID: order.ID,
	}

	switch record.Status {
	case bybitmodels.WithdrawalStatusPending.String(), bybitmodels.WithdrawalStatusBlockchainConfirmed.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusInProgress.String()}
	case bybitmodels.WithdrawalStatusFail.String(), bybitmodels.WithdrawalStatusReject.String(), bybitmodels.WithdrawalStatusCancelByUser.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusFailed.String()}
	case bybitmodels.WithdrawalStatusSuccess.String():
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusCompleted.String()}
		updateParams.Txid = pgtype.Text{Valid: true, String: record.TxHash}
	default:
		updateParams.Status = pgtype.Text{Valid: true, String: models.WithdrawalHistoryStatusRecovery.String()}
	}

	return s.updateExchangeWithdrawal(ctx, order.UserID, updateParams, tx)
}

func (s *Service) updateExchangeWithdrawal(
	ctx context.Context,
	usrID uuid.UUID,
	fields repo_exchange_withdrawal_history.UpdateParams,
	tx pgx.Tx,
) error {
	user, err := s.st.Users(repos.WithTx(tx)).GetByID(ctx, usrID)
	if err != nil {
		return fmt.Errorf("get user by id: %w", err)
	}

	if fields.Status.Valid {
		if err := s.handleOrderStatusChanged(ctx, user, fields); err != nil {
			return fmt.Errorf("handle order status changed: %w", err)
		}
	}

	return s.st.ExchangeWithdrawalHistory(repos.WithTx(tx)).Update(ctx, fields)
}

func (s *Service) handleOrderStatusChanged(ctx context.Context, usr *models.User, updateFields repo_exchange_withdrawal_history.UpdateParams) error {
	transferState, err := s.settingSvc.GetModelSetting(ctx, setting.TransfersStatus, usr)
	if err != nil {
		return fmt.Errorf("get transfer status setting: %w", err)
	}

	var transferStatusSetting string
	switch updateFields.Status.String {
	case models.WithdrawalHistoryStatusRecovery.String():
		if transferState.Value == setting.FlagValueEnabled {
			if updateFields.FailReason.Valid {
				if strings.Contains(updateFields.FailReason.String, "manual verification required") {
					// If transfer is in recovery state due to manual verification, we do not change the transfer status setting
					return nil
				}
			}
		}
	case models.WithdrawalHistoryStatusFailed.String():
		if transferState.Value == setting.FlagValueEnabled {
			if updateFields.FailReason.Valid {
				if strings.Contains(updateFields.FailReason.String, "connection timed out") ||
					strings.Contains(updateFields.FailReason.String, "connection reset by peer") ||
					strings.Contains(updateFields.FailReason.String, "connection refused") ||
					strings.Contains(updateFields.FailReason.String, "broken pipe") ||
					strings.Contains(updateFields.FailReason.String, "rate limit") {
					// If the transfer failed due to connection issues, we do not change the transfer status setting
					return nil
				}
				// Update setting only from manually enabled setting
				transferStatusSetting = setting.TransferStatusSystemSuspended
			}
		}
	case models.WithdrawalHistoryStatusCompleted.String():
		if transferState.Value == setting.TransferStatusSystemSuspended {
			// Update transfers state only if it was suspended by system
			transferStatusSetting = setting.FlagValueEnabled
		}
	}

	if transferStatusSetting != "" {
		if err := s.settingSvc.SetModelSetting(ctx, setting.UpdateDTO{
			Name:  setting.TransfersStatus,
			Value: transferStatusSetting,
			Model: setting.IModelSetting(usr),
		}); err != nil {
			return fmt.Errorf("set transfer status setting: %w", err)
		}
	}

	return nil
}

func (s *Service) RunWithdrawalQueue(ctx context.Context) {
	newWdTicker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case <-newWdTicker.C:
			s.processNewWithdrawals(ctx)
		}
	}
}

func (s *Service) RunWithdrawalUpdater(ctx context.Context) {
	oldWdTicker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-oldWdTicker.C:
			if err := s.processOldWithdrawals(ctx); err != nil {
				s.logger.Error("failed to process old withdrawals", err)
			}
		}
	}
}

func (s *Service) UpdateWithdrawalSetting(ctx context.Context, userID uuid.UUID, settingID uuid.UUID, isEnabled bool) (*models.ExchangeWithdrawalSetting, error) {
	return s.st.ExchangeWithdrawalSettings().UpdateIsEnabledByID(ctx, repo_exchange_withdrawal_settings.UpdateIsEnabledByIDParams{
		NewState: isEnabled,
		ID:       settingID,
		UserID:   userID,
	})
}

func (s *Service) processNewWithdrawals(ctx context.Context) {
	settings, err := s.st.ExchangeWithdrawalSettings().GetActive(ctx)
	if err != nil {
		s.logger.Error("failed to fetch withdrawal settings", err)
		return
	}

	userSettings := make(map[uuid.UUID][]*models.ExchangeWithdrawalSetting)
	for _, set := range settings {
		userSettings[set.UserID] = append(userSettings[set.UserID], set)
	}

	for userID, sets := range userSettings {
		go func() {
			s.processWithdrawals(ctx, userID, sets)
		}()
	}
}

func (s *Service) getWithdrawalState(ctx context.Context, userID, exchangeID uuid.UUID) (*models.ExchangeWithdrawalState, error) {
	params := repo_user_exchanges.GetByUserIDParams{UserID: userID, ExchangeID: exchangeID}
	ue, err := s.st.UserExchanges().GetByUserID(ctx, params)
	if err != nil {
		return nil, err
	}
	return &ue.WithdrawalState, nil
}

func (s *Service) processOldWithdrawals(ctx context.Context) error {
	lastOrders, err := s.st.ExchangeWithdrawalHistory().GetAllUnprocessed(ctx)
	if err != nil {
		return err
	}
	for _, order := range lastOrders {
		err := repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
			if order.Status == models.WithdrawalHistoryStatusInProgress {
				ex, err := s.st.Exchanges().GetByID(ctx, order.ExchangeID)
				if err != nil {
					s.logger.Error("failed to fetch exchange", err)
					return err
				}
				exClient, err := s.exManager.GetDriver(ctx, ex.Slug, order.UserID)
				if err != nil {
					s.logger.Error("failed to get exchange client", err)
					return err
				}
				if err := s.handleWithdrawal(ctx, ex.Slug, order, exClient, tx); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			s.logger.Error("failed to process withdrawal", err)
		}
	}
	return nil
}
