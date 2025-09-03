package handlers

import (
	"errors"
	"strconv"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/store_api_key_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/store_currency_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/store_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/store_webhook_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/store_whitelist_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/store_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/service/store"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	// Blank import for swaggen
	_ "github.com/dv-net/dv-merchant/internal/storage/repos/repo_stores"

	"github.com/gofiber/fiber/v3"
)

// createStore is a function to create store
//
//	@Summary		Create store
//	@Description	Create new store
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			register	body		store_request.CreateRequest	true	"Create store"
//	@Success		200			{object}	response.Result[store_response.StoreResponse]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Failure		503			{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/ [post]
//	@Security		BearerAuth
func (h Handler) createStore(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	dto := &store_request.CreateRequest{}

	if err := c.Bind().Body(dto); err != nil {
		return err
	}
	createdStore, err := h.services.StoreService.CreateStore(c.Context(), store.CreateStore{
		Name: dto.Name,
		Site: dto.Site,
	}, user)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusUnprocessableEntity)
	}

	res := converters.FromStoreModelToResponse(createdStore)
	return c.JSON(response.OkByData(res))
}

// loadUserStore is a function to get all user store
//
//	@Summary		Get store
//	@Description	Load all user store
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			register	body		store_request.CreateRequest	true	"Create store"
//	@Success		200			{object}	response.Result[[]store_response.StoreResponse]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		503			{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/ [get]
//	@Security		BearerAuth
func (h Handler) loadUserStore(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	stores, err := h.services.StoreService.GetStoresByUserID(c.Context(), user.ID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
	}
	res := converters.FromStoreModelToResponses(stores...)
	return c.JSON(response.OkByData(res))
}

// updateStore is a function to update store
//
//	@Summary		Update store
//	@Description	Update user store
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			register	body		store_request.UpdateRequest	true	"Update store"
//	@Success		200			{object}	response.Result[store_response.StoreResponse]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		503			{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id} [put]
//	@Security		BearerAuth
func (h Handler) updateStore(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}
	dto := &store_request.UpdateRequest{}
	if err := c.Bind().Body(dto); err != nil {
		return err
	}
	updatedStore, err := h.services.StoreService.UpdateStore(c.Context(), dto, targetStore.ID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	res := converters.FromStoreModelToResponse(updatedStore)
	return c.JSON(response.OkByData(res))
}

// loadStoreByID is a function to get store by id
//
//	@Summary		Load store
//	@Description	Load store
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Store ID"
//	@Success		200	{object}	response.Result[store_response.StoreResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		404	{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id} [get]
//	@Security		BearerAuth
func (h Handler) loadStoreByID(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}
	res := converters.FromStoreModelToResponse(targetStore)
	return c.JSON(response.OkByData(res))
}

// storeUnarchive is a function to init archive store
//
//	@Summary		Delete store
//	@Description	Delete store
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Store ID"
//	@Success		200	{object}	response.Result[any]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/unarchive [post]
//	@Security		BearerAuth
func (h Handler) storeUnarchive(c fiber.Ctx) error {
	targetStore, usr, err := h.validateAndLoadStoreWithUser(c)
	if err != nil {
		return err
	}

	dto := new(store_request.StoreArchiveRequest)
	if err = c.Bind().Body(dto); err != nil {
		return err
	}

	if err = h.services.StoreService.UnarchiveStore(c.Context(), store.ArchiveStoreDTO{
		OTP:     dto.OTP,
		User:    usr,
		StoreID: targetStore.ID,
	}); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("success"))
}

// archivedStoresList is a function to get archived stores list
//
//	@Summary		Archived stores list
//	@Description	Archived stores list
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[any]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/archived/list [get]
//	@Security		BearerAuth
func (h Handler) archivedStoresList(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	stores, err := h.services.StoreService.GetArchivedList(c.Context(), usr.ID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.FromStoreModelToResponses(stores...)))
}

// storeArchive is a function to init archive store
//
//	@Summary		Delete store
//	@Description	Delete store
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Store ID"
//	@Success		200	{object}	response.Result[any]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/archive [post]
//	@Security		BearerAuth
func (h Handler) storeArchive(c fiber.Ctx) error {
	targetStore, usr, err := h.validateAndLoadStoreWithUser(c)
	if err != nil {
		return err
	}

	dto := new(store_request.StoreArchiveRequest)
	if err = c.Bind().Body(dto); err != nil {
		return err
	}

	if err = h.services.StoreService.ArchiveStore(c.Context(), store.ArchiveStoreDTO{
		OTP:     dto.OTP,
		User:    usr,
		StoreID: targetStore.ID,
	}); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("success"))
}

