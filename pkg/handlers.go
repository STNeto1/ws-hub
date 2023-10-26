package pkg

import (
	"bytes"
	"log"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
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
		return
	}

	conn, ok := c.Locals("db").(*sqlx.DB)
	if !ok {
		log.Println("failed to get db")
		return
	}

	defer func() {
		unregisterAdmin <- c
		c.Close()
	}()

	registerAdmin <- c

	lastClientCount := len(clients)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		var buf bytes.Buffer
		if err := sendConnections(tmpl, &buf, c, len(clients)); err != nil {
			return
		}

		for {
			select {
			case <-time.After(time.Second * 5):
				if lastClientCount == len(clients) {
					continue
				}
				lastClientCount = len(clients)

				buf.Reset()
				if err := sendConnections(tmpl, &buf, c, len(clients)); err != nil {
					return
				}
			}
		}
	}()

	go func() {
		defer wg.Done()

		lastTopics, err := getLatestTopics(conn)
		if err != nil {
			log.Println("failed to get topics", err)
			return
		}

		var buf bytes.Buffer
		if err := sendTopics(tmpl, &buf, c, &lastTopics); err != nil {
			return
		}

		for {
			select {
			case <-time.After(time.Second * 5):
				topics, err := getLatestTopics(conn)
				if err != nil {
					log.Println("failed to get topics", err)
					continue
				}

				if lastTopics == nil || !equalSlice(lastTopics, topics) {
					buf.Reset()
					if err := sendTopics(tmpl, &buf, c, &topics); err != nil {
						log.Println("failed to send topics to client", err)
						return
					}
				}

				lastTopics = topics
			}
		}
	}()

	wg.Wait()

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

func sendTopics(tmpl *Template, buf *bytes.Buffer, conn *websocket.Conn, topics *[]string) error {
	if err := tmpl.Render(buf, "topics", fiber.Map{
		"topics": topics,
	}); err != nil {
		log.Println("failed to render template", err)
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, buf.Bytes())
}

func getLatestTopics(db *sqlx.DB) ([]string, error) {
	var topics []string
	if err := db.Select(&topics, "SELECT DISTINCT topic FROM logs ORDER BY created_at DESC LIMIT 10"); err != nil {
		return nil, err
	}

	return topics, nil
}

func equalSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, val := range a {
		if b[i] != val {
			return false
		}
	}

	return true
}
