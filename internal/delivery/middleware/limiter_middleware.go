package middleware

import (
	"time"

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
