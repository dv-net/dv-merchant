package handlers

import (
	"errors"
	"fmt"
	"net/http"

	_ "github.com/dv-net/dv-merchant/internal/delivery/http/responses/exchange_response" // Blank import for swaggen
	_ "github.com/dv-net/dv-merchant/internal/storage/storecmn"                          // Blank import for swaggen

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/exchange_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// @Summary		List exchanges
// @Description	Get all available exchanges with keys
// @Tags			Exchange
// @Accept			json
// @Produce		json
// @Success		200	{object}	response.Result[exchange_response.ExchangeListResponse]
// @Failure		422	{object}	apierror.Errors
// @Failure		503	{object}	apierror.Errors
// @Router			/v1/dv-admin/exchange/list [get]
// @Security		Bearer
func (h *Handler) exchangesList(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	res, err := h.services.ExchangeService.GetAvailableExchangesList(c.Context(), usr.ID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.NewExchangeListResponseFromDto(res)))
}

// @Summary		Set current exchange
// @Description	Set current exchange
// @Tags			Exchange
// @Accept			json
// @Produce		json
// @Success		200	{object}	response.Result[string]
// @Failure		422	{object}	apierror.Errors
// @Failure		503	{object}	apierror.Errors
// @Router			/v1/dv-admin/exchange/:slug/set [post]
// @Security		Bearer
func (h *Handler) setExchange(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}

	err = h.services.ExchangeService.SetCurrentExchange(c.Context(), usr.ID, slug)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}
	return c.JSON(response.OkByMessage("Current exchange successfully set"))
}

// @Summary		Update exchange keys
// @Description	Update or create exchange user keys
// @Tags			Exchange
// @Accept			json
// @Produce		json
// @Param			exchange_slug	path		string						true	"Exchange slug"
// @Param			update_keys		body		exchange_request.UpdateKeys	true	"Update exchange keys"
// @Success		200				{object}	response.Result[string]
// @Failure		422				{object}	apierror.Errors
// @Failure		503				{object}	apierror.Errors
// @Router			/v1/dv-admin/exchange/:slug/keys [post]
func (h *Handler) updateKeys(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &exchange_request.UpdateKeys{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}

	err = h.services.ExchangeService.SetExchangeKeys(c.Context(), usr.ID, slug, req.ToMap())
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("Exchange keys successfully updated"))
}

// @Summary		Delete exchange keys
// @Description	Delete exchange user keys
// @Tags			Exchange
// @Accept			json
// @Produce		json
// @Param			exchange_slug	path		string	true	"Exchange slug"
// @Success		200				{object}	response.Result[string]
// @Failure		422				{object}	apierror.Errors
// @Failure		503				{object}	apierror.Errors
// @Router			/v1/dv-admin/exchange/:slug/keys [delete]
func (h *Handler) deleteKeys(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}

	err = h.services.ExchangeService.DeleteExchangeKeys(c.Context(), usr.ID, slug)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("Exchange keys successfully deleted"))
}

// @Summary		Test exchange
// @Description	Route for ensure exchange correct set up
// @Tags			Exchange
// @Produce		json
// @Success		200	{object}	response.Result[string]
// @Failure		422	{object}	apierror.Errors
// @Failure		503	{object}	apierror.Errors
// @Router			/v1/dv-admin/exchange/:slug/test [get]
func (h *Handler) testConnection(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}

	err = h.services.ExchangeService.TestConnection(c.Context(), *usr, slug)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("ok"))
}

// @Summary		Test exchange API connection
// @Description	Test exchange API connection with provided credentials
// @Tags			Exchange
// @Accept			json
// @Produce		json
// @Param			test_connection	body		exchange_request.TestConnectionRequest	true	"Test connection request"
// @Success		200				{object}	response.Result[string]
// @Failure		422				{object}	apierror.Errors
// @Failure		503				{object}	apierror.Errors
// @Router			/v1/dv-admin/exchange/test [post]
// @Security		Bearer
func (h *Handler) testConnectionExternal(c fiber.Ctx) error {
	if _, err := loadAuthUser(c); err != nil {
		return err
	}

	req := &exchange_request.TestConnectionRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	slug := models.ExchangeSlug(req.Slug)
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}

	err := h.services.ExchangeService.TestConnectionRaw(c.Context(), slug, req.Credentials.Key, req.Credentials.Secret, req.Credentials.Passphrase)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("ok"))
}

