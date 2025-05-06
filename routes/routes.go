package routes

import (
	"go-rest-api/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, authController *controllers.AuthController) {
	// Auth routes
	app.Post("/api/register", authController.Register)
	app.Post("/api/login", authController.Login)
	app.Get("/api/user", authController.User)
	app.Post("/api/logout", authController.Logout)

	app.Get("/auth/microsoft", authController.MicrosoftLogin)
	app.Get("/auth/microsoft/callback", authController.MicrosoftCallback)

	app.Get("/login", authController.LoginPage)

	// Posts routes - convert your handlers to work with Fiber
	app.Get("/posts", func(c *fiber.Ctx) error {

		return c.SendString("Posts endpoint")
	})

	app.Get("/posts/:id", func(c *fiber.Ctx) error {
		return c.SendString("Get post endpoint")
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to the API")
	})
}
