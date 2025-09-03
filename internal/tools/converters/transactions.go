package converters

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/transaction_response"
	"github.com/dv-net/dv-merchant/internal/service/transactions"
	"github.com/dv-net/dv-merchant/internal/service/webhook"
)

func NewTransactionInfoResponseFromDto(dto transactions.TransactionInfoDto) transaction_response.TransactionInfoResponse {
	return transaction_response.TransactionInfoResponse{
		ID:               dto.ID,
		IsConfirmed:      dto.IsConfirmed,
		UserID:           dto.UserID,
		StoreID:          dto.StoreID,
		ReceiptID:        dto.ReceiptID,
		Wallet:           NewWalletInfoResponseFromDto(dto.Wallet),
		CurrencyID:       dto.CurrencyID,
		Blockchain:       dto.Blockchain,
		TxHash:           dto.TxHash,
		BcUniqKey:        dto.BcUniqKey,
		Type:             dto.Type,
		FromAddress:      dto.FromAddress,
		ToAddress:        dto.ToAddress,
		Amount:           dto.Amount,
		AmountUsd:        dto.AmountUsd,
		Fee:              dto.Fee,
		NetworkCreatedAt: dto.NetworkCreatedAt,
		WebhookHistory:   NewSendHistoryListResponseFromDto(dto.WebhookHistory),
		CreatedAt:        dto.CreatedAt,
		UpdatedAt:        dto.UpdatedAt,
	}
}

func NewSendHistoryListResponseFromDto(dtos []transactions.TransactionWhHistoryDto) []transaction_response.WhSendHistoryResponse {
	res := make([]transaction_response.WhSendHistoryResponse, 0, len(dtos))
	for _, dto := range dtos {
		res = append(res, NewWhSendHistoryResponseFromDto(dto))
	}

	return res
}

func NewWhSendHistoryResponseFromDto(dto transactions.TransactionWhHistoryDto) transaction_response.WhSendHistoryResponse {
	return transaction_response.WhSendHistoryResponse{
		ID:         dto.ID,
		StoreID:    dto.StoreID,
		WhType:     dto.WhType,
		URL:        dto.URL,
		IsSuccess:  dto.WhStatus == webhook.WebhookSendStatusSuccess,
		Request:    string(dto.Request),
		Response:   dto.Response,
		StatusCode: dto.ResponseStatusCode,
		CreatedAt:  dto.CreatedAt,
	}
}

func NewWalletInfoResponseFromDto(dto transactions.TransactionWalletInfoDto) transaction_response.WalletInfoResponse {
	return transaction_response.WalletInfoResponse{
		ID:              dto.ID,
		WalletStoreID:   dto.WalletStoreID,
		StoreExternalID: dto.StoreExternalID,
		WalletCreatedAt: dto.WalletCreatedAt,
		WalletUpdatedAt: dto.WalletUpdatedAt,
	}
}

func NewShortTransactionsListInfoFromDto(dtos []transactions.ShortTransactionInfo) transaction_response.ShortTransactionInfoListResponse {
	resConfirmed := make([]transaction_response.ShortTransactionResponse, 0, len(dtos))
	resUnconfirmed := make([]transaction_response.ShortTransactionResponse, 0, len(dtos))
	for _, dto := range dtos {
		if dto.IsConfirmed {
			resConfirmed = append(resConfirmed, transaction_response.ShortTransactionResponse{
				CurrencyCode: dto.CurrencyCode,
				Hash:         dto.Hash,
				Amount:       dto.Amount,
				AmountUSD:    dto.AmountUSD,
				Type:         dto.Type,
				CreatedAt:    dto.CreatedAt,
			})

			continue
		}

		resUnconfirmed = append(resUnconfirmed, transaction_response.ShortTransactionResponse{
			CurrencyCode: dto.CurrencyCode,
			Hash:         dto.Hash,
			Amount:       dto.Amount,
			AmountUSD:    dto.AmountUSD,
			Type:         dto.Type,
			CreatedAt:    dto.CreatedAt,
		})
	}

	return transaction_response.ShortTransactionInfoListResponse{
		Confirmed:   resConfirmed,
		Unconfirmed: resUnconfirmed,
	}
}
