package handlers

import (
	"fmt"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/setting_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/service/user"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	// Blank import for swaggen
	_ "github.com/dv-net/dv-merchant/internal/delivery/http/responses/settings_response"
	_ "github.com/dv-net/dv-merchant/internal/service/processing"

	"github.com/gofiber/fiber/v3"
)

// createOrUpdateRootSetting is a function to create or update root settings
//
//	@Summary		Create or update root settings
//	@Description	Create or update root settings
//	@Tags			Setting
//	@Accept			json
//	@Produce		json
//	@Param			register	body		setting_request.CreateRequest	true	"Create or update root setting"
//	@Success		200			{object}	response.Result[string]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		404			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Router			/v1/dv-admin/root-setting/ [post]
//	@Security		BearerAuth
func (h Handler) createOrUpdateRootSetting(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	request := &setting_request.CreateRequest{}
	if err := c.Bind().Body(request); err != nil {
		return err
	}

	profile := h.services.SystemService.GetProfile(c.Context())

	if profile == models.AppProfileDemo {
		return apierror.New().AddError(fmt.Errorf("demo setting not muttable")).SetHttpCode(fiber.StatusForbidden)
	}

	var code string
	if request.OTP != nil {
		code = *request.OTP
	}

	if err = h.services.UserService.SettingUpdate(c.Context(), usr, user.SettingUpdateDTO{
		OTP:   code,
		Name:  request.Name,
		Value: request.Value,
	}); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("success"))
}

// createOrUpdateUserSetting is a function to create or update root settings
//
//	@Summary		Create or update root settings
//	@Description	Create or update root settings
//	@Tags			Setting
//	@Accept			json
//	@Produce		json
//	@Param			register	body		setting_request.CreateRequest	true	"Create or update user setting"
//	@Success		200			{object}	response.Result[string]
//	@Success		202			{object}	response.Result[string]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		404			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Router			/v1/dv-admin/user-setting/ [post]
//	@Security		BearerAuth
func (h Handler) createOrUpdateUserSetting(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	request := &setting_request.CreateRequest{}
	if err := c.Bind().Body(request); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	var code string
	if request.OTP != nil {
		code = *request.OTP
	}

	if err = h.services.UserService.SettingUpdate(c.Context(), usr, user.SettingUpdateDTO{
		OTP:   code,
		Name:  request.Name,
		Value: request.Value,
		Model: setting.IModelSetting(usr),
	}); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("success"))
}

// getRootSettings is a function to get all root settings
//
//	@Summary		Get root settings
//	@Description	Get root settings
//	@Tags			Setting
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[[]settings_response.SettingResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Router			/v1/dv-admin/setting/ [get]
//	@Security		BearerAuth
func (h Handler) getRootSettings(c fiber.Ctx) error {
	settings, err := h.services.SettingService.GetRootSettings(c.Context())
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusUnprocessableEntity)
	}
	res := converters.FromSettingModelToResponses(settings...)
	return c.JSON(response.OkByData(res))
}

// listRootSettings is a function to get all available root settings
//
//	@Summary		List available root settings
//	@Description	List available root settings
//	@Tags			Setting
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[[]setting.Dto]
//	@Failure		401	{object}	apierror.Errors
//	@Router			/v1/dv-admin/setting/list [get]
//	@Security		BearerAuth
func (h Handler) listRootSettings(c fiber.Ctx) error {
	availableSettings, err := h.services.SettingService.GetRootSettingsList(c.Context())
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(availableSettings))
}

// getSettingByName is a function to get concrete user setting
//
//	@Summary		Get setting value
//	@Description	Get setting value
//	@Tags			Setting
//	@Produce		json
//	@Success		200	{object}	response.Result[settings_response.SettingResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Router			/v1/dv-admin/user-setting/{setting_name} [get]
//	@Security		BearerAuth
func (h *Handler) getSettingByName(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	res, err := h.services.SettingService.GetModelSetting(c.Context(), c.Params("setting_name"), usr)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.FromSettingModelToResponse(res)))
}

