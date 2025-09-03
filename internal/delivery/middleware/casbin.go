package middleware

import (
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/permission"

	"github.com/gofiber/fiber/v3"
)

func CasbinMiddleware(permSrv permission.IPermission, roles []models.UserRole) fiber.Handler {
	return permSrv.FiberMiddleware(roles...)
}
