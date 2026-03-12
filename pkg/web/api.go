//go:build web

package web

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func registerAPI(app *fiber.App, _ *gorm.DB) {
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
}
