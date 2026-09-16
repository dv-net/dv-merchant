package refund

import (
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_blocked_transactions"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type CreateRefundDTO struct {
	WalletID             uuid.UUID `json:"wallet_id"`
	BlockedTransactionID uuid.UUID `json:"blocked_transaction_id"`
	DestinationAddress   string    `json:"destination_address"`
	Email                string    `json:"email"`
} //	@name	CreateRefundDTO

type RejectRefundDTO struct {
	RefundRequestID uuid.UUID
	UserID          uuid.UUID
}

const CabinetBucketAvailable = "available"

type CabinetItem struct {
	BlockedTransactionID uuid.UUID
	TransactionID        uuid.UUID
	TxHash               string
	Blockchain           models.Blockchain
	CurrencyID           string
	CurrencyCode         string
	Amount               decimal.Decimal
	AmountUsd            decimal.NullDecimal
	FromAddress          string
	ToAddress            string
	RiskLevel            string
	Score                decimal.Decimal
	CreatedAt            pgtype.Timestamp
	RefundStatus         *string
	DestinationAddress   *string
}

// buildCabinet groups a wallet's blocked transactions by refund status. blocked rows
// already carry their underlying transaction/currency details via the enriching JOIN
// in GetAllByWalletID.
func buildCabinet(
	blocked []*repo_blocked_transactions.GetAllWithTxByWalletIDRow,
	refunds []*models.RefundRequest,
) map[string][]*CabinetItem {
	refundByBlockedTx := make(map[uuid.UUID]*models.RefundRequest, len(refunds))
	for _, r := range refunds {
		refundByBlockedTx[r.BlockedTransactionID] = r
	}

	grouped := make(map[string][]*CabinetItem)
	for _, b := range blocked {
		item := &CabinetItem{
			BlockedTransactionID: b.ID,
			TransactionID:        b.TransactionID,
			TxHash:               b.TxHash,
			Blockchain:           b.Blockchain,
			CurrencyID:           b.CurrencyID,
			CurrencyCode:         b.CurrencyCode,
			Amount:               b.Amount,
			AmountUsd:            b.AmountUsd,
			FromAddress:          b.FromAddress,
			ToAddress:            b.ToAddress,
			RiskLevel:            b.RiskLevel,
			Score:                b.Score,
			CreatedAt:            b.CreatedAt,
		}

		bucket := CabinetBucketAvailable
		if r, ok := refundByBlockedTx[b.ID]; ok {
			item.RefundStatus = &r.Status
			item.DestinationAddress = &r.DestinationAddress
			bucket = r.Status
		}
		grouped[bucket] = append(grouped[bucket], item)
	}
	return grouped
}

type RequestWithTxDTO struct {
	models.RefundRequest
	TransactionID uuid.UUID           `json:"transaction_id"`
	TxHash        string              `json:"tx_hash"`
	Amount        decimal.Decimal     `json:"amount"`
	AmountUsd     decimal.NullDecimal `json:"amount_usd"`
	CurrencyID    string              `json:"currency_id"`
	CurrencyCode  string              `json:"currency_code"`
	Blockchain    models.Blockchain   `json:"blockchain"`
	FromAddress   string              `json:"from_address"`
	ToAddress     string              `json:"to_address"`
	RiskLevel     string              `json:"risk_level"`
	Score         decimal.Decimal     `json:"score"`
} //	@name	RefundRequestWithTx
