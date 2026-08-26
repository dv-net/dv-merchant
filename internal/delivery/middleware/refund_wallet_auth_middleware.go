package middleware

import (
	"github.com/dv-net/dv-merchant/internal/service/auth"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/hash"

	"github.com/gofiber/fiber/v3"
)

func RefundWalletAuthMiddleware(auth auth.IAuth) fiber.Handler {
	return func(c fiber.Ctx) error {
		rawToken := c.Get("X-Refund-Token")
		if rawToken == "" {
			return apierror.New().AddError(fiber.ErrUnauthorized).SetHttpCode(fiber.StatusUnauthorized)
		}

		hashedToken := hash.SHA256(rawToken)

		w, err := auth.GetWalletByToken(c.Context(), hashedToken)
		if err != nil {
			return apierror.New().AddError(fiber.ErrUnauthorized).SetHttpCode(fiber.StatusUnauthorized)
		}

		c.Locals("refund_wallet", w)
		return c.Next()
	}
}
