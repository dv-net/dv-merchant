package handlers

import (
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/wallet_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/wallet_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/wallet"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"
	"github.com/dv-net/dv-merchant/internal/util"

	// Blank import for swaggen
	_ "github.com/dv-net/dv-merchant/internal/service/transactions"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type WalletShort struct {
	ID              uuid.UUID       `json:"id,omitempty"`
	StoreID         uuid.UUID       `json:"store_id"`
	StoreExternalID string          `json:"store_external_id"`
	Address         []*AddressShort `json:"address,omitempty"`
} //	@name	WalletShort

type AddressShort struct {
	CurrencyID string            `json:"currency_id"`
	Blockchain models.Blockchain `json:"blockchain"`
	Address    string            `json:"address"`
} //	@name	WalletShort

type AddressFull struct {
	WalletID   uuid.UUID         `json:"wallet_id"`
	CurrencyID string            `json:"currency_id"`
	Blockchain models.Blockchain `json:"blockchain"`
	Address    string            `json:"address"`
} //	@name	WalletFull

// createWalletWithAddresses creates/returns wallet with addressed pre-generated.
//
//	@Summary		Create/return wallet with addresses
//	@Description	Creates/returns wallet with addressed pre-generated.
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			json	body		wallet_request.CreateRequest	true	"CreateRequest"
//	@Success		200		{object}	response.Result[handlers.WalletShort]
//	@Failure		401		{object}	apierror.Errors
//	@Failure		500		{object}	apierror.Errors
//	@Router			/v1/dv-admin/wallet/addresses [post]
//	@Security		BearerAuth
func (h Handler) createWalletWithAddresses(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	request := &wallet_request.CreateRequest{}
	if err := c.Bind().Body(request); err != nil {
		return err
	}

	if request.Locale != nil && *request.Locale != "" {
		normalizedLocale := util.ParseLanguageTag(*request.Locale).String()
		request.Locale = &normalizedLocale
	}
	storeID, err := tools.ValidateUUID(request.StoreID)
	if err != nil {
		return err
	}
	store, err := h.services.StoreService.GetStoreByID(c.Context(), storeID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
	}
	if user.ID != store.UserID {
		return apierror.New().AddError(errors.New("this is not your store")).SetHttpCode(fiber.StatusUnauthorized)
	}
	dto := wallet.CreateStoreWalletWithAddressDTO{
		StoreID:         storeID,
		StoreExternalID: request.StoreExternalID,
		Email:           request.Email,
		Locale:          request.Locale,
	}
	amount := "0"
	if request.Amount.IsPositive() {
		amount = request.Amount.String()
	}

	walletWithAddress, err := h.services.WalletService.StoreWalletWithAddress(c.Context(), dto, amount)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	util.EnrichTopUpURLByParams(walletWithAddress.PayURL, util.TopUpParams{
		Amount:   request.Amount,
		Currency: request.Currency,
	})

	return c.JSON(response.OkByData(converters.FromWalletWithAddressModelToResponse(walletWithAddress)))
}

type WalletBalance struct {
	CurrencyID string `json:"currency_id"`
	Address    string `json:"address"`
	Blockchain string `json:"blockchain"`
	Balance    string `json:"balance"`
} //	@name	WalletBalance

// getWalletAddressesKeys creates an empty wallet or returns an existing one.
//
//	@Summary		Get wallet's private/public keys.
//	@Description	This endpoint returns private keys for each wallet's asset.
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//
//	@Param			totp	query		string			true	"TOTP auth code"
//	@Failure		401		{object}	apierror.Errors	"Unauthorized"
//	@Failure		500		{object}	apierror.Errors	"Internal Server Error"
//	@Router			/v1/dv-admin/wallet/addresses/keys [get]
//	@Security		BearerAuth
func (h Handler) getWalletAddressesKeys(c fiber.Ctx) error {
	u, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	ownerID := u.ProcessingOwnerID
	if !ownerID.Valid {
		return apierror.New().AddError(errors.New("empty owner processing id")).SetHttpCode(fiber.StatusBadRequest)
	}

	totp := c.Query("totp")
	if totp == "" {
		return apierror.New().AddError(errors.New("empty totp param")).SetHttpCode(fiber.StatusBadRequest)
	}

	request := &wallet_request.GetKeysRequest{
		OwnerID: ownerID.UUID,
		TOTP:    totp,
	}

	pairs, err := h.services.ProcessingOwnerService.GetOwnerPrivateKeys(c.Context(), request.OwnerID, request.TOTP)
	if err != nil {
		return apierror.New().AddError(fmt.Errorf("failed to get private keys %w", err)).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(pairs))
}

// getWalletSeeds creates an empty wallet or returns an existing one.
//
//	@Summary		Get wallet's seed phrase
//	@Description	This endpoint returns wallet's seed phrases
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			totp	query		string												true	"TOTP auth code"
//	@Success		200		{object}	response.Result[wallet_response.WalletSeedResponse]	"Successful operation"
//	@Failure		401		{object}	apierror.Errors										"Unauthorized"
//	@Failure		500		{object}	apierror.Errors										"Internal Server Error"
//	@Router			/v1/dv-admin/wallet/addresses/seeds [get]
//	@Security		BearerAuth
func (h Handler) getWalletSeeds(c fiber.Ctx) error {
	u, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	ownerID := u.ProcessingOwnerID
	if !ownerID.Valid {
		return apierror.New().AddError(errors.New("empty owner processing id")).SetHttpCode(fiber.StatusBadRequest)
	}

	totp := c.Query("totp")
	if totp == "" {
		return apierror.New().AddError(errors.New("empty totp param")).SetHttpCode(fiber.StatusBadRequest)
	}

	request := &wallet_request.GetSeedsRequest{
		OwnerID: ownerID.UUID,
		TOTP:    totp,
	}

	data, err := h.services.ProcessingOwnerService.GetOwnerSeed(c.Context(), request.OwnerID, request.TOTP)
	if err != nil {
		return apierror.New().AddError(fmt.Errorf("failed to get mnemonic phrase %w", err)).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(data))
}

// getWallets get wallet's with balance
//
//	@Summary		Get wallet's with balance
//	@Description	This endpoint returns wallet's with balance
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			json	body		wallet_request.GetWalletByStoreRequest																true	"GetWalletByStoreRequest"
//	@Success		200		{object}	response.Result[storecmn.FindResponseWithFullPagination[wallet_response.GetWalletBalanceResponse]]	"Successful operation"
//	@Failure		401		{object}	apierror.Errors																						"Unauthorized"
//	@Failure		500		{object}	apierror.Errors																						"Internal Server Error"
//	@Router			/v1/dv-admin/wallet [post]
//	@Security		BearerAuth
func (h Handler) getWallets(c fiber.Ctx) error {
	_, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	request := &wallet_request.GetWalletByStoreRequest{}

	if err := c.Bind().Body(request); err != nil {
		return err
	}

	rates, err := h.services.ExRateService.LoadRatesList(c.Context(), "okx")
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	wallets, err := h.services.BalanceService.GetWalletBalance(c.Context(), *request, rates)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(
		&storecmn.FindResponseWithFullPagination[wallet_response.GetWalletBalanceResponse]{
			Pagination: wallets.Pagination,
			Items:      converters.WalletBalanceModelsToResponses(wallets.Items...),
		}))
}

