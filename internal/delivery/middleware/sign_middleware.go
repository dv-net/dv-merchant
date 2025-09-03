package middleware

import (
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
	"github.com/dv-net/dv-merchant/internal/tools/hash"

	"github.com/gofiber/fiber/v3"
)

func SignMiddleware(service setting.ISettingService) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		body := ctx.Body()

		clientKey, err := service.GetRootSetting(ctx.Context(), setting.ProcessingClientKey)
		if err != nil {
			return apierror.New().AddError(fiber.NewError(fiber.StatusTeapot, "Key not install on settings")).SetHttpCode(fiber.StatusTeapot)
		}

		hashBody := hash.SHA256Signature(body, clientKey.Value)
		signHeader := ctx.Get("X-Sign")

		if signHeader != hashBody {
			return apierror.New().AddError(fiber.NewError(fiber.StatusTeapot, "Invalid signature")).SetHttpCode(fiber.StatusTeapot)
		}
		return ctx.Next()
	}
}
