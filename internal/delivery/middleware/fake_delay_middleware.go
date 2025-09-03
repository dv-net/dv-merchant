package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

func FakeDelayMiddleware(delay time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		time.Sleep(delay)
		return c.Next()
	}
}
