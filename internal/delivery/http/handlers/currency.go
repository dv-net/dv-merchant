package handlers

import (
	"errors"

	// Blank import for swaggen
	_ "github.com/dv-net/dv-merchant/internal/delivery/http/responses/currency_response"
	_ "github.com/dv-net/dv-merchant/internal/service/exrate"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/currency_request"
	"github.com/dv-net/dv-merchant/internal/models"
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
//	@Router			/v1/dv-admin/currencies [get]
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
//	@Router			/v1/dv-admin/currencies/{id} [get]
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
//	@Router			/v1/dv-admin/currencies/rate/{source} [get]
//	@Security		BearerAuth
func (h *Handler) getAllCurrencyRates(c fiber.Ctx) error {
	source := c.Params("source", "okx")
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	rates, err := h.services.ExRateService.GetAllCurrencyRate(c.Context(), source, user.RateScale)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByData(rates))
}

// updateRate is a function to Update rate scale
//
//	@Summary		Get Currency rate
//	@Description	Get Currency rate
//	@Tags			Currency
//	@Accept			json
//	@Produce		json
//	@Param			register	body		currency_request.UpdateCurrencyRateRequest	true	"Rate scale"
//	@Success		200			{object}	response.Result[string]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Router			/v1/dv-admin/currencies/rate/ [put]
//	@Security		BearerAuth
func (h Handler) updateRate(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	req := &currency_request.UpdateCurrencyRateRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	preparedSource := models.RateSource(req.RateSource)
	if !preparedSource.Valid() {
		return apierror.New().AddError(errors.New("invalid source")).SetHttpCode(fiber.StatusBadRequest)
	}

	err = h.services.UserService.UpdateRate(c.Context(), user, preparedSource, req.RateScale)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("Rete source successfully updated"))
}

func (h *Handler) initCurrencyRoutes(v1 fiber.Router) {
	currencies := v1.Group("/currencies")
	currencies.Get("/", h.getAllCurrencies)
	currencies.Get("/rate/", h.getAllCurrencyRates)
	currencies.Put("/rate/", h.updateRate)
	currencies.Get("/rate/:source", h.getAllCurrencyRates)
	currencies.Get("/:id", h.getCurrencyByID)
}