// @Summary		Get exchange balances
// @Description	Get exchange balances
// @Tags			Exchange
// @Produce		json
// @Success		200	{object}	response.Result[exchange_response.ExchangeBalanceResponse]
// @Failure		422	{object}	apierror.Errors
// @Failure		503	{object}	apierror.Errors
// @Router			/v1/dv-admin/exchange/:slug/balance [get]
// @Security		Bearer
func (h *Handler) getBalance(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}

	balances, err := h.services.ExchangeService.GetExchangeBalance(c.Context(), slug, *usr)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}
	return c.JSON(response.OkByData(converters.FromExchangeBalanceModelToResponse(balances)))
}

// @Summary		Update user exchange pairs
// @Description	Update user exchange pairs
// @Tags			Exchange
// @Accept			json
// @Produce		json
// @Param			update_pairs	body		exchange_request.UpdateExchangePairsRequest	true	"Update exchange pairs"
// @Success		200				{object}	response.Result[string]
// @Failure		422				{object}	apierror.Errors
// @Failure		503				{object}	apierror.Errors
// @Router			/v1/dv-admin/exchange/:slug/user-pairs [put]
// @Security		Bearer
func (h *Handler) updateUserExchangePairs(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &exchange_request.UpdateExchangePairsRequest{}
	err = c.Bind().Body(req)
	if err != nil {
		return err
	}

	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}

	err = h.services.ExchangeService.UpdateUserExchangePairs(c.Context(), usr.ID, slug, req)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("Exchange pairs successfully updated"))
}

// @Summary		Get user exchange pairs
// @Description	Get user exchange pairs
// @Tags			Exchange
// @Produce		json
// @Success		200	{object}	response.Result[exchange_response.ExchangeUserPairResponse]
// @Failure		422	{object}	apierror.Errors
// @Failure		503	{object}	apierror.Errors
// @Router			/v1/dv-admin/exchange/:slug/user-pairs [get]
// @Security		Bearer
func (h *Handler) getUserExchangePairs(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}

	pairs, err := h.services.ExchangeService.GetUserExchangePairs(c.Context(), usr.ID, slug)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.GetUserExchangePairsResponse(pairs)))
}

// @Summary		Get exchange pairs
// @Description	Get exchange pairs
// @Tags			Exchange
// @Produce		json
// @Success		200	{object}	response.Result[[]models.ExchangeSymbolDTO]
// @Failure		422	{object}	apierror.Errors
// @Failure		503	{object}	apierror.Errors
// @Router			/v1/exchange/:slug/pairs [get]
// @Security		Bearer
func (h *Handler) getExchangePairs(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}

	pairs, err := h.services.ExchangeService.GetExchangePairs(c.Context(), usr.ID, slug)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByData(pairs))
}

// @Summary		Update deposit addresses
// @Description	Update deposit addresses
// @Tags			Exchange
// @Produce		json
// @Success		200	{object}	response.Result[[]exchange_response.DepositUpdateResponse]
// @Failure		422	{object}	apierror.Errors
// @Failure		503	{object}	apierror.Errors
// @Router			/v1/exchange/:exchange_slug/deposit-update [get]
// @Security		Bearer
func (h *Handler) updateDepositAddresses(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}

	addr, err := h.services.ExchangeService.UpdateDepositAddresses(c.Context(), usr.ID, slug)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	rules, err := h.services.ExchangeRulesService.GetWithdrawalRules(c.Context(), slug, usr.ID.String())
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.UpdateDepositAddressesResponse(addr, rules)))
}

