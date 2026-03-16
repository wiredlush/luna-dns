//go:build web

package web

import (
	"bufio"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/wiredlush/luna-dns/pkg/blocklist"
)

const maxUploadSize = 15 << 20 // 15MB

func (s *Server) listBlocklist(c *fiber.Ctx) error {
	limit := 20
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(c.Query("offset")); err == nil && o >= 0 {
		offset = o
	}

	search := strings.TrimSpace(c.Query("search"))
	entries, total, err := s.db.SearchBlocklistEntries(search, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to list blocklist entries"})
	}

	return c.JSON(fiber.Map{"entries": entries, "total": total})
}

func (s *Server) createBlocklistEntry(c *fiber.Ctx) error {
	var req struct {
		Domain string `json:"domain"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	req.Domain = strings.TrimSpace(strings.ToLower(req.Domain))
	if req.Domain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Domain is required"})
	}

	entry, err := s.db.CreateBlocklistEntry(req.Domain)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Domain already exists"})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid domain format"})
	}

	s.syncBlocklist()

	return c.Status(fiber.StatusCreated).JSON(entry)
}

func (s *Server) deleteBlocklistEntry(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid id"})
	}

	if _, err := s.db.GetBlocklistEntry(uint(id)); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Entry not found"})
	}

	if err := s.db.DeleteBlocklistEntry(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete entry"})
	}

	s.syncBlocklist()

	return c.JSON(fiber.Map{"ok": true})
}

func (s *Server) clearBlocklist(c *fiber.Ctx) error {
	if err := s.db.DeleteAllBlocklistEntries(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to clear blocklist"})
	}

	s.syncBlocklist()

	return c.JSON(fiber.Map{"ok": true})
}

func (s *Server) uploadBlocklist(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File is required"})
	}

	if file.Size > maxUploadSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File too large (max 15MB)"})
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read file"})
	}
	defer f.Close()

	var domains []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		domain := blocklist.ParseLine(scanner.Text())
		if domain == "" {
			continue
		}
		domains = append(domains, strings.ToLower(domain))
	}

	if err := scanner.Err(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to parse file"})
	}

	if len(domains) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "No valid entries found in file"})
	}

	count, err := s.db.BulkCreateBlocklistEntries(domains)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to import entries"})
	}

	s.syncBlocklist()

	return c.JSON(fiber.Map{"imported": count})
}

func (s *Server) syncBlocklist() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.engine == nil || !s.engine.Running() {
		return
	}

	domains := s.loadBlocklistDomains()
	s.engine.SetBlocklist(domains)
}

func (s *Server) loadBlocklistDomains() []string {
	entries, err := s.db.ListAllBlocklistDomains()
	if err != nil {
		return nil
	}

	domains := make([]string, len(entries))
	for i, e := range entries {
		domains[i] = e.Domain
	}

	return domains
}