// generateNewApiKey is a function to generate new store api key
//
//	@Summary		Create store api key
//	@Description	Create store api key
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Store ID"
//	@Success		200	{object}	response.Result[store_response.StoreAPIKeyResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/apikey [post]
//	@Security		BearerAuth
func (h Handler) generateStoreAPIKey(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	storeAPIKey, err := h.services.StoreAPIKeyService.GenerateAPIKey(c.Context(), targetStore.ID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	go h.services.NotificationService.SendUser(
		c.Context(),
		models.NotificationTypeUserAccessKeyChanged,
		user,
		&notify.UserAccessKeyChangedNotification{
			Email:    user.Email,
			Language: user.Language,
		},
		&models.NotificationArgs{
			UserID:  &user.ID,
			StoreID: &targetStore.ID,
		},
	)

	res := converters.FromStoreAPIKeyModelToResponse(storeAPIKey)
	return c.JSON(response.OkByData(res))
}

// updateStatusStoreAPIKey is a function to update store api key status
//
//	@Summary		Update store api key status
//	@Description	Update store api key status
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string										true	"Store ID"
//	@Param			apiKeyId	path		string										true	"Apikey ID"
//	@Param			register	body		store_api_key_request.UpdateStatusRequest	true	"Update status store api key"
//	@Success		200			{object}	response.Result[store_response.StoreAPIKeyResponse]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		401			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/apikey/{apiKeyId}/status [put]
//	@Security		BearerAuth
func (h Handler) updateStatusStoreAPIKey(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}
	storeAPIKeyID, err := tools.ValidateUUID(c.Params("apiKeyId"))
	if err != nil {
		return err
	}

	dto := &store_api_key_request.UpdateStatusRequest{}
	if err := c.Bind().Body(dto); err != nil {
		return err
	}

	APIKey, err := h.services.StoreAPIKeyService.GetStoreAPIKeyByID(c.Context(), storeAPIKeyID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
	}

	if targetStore.ID != APIKey.StoreID {
		return apierror.New().AddError(errors.New("this is not key for this store")).SetHttpCode(fiber.StatusUnauthorized)
	}

	storeAPIKey, err := h.services.StoreAPIKeyService.UpdateStatusStoreAPIKey(c.Context(), storeAPIKeyID, dto.Status)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
	}

	res := converters.FromStoreAPIKeyModelToResponse(storeAPIKey)
	return c.JSON(response.OkByData(res))
}

// deleteStoreAPIKeys is a function to delete store api key status
//
//	@Summary		Delete store api key status
//	@Description	Delete store api key status
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string	true	"Store ID"
//	@Param			apiKeyId	path		string	true	"Apikey ID"
//	@Success		200			{object}	response.Result[string]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		401			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/apikey/{apiKeyId} [delete]
//	@Security		BearerAuth
func (h Handler) deleteStoreAPIKeys(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}

	storeAPIKeyID, err := tools.ValidateUUID(c.Params("apiKeyId"))
	if err != nil {
		return err
	}

	apiKey, err := h.services.StoreAPIKeyService.GetStoreAPIKeyByID(c.Context(), storeAPIKeyID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
	}

	if targetStore.ID != apiKey.StoreID {
		return apierror.New().AddError(errors.New("this is not key for this store")).SetHttpCode(fiber.StatusUnauthorized)
	}

	err = h.services.StoreAPIKeyService.DeleteAPIKey(c.Context(), storeAPIKeyID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByMessage("Store API key successfully deleted"))
}

// loadStoreAPIKeys is a function to Load store api key
//
//	@Summary		Load store api key
//	@Description	Load store api key
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Store ID"
//	@Success		200	{object}	response.Result[[]store_response.StoreAPIKeyResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/apikey/ [get]
//	@Security		BearerAuth
func (h Handler) loadStoreAPIKeys(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}

	storeAPIKeys, err := h.services.StoreAPIKeyService.GetAPIKeyByStoreID(c.Context(), targetStore.ID)
	if err != nil || len(storeAPIKeys) == 0 {
		return c.JSON(response.OkByData([]string{}))
	}

	res := converters.FromStoreAPIKeyModelToResponses(storeAPIKeys...)
	return c.JSON(response.OkByData(res))
}

