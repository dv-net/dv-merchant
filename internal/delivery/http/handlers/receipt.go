package handlers

import (
	"strconv"

	// Blank import for swaggen
	_ "github.com/dv-net/dv-merchant/internal/delivery/http/responses/receipt_response"
	_ "github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// loadReceiptById is a function to get receipt by id
//
//	@Summary		Load receipt
//	@Description	Load receipt
//	@Tags			Receipt
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uuid	true	"Receipt ID"
//	@Success		200	{object}	response.Result[receipt_response.ReceiptResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Router			/v1/dv-admin/receipts/{id} [get]
//	@Security		BearerAuth
func (h Handler) loadReceiptByID(c fiber.Ctx) error {
	_, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	id, err := tools.ValidateUUID(c.Params("id"))
	if err != nil {
		return err
	}

	receipt, err := h.services.ReceiptService.GetByID(c.Context(), id)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
	}

	res := converters.FromReceiptModelToResponse(receipt)

	return c.JSON(response.OkByData(res))
}

// loadReceipts is a function to Load receipts
//
//	@Summary		Load receipts
//	@Description	Load receipts
//	@Tags			Receipt
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int	false	"page number"
//	@Success		200		{object}	response.Result[[]receipt_response.ReceiptResponse]
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/dv-admin/receipts/ [get]
//	@Security		BearerAuth
func (h Handler) loadReceipts(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	pageParam := c.Query("page", "1")
	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = 1
	}
	receipts, err := h.services.ReceiptService.GetByUserID(c.Context(), user.ID, int32(page)) // #nosec
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
	}
	res := converters.FromReceiptModelToResponses(receipts...)

	return c.JSON(response.OkByData(res))
}

func (h Handler) initReceiptRoutes(group fiber.Router) {
	receipt := group.Group("/receipts")
	receipt.Get("/", h.loadReceipts)
	receipt.Get("/:id", h.loadReceiptByID)
}
