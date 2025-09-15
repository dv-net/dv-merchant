package middleware

import (
	"strings"

	"github.com/dv-net/dv-merchant/internal/service/auth"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/hash"

	"github.com/gofiber/fiber/v3"
)

func AuthMiddleware(auth auth.IAuth) fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return apierror.New().AddError(fiber.ErrUnauthorized).SetHttpCode(fiber.StatusUnauthorized)
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return apierror.New().AddError(fiber.ErrUnauthorized).SetHttpCode(fiber.StatusUnauthorized)
		}

		token := tokenParts[1]
		hashedToken := hash.SHA256(token)

		user, err := auth.GetUserByToken(c.Context(), hashedToken)
		if err != nil {
			return apierror.New().AddError(fiber.ErrUnauthorized).SetHttpCode(fiber.StatusUnauthorized)
		}

		c.Locals("user", user)
		c.Locals("token_hash", hashedToken)
		return c.Next()
	}
}
