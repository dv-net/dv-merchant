package handlers

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/dv-net/go-bip39"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/processing_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/processing_response"
	"github.com/dv-net/dv-merchant/internal/delivery/middleware"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/callback"
	_ "github.com/dv-net/dv-merchant/internal/service/processing" // Used by swaggo
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/service/wallet"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/response"
)

// processingCallback is a function to callback webhook from processing
//
//	@Summary		Processing callback
//	@Description	Processing callback
//	@Tags			Processing
//	@Accept			json
//	@Produce		json
//	@Param			register	body		processing_request.ProcessingWebhook	true	"Processing webhook"
//	@Failure		401			{object}	apierror.Errors
//	@Failure		403			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Failure		400			{object}	apierror.Errors
//	@Router			/v1/dv-admin/processing/callback [post]
//	@Security		BearerAuth
func (h Handler) processingCallback(c fiber.Ctx) error {
	// Processing the translation status
	statusCheckRequest := processing_request.TransferStatusWebhook{}
	err := c.Bind().Body(&statusCheckRequest)
	if err == nil && statusCheckRequest.Kind == models.WebhookKindTransferStatus {
		err = h.services.CallbackService.HandleUpdateTransferStatusCallback(c.Context(), statusCheckRequest)
		if err != nil {
			return apierror.New().AddError(errors.New("failed update transfer status")).SetHttpCode(fiber.StatusBadRequest)
		}
		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"success":    true,
			"request_id": statusCheckRequest.RequestID,
		})
	}

	// request binding
	req := &processing_request.ProcessingWebhook{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	var eWalletID *uuid.UUID
	switch req.WalletType {
	case models.WalletTypeHot:
		hotWalletID, err := uuid.Parse(req.ExternalWalletID)
		if err != nil {
			return apierror.New().AddError(errors.New("invalid external wallet id")).SetHttpCode(fiber.StatusBadRequest)
		}

		eWalletID = &hotWalletID
	case models.WalletTypeCold:
		// TODO: Handle cold wallets
		h.logger.Warn("deposit cold wallet not processed")
		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"success": true,
		})
	case models.WalletTypeProcessing:
		if req.Kind == models.WebhookKindDeposit {
			return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
				"success": true,
			})
		}
	}

	var dto interface{}

	switch req.Kind {
	case models.WebhookKindDeposit:
		// deposit
		dto = callback.DepositWebhookDto{
			Blockchain:       req.Blockchain,
			Hash:             req.Hash,
			NetworkCreatedAt: req.NetworkCreatedAt,
			FromAddress:      req.FromAddress,
			ToAddress:        req.ToAddress,
			Amount:           req.Amount,
			Fee:              req.Fee,
			ContractAddress:  req.ContractAddress,
			Status:           req.Status,
			IsSystem:         req.IsSystem,
			Confirmations:    req.Confirmations,
			TxUniqKey:        req.TxUniqKey,
			WalletType:       req.WalletType,
		}
	case models.WebhookKindTransfer:
		// transfer
		dto = callback.TransferWebhookDto{
			Blockchain:       req.Blockchain,
			Hash:             req.Hash,
			TransferID:       req.RequestID,
			NetworkCreatedAt: req.NetworkCreatedAt,
			FromAddress:      req.FromAddress,
			ToAddress:        req.ToAddress,
			Amount:           req.Amount,
			Fee:              req.Fee,
			ContractAddress:  req.ContractAddress,
			Status:           req.Status,
			IsSystem:         req.IsSystem,
			Confirmations:    req.Confirmations,
			WalletType:       req.WalletType,
			TxUniqKey:        req.TxUniqKey,
		}
	default:
		return apierror.New().AddError(errors.New("unknown transaction type")).SetHttpCode(fiber.StatusBadRequest)
	}

	currency, err := h.services.CurrencyService.GetCurrencyByBlockchainAndContract(c.Context(), req.Blockchain, req.ContractAddress)
	if errors.Is(err, pgx.ErrNoRows) {
		// skip unsupported contract
		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"success": true,
			"message": fmt.Sprintf(
				"unknown currency %s: %s (tx: %s, key: %s)",
				req.Blockchain,
				req.ContractAddress,
				req.Hash,
				req.TxUniqKey,
			),
		})
	}

	if err != nil {
		h.logger.Errorw("failed to fetch currency in processing callback", "error", err)
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	switch t := dto.(type) {
	case callback.DepositWebhookDto:
		t.ExternalWalletID = eWalletID
		t.Currency = currency

		err = h.services.CallbackService.HandleDepositCallback(c.Context(), t)
		if err != nil {
			return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
		}
	case callback.TransferWebhookDto:
		t.ExternalWalletID = eWalletID
		t.Currency = currency

		err = h.services.CallbackService.HandleTransferCallback(c.Context(), t)
		if err != nil {
			return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
		}
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success": true,
	})
}

