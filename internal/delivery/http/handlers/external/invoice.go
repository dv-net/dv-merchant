package external

import (
	"errors"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/invoice_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/invoice_response"
	"github.com/dv-net/dv-merchant/internal/dto"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/response"
	"github.com/gofiber/fiber/v3"
)

// CreateInvoice is a function to Create invoice
//
//	@Summary		Create invoice
//	@Description	Create invoice
//	@Tags			Invoice
//	@Accept			json
//	@Produce		json
//	@Param			request	body		invoice_request.CreateRequest						true	"Create invoice request"
//	@Success		200		{object}	response.Result[invoice_response.InvoiceResponse]	"Invoice created"
//	@Failure		401		{object}	apierror.Errors										"Unauthorized"
//	@Failure		404		{object}	apierror.Errors										"Not found"
//	@Failure		409		{object}	apierror.Errors										"Conflict"
//	@Failure		422		{object}	apierror.Errors										"Validation error"
//	@Failure		423		{object}	apierror.Errors										"Locked"
//	@Failure		500		{object}	apierror.Errors										"Internal server error"
//	@Router			/v1/external/invoice [post]
//	@Security		XApiKey
func (h *Handler) CreateInvoice(c fiber.Ctx) error {
	store, err := loadAuthStore(c)
	if err != nil {
		return err
	}

	usr, err := h.services.UserService.GetUserByID(c.Context(), store.UserID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	request := &invoice_request.CreateRequest{}
	if err := c.Bind().Body(request); err != nil {
		return err
	}

	if store.MinimalPayment.GreaterThan(request.AmountUSD) {
		return apierror.New().AddError(errors.New("amount_usd is less than minimal payment")).SetHttpCode(fiber.StatusBadRequest)
	}

	invoice, err := h.services.InvoiceService.CreateInvoice(c.Context(), dto.CreateInvoiceDTO{
		AmountUSD: request.AmountUSD,
		OrderID:   request.OrderID,
		Store:     store,
		User:      usr,
	})
	if err != nil {
		return h.handleError(err, "invoice")
	}

	return c.JSON(response.OkByData(invoice_response.NewFromModel(invoice)))
}

func (h *Handler) initInvoiceRoutes(v1 fiber.Router) {
	i := v1.Group("/invoice")
	i.Post("/", h.CreateInvoice)
}
