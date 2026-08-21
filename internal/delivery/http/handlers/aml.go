package handlers

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/aml_requests"
	"github.com/dv-net/dv-merchant/internal/service/aml"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/jackc/pgx/v5"

	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/aml_responses"

	// Blank imports for swagger gen
	_ "github.com/dv-net/dv-merchant/internal/storage/storecmn"

	"errors"
	"net/http"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// scoreTransaction is a function to send tx scoring in AML provider
//
//	@Summary		Score transaction in specific AML-provider
//	@Description	Score transaction in specific AML-provider
//	@Tags			AML
//	@Accept			json
//	@Produce		json
//	@Param			register	body		aml_requests.ScoreTxRequest	true	"Score transaction"
//	@Success		200			{object}	response.Result[string]
//	@Failure		422			{object}	apierror.Errors
//	@Failure		503			{object}	apierror.Errors
//	@Router			/v1/dv-admin/aml/score-transaction [post]
func (h *Handler) scoreTransaction(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &aml_requests.ScoreTxRequest{}
	if err = c.Bind().Body(req); err != nil {
		return err
	}

	_, err = h.services.AMLService.ScoreTransaction(c.Context(), usr, aml.CheckDTO{
		TxID:          req.TxID,
		CurrencyID:    req.CurrencyID,
		ProviderSlug:  req.ProviderSlug,
		Direction:     aml.Direction(req.Direction),
		OutputAddress: req.OutputAddress,
	})
	if err != nil {
		if errors.Is(err, aml.ErrUnsupportedCurrencies) ||
			errors.Is(err, aml.ErrUnsupportedProvider) ||
			errors.Is(err, aml.ErrInvalidAddress) {
			return apierror.New().AddError(err).SetHttpCode(http.StatusUnprocessableEntity)
		}
		return apierror.New().AddError(err).SetHttpCode(http.StatusInternalServerError)
	}

	return c.JSON(response.OkByMessage("OK"))
}

// updateAMLKeys updates or create AML-provider user keys
//
//	@Summary		Update AML-provider keys
//	@Description	Update or create AML-provider user keys
//	@Tags			AML
//	@Accept			json
//	@Produce		json
//	@Param			aml_provider_slug	path		string							true	"AML-provider slug"
//	@Param			update_keys			body		aml_requests.UpdateUserAMLKeys	true	"Update AML keys"
//	@Success		200					{object}	response.Result[string]
//	@Failure		422					{object}	apierror.Errors
//	@Failure		503					{object}	apierror.Errors
//	@Router			/v1/dv-admin/aml/{aml_provider_slug}/keys [post]
func (h *Handler) updateAMLKeys(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	slug := models.AMLSlug(c.Params("aml_provider_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("AML provider not found")).SetHttpCode(http.StatusNotFound)
	}

	req := &aml_requests.UpdateUserAMLKeys{}
	if err = c.Bind().Body(req); err != nil {
		return err
	}

	updatedKeys, err := h.services.AMLKeysService.UpdateUserKeys(c.Context(), usr, converters.ConvertAMLKeysRequestToDTO(slug, req))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.ConvertAmlKeysToResponseKeys(updatedKeys)))
}

// deleteAMLKeys removes AML-provider keys for user.
//
//	@Summary		Delete AML-provider keys
//	@Description	Delete AML-provider user keys
//	@Tags			AML
//	@Accept			json
//	@Produce		json
//	@Param			aml_provider_slug	path		string	true	"AML-provider slug"
//	@Success		200					{object}	response.Result[string]
//	@Failure		422					{object}	apierror.Errors
//	@Failure		503					{object}	apierror.Errors
//	@Router			/v1/dv-admin/aml/{aml_provider_slug}/keys [delete]
func (h *Handler) deleteAMLKeys(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	slug := models.AMLSlug(c.Params("aml_provider_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("AML provider not found")).SetHttpCode(http.StatusNotFound)
	}

	err = h.services.AMLKeysService.DeleteUserKeys(c.Context(), usr, slug)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("AML keys successfully deleted"))
}

// getAMLKeys returns auth keys for specific user
//
//	@Summary		Get AML-provider keys
//	@Description	Get AML-provider user keys
//	@Tags			AML
//	@Accept			json
//	@Produce		json
//	@Param			aml_provider_slug	path		string	true	"AML-provider slug"
//	@Success		200					{object}	response.Result[[]aml_responses.AMLKey]
//	@Failure		422					{object}	apierror.Errors
//	@Failure		503					{object}	apierror.Errors
//	@Router			/v1/dv-admin/aml/{aml_provider_slug}/keys [get]
func (h *Handler) getAMLKeys(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	slug := models.AMLSlug(c.Params("aml_provider_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("AML provider not found")).SetHttpCode(http.StatusNotFound)
	}

	keysDTO, err := h.services.AMLKeysService.GetKeys(c.Context(), usr, slug)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.ConvertAmlKeysToResponseKeys(keysDTO)))
}

