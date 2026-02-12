package handlers

import (
	"errors"
	"fmt"
	"net/http"

	errs "github.com/dv-net/dv-merchant/internal/delivery/http/errors"
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/auth_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/auth_response"
	"github.com/dv-net/dv-merchant/internal/delivery/middleware"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/service/user"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// register is a function to register a new user
//
//	@Summary		Register user
//	@Description	Register a new user
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			register	body		auth_request.RegisterRequest	true	"Register account"
//	@Success		200			{object}	response.Result[auth_response.RegisterUserResponse]
//	@Failure		422			{object}	apierror.Errors
//	@Failure		503			{object}	apierror.Errors
//	@Router			/v1/dv-admin/auth/register [post]
func (h *Handler) register(c fiber.Ctx) error {
	ctx := c.Context()

	registrationState, err := h.services.SettingService.GetRootSetting(ctx, setting.RegistrationState)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	if registrationState.Value == "disabled" {
		return apierror.New().AddError(fmt.Errorf("registration is disabled")).SetHttpCode(fiber.StatusBadRequest)
	}

	req := &auth_request.RegisterRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}
	dto := user.RequestToCreateUserDTO(req)
	dto.Role = models.UserRoleDefault

	regInfo, err := h.services.AuthService.RegisterUser(ctx, dto)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	token, err := h.services.AuthService.AuthByUser(ctx, regInfo.User)
	if err != nil {
		return apierror.New(errs.ErrNoMatchesFound).AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(auth_response.RegisterUserResponse{
		UserInfo: regInfo,
		Token:    token.FullToken,
	}))
}

// register is a function to register a new user
//
//	@Summary		Register root user
//	@Description	Register root user - available only for first user
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			register	body		auth_request.RegisterRequest	true	"Register account"
//	@Success		200			{object}	response.Result[auth_response.RegisterRootResponse]
//	@Failure		422			{object}	apierror.Errors
//	@Failure		503			{object}	apierror.Errors
//	@Router			/v1/dv-admin/auth/register-root [post]
func (h *Handler) registerRoot(c fiber.Ctx) error {
	req := &auth_request.RegisterRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	ctx := c.Context()

	usrs, err := h.services.UserService.GetAllUsers(ctx, 1)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	if len(usrs) > 0 {
		return apierror.New().AddError(errors.New("users already exists")).SetHttpCode(fiber.StatusBadRequest)
	}
	dto := user.RequestToCreateUserDTO(req)
	dto.Role = models.UserRoleRoot

	regInfo, err := h.services.AuthService.RegisterUser(ctx, dto)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	token, err := h.services.AuthService.AuthByUser(ctx, regInfo.User)
	if err != nil {
		return apierror.New(errs.ErrNoMatchesFound).AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(auth_response.RegisterRootResponse{
		Token: token.FullToken,
	}))
}

// login is a function to auth a user
//
//	@Summary		Auth user
//	@Description	Auth a user
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			register	body		auth_request.AuthRequest	true	"Register account"
//	@Success		200			{object}	response.Result[auth_response.AuthResponse]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Failure		503			{object}	apierror.Errors
//	@Router			/v1/dv-admin/auth/login [post]
func (h *Handler) login(c fiber.Ctx) error {
	dto := &auth_request.AuthRequest{}
	if err := c.Bind().Body(dto); err != nil {
		return err
	}
	ctx := c.Context()
	token, err := h.services.AuthService.Auth(ctx, *dto)

	if token == nil || err != nil {
		return apierror.New(errs.ErrNoMatchesFound).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(auth_response.AuthResponse{
		Token: token.FullToken,
	}))
}

func (h *Handler) initAuthRoutes(v1 fiber.Router) {
	auth := v1.Group("/auth")
	auth.Post("/register", h.register,
		middleware.TurnstileMiddleware(h.services.TurnstileVerifier),
		middleware.TimezoneNormalizer())
	auth.Post("/register-root", h.registerRoot,
		middleware.TurnstileMiddleware(h.services.TurnstileVerifier),
		middleware.TimezoneNormalizer())
	auth.Post("/login", h.login)
}
