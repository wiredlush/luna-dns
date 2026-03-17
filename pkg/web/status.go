//go:build web

package web

import (
	"github.com/gofiber/fiber/v2"
)

type statusResponse struct {
	DNS dnsStatus `json:"dns"`
}

type dnsStatus struct {
	Running bool `json:"running"`
}

func (s *Server) handleStatus(c *fiber.Ctx) error {
	resp := statusResponse{
		DNS: s.probeDNS(),
	}
	return c.JSON(resp)
}

func (s *Server) probeDNS() dnsStatus {
	eng := s.getEngine()
	return dnsStatus{Running: eng != nil && eng.Running()}
}
