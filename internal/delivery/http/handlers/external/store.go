package external

import (
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	// blank import for swaggen
	_ "github.com/dv-net/dv-merchant/internal/delivery/http/responses/currency_response"
	_ "github.com/dv-net/dv-merchant/internal/service/store"

	"github.com/gofiber/fiber/v3"
)

// storeCurrenciesExtended is a function to get all active store currencies with extended info
//
//	@Summary		Get extended list of store currencies
//	@Description	Get extended list of store currencies grouped by tokens and blockchains
//	@Tags			Store
//	@Accept			json
//	@Param			api_key	query		string	false	"Store API key"
//	@Success		200		{object}	response.Result[currency_response.CurrenciesExtendedResponse]
//	@Failure		401		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/external/store/currencies-extended [get]
//	@Security		XApiKey
func (h *Handler) storeCurrenciesExtended(c fiber.Ctx) error {
	store, err := loadAuthStore(c)
	if err != nil {
		return err
	}

	res, err := h.services.StoreService.GetStoreCurrencies(c.Context(), store.ID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.FromCurrencyModelsToExtendedResponse(res)))
}

// storeCurrencies is a function to get all active store currencies
//
//	@Summary		Get store currencies list
//	@Description	Get store currencies list
//	@Tags			Store
//	@Accept			json
//	@Param			api_key	query		string	false	"Store API key"
//	@Success		200		{object}	response.Result[[]currency_response.GetCurrencyResponse]
//	@Failure		401		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/external/store/currencies [get]
//	@Security		XApiKey
func (h *Handler) storeCurrencies(c fiber.Ctx) error {
	store, err := loadAuthStore(c)
	if err != nil {
		return err
	}

	res, err := h.services.StoreService.GetStoreCurrencies(c.Context(), store.ID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.FromCurrencyModelToResponses(res...)))
}

// storeCurrencyRate is a function to get all active store currencies
//
//	@Summary		Get store currency rate
//	@Description	Get store currency rate
//	@Tags			Store
//	@Accept			json
//	@Param			id		path		string	true	"Currency ID"
//	@Param			api_key	query		string	false	"Store API key"
//	@Success		200		{object}	response.Result[store.CurrencyRate]
//	@Failure		401		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/external/store/currencies/{id}/rate [get]
//	@Security		XApiKey
func (h *Handler) storeCurrencyRate(c fiber.Ctx) error {
	targetStore, err := loadAuthStore(c)
	if err != nil {
		return err
	}

	res, err := h.services.StoreCurrencyService.GetCurrencyWithRate(c.Context(), *targetStore, c.Params("id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(res))
}

func (h *Handler) initStoreRoutes(v3 fiber.Router) {
	storeRoutes := v3.Group("/store")
	storeRoutes.Get("/currencies-extended", h.storeCurrenciesExtended)
	storeRoutes.Get("/currencies", h.storeCurrencies)
	storeRoutes.Get("/currencies/:id/rate", h.storeCurrencyRate) // Deprecated remove after update lib
}
