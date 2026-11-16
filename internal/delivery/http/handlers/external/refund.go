package external

import (
	"errors"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/refund_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/refund_response"
	"github.com/dv-net/dv-merchant/internal/delivery/middleware"
	"github.com/dv-net/dv-merchant/internal/service/refund"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// estimateRefund returns the expected result of a refund of the store's blocked transaction. Nothing is created.
//
//	@Summary		Estimate a refund of a blocked transaction
//	@Description	Returns the estimated network fee and the amount to receive for a refund of a blocked transaction of the store. code_required is true when the address is not a sender of the deposit: send a confirmation code to the wallet email (send-code) and pass it on the refund creation. The fee is an estimate and may change by the time of sending
//	@Tags			Refund
//	@Accept			json
//	@Produce		json
//	@Param			api_key	query		string											false	"Store API key"
//	@Param			id		path		string											true	"Blocked transaction ID"
//	@Param			request	body		refund_request.ExternalRefundEstimateRequest	true	"Destination address"
//	@Success		200		{object}	response.Result[refund_response.ExternalRefundEstimateResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		409		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Router			/v1/external/blocked-transactions/{id}/refund/estimate [post]
//	@Security		XApiKey
func (h *Handler) estimateRefund(c fiber.Ctx) error {
	dto, err := h.parseStoreRefund(c, func(req *refund_request.ExternalRefundCreateRequest) error {
		body := &refund_request.ExternalRefundEstimateRequest{}
		if err := c.Bind().Body(body); err != nil {
			return err
		}
		req.DestinationAddress = body.DestinationAddress
		return nil
	})
	if err != nil {
		return err
	}

	estimate, err := h.services.RefundService.EstimateRefundByStore(c.Context(), dto)
	if err != nil {
		return h.prepareRefundError(err, "failed to estimate refund", dto)
	}

	return c.JSON(response.OkByData(refund_response.NewExternalRefundEstimateResponse(estimate)))
}

// sendRefundCode sends a one-time code confirming a refund to an address which is not a deposit sender.
// The code is sent to the wallet email (the payer) and is bound to the blocked transaction and the address.
//
//	@Summary		Send a refund confirmation code
//	@Description	Sends a one-time code to the wallet email. The code confirms a refund of the blocked transaction to the destination address which is not a sender of the deposit, pass it on the refund creation. The code can be sent again once a minute
//	@Tags			Refund
//	@Accept			json
//	@Produce		json
//	@Param			api_key	query		string											false	"Store API key"
//	@Param			id		path		string											true	"Blocked transaction ID"
//	@Param			request	body		refund_request.ExternalRefundSendCodeRequest	true	"Destination address"
//	@Success		200		{object}	response.Result[string]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		409		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		429		{object}	apierror.Errors
//	@Router			/v1/external/blocked-transactions/{id}/refund/send-code [post]
//	@Security		XApiKey
func (h *Handler) sendRefundCode(c fiber.Ctx) error {
	dto, err := h.parseStoreRefund(c, func(req *refund_request.ExternalRefundCreateRequest) error {
		body := &refund_request.ExternalRefundSendCodeRequest{}
		if err := c.Bind().Body(body); err != nil {
			return err
		}
		req.DestinationAddress = body.DestinationAddress
		return nil
	})
	if err != nil {
		return err
	}

	if err := h.services.RefundService.SendRefundCodeByStore(c.Context(), dto); err != nil {
		return h.prepareRefundError(err, "failed to send refund confirmation code", dto)
	}

	return c.JSON(response.OkByMessage("confirmation code has been sent to the wallet email"))
}

// createRefund creates a refund request of the store's blocked transaction.
//
//	@Summary		Create a refund of a blocked transaction
//	@Description	Creates a refund request of a blocked transaction of the store. A refund to a sender of the deposit is scheduled without a code and sent after the cooldown (status scheduled). A refund to another address requires the code sent to the wallet email (send-code) and is sent after the store owner approves it (status pending_review)
//	@Tags			Refund
//	@Accept			json
//	@Produce		json
//	@Param			api_key	query		string										false	"Store API key"
//	@Param			id		path		string										true	"Blocked transaction ID"
//	@Param			request	body		refund_request.ExternalRefundCreateRequest	true	"Destination address and confirmation code"
//	@Success		200		{object}	response.Result[refund_response.RefundRequestResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		401		{object}	apierror.Errors
//	@Failure		403		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		409		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Failure		429		{object}	apierror.Errors
//	@Router			/v1/external/blocked-transactions/{id}/refund [post]
//	@Security		XApiKey
func (h *Handler) createRefund(c fiber.Ctx) error {
	dto, err := h.parseStoreRefund(c, func(req *refund_request.ExternalRefundCreateRequest) error {
		return c.Bind().Body(req)
	})
	if err != nil {
		return err
	}

	ref, err := h.services.RefundService.CreateRefundByStore(c.Context(), dto)
	if err != nil {
		return h.prepareRefundError(err, "failed to create refund request", dto)
	}

	return c.JSON(response.OkByData(refund_response.NewRefundRequestResponse(ref)))
}

// getRefund returns a refund request of the store.
//
//	@Summary		Get a refund request
//	@Description	Returns a refund request of the store with its status
//	@Tags			Refund
//	@Produce		json
//	@Param			api_key	query		string	false	"Store API key"
//	@Param			id		path		string	true	"Refund request ID"
//	@Success		200		{object}	response.Result[refund_response.RefundRequestResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/external/refunds/{id} [get]
//	@Security		XApiKey
func (h *Handler) getRefund(c fiber.Ctx) error {
	store, err := loadAuthStore(c)
	if err != nil {
		return err
	}

	refundID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(errors.New("bad refund request id")).SetHttpCode(fiber.StatusBadRequest)
	}

	ref, err := h.services.RefundService.GetRefundByStore(c.Context(), store.ID, refundID)
	if err != nil {
		if errors.Is(err, refund.ErrRefundNotFound) {
			return apierror.New().AddError(refund.ErrRefundNotFound).SetHttpCode(fiber.StatusNotFound)
		}
		h.logger.Errorw("failed to get refund request", "store_id", store.ID, "refund_request_id", refundID, "error", err)
		return apierror.New().AddError(errors.New("failed to get refund request")).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(refund_response.NewRefundRequestResponse(ref)))
}

