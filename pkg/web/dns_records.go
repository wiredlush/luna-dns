//go:build web

package web

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/wiredlush/luna-dns/pkg/config"
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

	s.syncRecords()

	return c.Status(fiber.StatusCreated).JSON(recordResponse{
		ID: r.ID, Host: r.Host, IP: r.IP,
	})
}

func (s *Server) deleteRecord(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid id"})
	}

	if _, err := s.db.GetDnsRecord(uint(id)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Record not found"})
	}

	if err := s.db.DeleteDnsRecord(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete record"})
	}

	s.syncRecords()

	return c.JSON(fiber.Map{"ok": true})
}

func (s *Server) syncRecords() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.engine == nil || !s.engine.Running() {
		return
	}

	records, err := s.db.ListDnsRecords()
	if err != nil {
		return
	}

	hosts := make([]config.Host, len(records))
	for i, r := range records {
		hosts[i] = config.Host{Host: r.Host, IP: r.IP}
	}
	s.engine.SetHosts(hosts)
}