// getAMLKeys returns supported by AML-provider currencies
//
//	@Summary		Get supported by AML-provider currencies
//	@Description	Get supported by AML-provider currencies
//	@Tags			AML
//	@Accept			json
//	@Produce		json
//	@Param			aml_provider_slug	path		string	true	"AML-provider slug"
//	@Success		200					{object}	response.Result[[]models.CurrencyShort]
//	@Failure		422					{object}	apierror.Errors
//	@Failure		503					{object}	apierror.Errors
//	@Router			/v1/dv-admin/aml/{aml_provider_slug}/currencies [get]
func (h *Handler) getAMLCurrencies(c fiber.Ctx) error {
	_, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	slug := models.AMLSlug(c.Params("aml_provider_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("AML provider not found")).SetHttpCode(http.StatusNotFound)
	}

	currencies, err := h.services.AMLService.GetSupportedCurrencies(c.Context(), slug)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByData(currencies))
}

// getAMLKeys fetch AML-provider user keys.
//
//	@Summary		Get AML-provider checks history
//	@Description	Get AML-provider checks history
//	@Tags			AML
//	@Accept			json
//	@Produce		json
//	@Param			aml_provider_slug	path		string							true	"AML-provider slug"
//
//	@Param			string				query		aml_requests.GetHistoryRequest	true	"GetHistoryRequest"
//
//	@Success		200					{object}	response.Result[storecmn.FindResponseWithFullPagination[aml_responses.AmlHistoryResponse]]
//	@Failure		422					{object}	apierror.Errors
//	@Failure		503					{object}	apierror.Errors
//	@Router			/v1/dv-admin/aml/history [get]
func (h *Handler) amlHistory(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &aml_requests.GetHistoryRequest{}
	if err = c.Bind().Query(req); err != nil {
		return err
	}

	result, err := h.services.AMLService.GetCheckHistory(c.Context(), usr, aml.ChecksWithHistoryDTO{
		Slug:     req.ProviderSlug,
		DateFrom: req.DateFrom,
		DateTo:   req.DateTo,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.GetAMLCheckHistoryResponse(result)))
}

// getAMLStatistics returns today's AML check statistics for the current user.
//
//	@Summary		Get AML check statistics
//	@Description	Get today's completed, successful, and failed AML check counts
//	@Tags			AML
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[aml_responses.StatisticsResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Router			/v1/dv-admin/aml/statistics [get]
func (h *Handler) getAMLStatistics(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	statistics, err := h.services.AMLService.GetStatistics(c.Context(), usr.ID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByData(statistics))
}

// getAMLSignalingCategories returns the list of risk signal categories supported by an AML-provider.
//
//	@Summary		Get AML-provider signal categories
//	@Description	Get the list of risk signal categories a specific AML-provider can return in a check response (used to build a per-store ignore-list)
//	@Tags			AML
//	@Accept			json
//	@Produce		json
//	@Param			aml_provider_slug	path		string	true	"AML-provider slug"
//	@Success		200					{object}	response.Result[[]aml_responses.SignalCategoryResponse]
//	@Failure		400					{object}	apierror.Errors
//	@Failure		404					{object}	apierror.Errors
//	@Router			/v1/dv-admin/aml/{aml_provider_slug}/signals [get]
func (h *Handler) getAMLSignalingCategories(c fiber.Ctx) error {
	_, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	slug := models.AMLSlug(c.Params("aml_provider_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("AML provider not found")).SetHttpCode(http.StatusNotFound)
	}

	signalCategories, err := h.services.AMLService.GetSignalsCategorise(c.Context(), slug)
	if err != nil {
		return apierror.New().AddError(errors.New("failed to get signal categories")).SetHttpCode(http.StatusBadRequest)
	}

	resp := make([]aml_responses.SignalCategoryResponse, 0, len(signalCategories))
	for _, category := range signalCategories {
		resp = append(resp, aml_responses.SignalCategoryResponse{
			Category: category.Category,
			Label:    category.Label,
		})
	}

	return c.JSON(response.OkByData(resp))
}

// getAmlSettings returns AML settings for the current user
//
//	@Summary		Get AML settings
//	@Description	Get AML settings for the current user
//	@Tags			AML
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[aml_responses.AmlSettingsResponse]
//	@Failure		404	{object}	apierror.Errors
//	@Failure		503	{object}	apierror.Errors
//	@Router			/v1/dv-admin/aml/settings [get]
func (h *Handler) getAmlSettings(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	settings, err := h.services.AMLUserSettings.GetAmlSettings(c.Context(), usr.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(response.OkByData(struct{}{}))
		}
		return h.handleError(err, "aml settings")
	}
	return c.JSON(response.OkByData(aml_responses.NewAmlSettingsResponse(settings)))
}

// updateAmlSettings updates AML settings for the current user
//
//	@Summary		Update AML settings
//	@Description	Update AML settings for the current user
//	@Tags			AML
//	@Accept			json
//	@Produce		json
//	@Param			update_settings	body		aml_requests.UpdateAmlSettingsRequest	true	"Update AML settings"
//	@Success		200				{object}	response.Result[aml_responses.AmlSettingsResponse]
//	@Failure		404				{object}	apierror.Errors
//	@Failure		503				{object}	apierror.Errors
//	@Router			/v1/dv-admin/aml/settings [post]
func (h *Handler) updateAmlSettings(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	req := &aml_requests.UpdateAmlSettingsRequest{}
	if err = c.Bind().Body(req); err != nil {
		return err
	}

	slug := models.AMLSlug(req.ProviderSlug)
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("AML provider not found")).SetHttpCode(http.StatusNotFound)
	}

	settings, err := h.services.AMLUserSettings.UpdateAmlSettings(c.Context(), usr.ID, aml.UpdateAmlSettingsDTO{
		Enabled:      req.Enabled,
		ProviderSlug: &slug,
	})

	if err != nil {
		return h.handleError(err, "aml settings")
	}
	return c.JSON(response.OkByData(aml_responses.NewAmlSettingsResponse(settings)))
}

