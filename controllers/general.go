package controllers

import (
	"errors"
	"regexp"

	"github.com/blgolden/igendec/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/blgolden/controllers/session"
	"github.com/blgolden/igendec/users"
)

// Error strings to send to client
var (
	InternalServerErrorString = "Something went wrong, please try again later"
	ErrInternalServer         = errors.New(InternalServerErrorString)
)

// NameRegex is a regular expression that only allows alphanumberic characters, '-', and '_'
var NameRegex = regexp.MustCompile("^[a-zA-Z0-9_-]+$")

// Handler has the endpoint methods on it to allow
// better management of dependencies without global state
type Handler struct {
	UserBlacklist map[string]struct{}
	Session       *session.Sess
}

// NewHandler returns a new handler object
func NewHandler() *Handler {
	return &Handler{
		UserBlacklist: make(map[string]struct{}),
		Session:       session.New(),
	}
}

// NotFound is where the stack ends up if the request does not have an endpoint
func (h *Handler) NotFound(c *fiber.Ctx) error {
	c.Status(fiber.StatusNotFound).Render("errors/notfound", nil, "layout/primary")
	return nil
}

// RenderPrimary will render the given file with primary layout
func (h *Handler) RenderPrimary(htmlFile string, m fiber.Map, c *fiber.Ctx) error {
	if m == nil {
		m = make(map[string]interface{})
	}
	m["Authorised"] = h.Session.Exists(c)

	return c.Status(fiber.StatusOK).Render(htmlFile, m, "layout/primary")
}

// SignIn renders signin
func (h *Handler) SignIn(c *fiber.Ctx) error {
	switch c.Method() {
	case fiber.MethodPost:
		user, err := users.NewUser(c.FormValue("username")).Get()

		// Username doesn't exist
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Username or password incorrect")
		}

		// Password doesn't match
		if err = user.ComparePassword(c.FormValue("password")); err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Username or password incorrect")
		}

		// Check if they are on the blacklist
		if _, ok := h.UserBlacklist[user.Username]; ok {
			return c.Status(fiber.StatusUnauthorized).SendString("Not authenticated")
		}

		// Create h.Session for user
		h.Session.New(c, user)

		// Go to home page
		return h.Home(c)

	default:
		return h.RenderPrimary("signin", nil, c)
	}
}

// SignOut ends a h.Session for a user and redirects to home page
func (h *Handler) SignOut(c *fiber.Ctx) error {
	h.Session.Kill(c)
	return h.Home(c)
}

// Register expects a post request and will create a new user
func (h *Handler) Register(c *fiber.Ctx) error {
	switch c.Method() {
	case fiber.MethodPost:
		user := users.NewUser(c.FormValue("username"))

		// Check the user doesn't already exist
		if user.Exists() {
			return c.Status(fiber.StatusUnprocessableEntity).SendString("Username already exists")
		}

		// Check the passwords match
		if c.FormValue("password") != c.FormValue("password2") {
			return c.Status(fiber.StatusBadRequest).SendString("Passwords don't match")
		}

		// Validate the password
		if err := user.ValidateAndHashPassword(c.FormValue("password")); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		// Set details
		user.Firstname = c.FormValue("firstname")
		user.Surname = c.FormValue("surname")
		user.Email = c.FormValue("email")
		user.Location = c.FormValue("location")

		// Save to server
		if err := user.Save(); err != nil {
			logger.Warn("Failed to save user with error:%s", err)
			return c.Status(fiber.StatusInternalServerError).SendString(InternalServerErrorString)
		}

		// Create h.Session
		h.Session.New(c, user)
		return c.SendStatus(fiber.StatusOK)

	case fiber.MethodGet:
		m := make(fiber.Map)
		m["ShowRegister"] = true
		return h.RenderPrimary("signin", m, c)

	}
	return nil
}
