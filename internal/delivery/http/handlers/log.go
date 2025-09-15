package handlers

import (
	"errors"

	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/log_response"
	"github.com/dv-net/dv-merchant/internal/delivery/middleware"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// @Summary	Get available logs types
// @Description
// @Tags		Logs
// @Accept		json
// @Produce	json
// @Success	200	{object}	response.Result[log_response.GetLogTypesResponse]
// @Failure	422	{object}	apierror.Errors
// @Failure	503	{object}	apierror.Errors
// @Router		/v1/dv-admin/logs [get]
// @Security	Bearer
func (h *Handler) getLogsType(c fiber.Ctx) error {
	_, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	logs, err := h.services.LogService.GetAllTypes(c.Context())
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusServiceUnavailable)
	}

	return c.JSON(response.OkByData(converters.GetAllLogTypesResponse(logs)))
}

// @Summary	Get logs by slug
// @Description
// @Tags		Logs
// @Accept		json
// @Produce	json
// @Param		id	path		string	true	"Monitor UUID"
// @Success	200	{object}	response.Result[log_response.GetLogsResponse]
// @Failure	404	{object}	apierror.Errors
// @Failure	422	{object}	apierror.Errors
// @Failure	503	{object}	apierror.Errors
// @Router		/v1/dv-admin/logs/{slug} [get]
// @Security	Bearer
func (h *Handler) getLogsBySlug(c fiber.Ctx) error {
	_, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	slug := c.Params("slug")
	if slug == "" {
		return apierror.New().AddError(errors.New("missing correct monitor slug")).SetHttpCode(fiber.StatusBadRequest)
	}

	messages, err := h.services.LogService.GetLogsBySlug(c.Context(), slug)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusServiceUnavailable)
	}

	return c.JSON(response.OkByData(converters.GetAllLogsResponse(messages)))
}

// @Summary	Get last logs
// @Description
// @Tags		Logs
// @Accept		json
// @Produce	json
// @Success	200	{object}	response.Result[log_response.GetLastLogsResponse]
// @Failure	404	{object}	apierror.Errors
// @Failure	422	{object}	apierror.Errors
// @Failure	503	{object}	apierror.Errors
// @Router		/v1/dv-admin/logs/last [get]
// @Security	Bearer
func (h *Handler) getLogs(c fiber.Ctx) error {
	_, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	return c.JSON(response.OkByData(log_response.GetLastLogsResponse{
		Items: h.logger.LastLogs(),
	}))
}

func (h *Handler) initLogsRoutes(v1 fiber.Router) {
	logs := v1.Group("/logs", middleware.CasbinMiddleware(h.services.PermissionService, []models.UserRole{models.UserRoleRoot}))
	logs.Get("/", h.getLogsType)
	logs.Get("/last", h.getLogs)
	logs.Get("/:slug", h.getLogsBySlug)
}
