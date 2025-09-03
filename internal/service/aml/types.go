package aml

import (
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/pkg/aml"
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