// parseStoreRefund reads the store, the blocked transaction id from the path and the body by bind
func (h *Handler) parseStoreRefund(c fiber.Ctx, bind func(req *refund_request.ExternalRefundCreateRequest) error) (refund.StoreRefundDTO, error) {
	store, err := loadAuthStore(c)
	if err != nil {
		return refund.StoreRefundDTO{}, err
	}

	blockedTransactionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return refund.StoreRefundDTO{}, apierror.New().AddError(errors.New("bad blocked transaction id")).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &refund_request.ExternalRefundCreateRequest{}
	if err := bind(req); err != nil {
		return refund.StoreRefundDTO{}, err
	}

	return refund.StoreRefundDTO{
		StoreID:              store.ID,
		BlockedTransactionID: blockedTransactionID,
		DestinationAddress:   req.DestinationAddress,
		Code:                 req.Code,
	}, nil
}

// prepareRefundError logs the error and converts it to the http error.
// Only the sentinel error text is returned, the wrapped details stay in the log.
func (h *Handler) prepareRefundError(err error, msg string, dto refund.StoreRefundDTO) error {
	var (
		httpCode = fiber.StatusBadRequest
		respErr  = errors.New(msg)
	)

	switch {
	case errors.Is(err, refund.ErrRefundBlockedTxNotFound):
		httpCode, respErr = fiber.StatusNotFound, refund.ErrRefundBlockedTxNotFound
	case errors.Is(err, refund.ErrRefundAlreadyRequested), errors.Is(err, refund.ErrRefundAlreadyExists):
		httpCode, respErr = fiber.StatusConflict, refund.ErrRefundAlreadyRequested
	case errors.Is(err, refund.ErrRefundCodeNotRequired):
		httpCode, respErr = fiber.StatusConflict, refund.ErrRefundCodeNotRequired
	case errors.Is(err, refund.ErrRefundCodeRequired):
		httpCode, respErr = fiber.StatusForbidden, refund.ErrRefundCodeRequired
	case errors.Is(err, refund.ErrRefundCodeInvalid):
		httpCode, respErr = fiber.StatusForbidden, refund.ErrRefundCodeInvalid
	case errors.Is(err, refund.ErrRefundCodeCooldown):
		httpCode, respErr = fiber.StatusTooManyRequests, refund.ErrRefundCodeCooldown
	case errors.Is(err, refund.ErrRefundCodeAttempts):
		httpCode, respErr = fiber.StatusTooManyRequests, refund.ErrRefundCodeAttempts
	case errors.Is(err, refund.ErrRefundNoEmail):
		httpCode, respErr = fiber.StatusUnprocessableEntity, refund.ErrRefundNoEmail
	case errors.Is(err, refund.ErrRefundInvalidAddress):
		httpCode, respErr = fiber.StatusUnprocessableEntity, refund.ErrRefundInvalidAddress
	case errors.Is(err, refund.ErrRefundInsufficientFee):
		httpCode, respErr = fiber.StatusUnprocessableEntity, refund.ErrRefundInsufficientFee
	case errors.Is(err, refund.ErrRefundInsufficientFunds):
		httpCode, respErr = fiber.StatusUnprocessableEntity, refund.ErrRefundInsufficientFunds
	case errors.Is(err, refund.ErrRefundSourceInvalid):
		httpCode, respErr = fiber.StatusUnprocessableEntity, refund.ErrRefundSourceInvalid
	}

	log := h.logger.Warnw
	if httpCode == fiber.StatusBadRequest {
		// an unexpected error: processing, database or a validation without a dedicated code
		log = h.logger.Errorw
	}
	log(msg,
		"store_id", dto.StoreID,
		"blocked_transaction_id", dto.BlockedTransactionID,
		"http_code", httpCode,
		"error", err,
	)

	return apierror.New().AddError(respErr).SetHttpCode(httpCode)
}

func (h *Handler) initRefundRoutes(router fiber.Router) {
	// each call goes to processing
	limited := middleware.LimiterMiddleware(30, 60, middleware.WithSlidingWindow)

	router.Post("/blocked-transactions/:id/refund/estimate", limited, h.estimateRefund)
	router.Post("/blocked-transactions/:id/refund/send-code", limited, h.sendRefundCode)
	router.Post("/blocked-transactions/:id/refund", limited, h.createRefund)
	router.Get("/refunds/:id", h.getRefund)
}