// initProcessing is a function to register backend in processing
//
//	@Summary		Register in processing
//	@Description	Perform processing registration
//	@Tags			Processing
//	@Accept			json
//	@Produce		json
//	@Param			register	body		processing_request.InitializeRequest	true	"Initialize processing"
//	@Success		200			{object}	response.Result[processing_response.InitProcessingResponse]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		403			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Failure		400			{object}	apierror.Errors
//	@Router			/v1/dv-admin/processing/init [post]
//	@Security		BearerAuth
func (h Handler) initProcessing(c fiber.Ctx) error {
	_, err := loadAuthUser(c)
	if err != nil {
		return apierror.New().SetHttpCode(fiber.StatusBadRequest).AddError(err)
	}

	req := &processing_request.InitializeRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}
	callbackURL, err := c.GetRouteURL("processing_callback", nil)
	if err != nil {
		pError := errors.New("callback url route")
		return apierror.New().SetHttpCode(fiber.StatusBadRequest).AddError(pError)
	}

	keySetting, _ := h.services.SettingService.GetRootSetting(c.Context(), setting.ProcessingClientKey)
	if keySetting != nil {
		pError := errors.New("processing already initialized")
		return apierror.New().SetHttpCode(fiber.StatusBadRequest).AddError(pError)
	}

	_, err = h.services.SettingService.SetRootSetting(c.Context(), setting.ProcessingURL, req.Host)
	if err != nil {
		return apierror.New().SetHttpCode(fiber.StatusBadRequest).AddError(err)
	}

	baseURL := c.BaseURL()
	if matched, _ := regexp.MatchString(`^(http|https)://[a-zA-Z0-9.-]+(:\d+)?$`, req.MerchantPayFormDomain); matched {
		baseURL = req.MerchantPayFormDomain
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	_, err = h.services.SettingService.SetRootSetting(c.Context(), setting.MerchantPayFormDomain, req.MerchantDomain)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	if matched, _ := regexp.MatchString(`^(http|https)://[a-zA-Z0-9.-]+(:\d+)?$`, req.CallbackDomain); matched {
		baseURL = req.CallbackDomain
	}

	_, err = h.services.SettingService.SetRootSetting(c.Context(), setting.MerchantDomain, req.MerchantPayFormDomain)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	backendIP, err := h.services.SystemService.GetIP(c.Context())
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	settings, err := h.services.ProcessingClientService.CreateClient(c.Context(), baseURL+callbackURL, &backendIP, &req.MerchantDomain)
	if err != nil {
		pError := fmt.Errorf("creating processing client")
		return apierror.New().SetHttpCode(fiber.StatusBadRequest).AddError(pError)
	}

	res := processing_response.InitProcessingResponse{
		BaseURL:             settings.BaseURL,
		ProcessingClientID:  settings.ProcessingClientID,
		ProcessingClientKey: settings.ProcessingClientKey,
	}

	return c.JSON(response.OkByData(res))
}

