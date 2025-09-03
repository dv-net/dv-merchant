package processing

import (
	"context"
	"fmt"
	"sync/atomic"

	"connectrpc.com/connect"
	"github.com/google/uuid"

	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/interceptors"
	"github.com/dv-net/dv-merchant/internal/metrics"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/currency"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallet_addresses"
	"github.com/dv-net/dv-merchant/internal/util"
	"github.com/dv-net/dv-merchant/pkg/logger"

	clientv1 "github.com/dv-net/dv-processing/api/processing/client/v1"
	ownerv1 "github.com/dv-net/dv-processing/api/processing/owner/v1"
	transferv1 "github.com/dv-net/dv-processing/api/processing/transfer/v1"
	walletv1 "github.com/dv-net/dv-processing/api/processing/wallet/v1"
)

type IProcessingWallet interface {
	CreateOwnerHotWallet(ctx context.Context, params CreateOwnerHotWalletParams) (*HotWallet, error)
	AttachOwnerColdWallets(ctx context.Context, params AttachOwnerColdWalletsParams) error
	GetOwnerColdWallets(ctx context.Context, params GetOwnerColdWalletsParams) ([]BlockchainWalletsList, error)
	GetOwnerProcessingWallets(ctx context.Context, params GetOwnerProcessingWalletsParams) ([]WalletProcessing, error)
	GetOwnerProcessingWallet(ctx context.Context, params GetOwnerProcessingWalletsParams) (*WalletProcessing, error)
	MarkDirtyHotWallet(ctx context.Context, ownerID uuid.UUID, blockchain models.Blockchain, address string) error
	GetOwnerHotWalletKeys(ctx context.Context, user *models.User, otp string, params GetOwnerHotWalletKeysParams) (*GetOwnerHotWalletKeysData, error)
}

type IProcessingOwner interface {
	CreateOwner(ctx context.Context, outerID string, mnemonic string) (*RegisterOwnerInfo, error)
	GetOwnerPrivateKeys(ctx context.Context, ownerID uuid.UUID, otp string) (*GetOwnerPrivateKeysData, error)
	GetOwnerSeed(ctx context.Context, ownerID uuid.UUID, otp string) (*OwnerSeedData, error)
	DisableTwoFactorAuth(ctx context.Context, ownerID uuid.UUID, otp string) error
	ConfirmTwoFactorAuth(ctx context.Context, ownerID uuid.UUID, otp string) error
	GetTwoFactorAuthData(ctx context.Context, ownerID uuid.UUID) (TwoFactorAuthData, error)
	ProcessingSettings(ctx context.Context) (*Settings, error)
	ValidateTwoFactorToken(context.Context, uuid.UUID, string) error
}

type IProcessingTransfer interface {
	FundsWithdrawal(ctx context.Context, params FundsWithdrawalParams) (FundsWithdrawalResult, error)
	StatusFundsWithdrawal(ctx context.Context, requestID uuid.UUID) (status WithdrawalStatus, txHash string, err error)
}

type IProcessingClient interface {
	CreateClient(ctx context.Context, callbackURL string, backendAddress *string, merchantDomain *string) (*Settings, error)
	ChangeClient(ctx context.Context, ID uuid.UUID, callbackURL string) error
	GetCallbackURL(ctx context.Context, id uuid.UUID) (string, error)
	ProcessingSettings(ctx context.Context) (*Settings, error)
}

type IProcessingService interface {
	Reinitialize(ctx context.Context) error
	Initialized() bool
}

var (
	_ IProcessingOwner    = (*Service)(nil)
	_ IProcessingTransfer = (*Service)(nil)
	_ IProcessingService  = (*Service)(nil)
	_ IProcessingWallet   = (*Service)(nil)
)

type Service struct {
	log logger.Logger

	processingService *Processing

	currService     currency.ICurrency
	currConvService currconv.ICurrencyConvertor
	settingService  setting.ISettingService
	eventListener   event.IListener
	storage         storage.IStorage
	initialized     atomic.Bool
	m               *metrics.PrometheusMetrics
	appVersion      string
}