// loadStoreWebhooks is a function to Load store webhooks
//
//	@Summary		Load store webhooks
//	@Description	Load store webhooks
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Store ID"
//	@Success		200	{object}	response.Result[[]store_response.StoreWebhookResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/webhooks/ [get]
//	@Security		BearerAuth
func (h Handler) loadStoreWebhooks(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}

	storeWebhooks, err := h.services.StoreWebhooksService.GetStoreWebhookByStoreID(c.Context(), targetStore.ID)
	if err != nil || len(storeWebhooks) == 0 {
		return c.JSON(response.OkByData([]string{}))
	}
	res := converters.FromStoreWebhookModelToResponses(storeWebhooks...)
	return c.JSON(response.OkByData(res))
}

// loadStoreWebhook is a function to Load store webhooks
//
//	@Summary		Load store webhook
//	@Description	Load store webhook
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string	true	"Store ID"
//	@Param			webhookId	path		string	true	"Webhook ID"
//	@Success		200			{object}	response.Result[store_response.StoreWebhookResponse]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/webhooks/{webhookId} [get]
//	@Security		BearerAuth
func (h Handler) loadStoreWebhook(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}

	webhookID, err := tools.ValidateUUID(c.Params("webhookId"))
	if err != nil {
		return err
	}

	webhook, err := h.services.StoreWebhooksService.GetStoreWebhookByID(c.Context(), webhookID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
	}
	if targetStore.ID != webhook.StoreID {
		return apierror.New().AddError(errors.New("this webhook does not belong to this store")).SetHttpCode(fiber.StatusUnauthorized)
	}

	res := converters.FromStoreWebhookModelToResponse(webhook)
	return c.JSON(response.OkByData(res))
}

// createStoreWebhooks is a function to Create store webhooks
//
//	@Summary		Create store webhooks
//	@Description	Create store webhooks
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string								true	"Store ID"
//	@Param			register	body		store_webhook_request.CreateRequest	true	"Create store webhook"
//	@Success		200			{object}	response.Result[store_response.StoreWebhookResponse]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/webhooks/ [post]
//	@Security		BearerAuth
func (h Handler) createStoreWebhooks(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}

	dto := &store_webhook_request.CreateRequest{}
	if err := c.Bind().Body(dto); err != nil {
		return err
	}

	webhook, err := h.services.StoreWebhooksService.CreateStoreWebhooks(c.Context(), targetStore, dto)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.FromStoreWebhookModelToResponse(webhook)))
}

// updateStoreWebhooks is a function to Update store webhooks
//
//	@Summary		Update store webhooks
//	@Description	Update store webhooks
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string								true	"Store ID"
//	@Param			webhookId	path		string								true	"Webhook ID"
//	@Param			register	body		store_webhook_request.UpdateRequest	true	"Create store webhook"
//	@Success		200			{object}	response.Result[store_response.StoreWebhookResponse]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/webhooks/{webhookId} [put]
//	@Security		BearerAuth
func (h Handler) updateStoreWebhooks(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}

	webhookID, err := tools.ValidateUUID(c.Params("webhookId"))
	if err != nil {
		return err
	}

	dto := &store_webhook_request.UpdateRequest{}
	if err := c.Bind().Body(dto); err != nil {
		return err
	}

	webhook, err := h.services.StoreWebhooksService.GetStoreWebhookByID(c.Context(), webhookID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
	}
	if targetStore.ID != webhook.StoreID {
		return apierror.New().AddError(errors.New("this is webhook for this store")).SetHttpCode(fiber.StatusUnauthorized)
	}

	webhook, err = h.services.StoreWebhooksService.UpdateStoreWebhooks(c.Context(), webhookID, dto)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	res := converters.FromStoreWebhookModelToResponse(webhook)
	return c.JSON(response.OkByData(res))
}

