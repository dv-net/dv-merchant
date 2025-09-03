package handlers

import (
	"net/http"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/notification_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/notification_responses"
	"github.com/dv-net/dv-merchant/internal/delivery/middleware"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/notification_settings"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// Get available notifications list
//
//	@Summary		List available user notifications
//	@Description	List available user notifications
//	@Tags			Notifications
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[[]notification_responses.UserNotificationResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Router			/v1/notifications/list [get]
//	@Security		BearerAuth
func (h Handler) notificationsList(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	list, err := h.services.NotificationSettings.AvailableListByUser(c.Context(), usr)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.FromNotificationList(list)))
}

// Remove root setting
//
//	@Summary		List available user settings
//	@Description	List available user settings
//	@Tags			Notifications
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[string]
//	@Failure		401	{object}	apierror.Errors
//	@Router			/v1/notifications/{id} [put]
//	@Security		BearerAuth
func (h Handler) updateUserNotification(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	id, err := tools.ValidateUUID(c.Params("notification_id"))
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	dto := &notification_request.Update{}
	if err := c.Bind().Body(dto); err != nil {
		return err
	}

	if err := h.services.NotificationSettings.UpdateList(c.Context(), usr, []notification_settings.UpdateDTO{{
		ID:           id,
		EmailEnabled: dto.EmailEnabled,
		TgEnabled:    dto.TgEnabled,
	}}); err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("success"))
}

// Get available notifications list
//
//	@Summary		Update list user notifications
//	@Description	Update list user notifications
//	@Tags			Notifications
//	@Accept			json
//	@Produce		json
//	@Param			updateList	body		notification_request.UpdateList	true	"Update user notifications list"
//	@Success		200			{object}	response.Result[string]
//	@Failure		401			{object}	apierror.Errors
//	@Router			/v1/notifications/list/update [patch]
//	@Security		BearerAuth
func (h Handler) notificationsListUpdate(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &notification_request.UpdateList{}
	if err = c.Bind().Body(req); err != nil {
		return err
	}

	updateList := make([]notification_settings.UpdateDTO, 0, len(req.List))
	for _, v := range req.List {
		updateList = append(updateList, notification_settings.UpdateDTO{
			ID:           v.ID,
			EmailEnabled: v.EmailEnabled,
			TgEnabled:    v.TgEnabled,
		})
	}

	if err = h.services.NotificationSettings.UpdateList(c.Context(), usr, updateList); err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	return c.JSON(response.OkByMessage("success"))
}

// Test local SMTP server email sending
//
//	@Summary	Test local SMTP server email sending
//	@Tags		Notifications
//	@Accept		json
//	@Produce	json
//	@Param		testEmailRequest	body		notification_request.TestNotificationRequest	true	"Test email request"
//	@Success	200					{object}	response.Result[string]
//	@Failure	401					{object}	apierror.Errors
//	@Router		/v1/notifications/test [post]
//	@Security	BearerAuth
func (h Handler) testNotification(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &notification_request.TestNotificationRequest{}
	if err = c.Bind().Body(req); err != nil {
		return err
	}

	payload := &notify.UserTestEmailData{
		Language: usr.Language,
	}

	go h.services.NotificationService.SendSystemEmail(c.Context(), models.NotificationTypeUserTestEmail, req.Recipient, payload, &models.NotificationArgs{UserID: &usr.ID})

	return c.JSON(response.OkByMessage("success"))
}

// Get notification history
//
//	@Summary		Get user notification history
//	@Description	Get paginated notification history for the authenticated user with filtering options
//	@Tags			Notifications
//	@Accept			json
//	@Produce		json
//	@Param			page			query		uint32						false	"Page number"
//	@Param			page_size		query		uint32						false	"Page size"
//	@Param			types			query		[]models.NotificationType	false	"Notification types to filter by"
//	@Param			channels		query		[]models.DeliveryChannel	false	"Delivery channels to filter by"
//	@Param			created_from	query		string						false	"Filter from created date (RFC3339 format)"
//	@Param			created_to		query		string						false	"Filter to created date (RFC3339 format)"
//	@Param			sent_from		query		string						false	"Filter from sent date (RFC3339 format)"
//	@Param			sent_to			query		string						false	"Filter to sent date (RFC3339 format)"
//	@Param			ids				query		[]string					false	"Notification UUIDs to filter by"
//	@Param			destinations	query		[]string					false	"Destinations to filter by"
//	@Success		200				{object}	response.Result[storecmn.FindResponseWithFullPagination[notification_responses.NotificationHistoryResponse]]
//	@Failure		401				{object}	apierror.Errors
//	@Router			/v1/dv-admin/notifications/history [get]
//	@Security		BearerAuth
func (h Handler) getNotificationHistory(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &notification_request.GetNotificationHistoryRequest{}
	if err := c.Bind().Query(req); err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	history, err := h.services.NotificationService.GetHistoryByParams(c.Context(), usr, req)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(http.StatusBadRequest)
	}

	// Convert the response to use our custom response type
	convertedHistory := storecmn.FindResponseWithFullPagination[notification_responses.NotificationHistoryResponse]{
		Items:      converters.FromNotificationHistoryList(history.Items),
		Pagination: history.Pagination,
	}

	return c.JSON(response.OkByData(convertedHistory))
}

// Get available notification types
//
//	@Summary		Get available notification types
//	@Description	Get available notification types
//	@Tags			Notifications
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Result[notification_responses.NotificationTypeListResponse]
//	@Failure		401	{object}	apierror.Errors
//	@Router			/v1/dv-admin/notifications/types [get]
//	@Security		BearerAuth
func (h Handler) notificationsTypeList(c fiber.Ctx) error {
	_, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	types, err := h.services.NotificationService.GetTypesList(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(response.OkByData(converters.FromNotificationTypeList(types)))
}

func (h Handler) initNotificationRoutes(v1 fiber.Router) {
	notifications := v1.Group("/notifications")
	notifications.Put("/:notification_id", h.updateUserNotification)
	notifications.Get("/list", h.notificationsList)
	notifications.Patch("/list/update", h.notificationsListUpdate)
	notifications.Post("/test", h.testNotification)
	notifications.Get("/types", h.notificationsTypeList)

	history := notifications.Group("/history",
		middleware.CasbinMiddleware(
			h.services.PermissionService,
			[]models.UserRole{models.UserRoleRoot, models.UserRoleDefault},
		))
	history.Get("/", h.getNotificationHistory)
}
