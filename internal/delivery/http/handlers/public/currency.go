package public

import (
	// Blank import for swaggen
	_ "github.com/dv-net/dv-merchant/internal/delivery/http/responses/currency_response"
	_ "github.com/dv-net/dv-merchant/internal/models"
	_ "github.com/dv-net/dv-merchant/internal/service/exrate"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// getAllCurrencies is a function to get all Currencies data from database
//
//	@Summary		Get all Currencies
//	@Description	Get all Currencies
//	@Tags			Currency
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[[]currency_response.GetCurrencyResponse]
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/public/currencies [get]
//	@Security		BearerAuth
func (h *Handler) getAllCurrencies(c fiber.Ctx) error {
	currencies, err := h.services.CurrencyService.GetAllCurrency(c.Context())
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	res := converters.FromCurrencyModelToResponses(currencies...)
	return c.JSON(response.OkByData(res))
}

// getCurrencyByID is a function to get a Currency by ID
//
//	@Summary		Get Currency by ID
//	@Description	Get Currency by ID
//	@Tags			Currency
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Currency ID"
//	@Success		200	{object}	response.Result[currency_response.GetCurrencyResponse]
//	@Failure		404	{object}	apierror.Errors
//	@Failure		503	{object}	apierror.Errors
//	@Router			/v1/public/currencies/{id} [get]
//	@Security		BearerAuth
func (h *Handler) getCurrencyByID(c fiber.Ctx) error {
	id := c.Params("id")
	currency, err := h.services.CurrencyService.GetCurrencyByID(c.Context(), id)
	if err != nil {
		return apierror.New().SetHttpCode(fiber.StatusNotFound)
	}
	resp := converters.FromCurrencyModelToResponse(currency)

	return c.JSON(response.OkByData(resp))
}

// getAllCurrencyRates is a function to get a Currency rate
//
//	@Summary		Get Currency rate
//	@Description	Get Currency rate
//	@Tags			Currency
//	@Accept			json
//	@Produce		json
//	@Param			source	path		string	true	"Currency Source example binance"
//	@Success		200		{object}	response.Result[[]exrate.ExRate]
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/public/currencies/rate/{source} [get]
//	@Security		BearerAuth
func (h *Handler) getAllCurrencyRates(c fiber.Ctx) error {
	source := c.Params("source", "okx")

	rates, err := h.services.ExRateService.GetAllCurrencyRate(c.Context(), source)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByData(rates))
}

func (h *Handler) initCurrencyRoutes(v1 fiber.Router) {
	currencies := v1.Group("/currencies")
	currencies.Get("/", h.getAllCurrencies)
	currencies.Get("/rate/", h.getAllCurrencyRates)
	currencies.Get("/rate/:source", h.getAllCurrencyRates)
	currencies.Get("/:id", h.getCurrencyByID)
}
