package config

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

var Store *session.Store

// SetupSessionStore initializes the session store
func SetupSessionStore() {
	Store = session.New(session.Config{
		CookieName:     "session_id",
		Expiration:     24 * time.Hour,
		CookieHTTPOnly: true,
		CookiePath:     "/",
		CookieSameSite: "Lax",
	})
}

// GetSession retrieves the session for the given context
func GetSession(c *fiber.Ctx) (*session.Session, error) {
	return Store.Get(c)
}
