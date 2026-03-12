//go:build web

package web

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type userResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *Server) listUsers(c *fiber.Ctx) error {
	users, err := s.db.ListUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list users"})
	}

	res := make([]userResponse, len(users))
	for i, u := range users {
		res[i] = userResponse{ID: u.ID, Username: u.Username, CreatedAt: u.CreatedAt}
	}
	return c.JSON(res)
}

type createUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) createUser(c *fiber.Ctx) error {
	var req createUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username and password are required"})
	}

	user, err := s.db.CreateUser(req.Username, req.Password)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "username already exists"})
	}

	return c.Status(fiber.StatusCreated).JSON(userResponse{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
	})
}

func (s *Server) deleteUser(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	count, err := s.db.CountUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to check users"})
	}
	if count <= 1 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "cannot delete the last user"})
	}

	if err := s.db.DeleteUser(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete user"})
	}

	return c.JSON(fiber.Map{"ok": true})
}
