package pkg

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type errorMessagePayload struct {
	Error string `json:"error"`
}
type infoMessagePayload struct {
	Message string `json:"message"`
}

type client struct {
	isClosing bool
	mu        sync.Mutex
	topic     string
}

type registerClientTopic struct {
	conn  *websocket.Conn
	topic string
}

type broadcastMessage struct {
	topic   string
	message []byte
}

var clients = make(map[*websocket.Conn]*client)
var register = make(chan *websocket.Conn)
var registerClient = make(chan registerClientTopic)
var broadcast = make(chan broadcastMessage)
var unregister = make(chan *websocket.Conn)
var topics = make(map[string][]*websocket.Conn)

func RunHub() {
	for {
		select {
		case connection := <-register:
			clients[connection] = &client{}

		case rc := <-registerClient:
			client := clients[rc.conn]
			if client == nil {
				continue
			}
			client.topic = rc.topic

		case payload := <-broadcast:
			// Send the message to all clients
			for connection, c := range clients {
				if c.topic != payload.topic {
					continue
				}

				go func(connection *websocket.Conn, c *client) { // send to each client in parallel so we don't block on a slow client
					c.mu.Lock()
					defer c.mu.Unlock()

					if c.isClosing {
						return
					}

					if err := connection.WriteMessage(websocket.TextMessage, payload.message); err != nil {
						c.isClosing = true
						log.Println("write error:", err)

						connection.WriteMessage(websocket.CloseMessage, []byte{})
						connection.Close()
						unregister <- connection
					}
				}(connection, c)
			}

		case connection := <-unregister:
			// Remove the client from the hub
			delete(clients, connection)
		}
	}
}

type topicMessage struct {
	Topic string `json:"topic"`
}

func HandleMessage(c *websocket.Conn) {
	var topic topicMessage
	var msg json.RawMessage

	// When the function returns, unregister the client and close the connection
	defer func() {
		unregister <- c
		c.Close()
	}()

	// Register the client
	register <- c

	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read error:", err)
			}

			break
		}

		if messageType == websocket.TextMessage {
			if topic.Topic == "" {
				err := json.Unmarshal(message, &topic)
				if err != nil {
					sendErrorMessage(c, "Invalid payload")
					continue
				}

				if topic.Topic == "" {
					sendErrorMessage(c, "Invalid topic")
					continue
				}

				registerClient <- registerClientTopic{conn: c, topic: topic.Topic}
				sendInfoMessage(c, "connected")
				continue
			}

			err := json.Unmarshal(message, &msg)
			if err != nil {
				log.Println("error unmarshalling message:", err)

				// TODO: handle error, but for now just continue
				continue
			}

			broadcast <- broadcastMessage{topic: topic.Topic, message: message}
		}
	}
}

func sendErrorMessage(c *websocket.Conn, message string) {
	msg := errorMessagePayload{
		message,
	}

	jsonData, _ := json.Marshal(msg)

	if err := c.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		log.Println("write error:", err)
	}
}
func sendInfoMessage(c *websocket.Conn, message string) {
	msg := infoMessagePayload{
		message,
	}

	jsonData, _ := json.Marshal(msg)

	if err := c.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		log.Println("write error:", err)
	}
}
