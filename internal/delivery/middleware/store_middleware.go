package middleware

import (
	"errors"

	"github.com/dv-net/dv-merchant/internal/service/store"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"

	"github.com/gofiber/fiber/v3"
)

func StoreMiddleware(storeSvc store.IStore, whitelistSvc store.IStoreWhitelist) fiber.Handler {
	return func(c fiber.Ctx) error {
		type requestBody struct {
			APIKey string `json:"api_key"`
		}
		rBody := &requestBody{}
		_ = c.Bind().Body(rBody)
		authHeader := c.Get("X-Api-Key")
		authParameter := c.Query("api_key")
		if authHeader == "" && authParameter == "" && rBody.APIKey == "" {
			return apierror.New().AddError(fiber.ErrUnauthorized).SetHttpCode(fiber.StatusUnauthorized)
		}

		var key string
		switch {
		case authHeader != "":
			key = authHeader
		case authParameter != "":
			key = authParameter
		case rBody.APIKey != "":
			key = rBody.APIKey
		}

		authStore, err := storeSvc.GetStoreByStoreAPIKey(c.Context(), key)
		if err != nil {
			return apierror.New().AddError(fiber.ErrUnauthorized).SetHttpCode(fiber.StatusUnauthorized)
		}

		allowed, err := whitelistSvc.IsStoreIPAllowed(c.Context(), authStore.ID, c.IP())
		if err != nil {
			return apierror.New().AddError(errors.New("failed to check store whitelist")).SetHttpCode(fiber.StatusInternalServerError)
		}
		if !allowed {
			return apierror.New().AddError(errors.New("ip is not whitelisted")).SetHttpCode(fiber.StatusForbidden)
		}

		c.Locals("store", authStore)
		return c.Next()
	}
}
