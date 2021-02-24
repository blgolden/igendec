// Package session contains implementation of session library
// Uses singleton design approach so this is a stateful package
// Requires an init call to be made
package session

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/session/v2"
	"github.com/blgolden/igendec/users"
)

// Session errors
var (
	ErrNoSession = errors.New("no active session")
)

// Sess wraps an embedded fiber session so we can
// add some more specific methods to it
type Sess struct {
	*session.Session
}

// New creates a new session
func New() *Sess {
	return &Sess{session.New()}
}

// Store returns the store for the given context
func (s *Sess) Store(c *fiber.Ctx) *session.Store {
	return s.Get(c)
}

// New starts a session for a user
// A session contains the user struct for a user
func (s *Sess) New(c *fiber.Ctx, user *users.User) {
	store := s.Get(c)
	store.Set("username", user.Username)
	store.Save()
}

// Kill ends the session
func (s *Sess) Kill(c *fiber.Ctx) {
	s.Get(c).Destroy()
}

// Exists returns true if the context has a session
func (s *Sess) Exists(c *fiber.Ctx) bool {
	store := s.Get(c)
	return store.Get("username") != nil
}

// User returns the from the current session storage
func (s *Sess) User(c *fiber.Ctx) (*users.User, error) {
	store := s.Get(c)
	username := store.Get("username")
	if username == nil {
		return nil, ErrNoSession
	}
	return users.NewUser(username.(string)).Get()
}
