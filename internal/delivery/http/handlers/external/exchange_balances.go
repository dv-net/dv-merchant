package external

import (
	"errors"
	"net/http"

	_ "github.com/dv-net/dv-merchant/internal/delivery/http/responses/exchange_response" // Blank import for swaggen
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// @Summary		Get external exchange balances
// @Description	Get external exchange balances
// @Tags			Exchange
// @Accept			json
// @Produce		json
// @Param			api_key	query		string	false	"Store API key"
// @Success		200		{object}	response.Result[exchange_response.ExternalExchangeBalanceResponse]
// @Failure		401		{object}	apierror.Errors
// @Failure		500		{object}	apierror.Errors
// @Router			/v1/external/exchange-balances [get]
// @Security		XApiKey
func (h *Handler) getExternalExchangeBalances(c fiber.Ctx) error {
	store, err := loadAuthStore(c)
	if err != nil {
		return err
	}

	user, err := h.services.UserService.GetUserByID(c.Context(), store.UserID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	if user.ExchangeSlug == nil {
		return apierror.New().AddError(errors.New("user has no exchanges setup")).SetHttpCode(http.StatusBadRequest)
	}

	balances, err := h.services.ExchangeService.GetExchangeBalance(c.Context(), *user.ExchangeSlug, *user)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}
	return c.JSON(response.OkByData(converters.FromExchangeBalanceModelToResponse(balances)))
}

func (h *Handler) initExchangeBalances(v1 fiber.Router) {
	v1.Get("/exchange-balances", h.getExternalExchangeBalances)
}