// deleteStoreWebhooks is a function to Delete store webhooks
//
//	@Summary		Delete store webhooks
//	@Description	Delete store webhooks
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string	true	"Store ID"
//	@Param			webhookId	path		string	true	"Webhook ID"
//	@Success		200			{object}	response.Result[any]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		500			{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/webhooks/{webhookId} [delete]
//	@Security		BearerAuth
func (h Handler) deleteStoreWebhooks(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}

	webhookID, err := tools.ValidateUUID(c.Params("webhookId"))
	if err != nil {
		return err
	}

	webhook, err := h.services.StoreWebhooksService.GetStoreWebhookByID(c.Context(), webhookID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
	}
	if targetStore.ID != webhook.StoreID {
		return apierror.New().AddError(errors.New("this webhook does not belong to this store")).SetHttpCode(fiber.StatusUnauthorized)
	}

	err = h.services.StoreWebhooksService.DeleteStoreWebhooks(c.Context(), webhookID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByMessage("Store webhook successfully deleted"))
}

// loadStoreTransaction is a function to Load store transactions
//
//	@Summary		Load store transactions
//	@Description	Load store transactions
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"Store ID"
//	@Param			page	query		int		false	"page number"
//	@Success		200		{object}	response.Result[[]store_response.StoreTransactionResponse]
//	@Failure		401		{object}	apierror.Errors
//	@Failure		404		{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/transactions/ [get]
//	@Security		BearerAuth
func (h Handler) loadStoreTransaction(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	pageParam := c.Query("page", "1")
	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = 1
	}
	storeID, err := tools.ValidateUUID(c.Params("id"))
	if err != nil {
		return err
	}
	targetStore, err := h.services.StoreService.GetStoreByID(c.Context(), storeID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
	}
	if user.ID != targetStore.UserID {
		return apierror.New().AddError(errors.New("this is not your store")).SetHttpCode(fiber.StatusUnauthorized)
	}

	storeTransactions, err := h.services.TransactionService.GetStoreTransactions(c.Context(), targetStore.ID, int32(page)) // #nosec
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
	}

	res := converters.FromStoreTransactionModelToResponses(storeTransactions...)
	return c.JSON(response.OkByData(res))
}

// loadStoreCurrency is a function to Load store currency
//
//	@Summary		Load store currency
//	@Description	Load store currency
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Store ID"
//	@Success		200	{object}	response.Result[any]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/currencies/ [get]
//	@Security		BearerAuth
func (h Handler) loadStoreCurrencies(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}

	storeCurrencies, err := h.services.StoreCurrencyService.GetCurrenciesByStoreID(c.Context(), targetStore.ID)
	if err != nil || storeCurrencies == nil {
		return c.JSON(response.OkByData([]string{}))
	}
	res := make([]string, 0, len(storeCurrencies))
	for _, currency := range storeCurrencies {
		res = append(res, currency.CurrencyID)
	}
	return c.JSON(response.OkByData(res))
}

// updateStoreCurrency is a function to Update store currency
//
//	@Summary		Update store currency
//	@Description	Update store currency
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string									true	"Store ID"
//	@Param			register	body		store_currency_request.UpdateRequest	true	"Update store currency"
//	@Success		200			{object}	response.Result[string]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		400			{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/currencies/ [put]
//	@Security		BearerAuth
func (h Handler) updateStoreCurrency(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}

	req := &store_currency_request.UpdateRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	dto := &store.UpdateStoreCurrencyDTO{CurrencyIDs: req.CurrencyIDs}
	err = h.services.StoreCurrencyService.UpdateStoreCurrency(c.Context(), targetStore, dto)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByMessage("Store currency successfully updated"))
}

// loadStoreWhitelist is a function to Load store whitelist
//
//	@Summary		Load store whitelist
//	@Description	Load store whitelist
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Store ID"
//	@Success		200	{object}	response.Result[store_response.StoreWhitelistResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		400	{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/whitelists/ [get]
//	@Security		BearerAuth
func (h Handler) loadStoreWhitelist(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}

	storeWhitelist, err := h.services.StoreWhitelistService.GetStoreWhitelist(c.Context(), targetStore.ID)
	if err != nil || storeWhitelist == nil {
		return c.JSON(response.OkByData([]string{}))
	}

	return c.JSON(response.OkByData(converters.FromStoreWhitelistModelToResponses(storeWhitelist...)))
}

