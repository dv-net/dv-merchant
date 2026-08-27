package aml

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/transactions"
	"github.com/dv-net/dv-merchant/internal/service/wallet"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_checks"
	"github.com/dv-net/dv-merchant/pkg/aml"
	"github.com/dv-net/dv-merchant/pkg/aml/providers"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/dv-net/dv-processing/pkg/avalidator"

	"github.com/jackc/pgx/v5"
)

// slugMapping maps internal/models.AMLSlug ob providers.ProviderSlug.
var slugMapping = map[models.AMLSlug]aml.ProviderSlug{
	models.AMLSlugAMLBot:  aml.ProviderSlugAMLBot,
	models.AMLSlugBitOK:   aml.ProviderSlugBitOK,
	models.AMLSlugCoinKyt: aml.ProvideSlugCoinKyt,
}

// keyMapping maps internal/models.AmlKeyType to aml.AMLKeyType.
var keyMapping = map[models.AmlKeyType]aml.AuthKeyType{
	models.AmlKeyTypeAccessKeyID: aml.KeyAccessKeyID,
	models.AmlKeyTypeAccessKey:   aml.KeyAccessKey,
	models.AmlKeyTypeSecret:      aml.KeySecret,
	models.AmlKeyTypeAccessID:    aml.KeyAccessID,
	models.AmlKeyTypeAPIKey:      aml.KeyAPIKey,
}

type IService interface {
	ScoreTransaction(ctx context.Context, usr *models.User, dto CheckDTO) (*models.AmlCheck, error)
	AutoScoreDeposit(ctx context.Context, dto AutoScoreDepositDTO) (*models.AmlCheck, []aml.SignalContribution, error)
	GetStatistics(ctx context.Context, userID uuid.UUID) (*StatisticsDTO, error)
	GetAllActiveProviders() []models.AMLProvider
	GetSupportedCurrencies(ctx context.Context, slug models.AMLSlug) ([]*models.CurrencyShort, error)
	GetSignalsCategorise(ctx context.Context, slug models.AMLSlug) ([]aml.SignalCategory, error)
	ApplyVerdict(ctx context.Context, dto ApplyVerdictDTO) bool
}

var _ IService = (*Service)(nil)

type Service struct {
	st      storage.IStorage
	log     logger.Logger
	factory providers.ProviderFactory

	checkInProgress     *atomic.Bool
	checkStatusInterval time.Duration
	checkTimeout        time.Duration

	maxAttempts   int32
	eventListener event.IListener
	wallets       wallet.IWalletService
	blockedTx     transactions.IBlockedTransaction
}

func NewService(st storage.IStorage, factory providers.ProviderFactory, log logger.Logger, conf config.AML, eventListener event.IListener, wallets wallet.IWalletService, blockedTx transactions.IBlockedTransaction) *Service {
	return &Service{
		st:                  st,
		factory:             factory,
		log:                 log,
		checkInProgress:     &atomic.Bool{},
		checkStatusInterval: conf.CheckInterval,
		maxAttempts:         conf.MaxAttempts,
		checkTimeout:        conf.CheckTimeout,
		eventListener:       eventListener,
		wallets:             wallets,
		blockedTx:           blockedTx,
	}
}

func (s *Service) ApplyVerdict(ctx context.Context, dto ApplyVerdictDTO) bool {
	blocked, shouldMarkDirty := EvaluateRiskRules(dto.Check.Score, dto.Signals, dto.Rules)
	if shouldMarkDirty {
		usr, err := s.st.Users().GetByID(ctx, dto.UserID)
		if err != nil {
			s.log.Errorw("failed to get user for mark address dirty", "error", err)
		} else if markErr := s.wallets.MarkAddressDirty(ctx, usr, dto.ToAddress); markErr != nil {
			s.log.Errorw("failed to mark address as dirty", "error", markErr)
		}
	}

	if blocked {
		if !dto.WalletID.Valid {
			s.log.Errorw("cannot record blocked transaction: wallet id missing", "tx_id", dto.TransactionID)
			return blocked
		}

		riskLevel := models.AmlRiskLevel(models.AmlRiskLevelUndefined)
		if dto.Check.RiskLevel != nil {
			riskLevel = *dto.Check.RiskLevel
		}

		if _, err := s.blockedTx.CreateBlockedTransaction(ctx, transactions.CreateBlockedTransactionDTO{
			UserID:        dto.UserID,
			StoreID:       dto.StoreID,
			TransactionID: dto.TransactionID,
			AmlCheckID:    dto.Check.ID,
			WalletID:      dto.WalletID.UUID,
			RiskLevel:     string(riskLevel),
			Score:         dto.Check.Score,
		}); err != nil {
			s.log.Errorw("failed to record blocked transaction", "error", err, "tx_id", dto.TransactionID)
		}
	}

	return blocked
}