// @Summary		Get exchange withdrawal rules
// @Description	Get exchange withdrawal rules for all enabled currencies
// @Tags			Exchange
// @Produce		json
// @Success		200	{object}	response.Result[[][]exchange_response.ExchangeWithdrawalRulesResponse]
// @Failure		422	{object}	apierror.Errors
// @Failure		503	{object}	apierror.Errors
// @Router			/v1/exchange/:slug/withdrawal-rules [get]
// @Security		Bearer
func (h *Handler) getExchangeWithdrawalRules(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}

	rules, err := h.services.ExchangeService.GetExchangeWithdrawalRules(c.Context(), usr.ID, slug)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}
	return c.JSON(response.OkByData(converters.GetWithdrawalRulesResponse(rules)))
}

// @Summary		Get exchange deposit address
// @Description	Get exchange deposit address
// @Tags			Exchange
// @Produce		json
// @Success		200	{object}	response.Result[[]exchange_response.GetDepositAddressesResponse]
// @Failure		422	{object}	apierror.Errors
// @Failure		503	{object}	apierror.Errors
// @Router			/v1/exchange/deposit-addresses [get]
// @Security		Bearer
func (h *Handler) getExchangeDepositAddress(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	address, err := h.services.ExchangeService.GetDepositExchangeAddresses(c.Context(), usr.ID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	rules, err := h.services.ExchangeRulesService.GetAllWithdrawalRules(c.Context(), usr.ID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.GetDepositAddressesResponse(address, rules)))
}

// @Summary		Create withdrawal setting
// @Description	Creates withdrawal setting with minimum value in USD to withdrawal from exchange
// @Tags			Exchange
// @Accept			json
// @Produce		json
// @Param			withdrawal	body		exchange_request.CreateWithdrawalSettingRequest	true	"Create withdrawal"
// @Success		200			{object}	response.Result[string]
// @Failure		422			{object}	apierror.Errors
// @Failure		503			{object}	apierror.Errors
// @Router			/v1/exchange/:exchange_slug/withdrawal-setting [post]
// @Security		Bearer
func (h *Handler) createWithdrawalSetting(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}
	req := &exchange_request.CreateWithdrawalSettingRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	setting, err := h.services.ExchangeWithdrawalService.CreateWithdrawalSetting(c.Context(), usr.ID, slug, req)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}
	return c.JSON(response.OkByData(converters.GetWithdrawalSettingResponse(setting)))
}

// @Summary	Get all withdrawal settings
// @Tags		Exchange
// @Accept		json
// @Produce	json
// @Param		exchange_slug	path		string	true	"Exchange slug"
// @Success	200				{object}	response.Result[[]exchange_response.ExchangeWithdrawalSettingResponse]
// @Failure	422				{object}	apierror.Errors
// @Failure	503				{object}	apierror.Errors
// @Router		/v1/exchange/:exchange_slug/withdrawal-setting [get]
// @Security	Bearer
func (h *Handler) getWithdrawalSettings(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}

	settings, err := h.services.ExchangeWithdrawalService.GetWithdrawalSettings(c.Context(), usr.ID, slug)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}
	return c.JSON(response.OkByData(converters.GetWithdrawalSettingsResponse(settings)))
}

// @Summary	Get all withdrawal setting by id
// @Tags		Exchange
// @Accept		json
// @Produce	json
// @Param		exchange_slug	path		string	true	"Exchange slug"
// @Param		id				path		string	true	"Setting UUID"
// @Success	200				{object}	response.Result[exchange_response.ExchangeWithdrawalSettingResponse]
// @Failure	422				{object}	apierror.Errors
// @Failure	503				{object}	apierror.Errors
// @Router		/v1/exchange/{exchange_slug}/withdrawal-setting/{id} [get]
// @Security	Bearer
func (h *Handler) getWithdrawalSetting(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}
	id := c.Params("id")
	settingID, err := uuid.Parse(id)
	if err != nil {
		return apierror.New().AddError(errors.New("invalid setting uuid")).SetHttpCode(http.StatusBadRequest)
	}

	setting, err := h.services.ExchangeWithdrawalService.GetWithdrawalSetting(c.Context(), usr.ID, slug, settingID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}
	return c.JSON(response.OkByData(converters.GetWithdrawalSettingResponse(setting)))
}

