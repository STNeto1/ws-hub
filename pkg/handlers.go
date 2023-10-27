package pkg

import (
	"bytes"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type Container struct {
	Template *Template
	Db       *sqlx.DB
}

func NewContainer(tmpl *Template, db *sqlx.DB) *Container {
	return &Container{
		Template: tmpl,
		Db:       db,
	}
}

func (*Container) HandleIndex(c *fiber.Ctx) error {
	return c.Render("index.html", fiber.Map{
		"connections": len(clients),
	})
}

func (cnt *Container) HandleConnectionsWs(c *websocket.Conn) {
	defer func() {
		unregisterAdmin <- c
		c.Close()
	}()

	registerAdmin <- c

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()

		lastClientCount := len(clients)
		var buf bytes.Buffer
		if err := sendConnections(cnt.Template, &buf, c, len(clients)); err != nil {
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
				if err := sendConnections(cnt.Template, &buf, c, len(clients)); err != nil {
					return
				}
			}
		}
	}()

	go func() {
		defer wg.Done()

		lastTopics, err := getLatestTopics(cnt.Db)
		if err != nil {
			log.Println("failed to get topics", err)
			return
		}

		var buf bytes.Buffer
		if err := sendTopics(cnt.Template, &buf, c, &lastTopics); err != nil {
			return
		}

		for {
			select {
			case <-time.After(time.Second * 5):
				topics, err := getLatestTopics(cnt.Db)
				if err != nil {
					log.Println("failed to get topics", err)
					continue
				}

				if !slices.Equal(lastTopics, topics) {
					buf.Reset()
					if err := sendTopics(cnt.Template, &buf, c, &topics); err != nil {
						log.Println("failed to send topics to client", err)
						continue
					}

					lastTopics = topics
				}

			}
		}
	}()

	go func() {
		defer wg.Done()

		lastMessages, err := getLatestMessages(cnt.Db)
		if err != nil {
			log.Println("failed to get topics", err)
			return
		}

		var buf bytes.Buffer
		if err := sendMessages(cnt.Template, &buf, c, &lastMessages); err != nil {
			return
		}

		for {
			select {
			case <-time.After(time.Second * 5):
				messages, err := getLatestMessages(cnt.Db)
				if err != nil {
					log.Println("failed to get topics", err)
					continue
				}

				if !slices.EqualFunc(lastMessages, messages, LogPredicate) {
					buf.Reset()
					if err := sendMessages(cnt.Template, &buf, c, &messages); err != nil {
						log.Println("failed to send topics to client", err)
						continue
					}

					lastMessages = messages
				}

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

func sendMessages(tmpl *Template, buf *bytes.Buffer, conn *websocket.Conn, payload *[]Log) error {
	if err := tmpl.Render(buf, "messages", fiber.Map{
		"messages": payload,
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

func getLatestMessages(db *sqlx.DB) ([]Log, error) {
	var messages []Log
	if err := db.Select(&messages, "SELECT * FROM logs ORDER BY created_at DESC LIMIT 10"); err != nil {
		return nil, err
	}

	return messages, nil
}
