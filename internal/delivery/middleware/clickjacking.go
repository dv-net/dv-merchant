package middleware

import (
	"github.com/gofiber/fiber/v3"
)

func ClickjackingMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("X-Frame-Options", "SAMEORIGIN")
		c.Set("Content-Security-Policy", "frame-ancestors 'self'")
		return c.Next()
	}
}