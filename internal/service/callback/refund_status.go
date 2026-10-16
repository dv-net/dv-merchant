package callback

import (
	"context"
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/constants"
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/processing_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/refund"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_refund_requests"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// syncRefundRequestStatus moves the refund request of a refund transfer by the transfer status.
// Transfers without a refund request are skipped. The update is conditional, so repeated
// or late webhooks do not move a finished request back.
//
//   - processing / in_mempool / unconfirmed: scheduled -> processing (the cooldown is over)
//   - completed: scheduled, processing -> completed
//   - failed before the broadcast / canceled: scheduled, processing -> pending_review with the reason,
//     the transfer is detached so the admin can approve the refund again or reject it
//   - failed after the broadcast: scheduled, processing -> failed, the transfer stays attached,
//     the transaction may still be mined and is checked manually
func (s *Service) syncRefundRequestStatus(ctx context.Context, dto processing_request.TransferStatusWebhook, tx pgx.Tx) error {
	if dto.RequestID == nil {
		return nil
	}

	params := repo_refund_requests.UpdateStatusByTransferIDParams{
		TransferID:      uuid.NullUUID{UUID: *dto.RequestID, Valid: true},
		CurrentStatuses: []string{constants.RefundStatusScheduled, constants.RefundStatusProcessing},
	}

	switch dto.Status {
	case models.TransferStatusProcessing, models.TransferStatusInMempool, models.TransferStatusUnconfirmed:
		params.NewStatus = constants.RefundStatusProcessing
		params.CurrentStatuses = []string{constants.RefundStatusScheduled}
	case models.TransferStatusCompleted:
		params.NewStatus = constants.RefundStatusCompleted
	case models.TransferStatusFailed:
		// the refund transaction may be in the network and still be mined: the request must not be
		// approved again, processing keeps the deposit locked until it is checked manually
		if refundTransferBroadcast(dto) {
			params.NewStatus = constants.RefundStatusFailed
			params.LastError = refundErrorText("refund transfer failed after it was sent to the network, check the transaction manually", dto.ErrorMessage)
			break
		}

		params.NewStatus = constants.RefundStatusPendingReview
		params.ResetTransfer = true
		params.LastError = refundErrorText("refund transfer failed", dto.ErrorMessage)
	case models.TransferStatusCanceled:
		params.NewStatus = constants.RefundStatusPendingReview
		params.ResetTransfer = true
		params.LastError = refundErrorText("refund transfer was canceled", dto.ErrorMessage)
	default:
		return nil
	}

	updated, err := s.storage.RefundRequests(repos.WithTx(tx)).UpdateStatusByTransferID(ctx, params)
	if err != nil {
		return fmt.Errorf("update refund request by transfer status: %w", err)
	}

	if updated > 0 {
		s.log.Infow("refund request status updated by transfer status",
			"transfer_id", dto.RequestID.String(),
			"transfer_status", dto.Status,
			"refund_status", params.NewStatus,
		)
	}

	return nil
}

// refundTransferBroadcast checks if the failed refund transaction may be in the network
func refundTransferBroadcast(dto processing_request.TransferStatusWebhook) bool {
	if dto.Broadcast {
		return true
	}

	for _, sysTx := range dto.SystemTransactions {
		if sysTx.TxType == models.TransferTransactionTypeTransfer && sysTx.TxHash != "" {
			return true
		}
	}

	return false
}

func refundErrorText(reason string, message *string) pgtype.Text {
	text := reason
	if message != nil && *message != "" {
		text = reason + ": " + *message
	}

	return pgtype.Text{String: refund.TruncateRefundError(text), Valid: true}
}

// hotWalletTransferTxType returns the type of an outgoing hot wallet transaction:
// refund for transactions of refund transfers, transfer otherwise
func (s *Service) hotWalletTransferTxType(ctx context.Context, dto TransferWebhookDto, tx pgx.Tx) (models.TransactionsType, error) {
	if !dto.TransferID.Valid {
		return models.TransactionsTypeTransfer, nil
	}

	transfer, err := s.storage.Transfers(repos.WithTx(tx)).GetById(ctx, dto.TransferID.UUID)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.TransactionsTypeTransfer, nil
	}
	if err != nil {
		return "", fmt.Errorf("fetch transfer %s: %w", dto.TransferID.UUID, err)
	}

	if transfer.Kind == models.TransferKindRefund {
		return models.TransactionsTypeRefund, nil
	}

	return models.TransactionsTypeTransfer, nil
}
