package handlers

import (
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/service/notify"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/admin_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/admin_response"
	"github.com/dv-net/dv-merchant/internal/delivery/middleware"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	_ "github.com/dv-net/dv-merchant/internal/storage/storecmn" // swaggo
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"golang.org/x/text/language"

	"github.com/gofiber/fiber/v3"
)

// getUsers is a function to get all users
//
//	@Summary		Get all users
//	@Description	Get all users
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			string	query		admin_request.GetUsersRequest	true	"GetUsersRequest"
//	@Success		200		{object}	response.Result[storecmn.FindResponseWithFullPagination[admin_response.GetUsersResponse]]
//	@Failure		422		{object}	apierror.Errors
//	@Failure		503		{object}	apierror.Errors
//	@Router			/v1/admin/users [get]
//	@Security		BearerAuth
func (h *Handler) getUsers(c fiber.Ctx) error {
	_, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &admin_request.GetUsersRequest{}
	if err := c.Bind().Query(req); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	users, err := h.services.AdminService.GetAllUsersFiltered(c.Context(), *req)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(users))
}

// banUser is a function to ban user
//
//	@Summary		Issue ban to user
//	@Description	Issue ban to user
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			string	query		admin_request.BanUserRequest	true	"BanUserRequest"
//	@Success		200		{object}	response.Result[admin_response.BanUserResponse]
//	@Failure		422		{object}	apierror.Errors
//	@Failure		503		{object}	apierror.Errors
//	@Router			/v1/admin/ban [patch]
//	@Security		BearerAuth
func (h *Handler) banUser(c fiber.Ctx) error {
	_, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &admin_request.BanUserRequest{}
	if err := c.Bind().Query(req); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	user, err := h.services.UserService.GetUserByID(c.Context(), req.UserID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	res, err := h.services.AdminService.BanUserByID(c.Context(), user.ID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(res))
}

// unbanUser is a function to unban user
//
//	@Summary		Unban user
//	@Description	Unban user
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			string	query		admin_request.BanUserRequest	true	"BanUserRequest"
//	@Success		200		{object}	response.Result[admin_response.UnbanUserResponse]
//	@Failure		422		{object}	apierror.Errors
//	@Failure		503		{object}	apierror.Errors
//	@Router			/v1/admin/unban [patch]
//	@Security		BearerAuth
func (h *Handler) unbanUser(c fiber.Ctx) error {
	_, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &admin_request.UnbanUserRequest{}
	if err := c.Bind().Query(req); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	user, err := h.services.UserService.GetUserByID(c.Context(), req.UserID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	res, err := h.services.AdminService.UnbanUserByID(c.Context(), user.ID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(res))
}

// deleteUserRole is a function to delete role from user
//
//	@Summary		Remove user role
//	@Description	Request to remove user role
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			string	query		admin_request.RemoveUserRoleRequest	true	"RemoveUserRoleRequest"
//	@Success		200		{object}	response.Result[string]
//	@Failure		422		{object}	apierror.Errors
//	@Failure		503		{object}	apierror.Errors
//	@Router			/v1/admin/role [delete]
//	@Security		BearerAuth
func (h *Handler) deleteUserRole(c fiber.Ctx) error {
	_, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &admin_request.RemoveUserRoleRequest{}
	if err := c.Bind().Query(req); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	user, err := h.services.UserService.GetUserByID(c.Context(), req.UserID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	exists, err := h.services.PermissionService.DeleteUserRole(user.ID.String(), req.UserRole)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	if !exists {
		return apierror.New().AddError(errors.New("user does not have this role")).SetHttpCode(fiber.StatusNotFound) // NoContent
	}

	newRoles, err := h.services.PermissionService.UserRoles(user.ID.String())
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}
	if len(newRoles) == 0 {
		_, _ = h.services.PermissionService.AddUserRole(user.ID.String(), models.UserRoleDefault)
	}

	return c.JSON(response.OkByMessage("Role successfully deleted"))
}

// addUserRole is a function to add role to user
//
//	@Summary		Add user role
//	@Description	Request to add user role
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			string	body		admin_request.AddUserRoleRequest	true	"AddUserRoleRequest"
//	@Success		200		{object}	response.Result[admin_response.AddUserRoleResponse]
//	@Failure		422		{object}	apierror.Errors
//	@Failure		503		{object}	apierror.Errors
//	@Router			/v1/admin/role [post]
//	@Security		BearerAuth
func (h *Handler) addUserRole(c fiber.Ctx) error {
	_, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &admin_request.AddUserRoleRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	user, err := h.services.UserService.GetUserByID(c.Context(), req.UserID)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	exists, err := h.services.PermissionService.AddUserRole(user.ID.String(), req.UserRole)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	if !exists {
		return apierror.New().AddError(errors.New("user already has this role")).SetHttpCode(fiber.StatusConflict)
	}

	userRoles, err := h.services.PermissionService.UserRoles(user.ID.String())
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(&admin_response.AddUserRoleResponse{
		UserID:    user.ID,
		UserRoles: userRoles,
	}))
}

// inviteUserWithRole is a function invite support user
//
// @Summary		Invite user with role
// @Description	Invite user with role
// @Tags			Admin
// @Accept			json
// @Produce		json
// @Param			register	body		admin_request.InviteUserWithRoleRequest	true	"Invite user with specific role"
// @Success		200			{object}	response.Result[string]
// @Failure		400			{object}	apierror.Errors
// @Failure		401			{object}	apierror.Errors
// @Failure		403			{object}	apierror.Errors
// @Failure		422			{object}	apierror.Errors
// @Failure		500			{object}	apierror.Errors
// @Router			/v1/admin/invite [post]
// @Security		BearerAuth
func (h Handler) inviteUser(c fiber.Ctx) error {
	u, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	dto := &admin_request.InviteUserWithRoleRequest{}
	if err := c.Bind().Body(dto); err != nil {
		return err
	}

	_, err = h.services.PermissionService.UserRoles(u.ID.String())
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	// TODO: add hierarchy permission check to prevent inviting users with higher roles

	res, err := h.services.AdminService.InviteUserWithRole(c.Context(), dto)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	feURL, err := h.services.SettingService.GetRootSetting(c.Context(), setting.MerchantDomain)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	payload := &notify.UserInviteData{
		Language: language.English.String(),
		Role:     dto.Role.String(),
		Link:     fmt.Sprintf("%s/dv-admin/auth/accept-invite/%s/%s", feURL.Value, res.Token, res.NewUser.Email),
	}
	go h.services.NotificationService.SendUser(c.Context(), models.NotificationTypeUserInvite, res.NewUser, payload, &models.NotificationArgs{
		UserID: &u.ID,
	})

	return c.JSON(response.OkByMessage("User invite sent successfully"))
}

func (h *Handler) initAdminRoutes(v1 fiber.Router) {
	admin := v1.Group("/admin",
		middleware.AuthMiddleware(h.services.AuthService),
		middleware.CasbinMiddleware(h.services.PermissionService, []models.UserRole{models.UserRoleRoot}))
	admin.Get("/users", h.getUsers)
	admin.Post("/invite", h.inviteUser)
	admin.Patch("/ban", h.banUser)
	admin.Patch("/unban", h.unbanUser)
	admin.Delete("/role", h.deleteUserRole)
	admin.Post("/role", h.addUserRole)
}
