package pkg

import (
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/jmoiron/sqlx"
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
var setupClientTopic = make(chan registerClientTopic)

// Channel to broadcast a message to all clients in given topic
var broadcast = make(chan broadcastMessage)

//  Channel to unregister a client
var unregister = make(chan *websocket.Conn)

// Hub maintains the set of active clients and broadcasts messages to the
var admins = make(map[*websocket.Conn]*client)

// Channel to register "admins"
var registerAdmin = make(chan *websocket.Conn)

//  Channel to unregister a admin
var unregisterAdmin = make(chan *websocket.Conn)

// Channel to broadcast a message to all admins
var broadcastAdmin = make(chan []byte)

func RunHub() {
	for {
		select {
		case connection := <-register:
			clients[connection] = &client{}

		case rc := <-setupClientTopic:
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
			delete(clients, connection)
		}
	}
}

func RunAdminHub() {
	for {
		select {
		case connection := <-registerAdmin:
			admins[connection] = &client{}

		case payload := <-broadcastAdmin:
			for connection, c := range admins {
				go func(connection *websocket.Conn, c *client) { // send to each client in parallel so we don't block on a slow client
					c.mu.Lock()
					defer c.mu.Unlock()

					if c.isClosing {
						return
					}

					if err := connection.WriteMessage(websocket.TextMessage, payload); err != nil {
						c.isClosing = true
						log.Println("write error:", err)

						connection.WriteMessage(websocket.CloseMessage, []byte{})
						connection.Close()
						unregisterAdmin <- connection
					}
				}(connection, c)
			}

		case connection := <-unregisterAdmin:
			delete(admins, connection)
		}
	}
}

func RunWorker(conn *sqlx.DB) {
	query := `INSERT INTO logs (topic, message) VALUES (?, ?)`

	for {
		select {
		case payload := <-broadcast:

			if _, err := conn.Exec(query, payload.topic, payload.message); err != nil {
				log.Println("failed to insert into logs", err)
			}

		}
	}
}