// registerOwner register owner in processing
//
//	@Summary		Register root user in processing
//	@Description	Registers a root user in the processing system using provided mnemonic
//	@Tags			Processing
//	@Accept			json
//	@Produce		json
//	@Param			request	body		processing_request.MnemonicRequest	true	"Mnemonic to register root user"
//	@Success		200		{object}	response.Result[processing_response.OwnerProcessingResponse]
//	@Failure		400		{object}	apierror.Errors
//	@Failure		401		{object}	apierror.Errors
//	@Failure		403		{object}	apierror.Errors
//	@Failure		422		{object}	apierror.Errors
//	@Router			/v1/dv-admin/processing/register-owner [post]
//	@Security		BearerAuth
func (h *Handler) registerOwner(c fiber.Ctx) error {
	req := &processing_request.MnemonicRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}
	if !bip39.IsMnemonicValid(req.Mnemonic) {
		return apierror.New().AddError(errors.New("invalid mnemonic")).SetHttpCode(fiber.StatusUnprocessableEntity)
	}

	rootUsr, err := loadAuthUser(c)
	if err != nil {
		return apierror.New().SetHttpCode(fiber.StatusBadRequest).AddError(err)
	}

	if rootUsr.ProcessingOwnerID.Valid {
		return apierror.New().AddError(errors.New("root already registered in processing")).SetHttpCode(fiber.StatusBadRequest)
	}

	pOwnerInfo, err := h.services.UserService.RegisterUserInProcessing(c.Context(), rootUsr, req.Mnemonic)
	if err != nil {
		return apierror.New().SetHttpCode(fiber.StatusBadRequest).AddError(err)
	}

	return c.JSON(response.OkByData(&processing_response.OwnerProcessingResponse{
		OwnerID: pOwnerInfo.OwnerID.String(),
	}))
}

// updateProcessingCallback is a function to update callback URL from processing to backend
//
//	@Summary		Update callback URL from processing to backend
//	@Description	Update callback URL from processing to backend
//	@Tags			Processing
//	@Accept			json
//	@Produce		json
//	@Param			register	body		processing_request.UpdateProcessingCallbackDomain	true	"Chang processing callback domain"
//	@Success		200			{object}	response.Result[string]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		403			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Failure		400			{object}	apierror.Errors
//	@Router			/v1/dv-admin/processing/callback-url [post]
//	@Security		BearerAuth
func (h *Handler) updateProcessingCallback(c fiber.Ctx) error {
	clID, err := h.services.SettingService.GetRootSetting(c.Context(), setting.ProcessingClientID)
	if err != nil {
		return apierror.New().AddError(errors.New("undefined clientID")).SetHttpCode(fiber.StatusBadRequest)
	}
	preparedClID, err := uuid.Parse(clID.Value)
	if err != nil {
		return apierror.New().AddError(errors.New("invalid clientID")).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &processing_request.UpdateProcessingCallbackDomain{}
	if err = c.Bind().Body(req); err != nil {
		return err
	}

	matched, err := regexp.MatchString(`^(http|https)://[a-zA-Z0-9.-]+(:\d+)?$`, req.Domain)
	if !matched || err != nil {
		return apierror.New().AddError(errors.New("invalid domain")).SetHttpCode(fiber.StatusBadRequest)
	}
	preparedDomain := strings.TrimSuffix(req.Domain, "/")
	callbackURL, err := c.GetRouteURL("processing_callback", nil)
	if err != nil {
		pError := errors.New("callback url route")
		return apierror.New().SetHttpCode(fiber.StatusBadRequest).AddError(pError)
	}

	err = h.services.ProcessingClientService.ChangeClient(c.Context(), preparedClID, preparedDomain+callbackURL)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("success"))
}

// updateProcessingCallback is a function to update callback URL from processing to backend
//
//	@Summary		Update callback URL from processing to backend
//	@Description	Update callback URL from processing to backend
//	@Tags			Processing
//	@Produce		json
//	@Success		200	{object}	response.Result[processing_response.CallbackURLResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		403	{object}	apierror.Errors
//	@Failure		422	{object}	apierror.Errors
//	@Failure		400	{object}	apierror.Errors
//	@Router			/v1/dv-admin/processing/callback-url [get]
//	@Security		BearerAuth
func (h *Handler) getProcessingCallback(c fiber.Ctx) error {
	clID, err := h.services.SettingService.GetRootSetting(c.Context(), setting.ProcessingClientID)
	if err != nil {
		return apierror.New().AddError(errors.New("undefined clientID")).SetHttpCode(fiber.StatusBadRequest)
	}
	preparedClID, err := uuid.Parse(clID.Value)
	if err != nil {
		return apierror.New().AddError(errors.New("invalid clientID")).SetHttpCode(fiber.StatusBadRequest)
	}

	callbackURL, err := h.services.ProcessingClientService.GetCallbackURL(c.Context(), preparedClID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(processing_response.CallbackURLResponse{CallbackURL: callbackURL}))
}

