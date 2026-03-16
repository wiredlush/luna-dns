//go:build web

package web

import (
	"bufio"
	"log"
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

	if s.engine != nil && s.engine.Running() {
		s.engine.AddBlocklistEntry(req.Domain)
	}

	return c.Status(fiber.StatusCreated).JSON(entry)
}

func (s *Server) deleteBlocklistEntry(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid id"})
	}

	entry, err := s.db.GetBlocklistEntry(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Entry not found"})
	}

	if err := s.db.DeleteBlocklistEntry(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete entry"})
	}

	if s.engine != nil && s.engine.Running() {
		s.engine.RemoveBlocklistEntry(entry.Domain)
	}

	return c.JSON(fiber.Map{"ok": true})
}

func (s *Server) clearBlocklist(c *fiber.Ctx) error {
	if err := s.db.DeleteAllBlocklistEntries(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to clear blocklist"})
	}

	go s.syncBlocklist()

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

	go s.syncBlocklist()

	return c.JSON(fiber.Map{"imported": count})
}

func (s *Server) syncBlocklist() {
	s.mu.Lock()
	eng := s.engine
	s.mu.Unlock()

	if eng == nil || !eng.Running() {
		return
	}

	s.loadBlocklistIntoEngine()
}

func (s *Server) loadBlocklistIntoEngine() {
	log.Println("Loading blocklist entries from database...")
	builder := s.engine.NewBlocklistBuilder()
	s.db.IterateBlocklistDomains(1000, func(domains []string) error {
		builder.Add(domains)
		return nil
	})
	builder.Apply()
}