func New(
	ctx context.Context,
	logger logger.Logger,
	currService currency.ICurrency,
	currConvService currconv.ICurrencyConvertor,
	settingService setting.ISettingService,
	eventListener event.IListener,
	storage storage.IStorage,
	processingMetrics *metrics.PrometheusMetrics,
	appVersion string,
) *Service {
	svc := &Service{
		log:             logger,
		currService:     currService,
		currConvService: currConvService,
		settingService:  settingService,
		eventListener:   eventListener,
		storage:         storage,
		m:               processingMetrics,
		appVersion:      appVersion,
	}

	if err := svc.initServices(ctx); err != nil {
		logger.Warn("processing url not install")
	}

	svc.eventListener.Register(setting.RootSettingsChanged, svc.handleSettingChange)

	return svc
}

func (s *Service) Initialized() bool { return s.initialized.Load() }

func (s *Service) CreateClient(ctx context.Context, callbackURL string, backendAddress, merchantDomain *string) (*Settings, error) {
	if !s.Initialized() {
		return nil, ErrServiceNotInitialized
	}

	clientResponse, err := s.processingService.Client().Create(
		ctx,
		connect.NewRequest(&clientv1.CreateRequest{
			CallbackUrl:    callbackURL,
			BackendIp:      backendAddress,
			MerchantDomain: merchantDomain,
			BackendVersion: s.appVersion,
		}),
	)
	if clientResponse == nil || err != nil {
		return nil, fmt.Errorf("client registration: %w", err)
	}

	if _, err := s.settingService.SetRootSetting(ctx, setting.ProcessingClientKey, clientResponse.Msg.ClientKey); err != nil {
		return nil, fmt.Errorf("setting client_key url: %w", err)
	}
	if _, err := s.settingService.SetRootSetting(ctx, setting.ProcessingClientID, clientResponse.Msg.ClientId); err != nil {
		return nil, fmt.Errorf("setting client_id url: %w", err)
	}
	if _, err := s.settingService.SetRootSetting(ctx, setting.DvAdminSecretKey, clientResponse.Msg.AdminSecretKey); err != nil {
		return nil, fmt.Errorf("setting admin_secret: %w", err)
	}

	return s.ProcessingSettings(ctx)
}

func (s *Service) CreateOwner(ctx context.Context, externalID string, mnemonic string) (*RegisterOwnerInfo, error) {
	if !s.Initialized() {
		return nil, ErrServiceNotInitialized
	}

	pSetting, err := s.ProcessingSettings(ctx)
	if err != nil {
		return nil, err
	}
	createOwnerResp, err := s.processingService.Owner().Create(
		ctx,
		connect.NewRequest(&ownerv1.CreateRequest{
			ClientId:   pSetting.ProcessingClientID,
			ExternalId: externalID,
			Mnemonic:   mnemonic,
		}),
	)

	if createOwnerResp == nil {
		return nil, err
	}

	ownerID, err := uuid.Parse(createOwnerResp.Msg.GetId())
	if err != nil {
		return nil, err
	}

	return &RegisterOwnerInfo{
		OwnerID: ownerID,
	}, err
}

func (s *Service) ChangeClient(ctx context.Context, id uuid.UUID, callbackURL string) error {
	if !s.Initialized() {
		return ErrServiceNotInitialized
	}

	_, err := s.processingService.Client().UpdateCallbackURL(
		ctx,
		connect.NewRequest(&clientv1.UpdateCallbackURLRequest{ClientId: id.String(), CallbackUrl: callbackURL}),
	)

	return err
}

func (s *Service) GetCallbackURL(ctx context.Context, id uuid.UUID) (string, error) {
	if !s.Initialized() {
		return "", ErrServiceNotInitialized
	}

	callbackURL, err := s.processingService.Client().GetCallbackURL(ctx, connect.NewRequest(&clientv1.GetCallbackURLRequest{ClientId: id.String()}))
	if err != nil {
		return "", fmt.Errorf("get callback url: %w", err)
	}

	return callbackURL.Msg.GetCallbackUrl(), nil
}

