package invoice_response

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/constant"
	"github.com/dv-net/dv-merchant/internal/dto"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type InvoiceResponse struct {
	ID             uuid.UUID                 `json:"id"`
	AmountUSD      decimal.Decimal           `json:"amount_usd"`
	OrderID        string                    `json:"order_id"`
	ExpiredAt      time.Time                 `json:"expired_at"`
	Status         constant.InvoiceStatus    `json:"status"`
	InvoiceAddress []*InvoiceAddressResponse `json:"invoice_addresses"`
} // @name InvoiceResponse

type InvoiceAddressResponse struct {
	Address        string              `json:"address"`
	CurrencyID     string              `json:"currency_id"`
	Blockchain     models.Blockchain   `json:"blockchain"`
	RateAtCreation decimal.NullDecimal `json:"rate_at_creation"`
}

func NewFromModel(invoice *models.Invoice) InvoiceResponse {
	return InvoiceResponse{
		ID:        invoice.ID,
		AmountUSD: invoice.ExpectedAmountUsd,
		OrderID:   invoice.OrderID,
		Status:    invoice.Status,
		ExpiredAt: invoice.ExpiresAt.Time,
	}
}

func NewFromInvoiceInfoDTO(d *dto.InvoiceInfoDTO) InvoiceResponse {
	return InvoiceResponse{
		ID:             d.Invoice.ID,
		AmountUSD:      d.Invoice.ExpectedAmountUsd,
		OrderID:        d.Invoice.OrderID,
		Status:         d.Invoice.Status,
		ExpiredAt:      d.Invoice.ExpiresAt.Time,
		InvoiceAddress: NewFromInvoiceAddressDTO(d.InvoiceAddress),
	}
}

func NewFromInvoiceAddressDTO(d []*dto.InvoiceAddressDTO) []*InvoiceAddressResponse {
	resp := make([]*InvoiceAddressResponse, 0, len(d))
	for _, v := range d {
		resp = append(resp, &InvoiceAddressResponse{
			Address:        v.Address,
			CurrencyID:     v.CurrencyID,
			Blockchain:     v.Blockchain,
			RateAtCreation: v.RateAtCreation,
		})
	}
	return resp
}
