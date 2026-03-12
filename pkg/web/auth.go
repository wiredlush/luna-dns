//go:build web

package web

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type session struct {
	username string
	expires  time.Time
}

type sessionStore struct {
	mu       sync.RWMutex
	sessions map[string]session
}

func newSessionStore() *sessionStore {
	return &sessionStore{sessions: make(map[string]session)}
}

func (s *sessionStore) create(username string, ttl time.Duration) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	id := hex.EncodeToString(b)

	s.mu.Lock()
	s.sessions[id] = session{username: username, expires: time.Now().Add(ttl)}
	s.mu.Unlock()

	return id, nil
}

func (s *sessionStore) valid(id string) bool {
	s.mu.RLock()
	sess, ok := s.sessions[id]
	s.mu.RUnlock()
	if !ok {
		return false
	}

	if time.Now().After(sess.expires) {
		s.delete(id)
		return false
	}

	return true
}

func (s *sessionStore) delete(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

const sessionTTL = 1 * time.Hour

func (s *Server) handleLogin(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	user, err := s.db.FindByUsername(req.Username)
	if err != nil || !user.CheckPassword(req.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	sid, err := s.sessions.create(user.Username, sessionTTL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create session"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "session",
		Value:    sid,
		Path:     "/",
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Expires:  time.Now().Add(sessionTTL),
	})

	return c.JSON(fiber.Map{"ok": true})
}

func (s *Server) handleLogout(c *fiber.Ctx) error {
	if sid := c.Cookies("session"); sid != "" {
		s.sessions.delete(sid)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Expires:  time.Now().Add(-1 * time.Hour),
	})

	return c.JSON(fiber.Map{"ok": true})
}

func (s *Server) authMiddleware(c *fiber.Ctx) error {
	sid := c.Cookies("session")
	if sid == "" || !s.sessions.valid(sid) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	return c.Next()
}