func (s *Service) CreateOwnerHotWallet(ctx context.Context, params CreateOwnerHotWalletParams) (*HotWallet, error) {
	if !s.Initialized() {
		return nil, ErrServiceNotInitialized
	}

	blockchain, err := params.Blockchain.ToPb()
	if err != nil {
		return nil, err
	}
	req := &walletv1.CreateOwnerHotWalletRequest{
		OwnerId:          params.OwnerID.String(),
		ExternalWalletId: params.CustomerID,
		Blockchain:       blockchain,
	}
	if params.BitcoinAddressType != nil {
		req.BitcoinAddressType = util.Pointer(params.BitcoinAddressType.ToPb())
	}

	if params.LitecoinAddressType != nil {
		req.LitecoinAddressType = util.Pointer(params.LitecoinAddressType.ToPb())
	}

	resp, err := s.processingService.Wallet().CreateOwnerHotWallet(
		ctx,
		connect.NewRequest(req),
	)
	if err != nil {
		return nil, err
	}

	return &HotWallet{
		Address: resp.Msg.Address,
	}, nil
}

func (s *Service) AttachOwnerColdWallets(ctx context.Context, params AttachOwnerColdWalletsParams) error {
	if !s.Initialized() {
		return ErrServiceNotInitialized
	}

	blockchain, err := params.Blockchain.ToPb()
	if err != nil {
		return err
	}
	req := &walletv1.AttachOwnerColdWalletsRequest{
		OwnerId:    params.OwnerID.String(),
		Blockchain: blockchain,
		Addresses:  params.Addresses,
		Totp:       params.TOTP,
	}
	_, err = s.processingService.Wallet().AttachOwnerColdWallets(
		ctx,
		connect.NewRequest(req),
	)

	return err
}

func (s *Service) GetOwnerColdWallets(ctx context.Context, params GetOwnerColdWalletsParams) ([]BlockchainWalletsList, error) {
	if !s.Initialized() {
		return nil, ErrServiceNotInitialized
	}

	blockchain, err := params.Blockchain.ToPb()
	if err != nil {
		return nil, err
	}

	req := &walletv1.GetOwnerColdWalletsRequest{
		OwnerId:    params.OwnerID.String(),
		Blockchain: &blockchain,
	}
	resp, err := s.processingService.Wallet().GetOwnerColdWallets(
		ctx,
		connect.NewRequest(req),
	)
	if err != nil {
		return nil, err
	}

	res := make([]BlockchainWalletsList, 0, len(resp.Msg.Items))
	for _, v := range resp.Msg.Items {
		res = append(res, BlockchainWalletsList{
			Blockchain: models.Blockchain(v.Blockchain),
			Wallets:    v.Address,
		})
	}

	return res, nil
}

func (s *Service) GetOwnerProcessingWallet(ctx context.Context, params GetOwnerProcessingWalletsParams) (*WalletProcessing, error) {
	if !s.Initialized() {
		return nil, ErrServiceNotInitialized
	}

	req := &walletv1.GetOwnerProcessingWalletsRequest{
		OwnerId: params.OwnerID.String(),
		Tiny:    params.Tiny,
	}

	if params.Blockchain != nil {
		blockchain, err := params.Blockchain.ToPb()
		if err != nil {
			return nil, err
		}
		req.Blockchain = util.Pointer(blockchain)
	}

	resp, err := s.processingService.Wallet().GetOwnerProcessingWallets(
		ctx, connect.NewRequest(req),
	)
	if err != nil {
		return nil, err
	}

	wp := WalletProcessing{
		Address:    resp.Msg.Items[0].Address,
		Blockchain: models.ConvertToModel(resp.Msg.Items[0].Blockchain),
	}

	return util.Pointer(wp), nil
}

