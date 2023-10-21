package main

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/stneto1/ws-hub/pkg"
)

func main() {
	app := fiber.New(fiber.Config{})

	app.Use(logger.New())

	app.Get("/ws", websocket.New(pkg.HandleMessage, websocket.Config{}))

	go pkg.RunHub()

	app.Listen(":3000")
}
