package refund

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dv-net/dv-merchant/internal/constants"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_refund_requests"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"

	"github.com/dv-net/dv-processing/pkg/avalidator"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type IRefundService interface {
	CreateRefund(ctx context.Context, dto CreateRefundDTO) (*models.RefundRequest, error)
	GetCabinet(ctx context.Context, walletID uuid.UUID) (map[string][]*CabinetItem, error)
	GetPendingReview(ctx context.Context, storeID uuid.UUID) ([]*models.RefundRequest, error)
	RejectRefund(ctx context.Context, dto RejectRefundDTO) (*models.RefundRequest, error)
}

var ErrRefundAlreadyRequested = errors.New("a refund request already exists for this transaction")

type Service struct {
	logger  logger.Logger
	storage storage.IStorage
}

func NewService(log logger.Logger, st storage.IStorage) *Service {
	return &Service{
		logger:  log,
		storage: st,
	}
}

func (s *Service) CreateRefund(ctx context.Context, dto CreateRefundDTO) (*models.RefundRequest, error) {
	btx, err := s.storage.BlockedTransactions().GetById(ctx, dto.BlockedTransactionID)
	if err != nil {
		return nil, fmt.Errorf("fetch blocked transaction: %w", err)
	}

	if btx.WalletID != dto.WalletID {
		return nil, fmt.Errorf("wallet id mismatch: blocked transaction wallet id")
	}

	if _, err := s.storage.RefundRequests().GetByBlockedTransactionID(ctx, dto.BlockedTransactionID); err == nil {
		return nil, ErrRefundAlreadyRequested
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("check existing refund request: %w", err)
	}

	tx, err := s.storage.Transactions().GetById(ctx, btx.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("fetch transaction: %w", err)
	}

	if !avalidator.ValidateAddressByBlockchain(dto.DestinationAddress, tx.Blockchain.String()) {
		return nil, fmt.Errorf("invalid destination address for blockchain %s", tx.Blockchain.String())
	}

	ref, err := s.storage.RefundRequests().Create(ctx, repo_refund_requests.CreateParams{
		BlockedTransactionID: dto.BlockedTransactionID,
		WalletID:             dto.WalletID,
		StoreID:              btx.StoreID,
		DestinationAddress:   dto.DestinationAddress,
		Email:                dto.Email,
		Status:               constants.RefundStatusPendingReview,
	})
	if err != nil {
		return nil, fmt.Errorf("create refund request: %w", err)
	}
	return ref, nil
}

func (s *Service) GetCabinet(ctx context.Context, walletID uuid.UUID) (map[string][]*CabinetItem, error) {
	blocked, err := s.storage.BlockedTransactions().GetAllByWalletID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("fetch blocked transactions: %w", err)
	}

	refunds, err := s.storage.RefundRequests().GetAllByWalletID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("fetch refund requests: %w", err)
	}

	txByID := make(map[uuid.UUID]*models.Transaction, len(blocked))
	for _, b := range blocked {
		tx, err := s.storage.Transactions().GetById(ctx, b.TransactionID)
		if err != nil {
			return nil, fmt.Errorf("fetch transaction %s: %w", b.TransactionID, err)
		}
		txByID[b.TransactionID] = tx
	}

	return buildCabinet(blocked, refunds, txByID), nil
}

func (s *Service) GetPendingReview(ctx context.Context, storeID uuid.UUID) ([]*models.RefundRequest, error) {
	list, err := s.storage.RefundRequests().GetAllByStoreIDAndStatus(ctx, repo_refund_requests.GetAllByStoreIDAndStatusParams{
		StoreID: storeID,
		Status:  constants.RefundStatusPendingReview,
	})
	if err != nil {
		return nil, fmt.Errorf("fetch pending refund requests: %w", err)
	}
	return list, nil
}

func (s *Service) RejectRefund(ctx context.Context, dto RejectRefundDTO) (*models.RefundRequest, error) {
	ref, err := s.storage.RefundRequests().GetById(ctx, dto.RefundRequestID)
	if err != nil {
		return nil, fmt.Errorf("fetch refund request: %w", err)
	}

	if ref.StoreID != dto.StoreID {
		return nil, fmt.Errorf("store id mismatch: refund request store id")
	}

	if ref.Status != constants.RefundStatusPendingReview {
		return nil, fmt.Errorf("refund request is not pending review")
	}

	updated, err := s.storage.RefundRequests().Update(ctx, repo_refund_requests.UpdateParams{
		ID:         ref.ID,
		Status:     constants.RefundStatusRejected,
		TransferID: ref.TransferID,
		ReviewedAt: pgtypeutils.EncodeTime(time.Now()),
	})
	if err != nil {
		return nil, fmt.Errorf("reject refund request: %w", err)
	}
	return updated, nil
}
