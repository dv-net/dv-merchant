package aml

import (
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/pkg/aml"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Direction string

const (
	DirectionIn  Direction = "in"
	DirectionOut Direction = "out"
)

func (d Direction) ToAMLDirection() aml.Direction {
	switch d {
	case DirectionOut:
		return aml.DirectionOut
	default:
		return aml.DirectionIn
	}
}

type CheckDTO struct {
	TxID          string
	CurrencyID    string
	ProviderSlug  models.AMLSlug
	Direction     Direction
	OutputAddress string
}

type AutoScoreDepositDTO struct {
	TxID          uuid.UUID
	UserID        uuid.UUID
	TxHash        string
	CurrencyID    string
	ProviderSlug  *models.AMLSlug // nil = auto-select by user keys
	OutputAddress string
	DBTx          pgx.Tx // outer DB transaction from the deposit event; nil for manual checks
}
