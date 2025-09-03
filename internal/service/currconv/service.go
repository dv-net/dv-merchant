package currconv

import (
	"context"
	"fmt"
	"strings"

	"github.com/dv-net/dv-merchant/internal/service/exrate"

	"github.com/shopspring/decimal"
)

type ICurrencyConvertor interface {
	Convert(ctx context.Context, dto ConvertDTO) (decimal.Decimal, error)
}

func New(exrateSrc exrate.IExRateSource) ICurrencyConvertor {
	return &service{exrateSrc: exrateSrc}
}

type service struct {
	exrateSrc exrate.IExRateSource
}

type ConvertDTO struct {
	Source     string
	From       string
	To         string
	Amount     string
	StableCoin bool
	Scale      *decimal.Decimal
}

var _ ICurrencyConvertor = (*service)(nil)

func (srv *service) Convert(ctx context.Context, dto ConvertDTO) (v decimal.Decimal, err error) {
	from := strings.ToUpper(dto.From)
	to := strings.ToUpper(dto.To)
	var amountDec decimal.Decimal
	if amountDec, err = decimal.NewFromString(dto.Amount); err != nil {
		err = fmt.Errorf("parse amount to decimal: %w", err)
		return v, err
	}

	if from == to || dto.StableCoin {
		return amountDec, nil
	}
	var rate string
	if rate, err = srv.exrateSrc.GetCurrencyRate(ctx, dto.Source, from, to); err != nil {
		err = fmt.Errorf("get currency rate: %w", err)
		return v, err
	}
	var rateDec decimal.Decimal
	if rateDec, err = decimal.NewFromString(rate); err != nil {
		err = fmt.Errorf("parse currency rate to decimal: %w", err)
		return v, err
	}

	if dto.Scale != nil && !dto.StableCoin {
		multiplier := decimal.NewFromInt(100).Add(*dto.Scale).Div(decimal.NewFromInt(100))
		rateDec = rateDec.Mul(multiplier)
	}

	result := amountDec.Mul(rateDec)
	return result, nil
}