// getSummary get wallet's summary
//
//	@Summary		Get wallet's summary by currencies
//	@Description	This endpoint returns wallet's summary stats grouped by currencies
//	@Tags			Wallet
//	@Produce		json
//	@Success		200		{object}	response.Result[[]wallet.SummaryDTO]			"Successful operation"
//	@Failure		401		{object}	apierror.Errors									"Unauthorized"
//	@Failure		500		{object}	apierror.Errors									"Internal Server Error"
//	@Param			string	query		wallet_request.GetSummarizedUserWalletsRequest	true	"Get summarized user wallets query params"
//	@Router			/v1/dv-admin/wallet/summary [get]
//	@Security		BearerAuth
func (h *Handler) getSummary(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &wallet_request.GetSummarizedUserWalletsRequest{}
	if err := c.Bind().Query(req); err != nil {
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

// getHotWalletBalances get hot wallet's addresses total balance
//
//	@Summary		Get hot wallet's addresses total balance
//	@Description	Get hot wallet's addresses total balance
//	@Tags			Wallet
//	@Produce		json
//	@Success		200	{object}	response.Result[wallet_response.WalletAddressTotalUSDResponse]	"Successful operation"
//	@Failure		401	{object}	apierror.Errors													"Unauthorized"
//	@Failure		500	{object}	apierror.Errors													"Internal Server Error"
//	@Router			/v1/dv-admin/wallet/balances/hot [get]
//	@Security		BearerAuth
func (h *Handler) getHotWalletBalances(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	balances, err := h.services.BalanceService.GetHotWalletsTotalBalance(c.Context(), usr)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(&wallet_response.WalletAddressTotalUSDResponse{
		TotalUSD:  balances.TotalUSD.RoundDown(2),
		TotalDust: balances.TotalDust.RoundDown(2),
	}))
}

// getColdWalletBalances get cold wallet's addresses total balance
//
//	@Summary		Get cold wallet's addresses total balance
//	@Description	Get cold wallet's addresses total balance
//	@Tags			Wallet
//	@Produce		json
//	@Success		200	{object}	response.Result[wallet_response.WalletAddressTotalUSDResponse]	"Successful operation"
//	@Failure		401	{object}	apierror.Errors													"Unauthorized"
//	@Failure		500	{object}	apierror.Errors													"Internal Server Error"
//	@Router			/v1/dv-admin/wallet/balances/cold [get]
//	@Security		BearerAuth
func (h *Handler) getColdWalletBalances(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	balances, err := h.services.BalanceService.GetColdWalletsTotalBalance(c.Context(), usr)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(&wallet_response.WalletAddressTotalUSDResponse{
		TotalUSD: balances.TotalUSD.RoundDown(2),
	}))
}

