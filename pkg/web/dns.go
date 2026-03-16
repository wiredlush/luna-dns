//go:build web

package web

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/wiredlush/luna-dns/pkg/config"
	"github.com/wiredlush/luna-dns/pkg/engine"
)

type dnsConfigResponse struct {
	Addr     string `json:"addr"`
	Port     int    `json:"port"`
	Network  string `json:"network"`
	CacheTTL int64  `json:"cache_ttl"`
	Running  bool   `json:"running"`
}

type dnsConfigRequest struct {
	Addr     string `json:"addr"`
	Port     int    `json:"port"`
	Network  string `json:"network"`
	CacheTTL int64  `json:"cache_ttl"`
}

func (s *Server) getDnsConfig(c *fiber.Ctx) error {
	cfg, err := s.db.GetDnsConfig()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No dns configuration found"})
	}

	s.mu.Lock()
	running := s.engine != nil && s.engine.Running()
	s.mu.Unlock()

	return c.JSON(dnsConfigResponse{
		Addr:     cfg.Addr,
		Port:     cfg.Port,
		Network:  cfg.Network,
		CacheTTL: cfg.CacheTTL,
		Running:  running,
	})
}

func (s *Server) saveDnsConfig(c *fiber.Ctx) error {
	var req dnsConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	if req.Addr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Addr is required"})
	}
	if req.Port <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Port is required"})
	}
	if req.Network == "" {
		req.Network = "udp"
	}
	if req.Network != "udp" && req.Network != "tcp" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Network must be udp or tcp"})
	}
	if req.CacheTTL <= 0 {
		req.CacheTTL = 14400
	}

	if _, err := s.db.SaveDnsConfig(req.Addr, req.Port, req.Network, req.CacheTTL); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save configuration"})
	}

	s.mu.Lock()
	running := s.engine != nil && s.engine.Running()
	s.mu.Unlock()

	return c.JSON(dnsConfigResponse{Addr: req.Addr, Port: req.Port, Network: req.Network, CacheTTL: req.CacheTTL, Running: running})
}

func (s *Server) startDns(c *fiber.Ctx) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.engine != nil && s.engine.Running() {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Dns server is already running"})
	}

	if err := s.startEngine(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"ok": true, "running": true})
}

func (s *Server) restartDns(c *fiber.Ctx) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.engine != nil && s.engine.Running() {
		s.engine.Stop()
		s.engine = nil
	}

	if err := s.startEngine(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"ok": true, "running": true})
}

func (s *Server) loadRecords() []config.Host {
	records, err := s.db.ListDnsRecords()
	if err != nil {
		return nil
	}

	hosts := make([]config.Host, len(records))
	for i, r := range records {
		hosts[i] = config.Host{Host: r.Host, IP: r.IP}
	}

	return hosts
}

func (s *Server) loadForwarders() []config.DNS {
	forwarders, err := s.db.ListDnsForwarders()
	if err != nil {
		return nil
	}

	dnsServers := make([]config.DNS, len(forwarders))
	for i, f := range forwarders {
		dnsServers[i] = config.DNS{Addr: f.DialAddr(), Network: f.Network}
	}

	return dnsServers
}

func (s *Server) startEngine() error {
	cfg, err := s.db.GetDnsConfig()
	if err != nil {
		return fmt.Errorf("No dns configuration found, save a configuration first")
	}

	eng, err := engine.NewEngine(&config.Config{
		Addr:     cfg.ListenAddr(),
		Network:  cfg.Network,
		CacheTTL: cfg.CacheTTL,
		DNS:      s.loadForwarders(),
		Hosts:    s.loadRecords(),
	})
	if err != nil {
		return fmt.Errorf("Failed to create dns engine: %w", err)
	}

	if err := eng.StartBackground(); err != nil {
		return fmt.Errorf("Failed to start dns engine: %w", err)
	}

	s.engine = eng

	if domains := s.loadBlocklistDomains(); len(domains) > 0 {
		eng.SetBlocklist(domains)
	}

	s.db.SetDnsAutoStart(true)
	return nil
}

func (s *Server) stopDns(c *fiber.Ctx) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.engine == nil || !s.engine.Running() {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Dns server is not running"})
	}

	if err := s.engine.Stop(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to stop dns engine: " + err.Error()})
	}

	s.engine = nil
	s.db.SetDnsAutoStart(false)

	return c.JSON(fiber.Map{"ok": true, "running": false})
}
