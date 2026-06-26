package aml

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_check_history"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_checks"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/pkg/aml"
	"github.com/dv-net/dv-merchant/pkg/aml/providers"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/google/uuid"

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
	AutoScoreDeposit(ctx context.Context, dto AutoScoreDepositDTO) (*models.AmlCheck, error)
	GetCheckHistory(ctx context.Context, usr *models.User, dto ChecksWithHistoryDTO) (*storecmn.FindResponseWithFullPagination[*repo_aml_checks.FindRow], error)
	GetAllActiveProviders() []models.AMLSlug
	GetSupportedCurrencies(ctx context.Context, slug models.AMLSlug) ([]*models.CurrencyShort, error)
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
}

func NewService(st storage.IStorage, factory providers.ProviderFactory, log logger.Logger, conf config.AML, eventListener event.IListener) *Service {
	return &Service{
		st:                  st,
		factory:             factory,
		log:                 log,
		checkInProgress:     &atomic.Bool{},
		checkStatusInterval: conf.CheckInterval,
		maxAttempts:         conf.MaxAttempts,
		checkTimeout:        conf.CheckTimeout,
		eventListener:       eventListener,
	}
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

	provider, err := s.factory.GetClient(providerSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	amlSvc, auth, err := s.prepareServiceDataByUser(ctx, usr.ID, prepareParams{Slug: dto.ProviderSlug, ExternalID: dto.TxID})
	if err != nil {
		return nil, err
	}

	check, err := provider.InitCheckTransaction(ctx, aml.InitCheckDTO{
		TxID: dto.TxID,
		TokenData: aml.TokenData{
			Blockchain:      currData.AmlSupportedAsset.BlockchainName,
			ContractAddress: currData.AmlSupportedAsset.AssetIdentity,
		},
		Direction:     dto.Direction.ToAMLDirection(),
		OutputAddress: dto.OutputAddress,
	}, auth)
	if err != nil {
		return nil, err
	}

	createdAmlCheck, err := s.createCheck(ctx, usr.ID, *amlSvc, check, nil, nil)
	if err != nil {
		return nil, err
	}

	return createdAmlCheck, nil
}

func (s *Service) AutoScoreDeposit(ctx context.Context, dto AutoScoreDepositDTO) (*models.AmlCheck, error) {
	targetSlug, err := s.resolveProviderSlug(ctx, dto)
	if err != nil {
		return nil, err
	}

	providerSlug := slugMapping[targetSlug]

	currData, err := s.st.AmlSupportedAssets().GetBySlugAndCurrencyID(ctx, dto.CurrencyID, targetSlug)
	if err != nil {
		return nil, ErrUnsupportedCurrencies
	}

	if !avalidator.ValidateAddressByBlockchain(dto.OutputAddress, currData.Currency.Blockchain.String()) {
		return nil, fmt.Errorf("%w: '%s' for blockchain '%s'", ErrInvalidAddress, dto.OutputAddress, currData.Currency.Blockchain)
	}

	provider, err := s.factory.GetClient(providerSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	amlSvc, auth, err := s.prepareServiceDataByUser(ctx, dto.UserID, prepareParams{Slug: targetSlug, ExternalID: dto.TxHash})
	if err != nil {
		return nil, err
	}

	check, err := provider.InitCheckTransaction(ctx, aml.InitCheckDTO{
		TxID: dto.TxHash,
		TokenData: aml.TokenData{
			Blockchain:      currData.AmlSupportedAsset.BlockchainName,
			ContractAddress: currData.AmlSupportedAsset.AssetIdentity,
		},
		Direction:     aml.DirectionIn,
		OutputAddress: dto.OutputAddress,
	}, auth)
	if err != nil {
		return nil, err
	}

	createdAmlCheck, err := s.createCheck(ctx, dto.UserID, *amlSvc, check, &dto.TxID, dto.DBTx)
	if err != nil {
		return nil, err
	}
	return createdAmlCheck, nil
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

	for _, slug := range s.GetAllActiveProviders() {
		if _, err := s.st.AmlUserKeys().GetServiceCredentials(ctx, dto.UserID, slug); err != nil {
			continue
		}
		if _, err := s.st.AmlSupportedAssets().GetBySlugAndCurrencyID(ctx, dto.CurrencyID, slug); err != nil {
			continue
		}
		return slug, nil
	}

	return "", ErrNoProviderAvailable
}

func (s *Service) GetAllActiveProviders() []models.AMLSlug {
	providerSlugs := s.factory.GetAllRegisteredProviders()

	preparedResult := make([]models.AMLSlug, 0, len(providerSlugs))
	for _, providerSlug := range providerSlugs {
		for modelSlug, mappedSlug := range slugMapping {
			if mappedSlug == providerSlug {
				preparedResult = append(preparedResult, modelSlug)
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

	auth, err := s.prepareCredsBySlug(ctx, params.Slug, serviceData.Creds, params.ExternalID)
	if err != nil {
		return nil, nil, err
	}

	return serviceData.Service, auth, nil
}

func (s *Service) prepareCredsBySlug(
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

func (s *Service) createCheck(ctx context.Context, usrID uuid.UUID, service models.AmlService, check *aml.CheckResponse, txID *uuid.UUID, outerTx pgx.Tx) (*models.AmlCheck, error) {
	var createdCheck *models.AmlCheck

	insertFn := func(tx pgx.Tx) error {
		var err error

		var preparedRiskLvl *models.AmlRiskLevel
		if check.RiskLevel != nil {
			preparedRiskLvl, err = convertAmlRiskLevelToModel(*check.RiskLevel)
			if err != nil {
				return err
			}
		}
		params := repo_aml_checks.CreateParams{
			UserID:     usrID,
			ServiceID:  service.ID,
			ExternalID: check.ExternalID,
			Status:     convertAmlStatusToModel(check.Status),
			Score:      check.Score,
			RiskLevel:  preparedRiskLvl,
		}

		if txID != nil {
			params.TransactionID = uuid.NullUUID{UUID: *txID, Valid: true}
		}

		createdCheck, err = s.st.AmlChecks(repos.WithTx(tx)).Create(ctx, params)
		if err != nil {
			return err
		}

		if _, err = s.st.AmlCheckHistory(repos.WithTx(tx)).Create(ctx, repo_aml_check_history.CreateParams{
			AmlCheckID:      createdCheck.ID,
			RequestPayload:  check.Request,
			ServiceResponse: check.Response,
		}); err != nil {
			return err
		}

		if createdCheck.Status == models.AmlCheckStatusPending {
			return s.st.AmlCheckQueue(repos.WithTx(tx)).Create(ctx, usrID, createdCheck.ID)
		}

		return nil
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
