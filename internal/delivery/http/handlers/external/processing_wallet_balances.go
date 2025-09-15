package external

import (
	"errors"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/wallet_request"
	_ "github.com/dv-net/dv-merchant/internal/delivery/http/responses/wallet_response" // blank import for swaggo
	"github.com/dv-net/dv-merchant/internal/service/wallet"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// @Summary		Get external processing wallet balances
// @Description	Get external processing wallet balances
// @Tags			Withdrawal
// @Accept			json
// @Produce		json
// @Param			api_key	query		string	false	"Store API key"
// @Success		200		{object}	response.Result[wallet_response.ExternalProcessingWalletBalanceResponse]
// @Failure		401		{object}	apierror.Errors
// @Failure		500		{object}	apierror.Errors
// @Param			object	query		wallet_request.GetWalletAssetsRequest	true	"Get external processing wallet balances"
// @Router			/v1/external/processing-wallet-balances [get]
// @Security		XApiKey
func (h *Handler) getExternalProcessingWalletBalances(c fiber.Ctx) error {
	store, err := loadAuthStore(c)
	if err != nil {
		return err
	}

	user, err := h.services.UserService.GetUserByID(c.Context(), store.UserID)
	if err != nil {
		return err
	}

	ownerID := user.ProcessingOwnerID
	if !ownerID.Valid {
		return apierror.New().AddError(errors.New("empty owner processing id")).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &wallet_request.GetWalletAssetsRequest{}

	if err = c.Bind().Query(req); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	dto := wallet.GetProcessingWalletsDTO{
		OwnerID:     ownerID.UUID,
		Blockchains: req.Blockchains,
		Currencies:  req.Currencies,
	}

	assets, err := h.services.WalletBalanceService.GetProcessingBalances(c.Context(), dto)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(assets))
}

func (h *Handler) initProcessingWalletBalances(v1 fiber.Router) {
	v1.Get("/processing-wallet-balances", h.getExternalProcessingWalletBalances)
}
