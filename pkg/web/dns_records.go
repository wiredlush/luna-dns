//go:build web

package web

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type recordResponse struct {
	ID   uint   `json:"id"`
	Host string `json:"host"`
	IP   string `json:"ip"`
}

type recordRequest struct {
	Host string `json:"host"`
	IP   string `json:"ip"`
}

func (s *Server) listRecords(c *fiber.Ctx) error {
	records, err := s.db.ListDnsRecords()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list records"})
	}

	res := make([]recordResponse, len(records))
	for i, r := range records {
		res[i] = recordResponse{ID: r.ID, Host: r.Host, IP: r.IP}
	}

	return c.JSON(res)
}

func (s *Server) createRecord(c *fiber.Ctx) error {
	var req recordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Host == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Host is required"})
	}
	if req.IP == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "IP is required"})
	}

	r, err := s.db.CreateDnsRecord(req.Host, req.IP)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create record"})
	}

	if s.engine != nil && s.engine.Running() {
		s.engine.AddHostEntry(req.Host, req.IP)
	}

	return c.Status(fiber.StatusCreated).JSON(recordResponse{
		ID: r.ID, Host: r.Host, IP: r.IP,
	})
}

func (s *Server) deleteRecord(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid id"})
	}

	record, err := s.db.GetDnsRecord(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Record not found"})
	}

	if err := s.db.DeleteDnsRecord(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete record"})
	}

	if s.engine != nil && s.engine.Running() {
		s.engine.RemoveHostEntry(record.Host)
	}

	return c.JSON(fiber.Map{"ok": true})
}