// getWithdrawalWalletBalances get withdrawal wallet's addresses total balance
//
//	@Summary		Get withdrawal wallet's addresses total balance
//	@Description	Get withdrawal wallet's addresses total balance
//	@Tags			Wallet
//	@Produce		json
//	@Success		200	{object}	response.Result[wallet_response.WalletAddressTotalUSDResponse]	"Successful operation"
//	@Failure		401	{object}	apierror.Errors													"Unauthorized"
//	@Failure		500	{object}	apierror.Errors													"Internal Server Error"
//	@Router			/v1/dv-admin/wallet/balances/exchange-settings [get]
//	@Security		BearerAuth
func (h *Handler) getWithdrawalWalletBalances(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	balances, err := h.services.BalanceService.GetExchangeWalletsTotalBalance(c.Context(), usr)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(&wallet_response.WalletAddressTotalUSDResponse{
		TotalUSD: balances.TotalUSD.RoundDown(2),
	}))
}

// findWalletsWithTransactions find wallet's with transactions
//
//	@Summary		Find wallet's by address, id or  summary by currencies
//	@Description	This endpoint returns wallet's data with transactions
//	@Tags			Wallet
//	@Produce		json
//	@Success		200	{object}	response.Result[[]transactions.WalletWithTransactionsInfo]	"Successful operation"
//	@Failure		401	{object}	apierror.Errors												"Unauthorized"
//	@Failure		500	{object}	apierror.Errors												"Internal Server Error"
//	@Router			/v1/dv-admin/wallet/{searchParam} [get]
//	@Security		BearerAuth
func (h *Handler) findWalletsWithTransactions(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	result, err := h.services.WalletTransactionService.GetWalletInfoWithTransactionsByAddress(
		c.Context(),
		usr.ID,
		c.Params("searchParam"),
	)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(result))
}

// get wallet's info
//
//	@Summary		Find wallet's and get info grouped by blockchain
//	@Description	This endpoint returns wallet's data group by blockchains  **Deprecated**: Use `/v1/dv-admin/search/:searchParam instead
//	@Tags			Wallet
//	@Produce		json
//	@Success		200	{object}	response.Result[wallet.WithBlockchains]	"Successful operation"
//	@Failure		401	{object}	apierror.Errors							"Unauthorized"
//	@Failure		500	{object}	apierror.Errors							"Internal Server Error"
//	@Router			/v1/dv-admin/wallet/info/{searchParam} [get]
//	@Security		BearerAuth
func (h *Handler) getWalletInfo(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	wallets, err := h.services.WalletService.GetWalletsInfo(c.Context(), usr.ID, c.Params("searchParam"))
	if err != nil {
		if errors.Is(err, wallet.ErrServiceWalletNotFound) {
			return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
		}
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(wallets))
}

