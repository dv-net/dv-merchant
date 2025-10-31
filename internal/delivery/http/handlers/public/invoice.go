package public

import (
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/invoice_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/invoice_response"
	"github.com/dv-net/dv-merchant/internal/dto"
	"github.com/dv-net/dv-merchant/internal/service/store"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// getInvoice is a function to get invoice by id
//
//	@Summary		Get invoice
//	@Description	Get invoice by id
//	@Tags			Invoice
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Invoice id"
//	@Success		200	{object}	response.Result[invoice_response.InvoiceResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		410	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/public/invoice/{id} [get]
func (h *Handler) getInvoice(c fiber.Ctx) error {
	invoiceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(fmt.Errorf("bad invoice id")).SetHttpCode(fiber.StatusBadRequest)
	}

	iInfo, err := h.services.InvoiceService.GetInvoice(c.Context(), invoiceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.New().AddError(errors.New("invoice not found")).SetHttpCode(fiber.StatusNotFound)
		}
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	// get store
	_, err = h.services.StoreService.GetStoreByID(c.Context(), iInfo.Invoice.StoreID)
	if err != nil {
		if errors.Is(err, store.ErrStoreNotFound) {
			return apierror.New().AddError(errors.New("store not found")).SetHttpCode(fiber.StatusNotFound)
		}
		if errors.Is(err, store.ErrStoreDisabled) {
			return apierror.New().AddError(errors.New("store is disabled")).SetHttpCode(fiber.StatusGone)
		}
		return apierror.New().AddError(fmt.Errorf("something went wrong")).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByData(invoice_response.NewFromInvoiceInfoDTO(iInfo)))
}

// attachWallet is a function to attach wallet to invoice
//
//	@Summary		Attach wallet to invoice
//	@Description	Attach wallet to invoice
//	@Tags			Invoice
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string	true	"Invoice id"
//	@Param			currency_id	body		string	true	"Currency id"
//	@Success		200			{object}	response.Result[string]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		410			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/public/invoice/{id}/wallet [post]
func (h *Handler) attachWallet(c fiber.Ctx) error {
	invoiceID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(fmt.Errorf("bad invoice id")).SetHttpCode(fiber.StatusBadRequest)
	}
	req := &invoice_request.AttachWalletRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	iInfo, err := h.services.InvoiceService.GetInvoice(c.Context(), invoiceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apierror.New().AddError(errors.New("invoice not found")).SetHttpCode(fiber.StatusNotFound)
		}
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	// get store
	storeInfo, err := h.services.StoreService.GetStoreByID(c.Context(), iInfo.Invoice.StoreID)
	if err != nil {
		if errors.Is(err, store.ErrStoreNotFound) {
			return apierror.New().AddError(errors.New("store not found")).SetHttpCode(fiber.StatusNotFound)
		}
		if errors.Is(err, store.ErrStoreDisabled) {
			return apierror.New().AddError(errors.New("store is disabled")).SetHttpCode(fiber.StatusGone)
		}
		return apierror.New().AddError(fmt.Errorf("something went wrong")).SetHttpCode(fiber.StatusBadRequest)
	}

	curr, err := h.services.CurrencyService.GetCurrencyByID(c.Context(), req.CurrencyID)
	if err != nil {
		return apierror.New().AddError(errors.New("invalid currency")).SetHttpCode(fiber.StatusBadRequest)
	}
	address, err := h.services.InvoiceService.AttachAddress(c.Context(), iInfo.Invoice.ID, curr, storeInfo)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	iInfo.InvoiceAddress = []*dto.InvoiceAddressDTO{address}

	return c.JSON(response.OkByData(invoice_response.NewFromInvoiceInfoDTO(iInfo)))
}

func (h *Handler) initInvoiceRoutes(v1 fiber.Router) {
	group := v1.Group("/invoice")
	group.Get("/:id", h.getInvoice)
	group.Post("/:id/wallet", h.attachWallet)
}
