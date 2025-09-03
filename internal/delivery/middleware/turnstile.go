package middleware

import (
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/pkg/turnstile"

	"github.com/gofiber/fiber/v3"
)

func TurnstileMiddleware(verifier turnstile.Verifier) fiber.Handler {
	return func(c fiber.Ctx) error {
		type requestBody struct {
			CfTurnstile string `json:"cf-turnstile-response"` //nolint:tagliatelle
		}

		rBody := &requestBody{}
		_ = c.Bind().Body(rBody)

		if err := verifier.Verify(c.Context(), c.IP(), rBody.CfTurnstile); err != nil {
			return apierror.New().AddError(err).SetHttpCode(fiber.StatusUnauthorized)
		}

		return c.Next()
	}
}
