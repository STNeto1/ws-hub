package main

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/stneto1/ws-hub/pkg"
)

func main() {
	tmpl := pkg.CreateTemplate()
	app := fiber.New(fiber.Config{
		Views: tmpl,
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Cache-Control",
		AllowCredentials: true,
	}))
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("tmpl", tmpl)
		c.Locals("db", pkg.CreateDB())

		return c.Next()
	})

	app.Get("/", pkg.HandleIndex)
	app.Get("/ws", websocket.New(pkg.HandleMessage, websocket.Config{}))
	app.Get("/admin-ws", websocket.New(pkg.HandleConnectionsWs, websocket.Config{}))

	go pkg.RunHub()
	go pkg.RunAdminHub()
	go pkg.RunWorker(pkg.CreateDB())

	app.Listen(":3000")
}
