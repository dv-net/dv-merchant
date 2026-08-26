package handlers

import (
	"errors"

	"github.com/dv-net/dv-merchant/internal/service/refund"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"

	//swag go import
	_ "github.com/dv-net/dv-merchant/internal/models"
)

// loadStorePendingRefunds lists refund requests awaiting the merchant's decision
// (status pending_review) for one of their own stores.
//
//	@Summary		List pending refund requests for a store
//	@Description	Lists refund requests awaiting review for a store owned by the authenticated user
//	@Tags			Store,Refund
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Store ID"
//	@Success		200	{object}	response.Result[[]models.RefundRequest]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/refund-requests [get]
//	@Security		BearerAuth
func (h *Handler) loadStorePendingRefunds(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	storeID, err := tools.ValidateUUID(c.Params("id"))
	if err != nil {
		return err
	}

	targetStore, err := h.services.StoreService.GetStoreByID(c.Context(), storeID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
	}
	if usr.ID != targetStore.UserID {
		return apierror.New().AddError(errors.New("this is not your store")).SetHttpCode(fiber.StatusUnauthorized)
	}

	list, err := h.services.RefundService.GetPendingReview(c.Context(), targetStore.ID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(list))
}

// rejectStoreRefund declines a pending refund request for one of the caller's own stores.
// It never moves money — only RefundService.RejectRefund's state-machine check
// (must currently be pending_review) and a plain status update.
//
//	@Summary		Reject a refund request
//	@Description	Rejects a pending refund request for a store owned by the authenticated user
//	@Tags			Store,Refund
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string	true	"Store ID"
//	@Param			refundId	path		string	true	"Refund request ID"
//	@Success		200			{object}	response.Result[models.RefundRequest]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		401			{object}	apierror.Errors
//	@Failure		404			{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/refund-requests/{refundId}/reject [post]
//	@Security		BearerAuth
func (h *Handler) rejectStoreRefund(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	storeID, err := tools.ValidateUUID(c.Params("id"))
	if err != nil {
		return err
	}

	targetStore, err := h.services.StoreService.GetStoreByID(c.Context(), storeID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
	}
	if usr.ID != targetStore.UserID {
		return apierror.New().AddError(errors.New("this is not your store")).SetHttpCode(fiber.StatusUnauthorized)
	}

	refundID, err := tools.ValidateUUID(c.Params("refundId"))
	if err != nil {
		return err
	}

	ref, err := h.services.RefundService.RejectRefund(c.Context(), refund.RejectRefundDTO{
		RefundRequestID: refundID,
		StoreID:         targetStore.ID,
	})
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(ref))
}

func (h *Handler) initRefundAdminRoutes(v1 fiber.Router) {
	storeHandlers := v1.Group("/store")
	storeHandlers.Get("/:id/refund-requests", h.loadStorePendingRefunds)
	storeHandlers.Post("/:id/refund-requests/:refundId/reject", h.rejectStoreRefund)
}
