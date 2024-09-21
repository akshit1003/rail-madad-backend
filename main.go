package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"rail-madad/routes"
	_ "rail-madad/config"
)

func main() {
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		err := godotenv.Load()
		if err != nil {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}

	app := fiber.New()

	routes.SetupLoginRoutes(app)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, Fiber!")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" 
	}

	log.Fatal(app.Listen(":" + port))
}