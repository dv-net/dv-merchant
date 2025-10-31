package external

import (
	"errors"

	"github.com/dv-net/dv-merchant/internal/delivery/middleware"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/pkg/dbutils/pgerror"
	"github.com/jackc/pgx/v5"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	services *service.Services
}

func NewHandler(services *service.Services) *Handler {
	return &Handler{
		services: services,
	}
}

func loadAuthStore(c fiber.Ctx) (*models.Store, error) {
	store, ok := c.Locals("store").(*models.Store)
	if !ok {
		return nil, apierror.New().AddError(errors.New("undefined store")).SetHttpCode(fiber.StatusUnauthorized)
	}
	return store, nil
}

func (h *Handler) Init(api *fiber.App) {
	v1 := api.Group("api/v1")

	secured := v1.Group(
		"/external",
		middleware.StoreMiddleware(h.services.StoreService),
	)

	h.initWalletRoutes(secured)

	h.initStoreRoutes(secured)

	h.initWithdrawalRoutes(secured)

	h.initExchangeBalances(secured)

	h.initProcessingWalletBalances(secured)

	h.initInvoiceRoutes(secured)
}

func (h Handler) handleError(err error, modelName string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apierror.New().AddError(errors.New(modelName + " not found")).SetHttpCode(fiber.StatusNotFound)
	}
	var uniqueErr *pgerror.UniqueConstraintError
	if errors.As(err, &uniqueErr) {
		return apierror.New().AddError(uniqueErr).SetHttpCode(fiber.StatusUnprocessableEntity)
	}
	return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
}
