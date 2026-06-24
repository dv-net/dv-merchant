package server

import (
	"errors"
	"net/http"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/router"
	"github.com/dv-net/dv-merchant/internal/service"
	"github.com/dv-net/dv-merchant/internal/tools"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/gofiber/fiber/v3"
)

type Server struct {
	app    *fiber.App
	cfg    config.HTTPConfig
	logger logger.Logger
}

func NewServer(cfg config.HTTPConfig, services *service.Services, logger logger.Logger) *Server {
	app := fiber.New(fiber.Config{
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		StructValidator: tools.DefaultStructValidator(),
		ErrorHandler:    errorHandler,
	})

	router.NewRouter(cfg, services, logger).Init(app)

	return &Server{
		app:    app,
		cfg:    cfg,
		logger: logger,
	}
}

func (s *Server) Run() error {
	return s.app.Listen(":"+s.cfg.Port, fiber.ListenConfig{
		DisableStartupMessage: true,
	})
}

func (s *Server) Stop() error {
	return s.app.Shutdown()
}

func errorHandler(c fiber.Ctx, err error) error {
	var ae *apierror.Errors
	if errors.As(err, &ae) && ae.HttpCode != 0 {
		return c.Status(ae.HttpCode).JSON(ae)
	}

	var be *fiber.BindError
	if errors.As(err, &be) {
		return c.Status(fiber.StatusBadRequest).JSON(
			apierror.New().AddError(errors.New("invalid request")).SetHttpCode(fiber.StatusBadRequest),
		)
	}

	code := fiber.StatusInternalServerError
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	msg := errors.New(http.StatusText(code))
	return c.Status(code).JSON(apierror.New().AddError(msg).SetHttpCode(code))
}