// @Summary	Update withdrawal setting by id
// @Tags		Exchange
// @Accept		json
// @Produce	json
// @Param		exchange_slug	path		string										true	"Exchange slug"
// @Param		id				path		string										true	"Setting UUID"
// @Param		update_keys		body		exchange_request.UpdateWithdrawalSetting	true	"Update exchange settings"
// @Success	200				{object}	response.Result[exchange_response.ExchangeWithdrawalSettingResponse]
// @Failure	422				{object}	apierror.Errors
// @Failure	503				{object}	apierror.Errors
// @Router		/v1/exchange/{exchange_slug}/withdrawal-setting/{id} [patch]
// @Security	Bearer
func (h *Handler) updateWithdrawalSetting(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}

	id := c.Params("id")
	settingID, err := uuid.Parse(id)
	if err != nil {
		return apierror.New().AddError(errors.New("invalid setting uuid")).SetHttpCode(http.StatusBadRequest)
	}

	req := &exchange_request.UpdateWithdrawalSetting{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	setting, err := h.services.ExchangeWithdrawalService.UpdateWithdrawalSetting(c.Context(), usr.ID, settingID, req.Enabled)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.GetWithdrawalSettingResponse(setting)))
}

// @Summary	Delete withdrawal setting by id
// @Tags		Exchange
// @Accept		json
// @Produce	json
// @Param		exchange_slug	path		string	true	"Exchange slug"
// @Param		id				path		string	true	"Setting UUID"
// @Success	200				{object}	response.Result[string]
// @Failure	422				{object}	apierror.Errors
// @Failure	503				{object}	apierror.Errors
// @Router		/v1/exchange/{exchange_slug}/withdrawal-setting/{id} [delete]
// @Security	Bearer
func (h *Handler) deleteWithdrawalSetting(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}
	id := c.Params("id")
	settingID, err := uuid.Parse(id)
	if err != nil {
		return apierror.New().AddError(errors.New("invalid setting uuid")).SetHttpCode(http.StatusBadRequest)
	}

	err = h.services.ExchangeWithdrawalService.DeleteWithdrawalSetting(c.Context(), usr.ID, slug, settingID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}
	return c.JSON(response.OkByMessage("Withdrawal setting successfully deleted"))
}

// @Summary		Get withdrawals
// @Description	Get withdrawals
// @Tags			Exchange
// @Produce		json
// @Param			search	body		exchange_request.GetWithdrawalsRequest	true	"Get withdrawal request"
// @Success		200		{object}	response.Result[storecmn.FindResponseWithFullPagination[exchange_response.ExchangeWithdrawalHistoryResponse]]
// @Failure		422		{object}	apierror.Errors
// @Failure		503		{object}	apierror.Errors
// @Router			/v1/exchange/withdrawal [get]
// @Security		Bearer
func (h *Handler) getWithdrawals(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	req := &exchange_request.GetWithdrawalsRequest{}
	if err := c.Bind().Query(req); err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}
	if req.Slug != nil {
		slug := models.ExchangeSlug(*req.Slug)
		if !slug.Valid() {
			return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
		}
	}
	dto, err := h.services.ExchangeWithdrawalService.GetWithdrawalHistory(c.Context(), usr.ID, req)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}
	return c.JSON(response.OkByData(converters.GetWithdrawalsHistoryResponse(dto)))
}

// @Summary		Download withdrawals
// @Description	Download withdrawals
// @Tags			Exchange
// @Produce		json
// @Param			search	body		exchange_request.GetWithdrawalsExportedRequest	true	"Get withdrawal exported request"
// @Failure		422		{object}	apierror.Errors
// @Failure		503		{object}	apierror.Errors
// @Router			/v1/exchange/withdrawal/export [get]
// @Security		Bearer
func (h *Handler) downloadWithdrawals(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	req := &exchange_request.GetWithdrawalsExportedRequest{}
	if err := c.Bind().Query(req); err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}
	if req.Slug != nil {
		slug := models.ExchangeSlug(*req.Slug)
		if !slug.Valid() {
			return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
		}
	}
	history, err := h.services.ExchangeWithdrawalService.DownloadWithdrawalHistory(c.Context(), usr.ID, req)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	c.Response().Header.Set("Content-Type", "application/octet-stream")
	c.Response().Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.%s", "export_exchange_withdrawals", req.Format))
	return c.SendStream(history, history.Len())
}

