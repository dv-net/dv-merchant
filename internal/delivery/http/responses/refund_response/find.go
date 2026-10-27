package refund_response

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_refund_requests"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// RefundRequestDetailedResponse is a refund request with the blocked deposit, the AML check and the refund transfer
type RefundRequestDetailedResponse struct {
	ID                 uuid.UUID  `json:"id" format:"uuid"`
	WalletID           uuid.UUID  `json:"wallet_id" format:"uuid"`
	StoreID            uuid.UUID  `json:"store_id" format:"uuid"`
	DestinationAddress string     `json:"destination_address"`
	Status             string     `json:"status"`
	Email              string     `json:"email"`
	ReviewedAt         *time.Time `json:"reviewed_at" format:"date-time"`
	ExecuteAfter       *time.Time `json:"execute_after" format:"date-time"`
	LastError          *string    `json:"last_error"`
	CreatedAt          time.Time  `json:"created_at" format:"date-time"`
	UpdatedAt          *time.Time `json:"updated_at" format:"date-time"`

	BlockedTransaction RefundBlockedTransactionResponse `json:"blocked_transaction"`
	AmlCheck           RefundAmlCheckResponse           `json:"aml_check"`
	// Transfer is null until the refund is approved and sent to processing
	Transfer *RefundTransferResponse `json:"transfer"`
} //	@name	RefundRequestDetailedResponse

// RefundBlockedTransactionResponse is the deposit blocked by the AML check
type RefundBlockedTransactionResponse struct {
	ID            uuid.UUID           `json:"id" format:"uuid"`
	TransactionID uuid.UUID           `json:"transaction_id" format:"uuid"`
	TxHash        string              `json:"tx_hash"`
	Amount        decimal.Decimal     `json:"amount"`
	AmountUsd     decimal.NullDecimal `json:"amount_usd"`
	CurrencyID    string              `json:"currency_id"`
	CurrencyCode  string              `json:"currency_code"`
	Blockchain    string              `json:"blockchain"`
	FromAddress   string              `json:"from_address"`
	ToAddress     string              `json:"to_address"`
	// RiskLevel and Score are the values the deposit was blocked with
	RiskLevel   string          `json:"risk_level"`
	Score       decimal.Decimal `json:"score"`
	TxCreatedAt *time.Time      `json:"tx_created_at" format:"date-time"`
	BlockedAt   *time.Time      `json:"blocked_at" format:"date-time"`
} //	@name	RefundBlockedTransactionResponse

type RefundAmlCheckResponse struct {
	ID         uuid.UUID             `json:"id" format:"uuid"`
	Service    models.AMLSlug        `json:"service"`
	ExternalID string                `json:"external_id"`
	Status     models.AMLCheckStatus `json:"status"`
	Score      decimal.Decimal       `json:"score"`
	RiskLevel  *models.AmlRiskLevel  `json:"risk_level"`
	CreatedAt  *time.Time            `json:"created_at" format:"date-time"`
	UpdatedAt  *time.Time            `json:"updated_at" format:"date-time"`
} //	@name	RefundAmlCheckResponse

type RefundTransferResponse struct {
	ID        uuid.UUID              `json:"id" format:"uuid"`
	Number    *int64                 `json:"number"`
	Status    *models.TransferStatus `json:"status"`
	Stage     *models.TransferStage  `json:"stage"`
	Step      *string                `json:"step"`
	TxHash    *string                `json:"tx_hash"`
	Amount    decimal.NullDecimal    `json:"amount"`
	AmountUsd decimal.NullDecimal    `json:"amount_usd"`
	Message   *string                `json:"message"`
	CreatedAt *time.Time             `json:"created_at" format:"date-time"`
	UpdatedAt *time.Time             `json:"updated_at" format:"date-time"`
} //	@name	RefundTransferResponse

func NewRefundRequestDetailedResponse(row *repo_refund_requests.FindRow) *RefundRequestDetailedResponse {
	res := &RefundRequestDetailedResponse{
		ID:                 row.ID,
		WalletID:           row.WalletID,
		StoreID:            row.StoreID,
		DestinationAddress: row.DestinationAddress,
		Status:             row.Status,
		Email:              row.Email,
		ReviewedAt:         pgtypeutils.DecodeTime(row.ReviewedAt),
		ExecuteAfter:       pgtypeutils.DecodeTime(row.ExecuteAfter),
		LastError:          pgtypeutils.DecodeText(row.LastError),
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          pgtypeutils.DecodeTime(row.UpdatedAt),
		BlockedTransaction: RefundBlockedTransactionResponse{
			ID:            row.BlockedTransactionID,
			TransactionID: row.TransactionID,
			TxHash:        row.TxHash,
			Amount:        row.TxAmount,
			AmountUsd:     row.TxAmountUsd,
			CurrencyID:    row.TxCurrencyID,
			CurrencyCode:  row.TxCurrencyCode,
			Blockchain:    row.TxBlockchain.String(),
			FromAddress:   row.TxFromAddress,
			ToAddress:     row.TxToAddress,
			RiskLevel:     row.BlockRiskLevel,
			Score:         row.BlockScore,
			TxCreatedAt:   pgtypeutils.DecodeTime(row.TxCreatedAt),
			BlockedAt:     pgtypeutils.DecodeTime(row.BlockCreatedAt),
		},
		AmlCheck: RefundAmlCheckResponse{
			ID:         row.AmlCheckID,
			Service:    row.AmlServiceSlug,
			ExternalID: row.AmlCheckExternalID,
			Status:     row.AmlCheckStatus,
			Score:      row.AmlCheckScore,
			RiskLevel:  row.AmlCheckRiskLevel,
			CreatedAt:  pgtypeutils.DecodeTime(row.AmlCheckCreatedAt),
			UpdatedAt:  pgtypeutils.DecodeTime(row.AmlCheckUpdatedAt),
		},
	}

	if row.TransferID.Valid {
		res.Transfer = &RefundTransferResponse{
			ID:        row.TransferID.UUID,
			Number:    row.TransferNumber,
			Status:    row.TransferStatus,
			Stage:     row.TransferStage,
			Step:      row.TransferStep,
			TxHash:    row.TransferTxHash,
			Amount:    row.TransferAmount,
			AmountUsd: row.TransferAmountUsd,
			Message:   row.TransferMessage,
			CreatedAt: pgtypeutils.DecodeTime(row.TransferCreatedAt),
			UpdatedAt: pgtypeutils.DecodeTime(row.TransferUpdatedAt),
		}
	}

	return res
}

func NewRefundRequestsDetailedResponse(
	res *storecmn.FindResponseWithFullPagination[*repo_refund_requests.FindRow],
) *storecmn.FindResponseWithFullPagination[*RefundRequestDetailedResponse] {
	items := make([]*RefundRequestDetailedResponse, 0, len(res.Items))
	for _, row := range res.Items {
		items = append(items, NewRefundRequestDetailedResponse(row))
	}
	return &storecmn.FindResponseWithFullPagination[*RefundRequestDetailedResponse]{
		Items:      items,
		Pagination: res.Pagination,
	}
}
