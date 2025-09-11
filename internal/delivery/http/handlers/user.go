package handlers

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/user_response"
	"github.com/dv-net/dv-merchant/internal/dto"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/service/user"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/user_request"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// authUser is a function get user info for auth
//
//	@Summary		Auth user
//	@Description	Auth a user
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[user_response.GetUserInfoResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		401	{object}	apierror.Errors
//	@Failure		422	{object}	apierror.Errors
//	@Failure		503	{object}	apierror.Errors
//	@Router			/v1/dv-admin/user [get]
//	@Security		BearerAuth
func (h *Handler) authUser(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	roles, err := h.services.PermissionService.UserRoles(usr.ID.String())
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	quickStartGuideStatus, err := h.services.SettingService.GetModelSetting(c.Context(), setting.QuickStartGuideStatus, usr)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.FromUserModelToInfoResponse(usr, quickStartGuideStatus.Value, roles...)))
}

// confirmEmail is a function to confirm user email
//
//	@Summary		Auth user
//	@Description	Auth a user
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[string]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		401	{object}	apierror.Errors
//	@Failure		422	{object}	apierror.Errors
//	@Failure		503	{object}	apierror.Errors
//	@Router			/v1/dv-admin/user/email-confirmation [post]
//	@Security		BearerAuth
func (h *Handler) confirmEmail(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &user_request.ConfirmEmailRequest{}
	if err = c.Bind().Body(req); err != nil {
		return err
	}

	if err = h.services.UserCredentialsService.ConfirmEmail(c.Context(), usr, req.Code); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("ok"))
}

// updateUser is a function update user info
//
//	@Summary		Update user
//	@Description	Update user
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			register	body		user_request.UpdateRequest	true	"Update user info"
//
//	@Success		200			{object}	response.Result[user_response.GetUserInfoResponse]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		401			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Router			/v1/dv-admin/user [put]
//	@Security		BearerAuth
func (h *Handler) updateUser(c fiber.Ctx) error {
	authUser, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &user_request.UpdateRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}
	d := user.RequestToUpdateUserDTO(req, authUser.ID)

	usr, err := h.services.UserService.UpdateUser(c.Context(), authUser, d)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	roles, err := h.services.PermissionService.UserRoles(usr.ID.String())
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	quickStartGuideStatus, err := h.services.SettingService.GetModelSetting(c.Context(), setting.QuickStartGuideStatus, usr)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.FromUserModelToInfoResponse(usr, quickStartGuideStatus.Value, roles...)))
}

// changePassword is a function update user password
//
//	@Summary		Update user password
//	@Description	Update user password
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			register	body		user_request.ChangePasswordInternalRequest	true	"Update user password"
//	@Success		200			{object}	response.Result[string]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		401			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Router			/v1/dv-admin/user/change-password [post]
//	@Security		BearerAuth
func (h *Handler) changePassword(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}
	req := &user_request.ChangePasswordInternalRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	d := dto.RequestToChangePasswordDTO(req)
	tokenHash, ok := c.Locals("token_hash").(string)
	if !ok {
		return apierror.New().AddError(fiber.ErrUnauthorized).SetHttpCode(fiber.StatusUnauthorized)
	}

	if err = h.services.UserCredentialsService.ChangePassword(c.Context(), usr, d, tokenHash); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("Password successfully changed"))
}

// initEmailConfirmation is a function update user password
//
//	@Summary		Init email confirnation
//	@Description	Init email confirnation
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[string]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		401	{object}	apierror.Errors
//	@Failure		422	{object}	apierror.Errors
//	@Router			/v1/dv-admin/user/init-email-confirmation [post]
//	@Security		BearerAuth
func (h *Handler) initEmailConfirmation(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	if err = h.services.UserCredentialsService.InitEmailConfirmation(c.Context(), usr); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("ok"))
}

