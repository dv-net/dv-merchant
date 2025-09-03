package address

import (
	"context"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/google/uuid"
)

type IWalletAddressService interface {
	GetAddressesByWalletID(ctx context.Context, ID uuid.UUID) ([]*models.WalletAddress, error)
}

type Service struct {
	cfg               *config.Config
	storage           storage.IStorage
	logger            logger.Logger
	processingService processing.IProcessingWallet
}

func New(
	cfg *config.Config,
	storage storage.IStorage,
	logger logger.Logger,
	processingService processing.IProcessingWallet,
) *Service {
	return &Service{
		cfg:               cfg,
		storage:           storage,
		logger:            logger,
		processingService: processingService,
	}
}

func (s Service) GetAddressesByWalletID(ctx context.Context, id uuid.UUID) ([]*models.WalletAddress, error) {
	walletAddress, err := s.storage.WalletAddresses().GetWalletAddressesByWalletId(ctx, id)
	if err != nil {
		return nil, err
	}
	return walletAddress, nil
}
