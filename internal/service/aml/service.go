package aml

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_check_history"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_checks"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/pkg/aml"
	"github.com/dv-net/dv-merchant/pkg/aml/providers"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/dv-net/dv-processing/pkg/avalidator"

	"github.com/jackc/pgx/v5"
)

// slugMapping maps internal/models.AMLSlug ob providers.ProviderSlug.
var slugMapping = map[models.AMLSlug]aml.ProviderSlug{
	models.AMLSlugAMLBot: aml.ProviderSlugAMLBot,
	models.AMLSlugBitOK:  aml.ProviderSlugBitOK,
}

// keyMapping maps internal/models.AmlKeyType to aml.AMLKeyType.
var keyMapping = map[models.AmlKeyType]aml.AuthKeyType{
	models.AmlKeyTypeAccessKeyID: aml.KeyAccessKeyID,
	models.AmlKeyTypeAccessKey:   aml.KeyAccessKey,
	models.AmlKeyTypeSecret:      aml.KeySecret,
	models.AmlKeyTypeAccessID:    aml.KeyAccessID,
}

type IService interface {
	ScoreTransaction(ctx context.Context, usr *models.User, dto CheckDTO) (*models.AmlCheck, error)
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

	maxAttempts int32
}

func NewService(st storage.IStorage, factory providers.ProviderFactory, log logger.Logger, conf config.AML) *Service {
	return &Service{
		st:                  st,
		factory:             factory,
		log:                 log,
		checkInProgress:     &atomic.Bool{},
		checkStatusInterval: conf.CheckInterval,
		maxAttempts:         conf.MaxAttempts,
		checkTimeout:        conf.CheckTimeout,
	}
}

func (s *Service) ScoreTransaction(ctx context.Context, usr *models.User, dto CheckDTO) (*models.AmlCheck, error) {
	providerSlug, ok := slugMapping[dto.ProviderSlug]
	if !ok {
		return nil, fmt.Errorf("unsupported or disabled provider: %s", dto.ProviderSlug)
	}

	currData, err := s.st.AmlSupportedAssets().GetBySlugAndCurrencyID(ctx, dto.CurrencyID, dto.ProviderSlug)
	if err != nil {
		return nil, errors.New("currency is not supported")
	}

	if !avalidator.ValidateAddressByBlockchain(dto.OutputAddress, currData.Currency.Blockchain.String()) {
		return nil, fmt.Errorf("invalid address '%s' for blockchain '%s'", dto.OutputAddress, currData.Currency.Blockchain)
	}

	provider, err := s.factory.GetClient(providerSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	amlSvc, auth, err := s.prepareServiceDataByUser(ctx, usr, prepareParams{Slug: dto.ProviderSlug, ExternalID: dto.TxID})
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

	createdAmlCheck, err := s.createCheck(ctx, usr, *amlSvc, check)
	if err != nil {
		return nil, err
	}

	return createdAmlCheck, nil
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
			ID:         data.Currency.ID,
			Code:       data.Currency.Code,
			Precision:  data.Currency.Precision,
			Name:       data.Currency.Name,
			Blockchain: data.Currency.Blockchain,
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
	usr *models.User,
	params prepareParams,
	opts ...repos.Option,
) (*models.AmlService, aml.RequestAuthorizer, error) {
	serviceData, err := s.st.AmlUserKeys(opts...).GetServiceCredentials(ctx, usr.ID, params.Slug)
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

// createCheck creates check entity and enqueues it if polling is required by status
func (s *Service) createCheck(ctx context.Context, usr *models.User, service models.AmlService, check *aml.CheckResponse) (*models.AmlCheck, error) {
	var createdCheck *models.AmlCheck
	txErr := repos.BeginTxFunc(ctx, s.st.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		var err error

		var preparedRiskLvl *models.AmlRiskLevel
		if check.RiskLevel != nil {
			preparedRiskLvl, err = convertAmlRiskLevelToModel(*check.RiskLevel)
			if err != nil {
				return err
			}
		}

		createdCheck, err = s.st.AmlChecks(repos.WithTx(tx)).Create(ctx, repo_aml_checks.CreateParams{
			UserID:     usr.ID,
			ServiceID:  service.ID,
			ExternalID: check.ExternalID,
			Status:     convertAmlStatusToModel(check.Status),
			Score:      check.Score,
			RiskLevel:  preparedRiskLvl,
		})
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

		// Enqueue aml check if status requires
		if createdCheck.Status == models.AmlCheckStatusPending {
			return s.st.AmlCheckQueue(repos.WithTx(tx)).Create(ctx, usr.ID, createdCheck.ID)
		}

		return nil
	})
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
