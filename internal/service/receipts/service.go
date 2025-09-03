package receipts

import (
	"context"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currency"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_receipts"

	"github.com/google/uuid"
)

type IReceiptService interface {
	GetByID(ctx context.Context, ID uuid.UUID) (*models.Receipt, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, page int32) ([]*models.Receipt, error)
	GetAll(ctx context.Context, page int32) ([]*models.Receipt, error)
	Create(ctx context.Context, params repo_receipts.CreateParams, opts ...repos.Option) (*models.Receipt, error)
}

type Service struct {
	currency currency.ICurrency
	storage  storage.IStorage
}

const PageSize = 10

func New(storage storage.IStorage, currService currency.ICurrency) *Service {
	return &Service{
		currService,
		storage,
	}
}

func (s Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Receipt, error) {
	receipt, err := s.storage.Receipts().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func (s Service) GetByUserID(ctx context.Context, userID uuid.UUID, page int32) ([]*models.Receipt, error) {
	params := repo_receipts.GetByUserIdParams{
		UserID: userID,
		Limit:  PageSize,
		Offset: (page - 1) * PageSize,
	}

	receipt, err := s.storage.Receipts().GetByUserId(ctx, params)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func (s Service) GetAll(ctx context.Context, page int32) ([]*models.Receipt, error) {
	params := repo_receipts.GetAllParams{
		Limit:  PageSize,
		Offset: (page - 1) * PageSize,
	}

	receipt, err := s.storage.Receipts().GetAll(ctx, params)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func (s Service) Create(ctx context.Context, params repo_receipts.CreateParams, opts ...repos.Option) (*models.Receipt, error) {
	receipt, err := s.storage.Receipts(opts...).Create(ctx, params)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}
