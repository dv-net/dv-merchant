package withdraw

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transfers"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// refundData is the refund prepared from the blocked deposit
type refundData struct {
	user      *models.User
	tx        *models.Transaction
	amountUsd decimal.Decimal
	params    processing.FundsWithdrawalParams
}

// prepareRefund loads the blocked deposit of the refund request and builds the processing params:
// the whole deposit amount from the deposit address to the refund destination
func (s *service) prepareRefund(ctx context.Context, ref *models.RefundRequest, opts ...repos.Option) (*refundData, error) {
	if ref == nil {
		return nil, fmt.Errorf("refund request is required")
	}

	btx, err := s.storage.BlockedTransactions(opts...).GetById(ctx, ref.BlockedTransactionID)
	if err != nil {
		return nil, fmt.Errorf("fetch blocked transaction: %w", err)
	}

	if btx.WalletID != ref.WalletID {
		return nil, fmt.Errorf("blocked transaction does not belong to the refund wallet")
	}

	tx, err := s.storage.Transactions(opts...).GetById(ctx, btx.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("fetch transaction: %w", err)
	}

	user, err := s.storage.Users(opts...).GetByID(ctx, btx.UserID)
	if err != nil {
		return nil, fmt.Errorf("fetch user: %w", err)
	}

	if !user.ProcessingOwnerID.Valid {
		return nil, ErrProcessingUninitialized
	}

	curr, err := s.currencyService.GetCurrencyByID(ctx, tx.CurrencyID)
	if err != nil {
		return nil, fmt.Errorf("fetch currency: %w", err)
	}
	if curr.Blockchain == nil {
		return nil, ErrFiatCurrencyIsNotSupported
	}

	amountUsd := tx.AmountUsd.Decimal
	if !tx.AmountUsd.Valid {
		rate, err := s.currencyRate(ctx, user.RateSource.String(), curr)
		if err != nil {
			return nil, err
		}
		amountUsd = rate.Mul(tx.Amount)
	}

	return &refundData{
		user:      user,
		tx:        tx,
		amountUsd: amountUsd,
		params: processing.FundsWithdrawalParams{
			OwnerID:         user.ProcessingOwnerID.UUID,
			Blockchain:      tx.Blockchain,
			FromAddress:     []string{tx.ToAddress},
			ToAddress:       []string{ref.DestinationAddress},
			ContractAddress: curr.ContractAddress.String,
			WholeAmount:     false,
			Amount:          tx.Amount.String(),
			Purpose:         processing.TransferPurposeRefund,
			SourceTxHash:    &tx.TxHash,
		},
	}, nil
}

// CheckRefund checks in processing that the refund can be sent: the deposit is valid and
// the deposit address holds enough native coin to pay the network fee. Nothing is created.
func (s *service) CheckRefund(ctx context.Context, ref *models.RefundRequest) (*processing.CheckRefundResult, error) {
	data, err := s.prepareRefund(ctx, ref)
	if err != nil {
		return nil, err
	}

	result, err := s.processing.CheckRefund(ctx, data.params)
	if err != nil {
		return nil, fmt.Errorf("check refund in processing: %w", err)
	}

	return &result, nil
}

// CreateRefundTransfer creates the transfer of kind refund for the refund request. It is committed
// before the refund is sent to processing (SendRefund): the transfer id is the processing request id,
// so status webhooks and the reconciliation always find the request and its transfer.
func (s *service) CreateRefundTransfer(ctx context.Context, ref *models.RefundRequest, opts ...repos.Option) (*models.Transfer, error) {
	data, err := s.prepareRefund(ctx, ref, opts...)
	if err != nil {
		return nil, err
	}

	transfer, err := s.storage.Transfers(opts...).Create(ctx, repo_transfers.CreateParams{
		ID:            uuid.New(),
		UserID:        data.user.ID,
		Kind:          models.TransferKindRefund,
		CurrencyID:    data.tx.CurrencyID,
		Status:        models.TransferStatusNew,
		Stage:         models.ResolveTransferStageByStatus(models.TransferStatusNew),
		FromAddresses: data.params.FromAddress,
		ToAddresses:   data.params.ToAddress,
		Amount:        data.tx.Amount,
		AmountUsd:     data.amountUsd,
		Blockchain:    data.tx.Blockchain,
		// the request may be detached from a failed transfer, the transfer keeps the link
		RefundRequestID: uuid.NullUUID{UUID: ref.ID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create refund transfer: %w", err)
	}

	return transfer, nil
}

// SendRefund creates the refund transfer in processing with the transfer id as the request id.
// The whole deposit amount is refunded, processing subtracts the network fee for native coins.
// It must not be called inside a database transaction: the result of the call is ambiguous on
// network errors and is resolved by the caller with processing.GetTransferByRequestID.
func (s *service) SendRefund(ctx context.Context, ref *models.RefundRequest, transferID uuid.UUID, totp *string) (*processing.FundsWithdrawalResult, error) {
	data, err := s.prepareRefund(ctx, ref)
	if err != nil {
		return nil, err
	}

	params := data.params
	params.RequestID = transferID
	params.TOTP = totp

	result, err := s.processing.FundsWithdrawal(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create refund in processing: %w", err)
	}

	return &result, nil
}
