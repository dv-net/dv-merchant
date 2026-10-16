package public

import (
	"github.com/dv-net/dv-merchant/internal/service"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	services *service.Services
	logger   logger.Logger
}

func NewHandler(services *service.Services, logger logger.Logger) *Handler {
	return &Handler{
		services: services,
		logger:   logger,
	}
}

func (h *Handler) Init(api *fiber.App) {
	v1 := api.Group("api/v1")

	public := v1.Group("/public")

	h.initWalletRoutes(public)

	h.initUserRoutes(public)

	h.initCurrencyRoutes(public)

	h.initStoreRoutes(public)

	h.initMnemonic(public)

	h.initRefundRoutes(public)
}
