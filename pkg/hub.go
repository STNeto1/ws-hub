package pkg

import (
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

// Hub maintains the set of active clients and broadcasts messages to the
var clients = make(map[*websocket.Conn]*client)

// Channel to register clients
var register = make(chan *websocket.Conn)

// Channel to register a topic to a client
var registerClient = make(chan registerClientTopic)

// Channel to broadcast a message to all clients in given topic
var broadcast = make(chan broadcastMessage)

//  Channel to unregister a client
var unregister = make(chan *websocket.Conn)

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
