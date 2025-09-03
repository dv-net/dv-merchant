package handlers

import (
	"fmt"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/two_factor_auth"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/two_factor_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	// Blank import for swaggen
	_ "github.com/dv-net/dv-merchant/internal/service/processing"

	"github.com/gofiber/fiber/v3"
)

// getTwoFactorSecret returns owner's secret for TOTP setup.
//
//	@Summary		Retrieve 2FA authentication secret
//	@Description	Fetches the secret needed for setting up Time-based One-Time Password (TOTP).
//	@Tags			2fa
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[two_factor_response.GetTwoFactorAuthSecretResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Failure		403	{object}	apierror.Errors
//	@Failure		400	{object}	apierror.Errors
//	@Failure		422	{object}	apierror.Errors
//	@Router			/v1/dv-admin/2fa/ [get]
//	@Security		BearerAuth
func (h *Handler) getTwoFactorSecret(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	if !user.ProcessingOwnerID.Valid {
		return apierror.New().AddError(fmt.Errorf("user have no owner")).SetHttpCode(fiber.StatusUnprocessableEntity)
	}

	res, err := h.services.ProcessingOwnerService.GetTwoFactorAuthData(c.Context(), user.ProcessingOwnerID.UUID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	if res.IsConfirmed {
		res.Secret = ""
	}

	return c.JSON(response.OkByData(&two_factor_response.GetTwoFactorAuthSecretResponse{
		Secret:      res.Secret,
		IsConfirmed: res.IsConfirmed,
	}))
}

// confirmTwoFactor validates user provided TOTP against secret key, changing state of two-factor auth.
//
//	@Summary		Confirm 2FA authentication
//	@Description	Validates the provided TOTP code against the stored secret to activate 2FA.
//	@Tags			2fa
//	@Accept			json
//	@Produce		json
//	@Param			register	body		two_factor_auth.Confirm	true	"Confirmation data"
//	@Success		200			{object}	response.Result[string]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		403			{object}	apierror.Errors
//	@Failure		400			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Router			/v1/dv-admin/2fa/confirm [post]
//	@Security		BearerAuth
func (h *Handler) confirmTwoFactor(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	if !user.ProcessingOwnerID.Valid {
		return apierror.New().AddError(fmt.Errorf("user have no owner")).SetHttpCode(fiber.StatusUnprocessableEntity)
	}

	dto := &two_factor_auth.Confirm{}
	if err := c.Bind().Body(dto); err != nil {
		return err
	}

	confirmationErr := h.services.ProcessingOwnerService.ConfirmTwoFactorAuth(c.Context(), user.ProcessingOwnerID.UUID, dto.OTP)
	if confirmationErr != nil {
		return apierror.New().AddError(confirmationErr).SetHttpCode(fiber.StatusBadRequest)
	}

	// Send notification for successful 2FA activation
	go h.services.NotificationService.SendUser(
		c.Context(),
		models.NotificationTypeTwoFactorAuthentication,
		user,
		&notify.TwoFactorAuthenticationNotification{
			Email:     user.Email,
			Language:  user.Language,
			IsEnabled: true,
		},
		&models.NotificationArgs{
			UserID: &user.ID,
		},
	)

	return c.JSON(response.OkByMessage("2FA successfully confirmed"))
}

// disableTwoFactor disables two-factor auth completely, removing old secret.
//
//	@Summary		Disable 2FA authentication
//	@Description	Deactivates 2FA, effectively removing the current TOTP secret.
//	@Tags			2fa
//	@Accept			json
//	@Produce		json
//	@Param			register	body		two_factor_auth.ChangeStatus	true	"Status data"
//	@Success		200			{object}	response.Result[string]
//	@Failure		401			{object}	apierror.Errors
//	@Failure		403			{object}	apierror.Errors
//	@Failure		400			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Router			/v1/dv-admin/2fa/disable [post]
//	@Security		BearerAuth
func (h *Handler) disableTwoFactor(c fiber.Ctx) error {
	user, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	if !user.ProcessingOwnerID.Valid {
		return apierror.New().AddError(fmt.Errorf("user have no owner")).SetHttpCode(fiber.StatusUnprocessableEntity)
	}

	dto := &two_factor_auth.ChangeStatus{}
	if err := c.Bind().Body(dto); err != nil {
		return err
	}

	changeStatusErr := h.services.ProcessingOwnerService.DisableTwoFactorAuth(
		c.Context(),
		user.ProcessingOwnerID.UUID,
		dto.OTP,
	)
	if changeStatusErr != nil {
		return apierror.New().AddError(changeStatusErr).SetHttpCode(fiber.StatusBadRequest)
	}

	// Send notification for 2FA deactivation
	go h.services.NotificationService.SendUser(
		c.Context(),
		models.NotificationTypeTwoFactorAuthentication,
		user,
		&notify.TwoFactorAuthenticationNotification{
			Email:     user.Email,
			Language:  user.Language,
			IsEnabled: false,
		},
		&models.NotificationArgs{
			UserID: &user.ID,
		},
	)

	return c.JSON(response.OkByMessage("2FA successfully disabled"))
}

func (h *Handler) init2faRoutes(v1 fiber.Router) {
	ffa := v1.Group("/2fa")
	ffa.Get("/", h.getTwoFactorSecret)
	ffa.Post("/confirm", h.confirmTwoFactor)
	ffa.Post("/disable", h.disableTwoFactor)
}
