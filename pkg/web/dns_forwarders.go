//go:build web

package web

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/wiredlush/luna-dns/pkg/config"
)

type forwarderResponse struct {
	ID      uint   `json:"id"`
	Addr    string `json:"addr"`
	Port    int    `json:"port"`
	Network string `json:"network"`
}

type forwarderRequest struct {
	Addr    string `json:"addr"`
	Port    int    `json:"port"`
	Network string `json:"network"`
}

func (s *Server) listForwarders(c *fiber.Ctx) error {
	forwarders, err := s.db.ListDnsForwarders()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list forwarders"})
	}

	res := make([]forwarderResponse, len(forwarders))
	for i, f := range forwarders {
		res[i] = forwarderResponse{ID: f.ID, Addr: f.Addr, Port: f.Port, Network: f.Network}
	}
	return c.JSON(res)
}

func (s *Server) createForwarder(c *fiber.Ctx) error {
	var req forwarderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Addr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Addr is required"})
	}
	if req.Port <= 0 || req.Port > 65535 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Valid port is required"})
	}
	if req.Network != "udp" && req.Network != "tcp" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Network must be udp or tcp"})
	}

	f, err := s.db.CreateDnsForwarder(req.Addr, req.Port, req.Network)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create forwarder"})
	}

	s.syncForwarders()

	return c.Status(fiber.StatusCreated).JSON(forwarderResponse{
		ID: f.ID, Addr: f.Addr, Port: f.Port, Network: f.Network,
	})
}

func (s *Server) deleteForwarder(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid id"})
	}

	if _, err := s.db.GetDnsForwarder(uint(id)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Forwarder not found"})
	}

	if err := s.db.DeleteDnsForwarder(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete forwarder"})
	}

	s.syncForwarders()

	return c.JSON(fiber.Map{"ok": true})
}

func (s *Server) syncForwarders() {
	eng := s.getEngine()
	if eng == nil || !eng.Running() {
		return
	}

	forwarders, err := s.db.ListDnsForwarders()
	if err != nil {
		return
	}

	dnsServers := make([]config.DNS, len(forwarders))
	for i, f := range forwarders {
		dnsServers[i] = config.DNS{Addr: f.DialAddr(), Network: f.Network}
	}

	eng.SetForwarders(dnsServers)
}
