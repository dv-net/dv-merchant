package handlers

import (
	"errors"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/refund_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/refund_response"
	"github.com/dv-net/dv-merchant/internal/service/refund"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// loadUserPendingRefunds lists refund requests awaiting the merchant's decision
// (pending_review: approve / reject, scheduled: cancel) across all stores owned by the authenticated user.
//
//	@Summary		List pending refund requests
//	@Description	Lists refund requests awaiting review or scheduled for sending across all stores owned by the authenticated user, including blocked deposit details
//	@Tags			Store,Refund
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[[]refund_response.RefundRequestResponse]
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

	return c.JSON(response.OkByData(refund_response.NewRefundRequestResponses(list)))
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
//	@Success		200			{object}	response.Result[refund_response.RefundRequestResponse]
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
		return prepareRefundHTTPError(err)
	}

	return c.JSON(response.OkByData(refund_response.NewRefundRequestResponse(ref)))
}

// approveRefund sends a pending refund request to processing. The owner's TOTP is required:
// processing verifies it and sends the refund to the destination address immediately.
//
//	@Summary		Approve a refund request
//	@Description	Approves a pending refund request that belongs to a store owned by the authenticated user and sends the refund, confirmed by TOTP
//	@Tags			Store,Refund
//	@Accept			json
//	@Produce		json
//	@Param			refundId	path		string								true	"Refund request ID"
//	@Param			register	body		refund_request.ApproveRefundRequest	true	"TOTP confirmation"
//	@Success		200			{object}	response.Result[refund_response.RefundRequestResponse]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		401			{object}	apierror.Errors
//	@Failure		403			{object}	apierror.Errors
//	@Failure		409			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Router			/v1/dv-admin/refund-requests/{refundId}/approve [post]
//	@Security		BearerAuth
func (h *Handler) approveRefund(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	refundID, err := tools.ValidateUUID(c.Params("refundId"))
	if err != nil {
		return err
	}

	req := &refund_request.ApproveRefundRequest{}
	if err = c.Bind().Body(req); err != nil {
		return err
	}

	if err = req.Validate(); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	ref, err := h.services.RefundService.ApproveRefund(c.Context(), refund.ApproveRefundDTO{
		RefundRequestID: refundID,
		UserID:          usr.ID,
		TOTP:            req.TOTP,
	})
	if err != nil {
		return prepareRefundHTTPError(err)
	}

	return c.JSON(response.OkByData(refund_response.NewRefundRequestResponse(ref)))
}

// cancelRefund cancels a scheduled refund to the deposit sender before processing sends it.
// The request goes back to pending review.
//
//	@Summary		Cancel a scheduled refund
//	@Description	Cancels a scheduled refund to the deposit sender before it is sent, the request goes back to pending review
//	@Tags			Store,Refund
//	@Accept			json
//	@Produce		json
//	@Param			refundId	path		string	true	"Refund request ID"
//	@Success		200			{object}	response.Result[refund_response.RefundRequestResponse]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		401			{object}	apierror.Errors
//	@Failure		403			{object}	apierror.Errors
//	@Failure		409			{object}	apierror.Errors
//	@Router			/v1/dv-admin/refund-requests/{refundId}/cancel [post]
//	@Security		BearerAuth
func (h *Handler) cancelRefund(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	refundID, err := tools.ValidateUUID(c.Params("refundId"))
	if err != nil {
		return err
	}

	ref, err := h.services.RefundService.CancelRefund(c.Context(), refund.CancelRefundDTO{
		RefundRequestID: refundID,
		UserID:          usr.ID,
	})
	if err != nil {
		return prepareRefundHTTPError(err)
	}

	return c.JSON(response.OkByData(refund_response.NewRefundRequestResponse(ref)))
}

// executeRefundNow sends a scheduled refund to the deposit sender immediately instead of
// waiting for the cooldown. The owner's TOTP is required.
//
//	@Summary		Execute a scheduled refund now
//	@Description	Sends a scheduled refund to the deposit sender immediately without waiting for the cooldown, confirmed by TOTP
//	@Tags			Store,Refund
//	@Accept			json
//	@Produce		json
//	@Param			refundId	path		string								true	"Refund request ID"
//	@Param			register	body		refund_request.ApproveRefundRequest	true	"TOTP confirmation"
//	@Success		200			{object}	response.Result[refund_response.RefundRequestResponse]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		401			{object}	apierror.Errors
//	@Failure		403			{object}	apierror.Errors
//	@Failure		409			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Router			/v1/dv-admin/refund-requests/{refundId}/execute-now [post]
//	@Security		BearerAuth
func (h *Handler) executeRefundNow(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	refundID, err := tools.ValidateUUID(c.Params("refundId"))
	if err != nil {
		return err
	}

	req := &refund_request.ApproveRefundRequest{}
	if err = c.Bind().Body(req); err != nil {
		return err
	}

	if err = req.Validate(); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	ref, err := h.services.RefundService.ExecuteRefundNow(c.Context(), refund.ExecuteRefundNowDTO{
		RefundRequestID: refundID,
		UserID:          usr.ID,
		TOTP:            req.TOTP,
	})
	if err != nil {
		return prepareRefundHTTPError(err)
	}

	return c.JSON(response.OkByData(refund_response.NewRefundRequestResponse(ref)))
}

func prepareRefundHTTPError(err error) error {
	errCode := fiber.StatusBadRequest
	switch {
	case errors.Is(err, refund.ErrRefundNotOwned):
		errCode = fiber.StatusForbidden
	case errors.Is(err, refund.ErrRefundNotPendingReview),
		errors.Is(err, refund.ErrRefundAlreadyExists),
		errors.Is(err, refund.ErrRefundNotCancelable),
		errors.Is(err, refund.ErrRefundNotScheduled):
		errCode = fiber.StatusConflict
	case errors.Is(err, refund.ErrRefundInsufficientFee),
		errors.Is(err, refund.ErrRefundInsufficientFunds),
		errors.Is(err, refund.ErrRefundSourceInvalid),
		errors.Is(err, refund.ErrRefundTwoFactorRequired):
		errCode = fiber.StatusUnprocessableEntity
	}

	return apierror.New().AddError(err).SetHttpCode(errCode)
}

func (h *Handler) initRefundAdminRoutes(v1 fiber.Router) {
	storeHandlers := v1.Group("/refund-requests")
	storeHandlers.Get("/", h.loadUserPendingRefunds)
	storeHandlers.Post("/:refundId/reject", h.rejectRefund)
	storeHandlers.Post("/:refundId/approve", h.approveRefund)
	storeHandlers.Post("/:refundId/cancel", h.cancelRefund)
	storeHandlers.Post("/:refundId/execute-now", h.executeRefundNow)
}