// listAmlRiskRules returns AML risk rules for a specific provider
//
//	@Summary		Get AML risk rules
//	@Description	Get AML risk rules for a specific provider, merged with supported signal categories
//	@Tags			AML
//	@Accept			json
//	@Produce		json
//	@Param			aml_provider_slug	path		string	true	"AML-provider slug"
//	@Success		200					{object}	response.Result[[]aml_responses.RiskRuleResponse]
//	@Failure		400					{object}	apierror.Errors
//	@Failure		404					{object}	apierror.Errors
//	@Router			/v1/dv-admin/aml/{aml_provider_slug}/rules [get]
func (h *Handler) listAmlRiskRules(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	slug := models.AMLSlug(c.Params("aml_provider_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("AML provider not found")).SetHttpCode(http.StatusNotFound)
	}

	rules, err := h.services.AMLUserSettings.ListRiskRules(c.Context(), usr.ID, &slug)
	if err != nil {
		return h.handleError(err, "aml risk rules")
	}

	categories, err := h.services.AMLService.GetSignalsCategorise(c.Context(), slug)
	if err != nil {
		return h.handleError(err, "aml signal categories")
	}

	return c.JSON(response.OkByData(aml_responses.MergeRiskRules(categories, rules)))
}

// upsertAmlRiskRules creates or updates AML risk rules for a specific provider
//
//	@Summary		Upsert AML risk rules
//	@Description	Create or update AML risk rules for a specific provider
//	@Tags			AML
//	@Accept			json
//	@Produce		json
//	@Param			aml_provider_slug	path		string								true	"AML-provider slug"
//	@Param			upsert_rules		body		aml_requests.UpsertRiskRulesRequest	true	"Upsert risk rules"
//	@Success		200					{object}	response.Result[[]aml_responses.RiskRuleResponse]
//	@Failure		400					{object}	apierror.Errors
//	@Failure		404					{object}	apierror.Errors
//	@Router			/v1/dv-admin/aml/{aml_provider_slug}/rules [post]
func (h *Handler) upsertAmlRiskRules(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	slug := models.AMLSlug(c.Params("aml_provider_slug"))
	if !slug.Valid() {
		return apierror.New().AddError(errors.New("AML provider not found")).SetHttpCode(http.StatusNotFound)
	}

	req := &aml_requests.UpsertRiskRulesRequest{}
	if err = c.Bind().Body(req); err != nil {
		return err
	}

	dtos := make([]aml.RiskRuleDTO, 0, len(req.Rules))
	for _, r := range req.Rules {
		dtos = append(dtos, aml.RiskRuleDTO{RiskType: r.RiskType, Enabled: r.Enabled, Threshold: r.Threshold, Action: r.Action})
	}

	rules, err := h.services.AMLUserSettings.UpsertRiskRules(c.Context(), usr.ID, &slug, dtos)
	if err != nil {
		return h.handleError(err, "aml risk rules")
	}

	resp := make([]aml_responses.RiskRuleResponse, 0, len(rules))
	for _, r := range rules {
		resp = append(resp, aml_responses.NewRiskRuleResponse(r))
	}
	return c.JSON(response.OkByData(resp))
}

func (h *Handler) initAMLRoutes(v1 fiber.Router) {
	amlRoutes := v1.Group("/aml")
	amlRoutes.Post("/:aml_provider_slug/keys", h.updateAMLKeys)
	amlRoutes.Get("/:aml_provider_slug/keys", h.getAMLKeys)
	amlRoutes.Delete("/:aml_provider_slug/keys", h.deleteAMLKeys)
	amlRoutes.Get("/:aml_provider_slug/currencies", h.getAMLCurrencies)
	amlRoutes.Get("/:aml_provider_slug/signals", h.getAMLSignalingCategories)
	amlRoutes.Get("/history", h.amlHistory)
	amlRoutes.Get("/statistics", h.getAMLStatistics)
	amlRoutes.Post("/score-transaction", h.scoreTransaction)

	amlRoutes.Get("/settings", h.getAmlSettings)
	amlRoutes.Post("/settings", h.updateAmlSettings)
	amlRoutes.Get("/:aml_provider_slug/rules", h.listAmlRiskRules)
	amlRoutes.Post("/:aml_provider_slug/rules", h.upsertAmlRiskRules)
}
