package withdraw

import (
	"context"

	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/transfer_requests"
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/withdrawal_wallets_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/transfer_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transfers"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
)

type ITransferService interface {
	GetTransfers(ctx context.Context, userID uuid.UUID, req *withdrawal_wallets_request.TransferRequest) (*storecmn.FindResponseWithFullPagination[*transfer_response.GetTransferResponse], error)
	GetTransfersHistory(ctx context.Context, userID uuid.UUID, req *transfer_requests.TransferHistoryRequest) (*storecmn.FindResponseWithFullPagination[*transfer_response.GetTransferResponse], error)
	DeleteTransfers(ctx context.Context, user *models.User, req *transfer_requests.DeleteTransferRequest) error
}

func (s *service) GetTransfers(ctx context.Context, userID uuid.UUID, request *withdrawal_wallets_request.TransferRequest) (*storecmn.FindResponseWithFullPagination[*transfer_response.GetTransferResponse], error) {
	commonParams := storecmn.NewCommonFindParams()

	if request.PageSize != nil {
		commonParams.SetPageSize(request.PageSize)
	}
	if request.Page != nil {
		commonParams.SetPage(request.Page)
	}

	params := repo_transfers.GetTransfersByUserAndStatusParams{
		UserID:           userID,
		Stages:           request.Stages,
		Kinds:            request.Kinds,
		DateFrom:         request.DateFrom,
		CommonFindParams: *commonParams,
	}
	data, err := s.storage.Transfers().GetTransfersByUserAndStatus(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get prefetch wallet address: %w", err)
	}
	dto := make([]*transfer_response.GetTransferResponse, 0, len(data.Items))
	for _, transfer := range data.Items {
		dto = append(dto, &transfer_response.GetTransferResponse{
			ID:            transfer.ID.String(),
			Number:        transfer.Number,
			Stage:         transfer.Stage,
			UserID:        transfer.UserID.String(),
			Kind:          transfer.Kind,
			CurrencyID:    transfer.CurrencyID,
			Status:        transfer.Status,
			Step:          transfer.Step,
			FromAddresses: transfer.FromAddresses,
			ToAddresses:   transfer.ToAddresses,
			TxHash:        transfer.TxHash,
			Amount:        transfer.Amount,
			AmountUsd:     transfer.AmountUsd,
			Message:       transfer.Message,
			CreatedAt:     &transfer.CreatedAt.Time,
			UpdatedAt:     &transfer.UpdatedAt.Time,
		})
	}
	return &storecmn.FindResponseWithFullPagination[*transfer_response.GetTransferResponse]{
		Items:      dto,
		Pagination: data.Pagination,
	}, nil
}

func (s *service) GetTransfersHistory(
	ctx context.Context,
	userID uuid.UUID,
	request *transfer_requests.TransferHistoryRequest,
) (*storecmn.FindResponseWithFullPagination[*transfer_response.GetTransferResponse], error) {
	commonParams := storecmn.NewCommonFindParams()

	if request.PageSize != nil {
		commonParams.SetPageSize(request.PageSize)
	}
	if request.Page != nil {
		commonParams.SetPage(request.Page)
	}

	params := repo_transfers.GetTransferHistoryByAddressParams{
		UserID:           userID,
		Address:          request.Address,
		Stages:           []models.TransferStage{models.TransferStageCompleted, models.TransferStageInProgress},
		Kinds:            []models.TransferKind{models.TransferKindFromAddress, models.TransferKindFromProcessing},
		CommonFindParams: *commonParams,
	}
	data, err := s.storage.Transfers().GetTransfersHistoryByAddress(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get prefetch wallet address: %w", err)
	}
	dto := make([]*transfer_response.GetTransferResponse, 0, len(data.Items))
	for _, transfer := range data.Items {
		dto = append(dto, &transfer_response.GetTransferResponse{
			ID:            transfer.ID.String(),
			Number:        transfer.Number,
			Stage:         transfer.Stage,
			UserID:        transfer.UserID.String(),
			Kind:          transfer.Kind,
			CurrencyID:    transfer.CurrencyID,
			Status:        transfer.Status,
			Step:          transfer.Step,
			FromAddresses: transfer.FromAddresses,
			ToAddresses:   transfer.ToAddresses,
			TxHash:        transfer.TxHash,
			Amount:        transfer.Amount,
			AmountUsd:     transfer.AmountUsd,
			Message:       transfer.Message,
			CreatedAt:     &transfer.CreatedAt.Time,
			UpdatedAt:     &transfer.UpdatedAt.Time,
		})
	}
	return &storecmn.FindResponseWithFullPagination[*transfer_response.GetTransferResponse]{
		Items:      dto,
		Pagination: data.Pagination,
	}, nil
}

func (s *service) DeleteTransfers(ctx context.Context, user *models.User, req *transfer_requests.DeleteTransferRequest) error {
	params := make([]repo_transfers.BatchDeleteTransfersParams, 0, len(req.ID))
	for _, id := range req.ID {
		params = append(params, repo_transfers.BatchDeleteTransfersParams{
			ID:     id,
			UserID: user.ID,
		})
	}

	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		var err error
		batch := s.storage.Transfers(repos.WithTx(tx)).BatchDeleteTransfers(ctx, params)
		defer func() {
			if err := batch.Close(); err != nil {
				s.logger.Error("batch delete transfers close error", err)
			}
		}()

		batch.Exec(func(_ int, batchErr error) {
			if batchErr != nil {
				err = fmt.Errorf("batch delete transfers: %w", err)
				return
			}
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	return nil
}
