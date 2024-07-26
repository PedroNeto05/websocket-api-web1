package main

import (
	"log"
	"os"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

// Client represents a connected websocket client
type Client struct {
	ID   string
	Conn *websocket.Conn
}

// Message represents a message sent by a client
type Message struct {
	Id       string `json:"id"`
	UserId   string `json:"userId"`
	UserName string `json:"userName"`
	Content  string `json:"content"`
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

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

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

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar o arquivo .env: %v", err)
	}

	host := os.Getenv("HOST")
	port := "3001"
	adrr := host + ":" + port
	log.Fatal(app.Listen(adrr))
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
