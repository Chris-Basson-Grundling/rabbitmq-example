package main

import (
	"flag"
	"github.com/Chris-Basson-Grundling/rabbitmq-example/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"log"
)

var (
	port = flag.String("port", ":3000", "Port to listen on")
	prod = flag.Bool("prod", false, "Enable prefork in Production")
)

func main() {
	// Parse command-line flags
	flag.Parse()

	// Create fiber app
	app := fiber.New(fiber.Config{
		Prefork: *prod, // go run app.go -prod
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())

	v1 := app.Group("")

	// Setup static files
	app.Static("/", "./static/public")

	// Bind handlers
	v1.Get("/event", handlers.SseHandler)
	v1.Post("/consumer", handlers.AddConsumer)
	v1.Post("/producer", handlers.RunProducer)

	// Handle not founds
	app.Use(handlers.NotFound)

	// Listen on port 3000
	log.Fatal(app.Listen(*port))
}
