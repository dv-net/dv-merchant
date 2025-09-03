package public

import (
	"errors"
	"net/http"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/public_request"
	"github.com/dv-net/dv-merchant/internal/service/store"
	"github.com/dv-net/dv-merchant/internal/service/wallet"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"
	"github.com/dv-net/dv-merchant/internal/util"

	"github.com/gofiber/fiber/v3"
)

// getTopUpDataByStore is a function to get full wallet data
//
//	@Summary		Get top up data by store
//	@Description	Get wallet full data
//	@Tags			Wallet,Public
//	@Accept			json
//	@Produce		json
//	@Param			slug		path		string									true	"Slug of the shop"
//	@Param			external_id	path		string									true	"External client ID"
//	@Param			string		query		public_request.TopUpFormByStoreRequest	true	"TopUpFormByStoreRequest"
//	@Success		200			{object}	response.Result[public_request.GetWalletDto]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		410			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/public/store/topup/{store_id}/{external_id} [get]
func (h *Handler) getTopUpDataByStore(c fiber.Ctx) error {
	request := &public_request.TopUpFormByStoreRequest{}
	if err := c.Bind().Query(request); err != nil {
		return err
	}

	if request.Locale != nil && *request.Locale != "" {
		normalizedLocale := util.ParseLanguageTag(*request.Locale).String()
		request.Locale = &normalizedLocale
	}

	storeID, err := tools.ValidateUUID(c.Params("store_id"))
	if err != nil {
		return apierror.New().AddError(errors.New("invalid store id")).SetHttpCode(http.StatusBadRequest)
	}

	st, err := h.services.StoreService.GetStoreByID(c.Context(), storeID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	if !st.Status {
		return apierror.New().AddError(store.ErrStoreDisabled).SetHttpCode(fiber.StatusGone)
	}

	data, err := h.services.StoreService.PrepareTopUpDataByStore(c.Context(), wallet.CreateStoreWalletWithAddressDTO{
		StoreID:         storeID,
		StoreExternalID: c.Params("external_id"),
		UntrustedEmail:  request.Email,
		IP:              request.IP,
		Locale:          request.Locale,
	})
	if err != nil {
		code := fiber.StatusBadRequest
		if errors.Is(err, store.ErrStoreNotFound) {
			code = fiber.StatusNotFound
		} else if errors.Is(err, store.ErrStoreRequestLimitExceeded) {
			code = fiber.StatusForbidden
		}

		return apierror.New().AddError(err).SetHttpCode(code)
	}

	resp := converters.ConvertTopUpDataToResponse(data)
	return c.JSON(response.OkByData(resp))
}

func (h *Handler) initStoreRoutes(v1 fiber.Router) {
	group := v1.Group("/store")
	group.Get("/topup/:store_id/:external_id", h.getTopUpDataByStore)
}