func (s *Service) ScoreTransaction(ctx context.Context, usr *models.User, dto CheckDTO) (*models.AmlCheck, error) {
	providerSlug, ok := slugMapping[dto.ProviderSlug]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedProvider, dto.ProviderSlug)
	}

	currData, err := s.st.AmlSupportedAssets().GetBySlugAndCurrencyID(ctx, dto.CurrencyID, dto.ProviderSlug)
	if err != nil {
		return nil, ErrUnsupportedCurrencies
	}

	if !avalidator.ValidateAddressByBlockchain(dto.OutputAddress, currData.Currency.Blockchain.String()) {
		return nil, fmt.Errorf("%w: '%s' for blockchain '%s'", ErrInvalidAddress, dto.OutputAddress, currData.Currency.Blockchain)
	}

	if _, err := s.factory.GetClient(providerSlug); err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	amlSvc, _, err := s.prepareServiceDataByUser(ctx, usr.ID, prepareParams{Slug: dto.ProviderSlug, ExternalID: dto.TxID})
	if err != nil {
		return nil, err
	}

	createdAmlCheck, err := s.enqueueCheck(ctx, usr.ID, *amlSvc, aml.InitCheckDTO{
		TxID: dto.TxID,
		TokenData: aml.TokenData{
			Blockchain:      currData.AmlSupportedAsset.BlockchainName,
			ContractAddress: currData.AmlSupportedAsset.AssetIdentity,
		},
		Direction:     dto.Direction.ToAMLDirection(),
		OutputAddress: dto.OutputAddress,
	}, nil, nil)
	if err != nil {
		return nil, err
	}

	return createdAmlCheck, nil
}

func (s *Service) AutoScoreDeposit(ctx context.Context, dto AutoScoreDepositDTO) (*models.AmlCheck, []aml.SignalContribution, error) {
	targetSlug, err := s.resolveProviderSlug(ctx, dto)
	if err != nil {
		return nil, nil, err
	}

	providerSlug := slugMapping[targetSlug]

	currData, err := s.st.AmlSupportedAssets().GetBySlugAndCurrencyID(ctx, dto.CurrencyID, targetSlug)
	if err != nil {
		return nil, nil, ErrUnsupportedCurrencies
	}

	if !avalidator.ValidateAddressByBlockchain(dto.OutputAddress, currData.Currency.Blockchain.String()) {
		return nil, nil, fmt.Errorf("%w: '%s' for blockchain '%s'", ErrInvalidAddress, dto.OutputAddress, currData.Currency.Blockchain)
	}

	if _, err := s.factory.GetClient(providerSlug); err != nil {
		return nil, nil, fmt.Errorf("failed to get provider: %w", err)
	}

	amlSvc, _, err := s.prepareServiceDataByUser(ctx, dto.UserID, prepareParams{Slug: targetSlug, ExternalID: dto.TxHash})
	if err != nil {
		return nil, nil, err
	}

	createdAmlCheck, err := s.enqueueCheck(ctx, dto.UserID, *amlSvc, aml.InitCheckDTO{
		TxID: dto.TxHash,
		TokenData: aml.TokenData{
			Blockchain:      currData.AmlSupportedAsset.BlockchainName,
			ContractAddress: currData.AmlSupportedAsset.AssetIdentity,
		},
		Direction:     aml.DirectionIn,
		OutputAddress: dto.OutputAddress,
	}, &dto.TxID, dto.DBTx)
	if err != nil {
		return nil, nil, err
	}
	// no signals yet — the check hasn't run; they'll be available once the queued
	// check completes and CheckCompletedEvent is fired.
	return createdAmlCheck, nil, nil
}

// resolveProviderSlug returns the provider slug to use for AutoScoreDeposit.
// Returns ErrNoProviderAvailable when no suitable provider is found.
func (s *Service) resolveProviderSlug(ctx context.Context, dto AutoScoreDepositDTO) (models.AMLSlug, error) {
	if dto.ProviderSlug != nil {
		if _, ok := slugMapping[*dto.ProviderSlug]; !ok {
			return "", fmt.Errorf("%w: %s", ErrUnsupportedProvider, *dto.ProviderSlug)
		}
		return *dto.ProviderSlug, nil
	}

	for _, provider := range s.GetAllActiveProviders() {
		if _, err := s.st.AmlUserKeys().GetServiceCredentials(ctx, dto.UserID, provider.Slug); err != nil {
			continue
		}
		if _, err := s.st.AmlSupportedAssets().GetBySlugAndCurrencyID(ctx, dto.CurrencyID, provider.Slug); err != nil {
			continue
		}
		return provider.Slug, nil
	}

	return "", ErrNoProviderAvailable
}

