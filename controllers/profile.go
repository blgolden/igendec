package controllers

import (
	"github.com/gofiber/fiber/v2"
)

// Profile renders the profile page
func (h *Handler) Profile(c *fiber.Ctx) error {
	user, err := h.Session.User(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}
	m := make(fiber.Map)
	m = user.ToMap(m)
	return h.RenderPrimary("profile", m, c)
}

// UpdateProfile updates a users profile
func (h *Handler) UpdateProfile(c *fiber.Ctx) error {
	// Get user struct
	user, err := h.Session.User(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}

	user.Firstname = c.FormValue("firstname")
	user.Surname = c.FormValue("surname")
	user.Email = c.FormValue("email")
	user.Location = c.FormValue("location")

	if err = user.Update(); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}
	return c.SendStatus(fiber.StatusOK)
}

// UpdatePassword updates the password for a user
func (h *Handler) UpdatePassword(c *fiber.Ctx) error {
	// Get user struct
	user, err := h.Session.User(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}

	// Compare the old password with current to make sure they match
	if err = user.ComparePassword(c.FormValue("oldpassword")); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Old password does not match current password")
	}

	// Check the passwords match - this is also done client side
	password := c.FormValue("newpassword")
	if password != c.FormValue("newpassword2") {
		return c.Status(fiber.StatusBadRequest).SendString("Passwords don't match")
	}

	// Try validate and hash password
	if err = user.ValidateAndHashPassword(password); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	// Save the user
	if err = user.Update(); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
	}

	return c.SendStatus(fiber.StatusOK)
}