// @Summary		Get withdrawal by ID
// @Description	Get withdrawal by ID
// @Tags			Exchange
// @Produce		json
// @Param			id				path		string	true	"Withdrawal UUID"
// @Param			exchange_slug	path		string	true	"Exchange slug"
// @Success		200				{object}	response.Result[exchange_response.ExchangeWithdrawalHistoryResponse]
// @Failure		422				{object}	apierror.Errors
// @Failure		503				{object}	apierror.Errors
// @Router			/v1/exchange/{exchange_slug}/withdrawal/{id} [get]
func (h *Handler) getWithdrawalByID(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}
	id := c.Params("id")
	orderID, err := uuid.Parse(id)
	if err != nil {
		return apierror.New().AddError(errors.New("invalid order uuid")).SetHttpCode(http.StatusBadRequest)
	}
	dto, err := h.services.ExchangeWithdrawalService.GetWithdrawalByID(c.Context(), usr.ID, slug, orderID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}
	return c.JSON(response.OkByData(converters.GetWithdrawalHistoryResponse(dto)))
}

// @Summary		Get exchange chains
// @Description	Get exchange chains
// @Tags			Exchange
// @Produce		json
// @Success		200	{object}	response.Result[[]models.ExchangeChainShort]
// @Failure		422	{object}	apierror.Errors
// @Failure		503	{object}	apierror.Errors
// @Router			/v1/exchange/{exchange_slug}/chains [get]
func (h *Handler) getExchangeChains(c fiber.Ctx) error {
	_, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}

	chains, err := h.services.ExchangeService.GetExchangeChains(c.Context(), slug)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByData(chains))
}

// @Summary		Get user exchange orders history
// @Description	Get user exchange orders history
// @Tags			Exchange
// @Produce		json
// @Success		200	{object}	response.Result[storecmn.FindResponseWithFullPagination[exchange_response.ExchangeOrderHistoryResponse]]
// @Failure		422	{object}	apierror.Errors
// @Failure		503	{object}	apierror.Errors
// @Router			/v1/exchange/exchange-history [get]
// @Security		Bearer
func (h *Handler) getUserOrderHistory(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	request := &exchange_request.GetExchangeOrdersHistoryRequest{}
	if err := c.Bind().Query(request); err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}
	if request.Slug != nil {
		slug := models.ExchangeSlug(*request.Slug)
		if !slug.Valid() {
			return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
		}
	}

	history, err := h.services.ExchangeService.GetExchangeOrdersHistory(c.Context(), usr.ID, *request)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.GetOrdersHistoryResponse(history)))
}

// @Summary		Download user exchange orders history
// @Description	Download user exchange orders history
// @Tags			Exchange
// @Produce		json
// @Failure		422		{object}	apierror.Errors
// @Failure		503		{object}	apierror.Errors
// @Param			string	query		exchange_request.GetExchangeOrdersHistoryExportedRequest	true	"GetExchangeOrdersHistoryExportedRequest"
// @Router			/v1/exchange/exchange-history/export [get]
// @Security		Bearer
func (h *Handler) downloadUserOrderHistory(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	request := &exchange_request.GetExchangeOrdersHistoryExportedRequest{}
	if err := c.Bind().Query(request); err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}
	if request.Slug != nil {
		slug := models.ExchangeSlug(*request.Slug)
		if !slug.Valid() {
			return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
		}
	}
	history, err := h.services.ExchangeService.DownloadExchangeOrdersHistory(c.Context(), usr.ID, *request)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	c.Set("Content-Type", "application/octet-stream")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.%s", "user_exchange_history", request.Format))

	return c.SendStream(history, history.Len())
}