func (s *Service) GetOwnerProcessingWallets(ctx context.Context, params GetOwnerProcessingWalletsParams) ([]WalletProcessing, error) {
	if !s.Initialized() {
		return nil, ErrServiceNotInitialized
	}

	req := &walletv1.GetOwnerProcessingWalletsRequest{
		OwnerId: params.OwnerID.String(),
	}

	if params.Blockchain != nil {
		blockchain, err := params.Blockchain.ToPb()
		if err != nil {
			return nil, err
		}
		req.Blockchain = util.Pointer(blockchain)
	}

	resp, err := s.processingService.Wallet().GetOwnerProcessingWallets(
		ctx, connect.NewRequest(req),
	)
	if err != nil {
		return nil, err
	}

	res := make([]WalletProcessing, 0, len(resp.Msg.Items))
	for _, w := range resp.Msg.Items {
		wp := WalletProcessing{
			Address:    w.Address,
			Blockchain: models.ConvertToModel(w.Blockchain),
		}

		if w.Assets != nil {
			for _, asset := range w.Assets.Asset {
				wp.Assets = append(wp.Assets, &Asset{
					Identity: asset.Identity,
					Amount:   asset.Amount,
				})
			}
		}

		if w.BlockchainAdditionalData != nil {
			wp.AdditionalData = &BlockchainAdditionalData{}
			if w.BlockchainAdditionalData.TronData != nil {
				wp.AdditionalData.TronData = &TronData{
					AvailableEnergyForUse:    w.BlockchainAdditionalData.TronData.AvailableEnergyForUse,
					TotalEnergy:              w.BlockchainAdditionalData.TronData.TotalEnergy,
					AvailableBandwidthForUse: w.BlockchainAdditionalData.TronData.AvailableBandwidthForUse,
					TotalBandwidth:           w.BlockchainAdditionalData.TronData.TotalBandwidth,
					StackedTrx:               w.BlockchainAdditionalData.TronData.StackedTrx,
					StackedEnergy:            w.BlockchainAdditionalData.TronData.StackedEnergy,
					StackedEnergyTrx:         w.BlockchainAdditionalData.TronData.StackedEnergyTrx,
					StackedBandwidth:         w.BlockchainAdditionalData.TronData.StackedBandwidth,
					StackedBandwidthTrx:      w.BlockchainAdditionalData.TronData.StackedBandwidthTrx,
					TotalUsedEnergy:          w.BlockchainAdditionalData.TronData.TotalUsedEnergy,
					TotalUsedBandwidth:       w.BlockchainAdditionalData.TronData.TotalUsedBandwidth,
				}
			}
		}

		res = append(res, wp)
	}

	return res, nil
}

func (s *Service) FundsWithdrawal(ctx context.Context, params FundsWithdrawalParams) (FundsWithdrawalResult, error) {
	if !s.Initialized() {
		return FundsWithdrawalResult{}, ErrServiceNotInitialized
	}

	blockchain, err := params.Blockchain.ToPb()
	if err != nil {
		return FundsWithdrawalResult{}, err
	}

	req := &transferv1.CreateRequest{
		OwnerId:         params.OwnerID.String(),
		RequestId:       params.RequestID.String(),
		Blockchain:      blockchain,
		FromAddresses:   params.FromAddress,
		ToAddresses:     params.ToAddress,
		AssetIdentifier: params.ContractAddress,
		WholeAmount:     params.WholeAmount,
		Amount:          &params.Amount,
		Kind:            params.Kind,
	}

	resp, err := s.processingService.Transfers().Create(ctx, connect.NewRequest(req))
	if err != nil {
		return FundsWithdrawalResult{}, err
	}

	var status WithdrawalStatus
	switch resp.Msg.GetItem().Status {
	case transferv1.Status_STATUS_NEW:
		status = AcceptedWithdrawalStatus
	case transferv1.Status_STATUS_PENDING:
		status = PendingWithdrawalStatus
	case transferv1.Status_STATUS_FAILED:
		status = FailedWithdrawalStatus
	case transferv1.Status_STATUS_UNSPECIFIED:
	default:
		status = UnknownWithdrawalStatus
	}

	return FundsWithdrawalResult{
		WithdrawalStatus: status,
		TxHash:           resp.Msg.GetItem().TxHash,
		Message:          resp.Msg.GetItem().ErrorMessage,
	}, nil
}

