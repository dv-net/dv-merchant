package handlers

import (
	"context"

	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/system_response"
	"github.com/dv-net/dv-merchant/internal/delivery/middleware"
	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// info is a function to get system info
//
//	@Summary		Get system information
//	@Description	Get system information about initialization, root user existence, registration status
//	@Tags			System
//	@Produce		json
//	@Success		200	{object}	response.Result[system_response.SystemInfoResponse]
//	@Failure		400	{object}	apierror.Errors	"Bad request"
//	@Router			/v1/dv-admin/system/info [get]
func (h Handler) info(c fiber.Ctx) error {
	info, err := h.services.SystemService.GetInfo(c.Context())
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	resp := converters.SystemInfoModelToResponse(info)
	return c.JSON(response.OkByData(resp))
}

//	updater processing on server
//
// @Summary		Update processing
// @Description	Update processing
// @Tags			System
// @Produce		json
// @Success		200	{object}	response.Result[string]
// @Failure		400	{object}	apierror.Errors	"Bad request"
// @Router			/v1/dv-admin/system/update/processing [post]
func (h Handler) updateProcessing(c fiber.Ctx) error {
	ctx := context.Background()
	go func() {
		err := h.services.UpdaterService.UpdateProcessing(ctx)
		if err != nil {
			h.logger.Error("error updating processing status", err)
		}
	}()
	return c.JSON(response.OkByMessage("Success start processing update"))
}

//	updater backend on server
//
// @Summary		Update backend
// @Description	Update backend
// @Tags			System
// @Produce		json
// @Success		200	{object}	response.Result[string]
// @Failure		400	{object}	apierror.Errors	"Bad request"
// @Router			/v1/dv-admin/system/update/backend [post]
func (h Handler) updateBackend(c fiber.Ctx) error {
	ctx := context.Background()

	go func() {
		err := h.services.UpdaterService.UpdateBackend(ctx)
		if err != nil {
			h.logger.Error("error updating backend status", err)
		}
	}()

	return c.JSON(response.OkByMessage("Success start backend update"))
}

// loadNewVersions  is a function to load applications components versions
//
//	@Summary		Application versions info with no cache
//	@Description	Application versions info with no cache
//	@Tags			System
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[system_response.VersionResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Router			/v1/dv-admin/system/version [get]
//	@Security		BearerAuth
func (h *Handler) loadNewVersions(c fiber.Ctx) error {
	h.logger.Info("UPDATER DEBUG: Start checking new versions")
	versions, err := h.services.UpdaterService.CheckApplicationVersions(c.Context())
	if err != nil {
		h.logger.Info("UPDATER DEBUG: Checking new version completed without error")
		return c.JSON(response.OkByData(&system_response.VersionResponse{
			NewBackendVersion:    nil,
			NewProcessingVersion: nil,
		}))
	}
	h.logger.Error("UPDATER DEBUG: Checking new version completed without error", err)

	return c.JSON(response.OkByData(&system_response.VersionResponse{
		NewBackendVersion:    versions.BackendVersion,
		NewProcessingVersion: versions.ProcessingVersion,
	}))
}

func (h Handler) initPublicSystemRoutes(v3 fiber.Router) {
	public := v3.Group("/system")
	public.Get("/info", h.info)

	public.Post("/update/processing", h.updateProcessing,
		middleware.AuthMiddleware(h.services.AuthService),
		middleware.CasbinMiddleware(h.services.PermissionService, []models.UserRole{models.UserRoleRoot}),
	)

	public.Post("/update/backend", h.updateBackend,
		middleware.AuthMiddleware(h.services.AuthService),
		middleware.CasbinMiddleware(h.services.PermissionService, []models.UserRole{models.UserRoleRoot}),
	)

	public.Get("/versions", h.loadNewVersions,
		middleware.AuthMiddleware(h.services.AuthService),
		middleware.CasbinMiddleware(h.services.PermissionService, []models.UserRole{models.UserRoleDefault, models.UserRoleRoot, models.UserRoleSupport}),
	)
}
