package handlers

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/dictionaries_response"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// loadDictionaries  is a function to load dictionaries
//
//	@Summary		Dictionaries load
//	@Description	Dictionaries load
//	@Tags			Dictionary
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[dictionaries_response.GetDictionariesResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		500	{object}	apierror.Errors
//	@Router			/v1/dv-admin/dictionaries/ [get]
//	@Security		BearerAuth
func (h *Handler) loadDictionaries(c fiber.Ctx) error {
	d, err := h.services.DictionaryService.LoadDictionary(c.Context())
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	pInfo, err := h.services.ProcessingSystemService.GetProcessingSystemInfo(c.Context())
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	bVersion, bHash := h.services.SystemService.GetAppVersion(c.Context())

	generalSettings, err := h.services.SettingService.GetRootSettingsByNames(c.Context(), []string{setting.MerchantPayFormDomain})
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByData(&dictionaries_response.GetDictionariesResponse{
		AvailableRateSources:  d.AvailableSources,
		AvailableCurrencies:   d.AvailableCurrencies,
		AvailableAMLProviders: h.services.AMLService.GetAllActiveProviders(),
		BackendVersionHash:    bHash,
		BackendVersionTag:     bVersion,
		ProcessingVersionHash: pInfo.Hash,
		ProcessingVersionTag:  pInfo.Version,
		BackendAddress:        d.BackendAddress,
		GeneralSettings:       converters.FromSettingModelToResponses(generalSettings...),
	}))
}

func (h *Handler) initDictionariesRoutes(v1 fiber.Router) {
	dictionaries := v1.Group("/dictionaries")
	dictionaries.Get("/", h.loadDictionaries)
}
