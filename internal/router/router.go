package router

import (
	"github.com/dv-net/dv-merchant/frontend"
	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/delivery/http/handlers"
	"github.com/dv-net/dv-merchant/internal/delivery/http/handlers/external"
	"github.com/dv-net/dv-merchant/internal/delivery/http/handlers/public"
	"github.com/dv-net/dv-merchant/internal/delivery/middleware"
	"github.com/dv-net/dv-merchant/internal/service"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/etag"
)

const dvLanding = "https://dv.net"

type Router struct {
	config   config.HTTPConfig
	services *service.Services
	logger   logger.Logger
}

func NewRouter(conf config.HTTPConfig, services *service.Services, logger logger.Logger) *Router {
	return &Router{
		config:   conf,
		services: services,
		logger:   logger,
	}
}

func (r *Router) Init(app *fiber.App) {
	app.Use(etag.New())

	if r.config.Cors.Enabled {
		corsConfig := cors.ConfigDefault
		corsConfig.AllowMethods = []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodHead,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodPatch,
			fiber.MethodOptions,
		}

		if len(r.config.Cors.AllowedOrigins) > 0 {
			corsConfig.AllowOrigins = r.config.Cors.AllowedOrigins
		}

		app.Use(cors.New(corsConfig))
	}

	app.Use(middleware.CacheControlMiddleware())
	app.Use(middleware.ClickjackingMiddleware())

	app.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("pong")
	})

	r.initAPI(app)
	r.landingRedirect(app)
	frontend.InitStaticFiles(app)
}

func (r *Router) initAPI(app *fiber.App) {
	handlerV1 := handlers.NewHandler(r.services, r.logger)
	externalV1 := external.NewHandler(r.services)
	publicV1 := public.NewHandler(r.services)

	externalV1.Init(app)
	publicV1.Init(app)

	handlerV1.Init(app)
}

func (r *Router) landingRedirect(app *fiber.App) {
	app.Get("/", func(c fiber.Ctx) error {
		info, err := r.services.SystemService.GetInfo(c.Context())
		if err != nil {
			r.logger.Errorw("error getting info", "error", err)
			return c.SendStatus(fiber.StatusBadRequest)
		}

		if info.Initialized {
			return c.Redirect().To(dvLanding)
		}

		return c.Redirect().To("dv-admin")
	})
}
