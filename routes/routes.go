// Package routes is where we define the paths for the requests to take
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/blgolden/igendec/controllers"
)

// RegisterRoutes calls all the routes in the package to be registered
func RegisterRoutes(app *fiber.App, h *controllers.Handler) {
	Main(app, h)
	Create(app, h)
	Jobs(app, h)
	Profile(app, h)
}

// Main has all the default routes
func Main(app *fiber.App, h *controllers.Handler) {
	app.Get("/", h.Home)

	app.Get("/signin", h.SignIn)
	app.Post("/signin", h.SignIn)

	app.Get("/signout", h.SignOut)

	app.Get("/register", h.Register)
	app.Post("/register", h.Register)
}

// Create routes
func Create(app *fiber.App, h *controllers.Handler) {
	create := app.Group("/create")

	create.Get("/", h.Create)
	create.Post("/update", h.CreateUpdate)
	create.Post("/submit", h.CreateSubmit)
	create.Get("/build", h.CreateBuild)
	create.Post("/run", h.CreateRun)
}

// Profile routes
func Profile(app *fiber.App, h *controllers.Handler) {
	app.Get("/profile", h.Profile)

	app.Post("/updateprofile", h.UpdateProfile)

	app.Post("/updatepassword", h.UpdatePassword)
}

// Jobs routes
func Jobs(app *fiber.App, h *controllers.Handler) {
	jobs := app.Group("/jobs")
	jobs.Get("/", h.Jobs)
	jobs.Get("/info", h.JobsInfo)
	jobs.Get("/download", h.JobsDownload)
	jobs.Delete("/delete", h.JobsDelete)

	jobs.Get("/select", h.JobsSelect)
	jobs.Get("/select/database", h.JobsSelectDatabase)
	jobs.Post("/select/database/compare", h.JobsSelectDatabaseCompare)
}
