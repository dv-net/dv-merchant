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

func InitStaticFiles(app *fiber.App) {
	app.Get("/swagger.yaml", func(c fiber.Ctx) error {
		return c.SendFile("docs/swagger.yaml")
	})

	app.Get("/dv-admin/assets/:filename", func(c fiber.Ctx) error {
		return c.SendFile("dist/dv-admin/assets/"+c.Params("filename"), fiber.SendFile{
			FS:            StaticDir,
			CacheDuration: 0,
			Download:      true,
		})
	})

	app.Get("/dv-admin/static/*", func(c fiber.Ctx) error {
		path := c.Params("*")
		return c.SendFile("dist/dv-admin/static/"+path, fiber.SendFile{
			FS: StaticDir,
		})
	})

	app.Get("/pay/static/*", func(c fiber.Ctx) error {
		path := c.Params("*")
		return c.SendFile("dist/pay/static/"+path, fiber.SendFile{
			FS: StaticDir,
		})
	})

	app.Get("/pay/assets/:filename", func(c fiber.Ctx) error {
		return c.SendFile("dist/pay/assets/"+c.Params("filename"), fiber.SendFile{
			FS:            StaticDir,
			Download:      true,
			CacheDuration: 0,
		})
	})

	app.Get("/dv-admin/*", func(c fiber.Ctx) error {
		return c.SendFile(dvAdminEntrypointPath, fiber.SendFile{
			FS:            StaticDir,
			Download:      false,
			CacheDuration: 0,
		})
	})

	app.Get("/pay/*", func(c fiber.Ctx) error {
		return c.SendFile(dvPaymentEntrypointPath, fiber.SendFile{
			FS:            StaticDir,
			Download:      false,
			CacheDuration: 0,
		})
	})
}
