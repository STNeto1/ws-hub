package pkg

import (
	"encoding/json"
	"log"

	"github.com/gofiber/contrib/websocket"
)

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
