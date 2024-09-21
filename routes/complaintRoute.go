package routes

import (
    "github.com/gofiber/fiber/v2"
	"rail-madad/controllers"
)

func SetupLoginRoutes(app *fiber.App) {
    app.Post("/submit-complaint", controllers.SubmitPNR)
    app.Get("/get-complaints/:pnr", controllers.GetComplaints)
}