// updateStoreWhitelist is a function to Update store whitelist
//
//	@Summary		Update store whitelist
//	@Description	Update store whitelist
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string									true	"Store ID"
//	@Param			register	body		store_whitelist_request.UpdateRequest	true	"Update store whitelist"
//	@Success		200			{object}	response.Result[any]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		400			{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/whitelists/ [put]
//	@Security		BearerAuth
func (h Handler) updateStoreWhitelist(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}

	dto := &store_whitelist_request.UpdateRequest{}
	if err := c.Bind().Body(dto); err != nil {
		return err
	}

	storeWhitelists, err := h.services.StoreWhitelistService.CreateStoreWhitelist(c.Context(), targetStore.ID, dto.Ips)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	res := converters.FromStoreWhitelistModelToResponses(storeWhitelists...)
	return c.JSON(response.OkByData(res))
}

func (h Handler) validateAndLoadStore(c fiber.Ctx) (*models.Store, error) {
	user, err := loadAuthUser(c)
	if err != nil {
		return nil, err
	}
	storeID, err := tools.ValidateUUID(c.Params("id"))
	if err != nil {
		return nil, err
	}

	targetStore, err := h.services.StoreService.GetStoreByID(c.Context(), storeID)
	if err != nil {
		return nil, apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
	}
	if user.ID != targetStore.UserID {
		return nil, apierror.New().AddError(errors.New("this is not your store")).SetHttpCode(fiber.StatusUnauthorized)
	}
	return targetStore, nil
}

func (h Handler) validateAndLoadStoreWithUser(c fiber.Ctx) (*models.Store, *models.User, error) {
	user, err := loadAuthUser(c)
	if err != nil {
		return nil, nil, err
	}
	storeID, err := tools.ValidateUUID(c.Params("id"))
	if err != nil {
		return nil, nil, err
	}
	targetStore, err := h.services.StoreService.GetStoreByID(c.Context(), storeID)
	if err != nil {
		return nil, nil, apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
	}
	if user.ID != targetStore.UserID {
		return nil, nil, apierror.New().AddError(errors.New("this is not your store")).SetHttpCode(fiber.StatusUnauthorized)
	}
	return targetStore, user, nil
}

// patchStoreWhitelist is a function to Update store whitelist
//
//	@Summary		Patch store whitelist
//	@Description	Patch store whitelist
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id			path		uuid									true	"Store ID"
//	@Param			register	body		store_whitelist_request.PatchRequest	true	"Patch store whitelist"
//	@Success		200			{object}	response.Result[store_response.StoreWhitelistResponse]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		400			{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/whitelists/ [patch]
//	@Security		BearerAuth
func (h Handler) patchStoreWhitelist(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}

	dto := &store_whitelist_request.PatchRequest{}
	if err = c.Bind().Body(dto); err != nil {
		return err
	}

	storeWhitelists, err := h.services.StoreWhitelistService.PatchStoreWhitelist(c.Context(), targetStore.ID, dto.IP)
	if err != nil {
		return apierror.New().AddError(errors.New("ip already exists in whitelist")).SetHttpCode(fiber.StatusBadRequest)
	}

	res := converters.FromStoreWhitelistModelToResponses(storeWhitelists...)
	return c.JSON(response.OkByData(res))
}

// deleteStoreWhitelist is a function to Update store whitelist
//
//	@Summary		Delete store whitelist
//	@Description	Delete store whitelist
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Store ID"
//	@Param			ip	path		string	true	"Whitelist IP"
//	@Success		200	{object}	response.Result[any]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		400	{object}	apierror.Errors
//	@Router			/v1/store/{id}/whitelists/{ip} [delete]
//	@Security		BearerAuth
func (h Handler) deleteStoreWhitelist(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}
	ip := c.Params("ip")
	if ip == "" {
		return apierror.New().AddError(errors.New("ip is required")).SetHttpCode(fiber.StatusBadRequest)
	}

	err = h.services.StoreWhitelistService.DeleteStoreWhitelistByIP(c.Context(), targetStore.ID, ip)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByMessage("Store whitelist successfully deleted"))
}