// list is a function to get all available processings
//
//	@Summary		Get available processing
//	@Description	Get available processing addresses
//	@Tags			Processing
//	@Produce		json
//	@Success		200	{object}	response.Result[processing_response.ProcessingListResponse]
//	@Failure		422	{object}	apierror.Errors
//	@Failure		400	{object}	apierror.Errors
//	@Router			/v1/dv-admin/processing/list [get]
func (h *Handler) list(c fiber.Ctx) error {
	type MockResponseElem struct {
		IP       string `json:"ip"`
		Type     string `json:"type"`
		Location string `json:"location"`
	}

	resp := []MockResponseElem{
		{"http://127.0.0.1:9000", "aws", "RU"},
		{"http://127.0.0.1:9000", "aws", "DE"},
		{"http://127.0.0.1:9000", "local", "RU"},
	}

	return c.JSON(response.OkByData(resp))
}

// processingWallets gets all processing wallets
//
//	@Summary		Get available processing wallets
//	@Description	Get available processing wallets
//	@Tags			Processing
//	@Produce		json
//	@Success		200	{object}	response.Result[[]wallet.ProcessingWalletWithAssets]
//	@Failure		422	{object}	apierror.Errors
//	@Failure		400	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/dv-admin/processing/wallets [get]
//	@Security		BearerAuth
func (h *Handler) processingWallets(c fiber.Ctx) error {
	u, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	ownerID := u.ProcessingOwnerID
	if !ownerID.Valid {
		return apierror.New().AddError(errors.New("empty owner processing id")).SetHttpCode(fiber.StatusBadRequest)
	}

	dto := wallet.GetProcessingWalletsDTO{
		OwnerID: ownerID.UUID,
	}

	allBlockchains := models.AllBlockchain()

	preparedBlockchains := make([]models.Blockchain, 0, len(allBlockchains))
	if c.Query("blockchains") == "" {
		preparedBlockchains = allBlockchains
	} else {
		blockchains := strings.SplitSeq(c.Query("blockchains"), ",")
		for b := range blockchains {
			preparedBlockchains = append(preparedBlockchains, models.Blockchain(b))
		}
	}

	for _, blockchain := range preparedBlockchains {
		if err := blockchain.Valid(); err != nil {
			return apierror.New().AddError(fmt.Errorf("invalid blockchain: %s", blockchain)).SetHttpCode(fiber.StatusBadRequest)
		}

		if !blockchain.ShowProcessingWallets() {
			continue
		}

		dto.Blockchains = append(dto.Blockchains, blockchain)
	}

	// Fetch wallets with assets using params
	wassets, err := h.services.WalletBalanceService.GetProcessingBalances(c.Context(), dto)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	sort.Slice(wassets, func(i, j int) bool {
		return wassets[i].Address < wassets[j].Address
	})

	return c.JSON(response.OkByData(wassets))
}

func (h Handler) initProcessingRoutes(v1 fiber.Router) {
	processing := v1.Group("/processing")
	processing.Post(
		"/init",
		h.initProcessing,
		middleware.AuthMiddleware(h.services.AuthService),
		middleware.CasbinMiddleware(h.services.PermissionService, []models.UserRole{models.UserRoleRoot}),
	)
	processing.Get(
		"/list",
		h.list,
	)
	processing.Post(
		"/callback-url",
		h.updateProcessingCallback,
		middleware.AuthMiddleware(h.services.AuthService),
		middleware.CasbinMiddleware(h.services.PermissionService, []models.UserRole{models.UserRoleRoot}),
	)
	processing.Get(
		"/callback-url",
		h.getProcessingCallback,
		middleware.AuthMiddleware(h.services.AuthService),
		middleware.CasbinMiddleware(h.services.PermissionService, []models.UserRole{models.UserRoleRoot}),
	)

	processing.Post(
		"/register-owner",
		h.registerOwner,
		middleware.AuthMiddleware(h.services.AuthService),
	)

	processing.Post("/callback", h.processingCallback, middleware.SignMiddleware(h.services.SettingService)).Name("processing_callback")

	processing.Get("/wallets", h.processingWallets,
		middleware.AuthMiddleware(h.services.AuthService),
		middleware.CasbinMiddleware(
			h.services.PermissionService,
			[]models.UserRole{
				models.UserRoleDefault,
				models.UserRoleRoot,
				models.UserRoleSupport,
			},
		),
	)
}
