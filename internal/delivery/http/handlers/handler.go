package handlers

import (
	"errors"
	"reflect"

	"github.com/dv-net/dv-merchant/internal/delivery/middleware"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/binder"
	"github.com/google/uuid"
)

type Handler struct {
	services *service.Services
	logger   logger.Logger
}

func loadAuthUser(c fiber.Ctx) (*models.User, error) {
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return nil, apierror.New().AddError(errors.New("undefined user")).SetHttpCode(fiber.StatusUnauthorized)
	}
	return user, nil
}

func NewHandler(services *service.Services, logger logger.Logger) *Handler {
	return &Handler{
		services: services,
		logger:   logger,
	}
}

func (h *Handler) Init(api *fiber.App) {
	h.configureBinders()

	v1Admin := api.Group("api/v1/dv-admin")

	h.initAuthRoutes(v1Admin)

	h.initProcessingRoutes(v1Admin)

	h.initPublicSystemRoutes(v1Admin)

	securedV1Admin := v1Admin.Group(
		"/",
		middleware.AuthMiddleware(h.services.AuthService),
		middleware.CasbinMiddleware(
			h.services.PermissionService,
			[]models.UserRole{models.UserRoleDefault, models.UserRoleRoot, models.UserRoleSupport},
		),
	)

	h.initUserRoute(securedV1Admin)

	h.initStoreRoutes(securedV1Admin)

	h.initTransactionRoutes(securedV1Admin)

	h.initWalletRoutes(securedV1Admin)

	h.initReceiptRoutes(securedV1Admin)

	h.initSettingRoutes(securedV1Admin)

	h.init2faRoutes(securedV1Admin)

	h.initWithdrawalRoutes(securedV1Admin)

	h.initWithdrawalWalletsRoutes(securedV1Admin)

	h.initTransferRoutes(securedV1Admin)

	h.initAdminRoutes(v1Admin)

	h.initStatisticsRoutes(securedV1Admin)

	h.initExchangeRoutes(securedV1Admin)

	h.initWhRoutes(securedV1Admin)

	h.initDictionariesRoutes(securedV1Admin)

	h.initCurrencyRoutes(securedV1Admin)

	h.initLogsRoutes(securedV1Admin)

	h.initConsoleRoutes(securedV1Admin)

	h.initSearchRoutes(securedV1Admin)

	h.initNotificationRoutes(securedV1Admin)

	h.initAMLRoutes(securedV1Admin)
}

func (h *Handler) configureBinders() {
	binder.SetParserDecoder(binder.ParserConfig{
		IgnoreUnknownKeys: true,
		ParserType: []binder.ParserType{
			{
				CustomType: uuid.UUID{},
				Converter: func(value string) reflect.Value {
					if v, err := uuid.Parse(value); err == nil {
						return reflect.ValueOf(v)
					}

					return reflect.Value{}
				},
			},
		},
		ZeroEmpty: true,
	})
}
