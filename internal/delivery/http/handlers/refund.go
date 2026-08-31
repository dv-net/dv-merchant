package handlers

import (
	"github.com/dv-net/dv-merchant/internal/service/refund"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"

	// swag go import
	_ "github.com/dv-net/dv-merchant/internal/models"
)

// loadUserPendingRefunds lists refund requests awaiting the merchant's decision
// (status pending_review) across all stores owned by the authenticated user.
//
//	@Summary		List pending refund requests
//	@Description	Lists refund requests awaiting review across all stores owned by the authenticated user
//	@Tags			Store,Refund
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[[]models.RefundRequest]
//	@Failure		401	{object}	apierror.Errors
//	@Router			/v1/dv-admin/refund-requests [get]
//	@Security		BearerAuth
func (h *Handler) loadUserPendingRefunds(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	list, err := h.services.RefundService.GetPendingReviewByUser(c.Context(), usr.ID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(list))
}

// rejectRefund declines a pending refund request that belongs to one of the caller's
// own stores. It never moves money — only RefundService.RejectRefund's ownership and
// state-machine check (must currently be pending_review) and a plain status update.
//
//	@Summary		Reject a refund request
//	@Description	Rejects a pending refund request that belongs to a store owned by the authenticated user
//	@Tags			Store,Refund
//	@Accept			json
//	@Produce		json
//	@Param			refundId	path		string	true	"Refund request ID"
//	@Success		200			{object}	response.Result[models.RefundRequest]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		401			{object}	apierror.Errors
//	@Router			/v1/dv-admin/refund-requests/{refundId}/reject [post]
//	@Security		BearerAuth
func (h *Handler) rejectRefund(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	refundID, err := tools.ValidateUUID(c.Params("refundId"))
	if err != nil {
		return err
	}

	ref, err := h.services.RefundService.RejectRefund(c.Context(), refund.RejectRefundDTO{
		RefundRequestID: refundID,
		UserID:          usr.ID,
	})
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(ref))
}

func (h *Handler) initRefundAdminRoutes(v1 fiber.Router) {
	storeHandlers := v1.Group("/refund-requests")
	storeHandlers.Get("/", h.loadUserPendingRefunds)
	storeHandlers.Post("/:refundId/reject", h.rejectRefund)
}
