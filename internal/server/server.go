package server

import (
	"errors"

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
	code := fiber.StatusInternalServerError
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}
	var ae *apierror.Errors
	if errors.As(err, &ae) && ae.HttpCode != 0 {
		c.Status(ae.HttpCode)
		return c.JSON(ae)
	}

	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	return c.Status(code).SendString(err.Error())
}