func (s *Service) StatusFundsWithdrawal(_ context.Context, _ uuid.UUID) (status WithdrawalStatus, txHash string, err error) {
	return 0, "", fmt.Errorf("not implemented")
	// req := &transactionv1.StatusFundsWithdrawalRequest{RequestId: requestID.String()}
	// resp, err := s.transaction.StatusFundsWithdrawal(ctx, connect.NewRequest(req))
	// if resp == nil {
	// 	return UnknownWithdrawalStatus, "", err
	// }

	// var resultStatus WithdrawalStatus
	// switch resp.Msg.GetStatus() {
	// case commonv1.WithdrawalStatus_WITHDRAWAL_STATUS_ACCEPTED:
	// 	resultStatus = AcceptedWithdrawalStatus
	// case commonv1.WithdrawalStatus_WITHDRAWAL_STATUS_FAILED:
	// 	resultStatus = FailedWithdrawalStatus
	// case commonv1.WithdrawalStatus_WITHDRAWAL_STATUS_SUCCESS:
	// 	resultStatus = SuccessWithdrawalStatus
	// case commonv1.WithdrawalStatus_WITHDRAWAL_STATUS_UNSPECIFIED:
	// default:
	// 	resultStatus = UnknownWithdrawalStatus
	// }

	// return resultStatus, resp.Msg.GetTxHash(), err
}

func (s *Service) ProcessingSettings(ctx context.Context) (*Settings, error) {
	if !s.Initialized() {
		return nil, ErrServiceNotInitialized
	}

	processingURLSetting, err := s.settingService.GetRootSetting(ctx, setting.ProcessingURL)
	if err != nil {
		return nil, fmt.Errorf("processing base url not install")
	}
	clientIDSetting, err := s.settingService.GetRootSetting(ctx, setting.ProcessingClientID)
	if err != nil {
		return nil, fmt.Errorf("client ID not install")
	}
	clientKeySetting, err := s.settingService.GetRootSetting(ctx, setting.ProcessingClientKey)
	if err != nil {
		return nil, fmt.Errorf("client key not install")
	}

	processingSettings := &Settings{
		BaseURL:             processingURLSetting.Value,
		ProcessingClientID:  clientIDSetting.Value,
		ProcessingClientKey: clientKeySetting.Value,
	}

	return processingSettings, nil
}

func (s *Service) GetTwoFactorAuthData(ctx context.Context, ownerID uuid.UUID) (TwoFactorAuthData, error) {
	if !s.Initialized() {
		return TwoFactorAuthData{}, ErrServiceNotInitialized
	}

	res, err := s.processingService.Owner().GetTwoFactorAuthData(ctx, connect.NewRequest(&ownerv1.GetTwoFactorAuthDataRequest{
		OwnerId: ownerID.String(),
	}))
	if err != nil {
		return TwoFactorAuthData{}, fmt.Errorf("get 2fa data: %w", err)
	}

	return TwoFactorAuthData{
		Secret:      res.Msg.GetSecret(),
		IsConfirmed: res.Msg.GetIsConfirmed(),
	}, nil
}

func (s *Service) ConfirmTwoFactorAuth(ctx context.Context, ownerID uuid.UUID, otp string) error {
	if !s.Initialized() {
		return ErrServiceNotInitialized
	}

	_, err := s.processingService.Owner().ConfirmTwoFactorAuth(ctx, connect.NewRequest(&ownerv1.ConfirmTwoFactorAuthRequest{
		OwnerId: ownerID.String(),
		Totp:    otp,
	}))
	if err != nil {
		return fmt.Errorf("2fa confirmation: %w", err)
	}

	return nil
}

func (s *Service) DisableTwoFactorAuth(ctx context.Context, ownerID uuid.UUID, otp string) error {
	if !s.Initialized() {
		return ErrServiceNotInitialized
	}

	_, err := s.processingService.Owner().DisableTwoFactorAuth(ctx, connect.NewRequest(&ownerv1.DisableTwoFactorAuthRequest{
		OwnerId: ownerID.String(),
		Totp:    otp,
	}))
	if err != nil {
		return fmt.Errorf("disable 2FA: %w", err)
	}

	return nil
}

