package handlers

import (
	"net/http"

	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/dvadmin_response"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"
	"github.com/gofiber/fiber/v3"
)

// getDvnetOwnerData is a function to get system info
//
//	@Summary		Get system information
//	@Description	Get system information about initialization, root user existence, registration status
//	@Tags			System
//	@Produce		json
//	@Success		200	{object}	response.Result[dvadmin_response.OwnerData]
//	@Failure		400	{object}	apierror.Errors	"Bad request"
//	@Router			/v1/dv-admin/console/owner-data [get]
func (h Handler) getDvnetOwnerData(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	data, err := h.services.UserService.GetOwnerDataFromAdmin(c.Context(), user)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	resp := converters.FromAdminOwnerDataToResponse(data)
	return c.JSON(response.OkByData(resp))
}

// getAuthLink is a function to get dvnet authorization link
//
//	@Summary		Get link for auth in dvnet
//	@Description	Get link for auth in dvnet
//	@Tags			DVnet
//	@Produce		json
//	@Success		200	{object}	response.Result[dvadmin_response.AuthLinkResponse]
//	@Failure		400	{object}	apierror.Errors	"Bad request"
//	@Router			/v1/dv-admin/console/auth-link [get]
func (h Handler) getAuthLink(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	link, err := h.services.UserService.InitOwnerRegistration(c.Context(), user)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(dvadmin_response.AuthLinkResponse{Link: link})
}

func (h Handler) initConsoleRoutes(v3 fiber.Router) {
	public := v3.Group("/console")
	public.Get("/owner-data", h.getDvnetOwnerData)
	public.Get("/auth-link", h.getAuthLink)
}
