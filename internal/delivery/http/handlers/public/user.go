package public

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/public_request"
	"github.com/dv-net/dv-merchant/internal/delivery/middleware"
	"github.com/dv-net/dv-merchant/internal/service/user"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// forgotPassword submits user email for password reset, sending a reset link to the email.
//
//	@Summary		Request password reset link by email
//	@Description	Request password reset link by email
//	@Tags			User,Public
//	@Accept			json
//	@Produce		json
//	@Param			email	body		public_request.PublicUserForgotPasswordRequest	true	"User email"
//	@Success		200		{object}	response.Result[string]
//	@Router			/v1/public/user/forgot-password [post]
func (h *Handler) forgotPassword(c fiber.Ctx) error {
	dto := &public_request.PublicUserForgotPasswordRequest{}
	if err := c.Bind().Body(dto); err != nil {
		return err
	}

	if err := h.services.UserCredentialsService.InitPasswordReset(c.Context(), dto.Email); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("Password reset code will be sent if the email is registered"))
}

// resetPassword
//
//	@Summary		Reset user password
//	@Description	Reset user password
//	@Tags			User,Public
//	@Accept			json
//	@Produce		json
//	@Param			resetPassword	body		public_request.PublicResetPasswordRequest	true	"Public reset password request"
//	@Success		200				{object}	response.Result[string]
//	@Router			/v1/public/user/reset-password [post]
func (h *Handler) resetPassword(c fiber.Ctx) error {
	dto := &public_request.PublicResetPasswordRequest{}
	if err := c.Bind().Body(dto); err != nil {
		return err
	}

	if err := h.services.UserCredentialsService.ResetPassword(c.Context(), user.ResetPasswordDto{
		Code:            dto.Code,
		Email:           dto.Email,
		NewPassword:     dto.Password,
		ConfirmPassword: dto.PasswordConfirmation,
	}); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("Password reset successfully"))
}

// acceptInvite
//
//	@Summary		Accept admin invite link
//	@Description	Accept admin invite link
//	@Tags			User,Public
//	@Accept			json
//	@Produce		json
//	@Param			inviteRequest	body		public_request.PublicAcceptInviteRequest	true	"Invite request"
//	@Success		200				{object}	response.Result[string]
//	@Router			/v1/public/user/accept-invite [post]
func (h *Handler) acceptInvite(c fiber.Ctx) error {
	dto := &public_request.PublicAcceptInviteRequest{}
	if err := c.Bind().Body(dto); err != nil {
		return err
	}

	if err := h.services.UserService.AcceptInvite(c.Context(), dto.Email, dto.Token, dto.Password); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("Invite accepted successfully"))
}

func (h *Handler) initUserRoutes(v1 fiber.Router) {
	user := v1.Group("/user")
	user.Post("/forgot-password", h.forgotPassword,
		middleware.LimiterMiddleware(3, 60, middleware.WithSlidingWindow),
		middleware.FakeDelayMiddleware(2*time.Second),
	)
	user.Post("/reset-password", h.resetPassword,
		middleware.LimiterMiddleware(3, 60, middleware.WithSlidingWindow),
		middleware.FakeDelayMiddleware(2*time.Second),
	)
	user.Post("/accept-invite", h.acceptInvite)
}
