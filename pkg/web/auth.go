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
	username  string
	ip        string
	createdAt time.Time
	expires   time.Time
}

type sessionInfo struct {
	Username  string    `json:"username"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"created_at"`
	Expires   time.Time `json:"expires"`
	Current   bool      `json:"current"`
}

type sessionStore struct {
	mu       sync.RWMutex
	sessions map[string]session
}

func newSessionStore() *sessionStore {
	return &sessionStore{sessions: make(map[string]session)}
}

func (s *sessionStore) create(username, ip string, ttl time.Duration) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	id := hex.EncodeToString(b)

	now := time.Now()
	s.mu.Lock()
	s.sessions[id] = session{username: username, ip: ip, createdAt: now, expires: now.Add(ttl)}
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

func (s *sessionStore) username(id string) string {
	s.mu.RLock()
	sess, ok := s.sessions[id]
	s.mu.RUnlock()
	if !ok {
		return ""
	}
	return sess.username
}

func (s *sessionStore) delete(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

func (s *sessionStore) list(currentSID string) []sessionInfo {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	var result []sessionInfo
	for id, sess := range s.sessions {
		if now.After(sess.expires) {
			delete(s.sessions, id)
			continue
		}
		result = append(result, sessionInfo{
			Username:  sess.username,
			IP:        sess.ip,
			CreatedAt: sess.createdAt,
			Expires:   sess.expires,
			Current:   id == currentSID,
		})
	}
	return result
}

func (s *sessionStore) deleteByUsername(username, exceptSID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, sess := range s.sessions {
		if sess.username == username && id != exceptSID {
			delete(s.sessions, id)
		}
	}
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
		s.db.LogAudit(req.Username, "login_failed", "invalid credentials", c.IP())
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	sid, err := s.sessions.create(user.Username, c.IP(), sessionTTL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create session"})
	}

	s.db.LogAudit(user.Username, "login", "", c.IP())

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
	sid := c.Cookies("session")
	if sid != "" {
		username := s.sessions.username(sid)
		s.sessions.delete(sid)
		if username != "" {
			s.db.LogAudit(username, "logout", "", c.IP())
		}
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

func (s *Server) listSessions(c *fiber.Ctx) error {
	return c.JSON(s.sessions.list(c.Cookies("session")))
}

func (s *Server) logoutAll(c *fiber.Ctx) error {
	sid := c.Cookies("session")
	username := s.sessions.username(sid)
	if username == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	s.sessions.deleteByUsername(username, sid)
	s.db.LogAudit(username, "logout_all", "", c.IP())

	return c.JSON(fiber.Map{"ok": true})
}

func (s *Server) authMiddleware(c *fiber.Ctx) error {
	sid := c.Cookies("session")
	if sid == "" || !s.sessions.valid(sid) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	return c.Next()
}
