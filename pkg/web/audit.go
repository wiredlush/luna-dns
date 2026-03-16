//go:build web

package web

import (
	"github.com/gofiber/fiber/v2"
)

func (s *Server) listAuditLogs(c *fiber.Ctx) error {
	logs, err := s.db.ListAuditLogs()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch audit logs"})
	}

	return c.JSON(logs)
}