// @Summary		Toggle exchange withdrawals
// @Description	Toggle exchange withdrawals
// @Tags			Exchange
// @Produce		json
// @Param			exchange_slug	path		string												true	"Exchange slug"
// @Param			state			body		exchange_request.ToggleExchangeWithdrawalsRequest	true	"Toggle withdrawals request"
// @Success		200				{object}	response.Result[string]
// @Failure		422				{object}	apierror.Errors
// @Failure		503				{object}	apierror.Errors
// @Router			/v1/exchange/{exchange_slug}/toggle-withdrawals [post]
// @Security		Bearer
func (h *Handler) toggleExchangeWithdrawals(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}
	req := &exchange_request.ToggleExchangeWithdrawalsRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	state, err := h.services.ExchangeService.ChangeExchangeWithdrawalState(c.Context(), slug, usr.ID, models.ExchangeWithdrawalState(req.NewState))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByData(state.WithdrawalState))
}

// @Summary		Toggle exchange swaps
// @Description	Toggle exchange swaps
// @Tags			Exchange
// @Produce		json
// @Param			exchange_slug	path		string										true	"Exchange slug"
// @Param			state			body		exchange_request.ToggleExchangeSwapsRequest	true	"Toggle swaps request"
// @Success		200				{object}	response.Result[string]
// @Failure		422				{object}	apierror.Errors
// @Failure		503				{object}	apierror.Errors
// @Router			/v1/exchange/{exchange_slug}/toggle-swaps [post]
// @Security		Bearer
func (h *Handler) toggleExchangeSwaps(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	slug := models.ExchangeSlug(c.Params("exchange_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("exchange not found")).SetHttpCode(http.StatusNotFound)
	}
	req := &exchange_request.ToggleExchangeSwapsRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	state, err := h.services.ExchangeService.ChangeExchangeSwapsState(c.Context(), slug, usr.ID, models.ExchangeSwapState(req.NewState))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByData(state.WithdrawalState))
}

func (h *Handler) initExchangeRoutes(router fiber.Router) {
	g := router.Group("/exchange")
	g.Get("/list", h.exchangesList)
	g.Get("/deposit-addresses", h.getExchangeDepositAddress)
	g.Post("/:exchange_slug/keys", h.updateKeys)
	g.Delete("/:exchange_slug/keys", h.deleteKeys)
	g.Get("/:exchange_slug/test", h.testConnection)
	g.Post("/test", h.testConnectionExternal)
	g.Get("/:exchange_slug/balance", h.getBalance)
	g.Post("/:exchange_slug/set", h.setExchange)
	g.Get("/exchange-history", h.getUserOrderHistory)
	g.Get("/exchange-history/export", h.downloadUserOrderHistory)
	g.Put("/:exchange_slug/user-pairs", h.updateUserExchangePairs)
	g.Get("/:exchange_slug/user-pairs", h.getUserExchangePairs)
	g.Get("/:exchange_slug/withdrawal-rules", h.getExchangeWithdrawalRules)
	g.Get("/:exchange_slug/pairs", h.getExchangePairs)
	g.Get("/:exchange_slug/deposit-update", h.updateDepositAddresses)
	g.Post("/:exchange_slug/withdrawal-setting", h.createWithdrawalSetting)
	g.Get("/:exchange_slug/withdrawal-setting", h.getWithdrawalSettings)
	g.Get("/:exchange_slug/withdrawal-setting/:id", h.getWithdrawalSetting)
	g.Patch("/:exchange_slug/withdrawal-setting/:id", h.updateWithdrawalSetting)
	g.Delete("/:exchange_slug/withdrawal-setting/:id", h.deleteWithdrawalSetting)
	g.Get("/withdrawal", h.getWithdrawals)
	g.Get("/withdrawal/export", h.downloadWithdrawals)
	g.Get("/:exchange_slug/withdrawal/:id", h.getWithdrawalByID)
	g.Get("/:exchange_slug/chains", h.getExchangeChains)
	g.Post("/:exchange_slug/toggle-withdrawals", h.toggleExchangeWithdrawals)
	g.Post("/:exchange_slug/toggle-swaps", h.toggleExchangeSwaps)
}
