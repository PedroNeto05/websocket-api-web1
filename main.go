package main

import (
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// Client represents a connected websocket client
type Client struct {
	ID   string
	Conn *websocket.Conn
}

// Message represents a message sent by a client
type Message struct {
	Id      string `json:"id"`
	UserID  string `json:"userId"`
	Content string `json:"content"`
}

var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan Message)
)

func main() {
	app := fiber.New()

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		defer func() {
			delete(clients, c)
			c.Close()
		}()

		clients[c] = true

		for {
			var msg Message
			if err := c.ReadJSON(&msg); err != nil {
				log.Println("Error reading json:", err)
				delete(clients, c)
				break
			}

			broadcast <- msg
		}
	}))

	go handleMessages()

	log.Fatal(app.Listen(":3001"))
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			if err := client.WriteJSON(msg); err != nil {
				log.Printf("Websocket error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
