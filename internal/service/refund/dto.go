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

const CabinetBucketAvailable = "available"

type CabinetItem struct {
	BlockedTransactionID uuid.UUID
	TransactionID        uuid.UUID
	TxHash               string
	Blockchain           models.Blockchain
	CurrencyID           string
	RiskLevel            string
	Score                decimal.Decimal
	CreatedAt            pgtype.Timestamp
	RefundStatus         *string
	DestinationAddress   *string
}

// buildCabinet groups a wallet's blocked transactions by refund status. txByID looks
// up the underlying transaction (for tx hash/blockchain/currency) by TransactionID —
// callers fetch it themselves since blocked_transactions doesn't carry those fields.
func buildCabinet(blocked []*models.BlockedTransaction, refunds []*models.RefundRequest, txByID map[uuid.UUID]*models.Transaction) map[string][]*CabinetItem {
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
