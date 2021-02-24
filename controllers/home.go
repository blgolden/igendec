package controllers

import (
	"github.com/gofiber/fiber/v2"
)

// Home is the endpoint for getting the home page
func (h *Handler) Home(c *fiber.Ctx) error {
	return h.RenderPrimary("home", nil, c)
}
