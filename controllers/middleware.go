package controllers

import (
	"github.com/gofiber/fiber/v2"
)

// Authorise Middleware:
// Authorises a user before going to any page
// Otherwise, renders the sign in page
func (h *Handler) Authorise(c *fiber.Ctx) error {
	if !isExceptionRoute(c.Path()) && !h.Session.Exists(c) {
		return c.Redirect("/signin")
	}
	return c.Next()
}

// IsExceptionRoute is true if this route doesn't need authentication
func isExceptionRoute(route string) bool {
	return route == "/signin" ||
		route == "/" ||
		route == "/register"
}