func (s *Service) GetAllActiveProviders() []models.AMLProvider {
	providerSlugs := s.factory.GetAllRegisteredProviders()

	preparedResult := make([]models.AMLProvider, 0, len(providerSlugs))
	for _, providerSlug := range providerSlugs {
		for modelSlug, mappedSlug := range slugMapping {
			if mappedSlug == providerSlug {
				preparedResult = append(preparedResult, models.AMLProvider{
					Slug:  modelSlug,
					Label: modelSlug.Label(),
				})
				break
			}
		}
	}

	return preparedResult
}

func (s *Service) GetSupportedCurrencies(ctx context.Context, slug models.AMLSlug) ([]*models.CurrencyShort, error) {
	if err := s.ensureProviderEnabled(slug); err != nil {
		return nil, err
	}

	res, err := s.st.AmlSupportedAssets().GetAllBySlug(ctx, slug)
	if err != nil {
		return nil, ErrUnsupportedCurrencies
	}

	preparedCurrs := make([]*models.CurrencyShort, 0, len(res))
	for _, data := range res {
		preparedCurrs = append(preparedCurrs, &models.CurrencyShort{
			ID:           data.Currency.ID,
			Code:         data.Currency.Code,
			Precision:    data.Currency.Precision,
			Name:         data.Currency.Name,
			Blockchain:   data.Currency.Blockchain,
			IsStableCoin: data.Currency.IsStablecoin,
		})
	}

	return preparedCurrs, nil
}

func (s *Service) GetSignalsCategorise(_ context.Context, slug models.AMLSlug) ([]aml.SignalCategory, error) {
	if err := s.ensureProviderEnabled(slug); err != nil {
		return nil, err
	}

	client, err := s.factory.GetClient(slugMapping[slug])
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	lister, ok := client.(aml.SignalCategoryLister)
	if !ok {
		return []aml.SignalCategory{}, nil
	}

	return lister.SignalCategories(), nil
}

type prepareParams struct {
	Slug       models.AMLSlug
	ExternalID string
}

func (s *Service) prepareServiceDataByUser(
	ctx context.Context,
	usrID uuid.UUID,
	params prepareParams,
	opts ...repos.Option,
) (*models.AmlService, aml.RequestAuthorizer, error) {
	serviceData, err := s.st.AmlUserKeys(opts...).GetServiceCredentials(ctx, usrID, params.Slug)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get service credentials: %w", err)
	}

	auth, err := s.prepareCreedsBySlug(ctx, params.Slug, serviceData.Creds, params.ExternalID)
	if err != nil {
		return nil, nil, err
	}

	return serviceData.Service, auth, nil
}

func (s *Service) prepareCreedsBySlug(
	ctx context.Context,
	slug models.AMLSlug,
	creds map[models.AmlKeyType]string,
	externalID string,
) (aml.RequestAuthorizer, error) {
	providerSlug, ok := slugMapping[slug]
	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", slug)
	}

	mappedCreds := make(map[aml.AuthKeyType]string, len(creds))
	for modelKey, value := range creds {
		if amlKey, ok := keyMapping[modelKey]; ok {
			mappedCreds[amlKey] = value
		}
	}

	return s.factory.CreateAuthorizer(ctx, providerSlug, mappedCreds, externalID)
}

const pendingExternalIDPrefix = "pending:"

func isPendingExternalID(externalID string) bool {
	return strings.HasPrefix(externalID, pendingExternalIDPrefix)
}

func (s *Service) enqueueCheck(ctx context.Context, usrID uuid.UUID, service models.AmlService, dto aml.InitCheckDTO, txID *uuid.UUID, outerTx pgx.Tx) (*models.AmlCheck, error) {
	payload, err := json.Marshal(dto)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal check payload: %w", err)
	}

	var createdCheck *models.AmlCheck

	insertFn := func(tx pgx.Tx) error {
		params := repo_aml_checks.CreateParams{
			UserID:     usrID,
			ServiceID:  service.ID,
			ExternalID: pendingExternalIDPrefix + uuid.NewString(),
			Status:     models.AmlCheckStatusPending,
			Score:      decimal.Zero,
		}

		if txID != nil {
			params.TransactionID = uuid.NullUUID{UUID: *txID, Valid: true}
		}

		var err error
		createdCheck, err = s.st.AmlChecks(repos.WithTx(tx)).Create(ctx, params)
		if err != nil {
			return err
		}

		return s.st.AmlCheckQueue(repos.WithTx(tx)).Create(ctx, usrID, createdCheck.ID, payload)
	}

	var txErr error
	if outerTx != nil {
		txErr = insertFn(outerTx)
	} else {
		txErr = repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, insertFn)
	}
	if txErr != nil {
		return nil, txErr
	}

	return createdCheck, nil
}

func (s *Service) ensureProviderEnabled(slug models.AMLSlug) error {
	amlSLug, ok := slugMapping[slug]
	if !ok {
		return ErrUnsupportedProvider
	}

	_, err := s.factory.GetClient(amlSLug)
	if err != nil {
		return ErrUnsupportedProvider
	}

	return nil
}