func (s *Service) GetOwnerPrivateKeys(ctx context.Context, ownerID uuid.UUID, otp string) (*GetOwnerPrivateKeysData, error) {
	if !s.Initialized() {
		return nil, ErrServiceNotInitialized
	}

	req := &ownerv1.GetPrivateKeysRequest{
		OwnerId: ownerID.String(),
		Totp:    otp,
	}
	resp, err := s.processingService.Owner().GetPrivateKeys(
		ctx,
		connect.NewRequest(req),
	)
	if err != nil {
		return nil, fmt.Errorf("owner private keys: %w", err)
	}

	data := &GetOwnerPrivateKeysData{
		make(map[string]*KeyPairSequence),
	}

	for blockchain, pair := range resp.Msg.GetKeys() {
		if _, exists := data.Keys[blockchain]; !exists {
			data.Keys[blockchain] = &KeyPairSequence{}
			for _, p := range pair.GetPairs() {
				data.Keys[blockchain].Pairs = append(data.Keys[blockchain].Pairs, KeyPair{
					PublicKey:  p.PublicKey,
					PrivateKey: p.PrivateKey,
					Address:    p.Address,
					Kind:       p.Kind,
				})
			}
		}
	}

	return data, nil
}

func (s *Service) GetOwnerHotWalletKeys(ctx context.Context, user *models.User, otp string, params GetOwnerHotWalletKeysParams) (*GetOwnerHotWalletKeysData, error) {
	if !s.Initialized() {
		return nil, ErrServiceNotInitialized
	}

	// Get all wallet addresses for the user
	allWalletAddresses, err := s.storage.WalletAddresses().GetWalletAddressesByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("get wallet addresses: %w", err)
	}

	// Apply include/exclude logic
	var walletAddresses []*models.WalletAddress

	// If specific IDs are provided (include list), filter to only those
	if len(params.WalletAddressIDs) > 0 {
		includeMap := make(map[uuid.UUID]struct{})
		for _, id := range params.WalletAddressIDs {
			includeMap[id] = struct{}{}
		}

		for _, addr := range allWalletAddresses {
			if _, ok := includeMap[addr.ID]; ok {
				walletAddresses = append(walletAddresses, addr)
			}
		}
	} else {
		// If no include list, use all addresses
		walletAddresses = allWalletAddresses
	}

	// Apply exclude logic
	if len(params.ExcludedWalletAddressesIDs) > 0 {
		excludeMap := make(map[uuid.UUID]bool)
		for _, id := range params.ExcludedWalletAddressesIDs {
			excludeMap[id] = true
		}

		var filteredAddresses []*models.WalletAddress
		for _, addr := range walletAddresses {
			if !excludeMap[addr.ID] {
				filteredAddresses = append(filteredAddresses, addr)
			}
		}
		walletAddresses = filteredAddresses
	}

	addresses := make([]string, 0, len(walletAddresses))
	for _, w := range walletAddresses {
		addresses = append(addresses, w.Address)
	}

	// Prepare request
	req := &ownerv1.GetHotWalletKeysRequest{
		OwnerId:         user.ProcessingOwnerID.UUID.String(),
		WalletAddresses: addresses,
		Otp:             otp,
	}

	// Fetch data from processing service
	resp, err := s.processingService.Owner().GetHotWalletKeys(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, fmt.Errorf("get hot wallet keys: %w", err)
	}

	// Populate AllSelectedWallets for logging purposes
	allSelectedWallets := make([]*repo_wallet_addresses.FilterOwnerWalletAddressesRow, 0, len(walletAddresses))
	for _, w := range walletAddresses {
		allSelectedWallets = append(allSelectedWallets, &repo_wallet_addresses.FilterOwnerWalletAddressesRow{
			Address:           w.Address,
			UserID:            w.UserID,
			WalletAddressesID: w.ID,
		})
	}

	data := &GetOwnerHotWalletKeysData{
		Entries:            make([]HotWalletKeyPair, 0, len(resp.Msg.GetEntries())),
		AllSelectedWallets: allSelectedWallets,
	}

	for _, entry := range resp.Msg.GetEntries() {
		hotWallet := HotWalletKeyPair{
			Name:  entry.Name.String(),
			Items: make([]HotKeyPair, 0, len(entry.GetItems())),
		}
		for _, item := range entry.GetItems() {
			hotWallet.Items = append(hotWallet.Items, HotKeyPair{
				PublicKey:  item.GetPublicKey(),
				PrivateKey: item.GetPrivateKey(),
				Address:    item.GetAddress(),
			})
		}
		data.Entries = append(data.Entries, hotWallet)
	}
	return data, nil
}