// list user setting is a function to get all available user settings
//
//	@Summary		List available user settings
//	@Description	List available user settings
//	@Tags			Setting
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[[]setting.Dto]
//	@Failure		401	{object}	apierror.Errors
//	@Router			/v1/dv-admin/user-setting/list [get]
//	@Security		BearerAuth
func (h Handler) getUserSettingList(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	availableSettings, err := h.services.SettingService.GetAvailableModelSettings(c.Context(), user)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByData(availableSettings))
}

// getStoreSettingList is a function to get all available store settings
//
//	@Summary		List available store settings
//	@Description	List available store settings
//	@Tags			Setting
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[[]setting.Dto]
//	@Failure		401	{object}	apierror.Errors
//	@Router			/v1/dv-admin/store-setting/list/{id} [get]
//	@Security		BearerAuth
func (h Handler) getStoreSettingList(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}

	availableSettings, err := h.services.SettingService.GetAvailableStoreModelSettings(c.Context(), setting.IModelSetting(targetStore))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	return c.JSON(response.OkByData(availableSettings))
}

// createOrUpdateStoreSetting is a function to create or update store settings
//
//	@Summary		Create or update store settings
//	@Description	Create or update store settings
//	@Tags			Setting
//	@Accept			json
//	@Produce		json
//	@Param			register	body		setting_request.CreateRequest	true	"Create or update store setting"
//	@Success		200			{object}	response.Result[string]
//	@Success		202			{object}	response.Result[string]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		404			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Router			/v1/dv-admin/store-setting/{id} [post]
//	@Security		BearerAuth
func (h Handler) createOrUpdateStoreSetting(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}

	request := &setting_request.CreateRequest{}
	if err := c.Bind().Body(request); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	if err = h.services.SettingService.SetStoreModelSetting(c.Context(), setting.UpdateDTO{
		Name:  request.Name,
		Value: *request.Value,
		Model: setting.IModelSetting(targetStore),
	}); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("success"))
}

// getStoreSettingByName is a function to get concrete store setting
//
//	@Summary		Get store setting value
//	@Description	Get store setting value
//	@Tags			Setting
//	@Produce		json
//	@Success		200	{object}	response.Result[settings_response.SettingResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Router			/v1/dv-admin/store-setting/{id}/{setting_name} [get]
//	@Security		BearerAuth
func (h Handler) getStoreSettingByName(c fiber.Ctx) error {
	targetStore, err := h.validateAndLoadStore(c)
	if err != nil {
		return err
	}

	res, err := h.services.SettingService.GetStoreModelSetting(c.Context(), c.Params("setting_name"), setting.IModelSetting(targetStore))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.FromSettingModelToResponse(res)))
}

func (h Handler) initSettingRoutes(v1 fiber.Router) {
	rootSettings := v1.Group("/root-setting", h.services.PermissionService.FiberMiddleware(models.UserRoleRoot))
	rootSettings.Post("/", h.createOrUpdateRootSetting)
	rootSettings.Get("/", h.getRootSettings)
	rootSettings.Get("/list", h.listRootSettings)

	userSettings := v1.Group("/user-setting", h.services.PermissionService.FiberMiddleware(
		[]models.UserRole{
			models.UserRoleRoot,
			models.UserRoleDefault,
		}...,
	))
	userSettings.Get("/list", h.getUserSettingList)
	userSettings.Post("/", h.createOrUpdateUserSetting)
	userSettings.Get("/:setting_name", h.getSettingByName)

	storeSettings := v1.Group("/store-setting", h.services.PermissionService.FiberMiddleware(
		[]models.UserRole{
			models.UserRoleRoot,
			models.UserRoleDefault,
		}...,
	))
	storeSettings.Get("/list/:id", h.getStoreSettingList)
	storeSettings.Post("/:id", h.createOrUpdateStoreSetting)
	storeSettings.Get("/:id/:setting_name", h.getStoreSettingByName)
}
