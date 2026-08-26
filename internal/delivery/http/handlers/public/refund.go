package public

import (
	"errors"
	"time"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/public_request"
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
		return apierror.New().AddError(errors.New("failed create refund request")).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(ref))
}

// refundCabinet lists the authenticated wallet's blocked transactions, each paired
// with the refund request already filed against it (if any).
//
//	@Summary		List a wallet's blocked transactions
//	@Description	Lists blocked transactions for the authenticated wallet, with refund status if a claim was filed
//	@Tags			Refund,Public
//	@Produce		json
//	@Success		200	{object}	response.Result[map[string][]public_request.CabinetItemResponse]
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

	result := make(map[string][]public_request.CabinetItemResponse, len(grouped))
	for bucket, items := range grouped {
		bucketItems := make([]public_request.CabinetItemResponse, 0, len(items))
		for _, item := range items {
			bucketItems = append(bucketItems, public_request.CabinetItemResponse{
				BlockedTransactionID: item.BlockedTransactionID,
				TransactionID:        item.TransactionID,
				TxHash:               item.TxHash,
				Blockchain:           item.Blockchain,
				CurrencyID:           item.CurrencyID,
				RiskLevel:            item.RiskLevel,
				Score:                item.Score,
				CreatedAt:            pgtypeutils.DecodeTime(item.CreatedAt),
				RefundStatus:         item.RefundStatus,
				DestinationAddress:   item.DestinationAddress,
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
	r.Get("/cabinet",
		middleware.RefundWalletAuthMiddleware(h.services.AuthService),
		h.refundCabinet)
}
