//go:build web

package web

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) listAuditLogs(c *fiber.Ctx) error {
	limit := 50
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		limit = l
	}

	logs, err := s.db.ListAuditLogs(limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch audit logs"})
	}

	return c.JSON(logs)
}
