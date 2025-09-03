package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

func CacheControlMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		if strings.Contains(c.Path(), "api/") {
			c.Set("Cache-Control", "no-store")
			return c.Next()
		}
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		c.Set("last-modified", time.Now().UTC().Format(http.TimeFormat))

		c.Request().Header.Del("If-modified-since")
		return c.Next()
	}
}
