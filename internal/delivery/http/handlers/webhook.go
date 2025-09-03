package handlers

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/webhook"
	"github.com/dv-net/dv-merchant/internal/models"

	// Blank import for swaggen
	_ "github.com/dv-net/dv-merchant/internal/delivery/http/responses/webhook_response"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/converters"
	"github.com/dv-net/dv-merchant/internal/tools/response"

	"github.com/gofiber/fiber/v3"
)

// history get webhook history
//
//	@Summary		Get webhooks history by stores
//	@Description	This endpoint returns webhook history
//	@Tags			WebhookHistory
//	@Accept			json
//	@Produce		json
//	@Param			string	query		webhook.GetWhHistoryRequest							true	"GetWhHistoryRequest"
//	@Success		200		{object}	response.Result[webhook_response.WhHistoryResponse]	"Successful operation"
//	@Failure		401		{object}	apierror.Errors										"Unauthorized"
//	@Failure		500		{object}	apierror.Errors										"Internal Server Error"
//	@Router			/v1/dv-admin/webhook/history [get]
//	@Security		BearerAuth
func (h *Handler) history(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &webhook.GetWhHistoryRequest{}
	if err = c.Bind().Query(req); err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	res, err := h.services.WebHookService.GetHistory(c.Context(), *usr, req.StoreUUIDs, req.Page, req.PageSize)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.FromHistoryDataListToResponseList(res)))
}

// sendTestWh get send mock webhook
//
//	@Summary		Emulate webhook to store
//	@Description	This endpoint emulates store webhook
//	@Tags			StoreWebhook
//	@Accept			json
//	@Produce		json
//	@Param			register	body		webhook.TestWebhookRequest							true	"TestWebhookRequest"
//	@Success		200			{object}	response.Result[webhook_response.SendTestResult]	"Successful operation"
//	@Failure		401			{object}	apierror.Errors										"Unauthorized"
//	@Failure		500			{object}	apierror.Errors										"Internal Server Error"
//	@Router			/v1/dv-admin/webhook/send-test [post]
//	@Security		BearerAuth
func (h *Handler) sendTestWh(c fiber.Ctx) error {
	usr, err := loadAuthUser(c)
	if err != nil {
		return err
	}

	req := &webhook.TestWebhookRequest{}
	if err := c.Bind().Body(req); err != nil {
		return err
	}

	res, err := h.services.StoreWebhooksService.SendMockWebhook(
		c.Context(),
		usr,
		req.WhID,
		models.WebhookEvent(req.EventType),
	)
	if err != nil {
		return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
	}

	return c.JSON(response.OkByData(converters.FromSendResultDtoToResponse(res)))
}

func (h *Handler) initWhRoutes(router fiber.Router) {
	wh := router.Group("/webhook")
	wh.Get("/history", h.history)
	wh.Post("/send-test", h.sendTestWh)
}
