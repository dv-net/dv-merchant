package external

import (
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/util"

	_ "github.com/dv-net/dv-merchant/internal/delivery/http/responses/wallet_response" // swaggo
	"github.com/dv-net/dv-merchant/internal/service/wallet"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/wallet_request"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// createWalletWithAddress is a function to Create wallet with address
//
//	@Summary		Create wallet with address
//	@Description	Create wallet with address
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			api_key		query		string									false	"Store API key"
//	@Param			register	body		wallet_request.ExternalCreateRequest	true	"Create wallet"
//	@Success		200			{object}	response.Result[wallet_response.CreateWalletExternalResponse]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/external/wallet [post]
//	@Security		XApiKey
func (h *Handler) createWalletWithAddressByBody(c fiber.Ctx) error {
	store, err := loadAuthStore(c)
	if err != nil {
		return err
	}

	request := &wallet_request.ExternalCreateRequest{}
	if err := c.Bind().Body(request); err != nil {
		return err
	}

	if request.Locale != nil && *request.Locale != "" {
		normalizedLocale := util.ParseLanguageTag(*request.Locale).String()
		request.Locale = &normalizedLocale
	}

	amountUSD := store.MinimalPayment
	if request.Amount.IsPositive() {
		amountUSD = request.Amount
	}

	walletWithAddress, err := h.services.WalletService.StoreWalletWithAddress(
		c.Context(),
		h.prepareDtoFromRequest(request, store),
		amountUSD.String(),
	)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	util.EnrichTopUpURLByParams(walletWithAddress.PayURL, util.TopUpParams{
		Amount:   request.Amount,
		Currency: request.Currency,
	})

	return c.JSON(response.OkByData(converters.FromWalletWithAddressModelToResponse(walletWithAddress)))
}

// createWalletWithAddress is a function to Create wallet with address
//
//	@Summary		Create wallet with address
//	@Description	Create wallet with address
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			api_key	query		string									false	"Store API key"
//	@Param			string	query		wallet_request.ExternalCreateRequest	true	"Create wallet"
//	@Success		200		{object}	response.Result[wallet_response.CreateWalletExternalResponse]
//	@Failure		401		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/external/wallet [get]
//	@Security		XApiKey
func (h *Handler) createWalletWithAddressByQuery(c fiber.Ctx) error {
	store, err := loadAuthStore(c)
	if err != nil {
		return err
	}

	request := &wallet_request.ExternalCreateRequest{}
	if err = c.Bind().Query(request); err != nil {
		return err
	}

	if request.Locale != nil && *request.Locale != "" {
		normalizedLocale := util.ParseLanguageTag(*request.Locale).String()
		request.Locale = &normalizedLocale
	}

	amountUSD := store.MinimalPayment
	if request.Amount.IsPositive() {
		amountUSD = request.Amount
	}

	walletWithAddress, err := h.services.WalletService.StoreWalletWithAddress(
		c.Context(),
		h.prepareDtoFromRequest(request, store),
		amountUSD.String(),
	)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	util.EnrichTopUpURLByParams(walletWithAddress.PayURL, util.TopUpParams{
		Amount:   request.Amount,
		Currency: request.Currency,
	})
	return c.JSON(response.OkByData(converters.FromWalletWithAddressModelToResponse(walletWithAddress)))
}

func (h *Handler) prepareDtoFromRequest(req *wallet_request.ExternalCreateRequest, store *models.Store) wallet.CreateStoreWalletWithAddressDTO {
	return wallet.CreateStoreWalletWithAddressDTO{
		StoreID:         store.ID,
		StoreExternalID: req.StoreExternalID,
		IP:              req.IP,
		Email:           req.Email,
		UntrustedEmail:  req.Email,
		Locale:          req.Locale,
	}
}

// @Summary		Get external hot wallet balances
// @Description	Get external hot wallet balances
//
// @Tags			Store
// @Produce		json
// @Param			api_key	query		string											false	"Store API key"
// @Success		200		{object}	response.Result[[]wallet.SummaryDTO]			"Successful operation"
// @Failure		401		{object}	apierror.Errors									"Unauthorized"
// @Failure		500		{object}	apierror.Errors									"Internal Server Error"
// @Param			string	query		wallet_request.GetSummarizedUserWalletsRequest	true	"Get summarized user wallets query params"
// @Router			/v1/external/wallet/balance/hot [get]
// @Security		XApiKey
func (h *Handler) getHotWalletBalances(c fiber.Ctx) error {
	store, err := loadAuthStore(c)
	if err != nil {
		return err
	}

	req := &wallet_request.GetSummarizedUserWalletsRequest{}
	if err := c.Bind().Query(req); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	usr, err := h.services.UserService.GetUserByID(c.Context(), store.UserID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	rates, err := h.services.ExRateService.LoadRatesList(c.Context(), usr.RateSource.String())
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	summary, err := h.services.WalletService.SummarizeUserWalletsByCurrency(c.Context(), usr.ID, rates, req.MinBalance)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(summary))
}

func (h *Handler) initWalletRoutes(v1 fiber.Router) {
	wallet := v1.Group("/wallet")
	wallet.Post("/", h.createWalletWithAddressByBody)
	wallet.Get("/", h.createWalletWithAddressByQuery)
	wallet.Get("/balance/hot", h.getHotWalletBalances)
}