func (s *Service) GetOwnerSeed(ctx context.Context, ownerID uuid.UUID, otp string) (*OwnerSeedData, error) {
	if !s.Initialized() {
		return nil, ErrServiceNotInitialized
	}

	req := &ownerv1.GetSeedsRequest{
		OwnerId: ownerID.String(),
		Totp:    otp,
	}

	resp, err := s.processingService.Owner().GetSeeds(
		ctx,
		connect.NewRequest(req),
	)
	if err != nil {
		return nil, fmt.Errorf("owner seeds: %w", err)
	}

	data := &OwnerSeedData{
		PassPhrase: resp.Msg.GetPassPhrase(),
		Mnemonic:   resp.Msg.GetMnemonic(),
	}

	return data, nil
}

// MarkDirtyHotWallet
func (s *Service) MarkDirtyHotWallet(ctx context.Context, ownerID uuid.UUID, blockchain models.Blockchain, address string) error {
	if !s.Initialized() {
		return ErrServiceNotInitialized
	}

	blockchainPb, err := blockchain.ToPb()
	if err != nil {
		return err
	}

	_, err = s.processingService.Wallet().MarkDirtyHotWallet(ctx, connect.NewRequest(&walletv1.MarkDirtyHotWalletRequest{
		OwnerId:    ownerID.String(),
		Blockchain: blockchainPb,
		Address:    address,
	}))

	return err
}

// ValidateTwoFactorToken validates the two-factor token for the owner.
func (s *Service) ValidateTwoFactorToken(ctx context.Context, ownerID uuid.UUID, token string) error {
	if !s.Initialized() {
		return ErrServiceNotInitialized
	}

	_, err := s.processingService.Owner().ValidateTwoFactorToken(ctx, connect.NewRequest(&ownerv1.ValidateTwoFactorTokenRequest{
		OwnerId: ownerID.String(),
		Totp:    token,
	}))
	if err != nil {
		return fmt.Errorf("validate 2FA token: %w", err)
	}

	return nil
}

// Reinitialize reinit service if changed baseUrl
func (s *Service) Reinitialize(ctx context.Context) error {
	s.log.Info("Reinitializing processing service")

	s.initialized.Store(false)

	if err := s.initServices(ctx); err != nil {
		return fmt.Errorf("failed to init services: %w", err)
	}

	return nil
}

func (s *Service) initServices(ctx context.Context) error {
	processingURL, err := s.settingService.GetRootSetting(ctx, setting.ProcessingURL)
	if err != nil {
		return fmt.Errorf("failed to get processing url: %w", err)
	}

	options := connect.WithOptions(
		connect.WithInterceptors(
			interceptors.NewSignInterceptor(s.settingService),
			interceptors.NewProcessingMetric(s.m),
		),
	)

	s.processingService = NewProcessing(processingURL.Value, options)

	s.initialized.Store(true)

	return nil
}

func (s *Service) handleSettingChange(ev event.IEvent) error {
	if settingEvent, ok := ev.(*setting.RootSettingChangedEvent); ok && settingEvent.SettingName == setting.ProcessingURL {
		return s.Reinitialize(context.Background())
	}

	return nil
}
