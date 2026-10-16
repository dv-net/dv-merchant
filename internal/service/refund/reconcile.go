package refund

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/dv-net/dv-merchant/internal/constants"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_refund_requests"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	reconcileInterval = time.Minute
	// reconcileGrace is longer than any processing call: a request older than it is not being sent
	// right now, so processing has either created its transfer or will never create it
	reconcileGrace    = 10 * time.Minute
	reconcileMaxItems = 100
)

// RunReconciler periodically resolves refund requests which are in flight for too long
func (s *Service) RunReconciler(ctx context.Context) {
	var inProgress atomic.Bool

	ticker := time.NewTicker(reconcileInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !inProgress.CompareAndSwap(false, true) {
				continue
			}

			if err := s.ReconcileInFlight(ctx); err != nil {
				s.logger.Errorw("refund reconciliation failed", "error", err)
			}

			inProgress.Store(false)
		}
	}
}

// ReconcileInFlight syncs refund requests in processing / scheduled with their transfers in processing.
// It covers lost status webhooks and refunds whose processing call result was unknown.
func (s *Service) ReconcileInFlight(ctx context.Context) error {
	requests, err := s.storage.RefundRequests().GetInFlightForReconcile(ctx, repo_refund_requests.GetInFlightForReconcileParams{
		Statuses:      []string{constants.RefundStatusProcessing, constants.RefundStatusScheduled},
		UpdatedBefore: pgtypeutils.EncodeTime(time.Now().Add(-reconcileGrace)),
		MaxItems:      reconcileMaxItems,
	})
	if err != nil {
		return fmt.Errorf("fetch in flight refund requests: %w", err)
	}

	for _, ref := range requests {
		err := s.reconcileOne(ctx, ref)
		if errors.Is(err, processing.ErrServiceNotInitialized) {
			return nil
		}
		if err != nil {
			s.logger.Errorw("failed to reconcile refund request",
				"refund_request_id", ref.ID,
				"transfer_id", ref.TransferID.UUID,
				"error", err,
			)
		}
	}

	return nil
}

func (s *Service) reconcileOne(ctx context.Context, ref *models.RefundRequest) error {
	info, err := s.processing.GetTransferByRequestID(ctx, ref.TransferID.UUID)
	if errors.Is(err, processing.ErrTransferNotFound) {
		if ref.Status != constants.RefundStatusProcessing {
			return fmt.Errorf("scheduled refund transfer %s is not found in processing", ref.TransferID.UUID)
		}

		s.logger.Infow("refund transfer was not created in processing, returning the request to review",
			"refund_request_id", ref.ID,
			"transfer_id", ref.TransferID.UUID,
		)

		return s.revertUnsent(ctx, ref.ID, ref.TransferID.UUID, "refund was not created in processing")
	}
	if err != nil {
		return fmt.Errorf("get transfer from processing: %w", err)
	}

	params := repo_refund_requests.UpdateStatusByTransferIDParams{
		TransferID:      ref.TransferID,
		CurrentStatuses: []string{constants.RefundStatusScheduled, constants.RefundStatusProcessing},
	}

	switch info.Status {
	case models.TransferStatusNew:
		// created, but the result was not saved: scheduled with the cooldown or waiting for execution
		_, err := s.markSent(ctx, ref.ID, ref.TransferID, info.ExecuteAfter)
		return err
	case models.TransferStatusPending, models.TransferStatusProcessing,
		models.TransferStatusInMempool, models.TransferStatusUnconfirmed, models.TransferStatusFrozen:
		// sending: the request is processing, updated_at is refreshed to check it later again
		params.NewStatus = constants.RefundStatusProcessing
		params.LastError = ref.LastError
	case models.TransferStatusCompleted:
		params.NewStatus = constants.RefundStatusCompleted
	case models.TransferStatusFailed:
		if info.Broadcast {
			params.NewStatus = constants.RefundStatusFailed
			params.LastError = reconcileErrorText("refund transfer failed after it was sent to the network, check the transaction manually", info.ErrorMessage)
			break
		}

		params.NewStatus = constants.RefundStatusPendingReview
		params.ResetTransfer = true
		params.LastError = reconcileErrorText("refund transfer failed", info.ErrorMessage)
	case models.TransferStatusCanceled:
		params.NewStatus = constants.RefundStatusPendingReview
		params.ResetTransfer = true
		params.LastError = reconcileErrorText("refund transfer was canceled", info.ErrorMessage)
	default:
		return fmt.Errorf("unknown transfer status %q", info.Status)
	}

	updated, err := s.storage.RefundRequests().UpdateStatusByTransferID(ctx, params)
	if err != nil {
		return fmt.Errorf("update refund request by transfer status: %w", err)
	}

	if updated > 0 && params.NewStatus != ref.Status {
		s.logger.Infow("refund request reconciled with processing",
			"refund_request_id", ref.ID,
			"transfer_id", ref.TransferID.UUID,
			"transfer_status", info.Status,
			"refund_status", params.NewStatus,
		)
	}

	return nil
}

func reconcileErrorText(reason, message string) pgtype.Text {
	text := reason
	if message != "" {
		text = reason + ": " + message
	}

	return pgtype.Text{String: TruncateRefundError(text), Valid: true}
}
