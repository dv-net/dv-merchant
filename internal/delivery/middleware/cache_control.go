package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

// cachedStaticPathPrefixes are content-hashed static assets served with a
// long-lived Cache-Control by frontend.InitStaticFiles; this middleware must
// not overwrite their caching headers.
var cachedStaticPathPrefixes = []string{
	"/dv-admin/assets/",
	"/dv-admin/static/",
	"/pay/assets/",
	"/pay/static/",
}

func CacheControlMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		if strings.Contains(c.Path(), "api/") {
			c.Set("Cache-Control", "no-store")
			return c.Next()
		}
		for _, prefix := range cachedStaticPathPrefixes {
			if strings.HasPrefix(c.Path(), prefix) {
				return c.Next()
			}
		}
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		c.Set("last-modified", time.Now().UTC().Format(http.TimeFormat))

		c.Request().Header.Del("If-modified-since")
		return c.Next()
	}
}
