package refund

import (
	"github.com/dv-net/dv-merchant/internal/models"
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

// RefundRequestDetails is a pending/reviewed refund request enriched with the
// underlying blocked deposit transaction data for admin UI.
type RefundRequestDetails struct {
	ID                   uuid.UUID
	BlockedTransactionID uuid.UUID
	WalletID             uuid.UUID
	StoreID              uuid.UUID
	TransferID           uuid.NullUUID
	DestinationAddress   string
	Status               string
	Email                string
	ReviewedAt           pgtype.Timestamp
	CreatedAt            pgtype.Timestamp
	UpdatedAt            pgtype.Timestamp

	TransactionID uuid.UUID
	TxHash        string
	Amount        decimal.Decimal
	AmountUsd     decimal.NullDecimal
	CurrencyID    string
	CurrencyCode  string
	Blockchain    models.Blockchain
	FromAddress   string
	ToAddress     string
	RiskLevel     string
	Score         decimal.Decimal
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

// buildCabinet groups a wallet's blocked transactions by refund status. txByID looks
// up the underlying transaction (for tx hash/blockchain/currency) by TransactionID —
// callers fetch it themselves since blocked_transactions doesn't carry those fields.
func buildCabinet(
	blocked []*models.BlockedTransaction,
	refunds []*models.RefundRequest,
	txByID map[uuid.UUID]*models.Transaction,
	currencyCodeByID map[string]string,
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
			RiskLevel:            b.RiskLevel,
			Score:                b.Score,
			CreatedAt:            b.CreatedAt,
		}
		if tx, ok := txByID[b.TransactionID]; ok {
			item.TxHash = tx.TxHash
			item.Blockchain = tx.Blockchain
			item.CurrencyID = tx.CurrencyID
			item.CurrencyCode = currencyCodeByID[tx.CurrencyID]
			item.Amount = tx.Amount
			item.AmountUsd = tx.AmountUsd
			item.FromAddress = tx.FromAddress
			item.ToAddress = tx.ToAddress
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