// getHotWalletKeys get hot wallet's keys
//
//	@Summary		Get hot wallet's keys
//	@Description	Get hot wallet's keys
//	@Tags			Wallet
//	@Produce		json
//	@Param			json	body		wallet_request.GetHotWalletKeysRequest	true	"GetHotWalletKeysRequest"
//	@Failure		401		{object}	apierror.Errors							"Unauthorized"
//	@Failure		500		{object}	apierror.Errors							"Internal Server Error"
//	@Router			/v1/dv-admin/wallet/keys/hot [post]
//	@Security		BearerAuth
func (h *Handler) getHotWalletKeys(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	request := &wallet_request.GetHotWalletKeysRequest{}
	if err := c.Bind().Body(request); err != nil {
		return err
	}
	dto := wallet.LoadPrivateKeyDTO{
		User:                       usr,
		Otp:                        request.TOTP,
		WalletAddressIDs:           request.WalletAddressIDs,
		ExcludedWalletAddressesIDs: request.ExcludedWalletAddressIDs,
		FileType:                   request.FileType,
		IP:                         c.IP(),
	}
	data, err := h.services.WalletService.LoadPrivateAddresses(c.Context(), dto)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	c.Response().Header.Set("Content-Type", "application/octet-stream")
	c.Response().Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", "hot_wallet_keys."+request.FileType))
	return c.SendStream(data, data.Len())
}

// Restore wallet's transactions
//
//	@Summary		Restore missed wallet transaction
//	@Description	This endpoint returns wallet's data group by blockchains
//	@Tags			Wallet
//	@Success		200	{object}	response.Result[string]
//	@Failure		401	{object}	apierror.Errors	"Unauthorized"
//	@Failure		500	{object}	apierror.Errors	"Internal Server Error"
//	@Router			/v1/dv-admin/wallet/restore/{walletId} [POST]
//	@Security		BearerAuth
func (h *Handler) restoreTx(c fiber.Ctx) error {
	_, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	walletID, err := uuid.Parse(c.Params("walletId"))
	if err != nil {
		return err
	}

	if err := h.services.WalletRestorer.RestoreWallet(c.Context(), walletID); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("ok"))
}

// Convert wallet's address
//
//	@Summary		Convert wallet's address
//	@Description	This endpoint convert legacy address to new format
//	@Tags			Wallet
//	@Param			json	body		wallet_request.AddressConverterRequest						true	"AddressConverterRequest"
//	@Success		200		{object}	response.Result[wallet_response.ConvertedAddressResponse]	"Successful operation"
//	@Failure		401		{object}	apierror.Errors												"Unauthorized"
//	@Failure		500		{object}	apierror.Errors												"Internal Server Error"
//	@Router			/v1/dv-admin/wallet/addresses/converter [post]
//	@Security		BearerAuth
func (h *Handler) addressConverter(c fiber.Ctx) error {
	_, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	req := &wallet_request.AddressConverterRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	addr, err := h.services.WalletConverter.ConvertLegacyAddressToNewFormat(req.Address, req.Blockchain)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByData(wallet_response.ConvertedAddressResponse{
		Address: addr.Address,
		Legacy:  addr.LegacyAddress,
	}))
}

func (h Handler) initWalletRoutes(v1 fiber.Router) {
	walletRouter := v1.Group("/wallet")
	walletRouter.Post("/", h.getWallets)
	walletRouter.Get("/summary", h.getSummary)
	walletRouter.Post("/addresses", h.createWalletWithAddresses)
	walletRouter.Get("/addresses/keys", h.getWalletAddressesKeys)
	walletRouter.Get("/addresses/seeds", h.getWalletSeeds)
	walletRouter.Post("/addresses/converter", h.addressConverter)
	walletRouter.Get("/balances/hot", h.getHotWalletBalances)
	walletRouter.Get("/balances/cold", h.getColdWalletBalances)
	walletRouter.Get("/balances/exchange-settings", h.getWithdrawalWalletBalances)
	walletRouter.Get("/info/:searchParam", h.getWalletInfo)
	walletRouter.Post("/restore/:walletId", h.restoreTx)
	walletRouter.Get("/:searchParam", h.findWalletsWithTransactions)
	walletRouter.Post("/keys/hot", h.getHotWalletKeys)
}
