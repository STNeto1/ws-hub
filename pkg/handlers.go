package pkg

import (
	"bytes"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func HandleIndex(c *fiber.Ctx) error {
	return c.Render("index.html", fiber.Map{
		"connections": len(clients),
	})
}

func HandleConnectionsWs(c *websocket.Conn) {
	tmpl, ok := c.Locals("tmpl").(*Template)
	if !ok {
		log.Println("failed to get template")
		c.Close()
		return
	}

	defer func() {
		unregisterAdmin <- c
		c.Close()
	}()

	registerAdmin <- c

	var buf bytes.Buffer
	if err := sendConnections(tmpl, &buf, c, len(clients)); err != nil {
		return
	}

	lastClientCount := len(clients)

	for {
		if lastClientCount == len(clients) {
			time.Sleep(time.Second * 5)
			continue
		}
		lastClientCount = len(clients)

		buf.Reset()
		if err := sendConnections(tmpl, &buf, c, len(clients)); err != nil {
			return
		}

		time.Sleep(time.Second * 5)
	}

}

func sendConnections(tmpl *Template, buf *bytes.Buffer, conn *websocket.Conn, count int) error {
	if err := tmpl.Render(buf, "connections", fiber.Map{
		"connections": count,
	}); err != nil {
		log.Println("failed to render template", err)
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, buf.Bytes())

}
