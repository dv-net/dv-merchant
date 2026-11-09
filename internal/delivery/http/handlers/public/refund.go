package public

import (
	"errors"
	"time"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/public_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/refund_response"
	"github.com/dv-net/dv-merchant/internal/delivery/middleware"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/refund"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/response"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// refundLookup sends a verification code to the wallet's confirmed email if
// wallet_id/store_id/email match. The response is intentionally identical
// regardless of whether a match was found, to avoid leaking which part failed.
//
//	@Summary		Request a refund verification code
//	@Description	Sends an OTP code to the wallet's confirmed email
//	@Tags			Refund,Public
//	@Accept			json
//	@Produce		json
//	@Param			request	body		public_request.RefundLookupRequest	true	"RefundLookupRequest"
//	@Success		200		{object}	response.Result[string]
//	@Failure		400		{object}	apierror.Errors
//	@Router			/v1/public/refund/lookup [post]
func (h *Handler) refundLookup(c fiber.Ctx) error {
	req := &public_request.RefundLookupRequest{}
	if err := c.Bind().Body(req); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	_ = h.services.AuthService.SendWalletVerificationCode(c.Context(), req.WalletID, req.StoreID, req.Email)

	return c.JSON(response.OkByMessage("if the details are correct, a verification code has been sent"))
}

// refundVerify verifies the OTP code sent by refundLookup and, on success, issues
// a short-lived wallet session token used by subsequent refund cabinet endpoints.
//
//	@Summary		Verify a refund verification code
//	@Description	Verifies the OTP code and issues a wallet session token
//	@Tags			Refund,Public
//	@Accept			json
//	@Produce		json
//	@Param			request	body		public_request.RefundVerifyRequest	true	"RefundVerifyRequest"
//	@Success		200		{object}	response.Result[public_request.RefundVerifyResponse]
//	@Failure		401		{object}	apierror.Errors
//	@Router			/v1/public/refund/verify [post]
func (h *Handler) refundVerify(c fiber.Ctx) error {
	req := &public_request.RefundVerifyRequest{}
	if err := c.Bind().Body(req); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	token, err := h.services.AuthService.VerifyWalletCode(c.Context(), req.WalletID, req.StoreID, req.Email, req.Code)
	if err != nil {
		return apierror.New().AddError(errors.New("invalid or expired code")).SetHttpCode(fiber.StatusUnauthorized)
	}

	return c.JSON(response.OkByData(public_request.RefundVerifyResponse{Token: token.FullToken}))
}

// refundClaim creates a refund request for one of the caller's blocked transactions.
// wallet_id is taken from the authenticated session (RefundWalletAuthMiddleware /
// Locals("refund_wallet")), never from client input — otherwise a valid session for
// one wallet could be used to claim another wallet's blocked transaction.
//
//	@Summary		Create a refund request
//	@Description	Creates a refund request for a blocked transaction belonging to the authenticated wallet
//	@Tags			Refund,Public
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Blocked transaction ID"
//	@Param			request	body		public_request.RefundClaimRequest	true	"RefundClaimRequest"
//	@Success		200		{object}	response.Result[models.RefundRequest]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		409		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Router			/v1/public/refund/blocked-transactions/{id}/claim [post]
func (h *Handler) refundClaim(c fiber.Ctx) error {
	w, err := loadRefundWallet(c)
	if err != nil {
		return err
	}

	blockedTransactionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(errors.New("bad blocked transaction id")).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &public_request.RefundClaimRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	ref, err := h.services.RefundService.CreateRefund(c.Context(), refund.CreateRefundDTO{
		WalletID:             w.ID,
		BlockedTransactionID: blockedTransactionID,
		DestinationAddress:   req.DestinationAddress,
		Email:                w.Email.String,
	})
	if err != nil {
		return h.prepareRefundClaimError(err, "failed to create refund request", w.ID, blockedTransactionID)
	}

	return c.JSON(response.OkByData(ref))
}

// refundEstimate returns the network fee and the amount the payer receives for a refund of one
// of the caller's blocked transactions to the destination address. Nothing is created.
// It runs the same checks as refundClaim, wallet_id is taken from the authenticated session.
//
//	@Summary		Estimate a refund
//	@Description	Returns the estimated network fee and the amount to receive for a refund of a blocked transaction belonging to the authenticated wallet. The fee is an estimate and may change by the time of sending
//	@Tags			Refund,Public
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Blocked transaction ID"
//	@Param			request	body		public_request.RefundClaimRequest	true	"RefundClaimRequest"
//	@Success		200		{object}	response.Result[refund_response.RefundEstimateResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Failure		409		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Router			/v1/public/refund/blocked-transactions/{id}/estimate [post]
func (h *Handler) refundEstimate(c fiber.Ctx) error {
	w, err := loadRefundWallet(c)
	if err != nil {
		return err
	}

	blockedTransactionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apierror.New().AddError(errors.New("bad blocked transaction id")).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &public_request.RefundClaimRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	estimate, err := h.services.RefundService.EstimateRefund(c.Context(), refund.EstimateRefundDTO{
		WalletID:             w.ID,
		BlockedTransactionID: blockedTransactionID,
		DestinationAddress:   req.DestinationAddress,
	})
	if err != nil {
		return h.prepareRefundClaimError(err, "failed to estimate refund", w.ID, blockedTransactionID)
	}

	return c.JSON(response.OkByData(refund_response.NewRefundEstimateResponse(estimate)))
}

// refundToSender creates a refund request to a deposit sender without the email verification.
// It is allowed only when the store owner enabled it (user setting refund_auth_mode = custom_address_only).
// The payer sends the address, processing checks it against the deposit senders in the blockchain.
//
//	@Summary		Create a refund to the deposit sender
//	@Description	Creates a refund request of a blocked transaction to the address it was sent from, without the email verification. The address must be one of the deposit senders (any input address for btc like blockchains), otherwise the email verification is required. Available when the store owner allows it, see refund_to_sender_without_auth of the wallet data
//	@Tags			Refund,Public
//	@Accept			json
//	@Produce		json
//	@Param			id						path		string								true	"Wallet ID"
//	@Param			blocked_transaction_id	path		string								true	"Blocked transaction ID"
//	@Param			request					body		public_request.RefundClaimRequest	true	"RefundClaimRequest"
//	@Success		200						{object}	response.Result[models.RefundRequest]
//	@Failure		400						{object}	apierror.Errors
//	@Failure		403						{object}	apierror.Errors
//	@Failure		404						{object}	apierror.Errors
//	@Failure		409						{object}	apierror.Errors
//	@Failure		422						{object}	apierror.Errors
//	@Router			/v1/public/wallet/{id}/blocked-transactions/{blocked_transaction_id}/refund-to-sender [post]
func (h *Handler) refundToSender(c fiber.Ctx) error {
	dto, err := parseRefundToSenderParams(c)
	if err != nil {
		return err
	}

	ref, err := h.services.RefundService.CreateRefundToSender(c.Context(), dto)
	if err != nil {
		return h.prepareRefundClaimError(err, "failed to create refund request", dto.WalletID, dto.BlockedTransactionID)
	}

	return c.JSON(response.OkByData(ref))
}

// refundToSenderEstimate returns the network fee and the amount the payer receives for a refund
// to a deposit sender without the email verification. Nothing is created.
//
//	@Summary		Estimate a refund to the deposit sender
//	@Description	Returns the estimated network fee and the amount to receive for a refund of a blocked transaction to the address it was sent from. The address must be one of the deposit senders. Available when the store owner allows refunds without the email verification
//	@Tags			Refund,Public
//	@Accept			json
//	@Produce		json
//	@Param			id						path		string								true	"Wallet ID"
//	@Param			blocked_transaction_id	path		string								true	"Blocked transaction ID"
//	@Param			request					body		public_request.RefundClaimRequest	true	"RefundClaimRequest"
//	@Success		200						{object}	response.Result[refund_response.RefundEstimateResponse]
//	@Failure		400						{object}	apierror.Errors
//	@Failure		403						{object}	apierror.Errors
//	@Failure		404						{object}	apierror.Errors
//	@Failure		409						{object}	apierror.Errors
//	@Failure		422						{object}	apierror.Errors
//	@Router			/v1/public/wallet/{id}/blocked-transactions/{blocked_transaction_id}/refund-to-sender/estimate [post]
func (h *Handler) refundToSenderEstimate(c fiber.Ctx) error {
	dto, err := parseRefundToSenderParams(c)
	if err != nil {
		return err
	}

	estimate, err := h.services.RefundService.EstimateRefundToSender(c.Context(), dto)
	if err != nil {
		return h.prepareRefundClaimError(err, "failed to estimate refund", dto.WalletID, dto.BlockedTransactionID)
	}

	return c.JSON(response.OkByData(refund_response.NewRefundEstimateResponse(estimate)))
}

func parseRefundToSenderParams(c fiber.Ctx) (refund.ToSenderDTO, error) {
	walletID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return refund.ToSenderDTO{}, apierror.New().AddError(errors.New("bad wallet id")).SetHttpCode(fiber.StatusBadRequest)
	}

	blockedTransactionID, err := uuid.Parse(c.Params("blocked_transaction_id"))
	if err != nil {
		return refund.ToSenderDTO{}, apierror.New().AddError(errors.New("bad blocked transaction id")).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &public_request.RefundClaimRequest{}
	if err := c.Bind().Body(req); err != nil {
		return refund.ToSenderDTO{}, err
	}

	return refund.ToSenderDTO{
		WalletID:             walletID,
		BlockedTransactionID: blockedTransactionID,
		DestinationAddress:   req.DestinationAddress,
	}, nil
}

// prepareRefundClaimError logs the claim (or estimate) error and converts it to the http error.
// Only the sentinel error text is returned to the payer, the wrapped details stay in the log.
func (h *Handler) prepareRefundClaimError(err error, msg string, walletID, blockedTransactionID uuid.UUID) error {
	var (
		httpCode = fiber.StatusBadRequest
		respErr  = errors.New(msg)
	)

	switch {
	case errors.Is(err, refund.ErrRefundBlockedTxNotFound):
		httpCode, respErr = fiber.StatusNotFound, refund.ErrRefundBlockedTxNotFound
	case errors.Is(err, refund.ErrRefundAlreadyRequested), errors.Is(err, refund.ErrRefundAlreadyExists):
		httpCode, respErr = fiber.StatusConflict, refund.ErrRefundAlreadyRequested
	case errors.Is(err, refund.ErrRefundInvalidAddress):
		httpCode, respErr = fiber.StatusUnprocessableEntity, refund.ErrRefundInvalidAddress
	case errors.Is(err, refund.ErrRefundInsufficientFee):
		httpCode, respErr = fiber.StatusUnprocessableEntity, refund.ErrRefundInsufficientFee
	case errors.Is(err, refund.ErrRefundInsufficientFunds):
		httpCode, respErr = fiber.StatusUnprocessableEntity, refund.ErrRefundInsufficientFunds
	case errors.Is(err, refund.ErrRefundSourceInvalid):
		httpCode, respErr = fiber.StatusUnprocessableEntity, refund.ErrRefundSourceInvalid
	case errors.Is(err, refund.ErrRefundNotSender):
		httpCode, respErr = fiber.StatusUnprocessableEntity, refund.ErrRefundNotSender
	case errors.Is(err, refund.ErrRefundAuthRequired):
		httpCode, respErr = fiber.StatusForbidden, refund.ErrRefundAuthRequired
	}

	log := h.logger.Warnw
	if httpCode == fiber.StatusBadRequest {
		// an unexpected error: processing, database or a validation without a dedicated code
		log = h.logger.Errorw
	}
	log(msg,
		"wallet_id", walletID,
		"blocked_transaction_id", blockedTransactionID,
		"http_code", httpCode,
		"error", err,
	)

	return apierror.New().AddError(respErr).SetHttpCode(httpCode)
}

// refundCabinet lists the authenticated wallet's blocked transactions, each paired
// with the refund request already filed against it (if any).
//
//	@Summary		List a wallet's blocked transactions
//	@Description	Lists blocked transactions for the authenticated wallet, with refund status if a claim was filed
//	@Tags			Refund,Public
//	@Produce		json
//	@Success		200	{object}	response.Result[map[string][]refund_response.CabinetItemResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Router			/v1/public/refund/cabinet [get]
func (h *Handler) refundCabinet(c fiber.Ctx) error {
	w, err := loadRefundWallet(c)
	if err != nil {
		return err
	}

	grouped, err := h.services.RefundService.GetCabinet(c.Context(), w.ID)
	if err != nil {
		return apierror.New().AddError(errors.New("failed to fetch cabinet")).SetHttpCode(fiber.StatusBadRequest)
	}

	result := make(map[string][]refund_response.CabinetItemResponse, len(grouped))
	for bucket, items := range grouped {
		bucketItems := make([]refund_response.CabinetItemResponse, 0, len(items))
		for _, item := range items {
			bucketItems = append(bucketItems, refund_response.CabinetItemResponse{
				BlockedTransactionID: item.BlockedTransactionID,
				TransactionID:        item.TransactionID,
				TxHash:               item.TxHash,
				Blockchain:           item.Blockchain,
				CurrencyID:           item.CurrencyID,
				CurrencyCode:         item.CurrencyCode,
				Amount:               item.Amount,
				AmountUsd:            item.AmountUsd,
				FromAddress:          item.FromAddress,
				ToAddress:            item.ToAddress,
				RiskLevel:            item.RiskLevel,
				Score:                item.Score,
				BlockReasons:         refund_response.NewCabinetBlockReasonsResponse(item.BlockReasons),
				CreatedAt:            pgtypeutils.DecodeTime(item.CreatedAt),
				RefundStatus:         item.RefundStatus,
				DestinationAddress:   item.DestinationAddress,
				RefundExecuteAfter:   pgtypeutils.DecodeTime(item.RefundExecuteAfter),
			})
		}
		result[bucket] = bucketItems
	}

	return c.JSON(response.OkByData(result))
}

func loadRefundWallet(c fiber.Ctx) (*models.Wallet, error) {
	w, ok := c.Locals("refund_wallet").(*models.Wallet)
	if !ok {
		return nil, apierror.New().AddError(errors.New("undefined wallet")).SetHttpCode(fiber.StatusUnauthorized)
	}
	return w, nil
}

func (h *Handler) initRefundRoutes(v1 fiber.Router) {
	r := v1.Group("/refund")
	r.Post("/lookup",
		middleware.LimiterMiddleware(3, 60, middleware.WithSlidingWindow),
		middleware.FakeDelayMiddleware(2*time.Second),
		h.refundLookup)
	r.Post("/verify",
		middleware.LimiterMiddleware(5, 60, middleware.WithSlidingWindow),
		middleware.FakeDelayMiddleware(2*time.Second),
		h.refundVerify)
	r.Post("/blocked-transactions/:id/claim",
		middleware.RefundWalletAuthMiddleware(h.services.AuthService),
		h.refundClaim)
	// each call goes to processing
	r.Post("/blocked-transactions/:id/estimate",
		middleware.LimiterMiddleware(10, 60, middleware.WithSlidingWindow),
		middleware.RefundWalletAuthMiddleware(h.services.AuthService),
		h.refundEstimate)
	r.Get("/cabinet",
		middleware.RefundWalletAuthMiddleware(h.services.AuthService),
		h.refundCabinet)
}
