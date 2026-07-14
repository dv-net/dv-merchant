package store

import (
	"context"
	"errors"

	"github.com/dv-net/dv-merchant/internal/constants"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_stores"

	"github.com/jackc/pgx/v5/pgtype"
)

type IVerify interface {
	VerifyStore(ctx context.Context, dto VerifyStoreDTO) (*models.Store, error)
	RejectStore(ctx context.Context, dto RejectStoreDTO) (*models.Store, error)
	ResendStoreVerification(ctx context.Context, dto ResendStoreVerificationDTO) (*models.Store, error)
}

var _ IVerify = (*Service)(nil)

var ErrRejectionReasonRequired = errors.New("rejection reason is required")

func (s *Service) VerifyStore(ctx context.Context, dto VerifyStoreDTO) (*models.Store, error) {
	store, err := s.storage.Stores().UpdateVerificationStatus(ctx, repo_stores.UpdateVerificationStatusParams{
		VerificationStatus: constants.StoreVerificationStatusSuccess.String(),
		AdminID:            dto.Admin.ID,
		StoreID:            dto.StoreID,
	})
	if err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Service) RejectStore(ctx context.Context, dto RejectStoreDTO) (*models.Store, error) {
	if dto.Reason == "" {
		return nil, ErrRejectionReasonRequired
	}

	store, err := s.storage.Stores().UpdateVerificationStatus(ctx, repo_stores.UpdateVerificationStatusParams{
		VerificationStatus: constants.StoreVerificationStatusRejected.String(),
		AdminID:            dto.Admin.ID,
		RejectionReason:    pgtype.Text{String: dto.Reason, Valid: true},
		StoreID:            dto.StoreID,
	})
	if err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Service) ResendStoreVerification(ctx context.Context, dto ResendStoreVerificationDTO) (*models.Store, error) {
	store, err := s.storage.Stores().ResendStoreVerification(ctx, repo_stores.ResendStoreVerificationParams{
		VerificationStatus: constants.StoreVerificationStatusPending.String(),
		StoreID:            dto.StoreID,
	})
	if err != nil {
		return nil, err
	}
	return store, nil
}
