package handlers

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/aml_requests"
	"github.com/dv-net/dv-merchant/internal/service/aml"
	"github.com/dv-net/dv-merchant/internal/tools/converters"

	// Blank imports for swagger gen
	_ "github.com/dv-net/dv-merchant/internal/delivery/http/responses/aml_responses"
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

func (h *Handler) initAMLRoutes(v1 fiber.Router) {
	amlRoutes := v1.Group("/aml")
	amlRoutes.Post("/:aml_provider_slug/keys", h.updateAMLKeys)
	amlRoutes.Get("/:aml_provider_slug/keys", h.getAMLKeys)
	amlRoutes.Delete("/:aml_provider_slug/keys", h.deleteAMLKeys)
	amlRoutes.Get("/:aml_provider_slug/currencies", h.getAMLCurrencies)
	amlRoutes.Get("/history", h.amlHistory)
	amlRoutes.Post("/score-transaction", h.scoreTransaction)
}
