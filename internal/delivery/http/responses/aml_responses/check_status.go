package aml_responses

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/refund_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AmlCheckStatusResponse struct {
	TransactionID uuid.UUID             `json:"transaction_id"`
	TxHash        string                `json:"tx_hash"`
	Blockchain    models.Blockchain     `json:"blockchain"`
	Status        models.AMLCheckStatus `json:"status"`
	InProgress    bool                  `json:"in_progress"`
	Score         decimal.Decimal       `json:"score"`
	RiskLevel     *models.AmlRiskLevel  `json:"risk_level"`
	CreatedAt     *time.Time            `json:"created_at"`
	UpdatedAt     *time.Time            `json:"updated_at"`
	// BlockReasons are the AML rules which blocked the transaction: null when it is not blocked,
	// empty when it was blocked before the reasons were recorded
	BlockReasons []refund_response.CabinetBlockReasonResponse `json:"block_reasons"`
} //	@name	AmlCheckStatusResponse

// NewAmlCheckStatusResponse builds the check status, blockReasons are set only for a blocked transaction
func NewAmlCheckStatusResponse(tx *models.Transaction, check *models.AmlCheck, blockReasons models.AmlBlockReasons, blocked bool) *AmlCheckStatusResponse {
	res := &AmlCheckStatusResponse{
		TransactionID: tx.ID,
		TxHash:        tx.TxHash,
		Blockchain:    tx.Blockchain,
		Status:        check.Status,
		InProgress:    check.Status == models.AmlCheckStatusPending,
		Score:         check.Score,
		RiskLevel:     check.RiskLevel,
		CreatedAt:     pgtypeutils.DecodeTime(check.CreatedAt),
		UpdatedAt:     pgtypeutils.DecodeTime(check.UpdatedAt),
	}
	if blocked {
		res.BlockReasons = refund_response.NewCabinetBlockReasonsResponse(blockReasons)
	}
	return res
}