// initEmailChange is a function update user password
//
//	@Summary		Init email change
//	@Description	Init email change
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[string]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		401	{object}	apierror.Errors
//	@Failure		422	{object}	apierror.Errors
//	@Router			/v1/dv-admin/user/init-email-change [post]
//	@Security		BearerAuth
func (h *Handler) initEmailChange(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	if err = h.services.UserCredentialsService.InitEmailChange(c.Context(), usr); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("ok"))
}

// confirmEmailChange is a function update user password
//
//	@Summary		Change email
//	@Description	Change email
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			register	body		user_request.ChangeEmailRequest	true	"Confirm email"
//	@Success		200			{object}	response.Result[string]
//	@Failure		400			{object}	apierror.Errors
//	@Failure		401			{object}	apierror.Errors
//	@Failure		422			{object}	apierror.Errors
//	@Router			/v1/dv-admin/user/change-email [post]
//	@Security		BearerAuth
func (h *Handler) confirmEmailChange(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &user_request.ChangeEmailRequest{}
	if err = c.Bind().Body(req); err != nil {
		return err
	}

	err = h.services.UserCredentialsService.ConfirmEmailChange(c.Context(), usr, user.ChangeEmailConfirmationDto{
		NewEmail:             req.NewEmail,
		NewEmailConfirmation: req.NewEmailConfirmation,
		Code:                 req.Code,
	})
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("ok"))
}

// generateTgLink is a function to generate tg link
//
//	@Summary		Generate telegram link
//	@Description	Generate telegram link
//	@Tags			User
//	@Produce		json
//	@Success		200	{object}	response.Result[user_response.TgLinkResponse]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		401	{object}	apierror.Errors
//	@Failure		422	{object}	apierror.Errors
//	@Router			/v1/dv-admin/user/tg-link [post]
//	@Security		BearerAuth
func (h *Handler) generateTgLink(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	link, err := h.services.UserService.GenerateTgLink(c.Context(), usr)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(user_response.TgLinkResponse{Link: link}))
}

// tgUnlinkInit is a function to init tg unlink
//
//	@Summary		Init telegram unlink
//	@Description	Init telegram unlink
//	@Tags			User
//	@Produce		json
//	@Success		200	{object}	response.Result[string]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		401	{object}	apierror.Errors
//	@Failure		422	{object}	apierror.Errors
//	@Router			/v1/dv-admin/user/tg-unlink/init [post]
//	@Security		BearerAuth
func (h *Handler) tgUnlinkInit(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	if err = h.services.UserService.InitTgInUnlink(c.Context(), usr); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("success"))
}

// tgUnlinkConfirm is a function to confirm tg unlink
//
//	@Summary		Confirm telegram unlink
//	@Description	Confirm telegram unlink
//	@Tags			User
//	@Produce		json
//	@Success		200	{object}	response.Result[string]
//	@Failure		400	{object}	apierror.Errors
//	@Failure		401	{object}	apierror.Errors
//	@Failure		422	{object}	apierror.Errors
//	@Router			/v1/dv-admin/user/tg-unlink/confirm [post]
//	@Security		BearerAuth
func (h *Handler) tgUnlinkConfirm(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &user_request.ConfirmTgUnlinkRequest{}
	if err = c.Bind().Body(req); err != nil {
		return err
	}

	if err = h.services.UserService.ConfirmTgUnlink(c.Context(), usr, req.Code); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("success"))
}

func (h *Handler) initUserRoute(v1 fiber.Router) {
	users := v1.Group("/user")
	users.Get("/", h.authUser)
	users.Put("/", h.updateUser)
	users.Post("/change-password", h.changePassword)
	users.Post("/confirm-email", h.confirmEmail)
	users.Post("/init-email-confirmation", h.initEmailConfirmation)
	users.Post("/init-email-change", h.initEmailChange)
	users.Post("/change-email", h.confirmEmailChange)
	users.Post("/tg-link", h.generateTgLink)
	users.Post("/tg-unlink/init", h.tgUnlinkInit)
	users.Post("/tg-unlink/confirm", h.tgUnlinkConfirm)
}
