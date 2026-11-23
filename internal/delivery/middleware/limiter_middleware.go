package middleware

import (
	"strings"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

type LimiterOption func(c *limiter.Config)

func LimiterMiddleware(maxTries, exp int, options ...LimiterOption) fiber.Handler {
	cfg := &limiter.Config{
		Max:        maxTries,
		Expiration: time.Duration(exp) * time.Second,
	}

	for _, fn := range options {
		fn(cfg)
	}

	return limiter.New(*cfg)
}

func WithSlidingWindow(c *limiter.Config) {
	c.LimiterMiddleware = limiter.SlidingWindow{}
}

// WithAuthUserKey limits per authenticated user (requires AuthMiddleware before it),
// falling back to the client IP if no user is set.
func WithAuthUserKey(c *limiter.Config) {
	c.KeyGenerator = func(ctx fiber.Ctx) string {
		if usr, ok := ctx.Locals("user").(*models.User); ok && usr != nil {
			return "user:" + usr.ID.String()
		}
		return "ip:" + ctx.IP()
	}
}

func WithIPAndBodyEmailKey(c *limiter.Config) {
	c.KeyGenerator = func(ctx fiber.Ctx) string {
		var body struct {
			Email string `json:"email"`
		}
		_ = json.Unmarshal(ctx.Body(), &body)
		return ctx.IP() + "|" + strings.ToLower(strings.TrimSpace(body.Email))
	}
}