// deleteStoreSecret is a function to remove store secret key
//
//	@Summary		Remove store webhook secret
//	@Description	Remove store webhook secret
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Store ID"
//	@Success		200	{object}	string
//	@Failure		401	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/secret [delete]
//	@Security		BearerAuth
func (h *Handler) deleteStoreSecret(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}

	if err = h.services.StoreSecretService.RemoveSecretByStore(c.Context(), targetStore.ID); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("store secret successfully deleted"))
}

// getStoreSecret is a function to fetch store secret for webhook sign
//
//	@Summary		Fetch store webhook secret
//	@Description	Fetch store webhook secret
//	@Tags			Store
//	@Produce		json
//	@Param			id	path		string	true	"Store ID"
//	@Success		200	{object}	response.Result[store_response.StoreSecretResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/secret [get]
//	@Security		BearerAuth
func (h *Handler) getStoreSecret(c fiber.Ctx) error {
	st, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}

	secret, err := h.services.StoreSecretService.GetSecretByStore(c.Context(), st.ID)
	if errors.Is(err, store.ErrStoreSecretNotFound) {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusNotFound)
	}
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(store_response.StoreSecretResponse{Secret: secret}))
}

// generateStoreSecret is a function to generate store secret for webhook sign
//
//	@Summary		Generate store webhook secret
//	@Description	Generate store webhook secret
//	@Tags			Store
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Store ID"
//	@Success		200	{object}	response.Result[store_response.StoreSecretResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/dv-admin/store/{id}/secret [post]
//	@Security		BearerAuth
func (h Handler) generateStoreSecret(c fiber.Ctx) error {
	st, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	secret, err := h.services.StoreSecretService.GenerateNewSecret(c.Context(), st.ID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	go h.services.NotificationService.SendUser(
		c.Context(),
		models.NotificationTypeUserAccessKeyChanged,
		user,
		&notify.UserAccessKeyChangedNotification{
			Email:    user.Email,
			Language: user.Language,
		},
		&models.NotificationArgs{
			UserID:  &user.ID,
			StoreID: &st.ID,
		},
	)

	return c.JSON(response.OkByData(store_response.StoreSecretResponse{Secret: secret}))
}

func (h *Handler) initStoreRoutes(v1 fiber.Router) {
	storeHandlers := v1.Group("/store")
	storeHandlers.Post("/", h.createStore)
	storeHandlers.Get("/", h.loadUserStore)
	storeHandlers.Get("/archived", h.archivedStoresList)
	storeHandlers.Put("/:id", h.updateStore)
	storeHandlers.Get("/:id", h.loadStoreByID)
	storeHandlers.Post("/:id/archive", h.storeArchive)
	storeHandlers.Post("/:id/unarchive", h.storeUnarchive)
	storeHandlers.Post("/:id/apikey", h.generateStoreAPIKey)
	storeHandlers.Get("/:id/apikey", h.loadStoreAPIKeys)
	storeHandlers.Put("/:id/apikey/:apiKeyId/status", h.updateStatusStoreAPIKey)
	storeHandlers.Delete("/:id/apikey/:apiKeyId", h.deleteStoreAPIKeys)
	storeHandlers.Post("/:id/secret", h.generateStoreSecret)
	storeHandlers.Get("/:id/secret", h.getStoreSecret)
	storeHandlers.Delete("/:id/secret", h.deleteStoreSecret)
	storeHandlers.Get("/:id/webhooks", h.loadStoreWebhooks)
	storeHandlers.Get("/:id/webhooks/:webhookId", h.loadStoreWebhook)
	storeHandlers.Post("/:id/webhooks", h.createStoreWebhooks)
	storeHandlers.Put("/:id/webhooks/:webhookId", h.updateStoreWebhooks)
	storeHandlers.Delete("/:id/webhooks/:webhookId", h.deleteStoreWebhooks)
	storeHandlers.Get("/:id/transactions", h.loadStoreTransaction)
	storeHandlers.Get("/:id/currencies", h.loadStoreCurrencies)
	storeHandlers.Put("/:id/currencies", h.updateStoreCurrency)
	storeHandlers.Get("/:id/whitelists", h.loadStoreWhitelist)
	storeHandlers.Put("/:id/whitelists", h.updateStoreWhitelist)
	storeHandlers.Patch("/:id/whitelists", h.patchStoreWhitelist)
	storeHandlers.Delete("/:id/whitelists/:ip", h.deleteStoreWhitelist)
}
