package frontend

import (
	"embed"

	"github.com/gofiber/fiber/v3"
)

//go:embed all:dist
var StaticDir embed.FS

const (
	dvAdminEntrypointPath   = "dist/dv-admin/index.html"
	dvPaymentEntrypointPath = "dist/pay/index.html"
)

// assetMaxAge is the Cache-Control max-age (in seconds) for content-hashed
// static assets (e.g. @dv.net-c12345.js), which are safe to cache for a long time.
const assetMaxAge = 31536000 // 1 year

func InitStaticFiles(app *fiber.App) {
	app.Get("/swagger.yaml", func(c fiber.Ctx) error {
		return c.SendFile("docs/swagger.yaml")
	})

	app.Get("/dv-admin/assets/:filename", func(c fiber.Ctx) error {
		return c.SendFile("dist/dv-admin/assets/"+c.Params("filename"), fiber.SendFile{
			FS:       StaticDir,
			Compress: true,
			MaxAge:   assetMaxAge,
		})
	})

	app.Get("/dv-admin/static/*", func(c fiber.Ctx) error {
		path := c.Params("*")
		return c.SendFile("dist/dv-admin/static/"+path, fiber.SendFile{
			FS:       StaticDir,
			Compress: true,
			MaxAge:   assetMaxAge,
		})
	})

	app.Get("/pay/static/*", func(c fiber.Ctx) error {
		path := c.Params("*")
		return c.SendFile("dist/pay/static/"+path, fiber.SendFile{
			FS:       StaticDir,
			Compress: true,
			MaxAge:   assetMaxAge,
		})
	})

	app.Get("/pay/assets/:filename", func(c fiber.Ctx) error {
		return c.SendFile("dist/pay/assets/"+c.Params("filename"), fiber.SendFile{
			FS:       StaticDir,
			Compress: true,
			MaxAge:   assetMaxAge,
		})
	})

	app.Get("/dv-admin/*", func(c fiber.Ctx) error {
		return c.SendFile(dvAdminEntrypointPath, fiber.SendFile{
			FS:       StaticDir,
			Compress: true,
		})
	})

	app.Get("/pay/*", func(c fiber.Ctx) error {
		return c.SendFile(dvPaymentEntrypointPath, fiber.SendFile{
			FS:       StaticDir,
			Compress: true,
		})
	})
}
