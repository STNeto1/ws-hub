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

	container := pkg.NewContainer(tmpl, pkg.CreateDB())

	app := fiber.New(fiber.Config{
		Views: tmpl,
	})
	app.Static("/assets", "./public")

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Cache-Control",
		AllowCredentials: true,
	}))

	app.Get("/", container.HandleIndex)
	app.Get("/ws", websocket.New(container.HandleMessage, websocket.Config{}))
	app.Get("/admin-ws", websocket.New(container.HandleConnectionsWs, websocket.Config{}))

	go pkg.RunHub()
	go pkg.RunAdminHub()
	go pkg.RunWorker(pkg.CreateDB())

	app.Listen(":3000")
}
