package server

import (
	"log"

	"github.com/gorilla/websocket"
)

// Client представляет подключённого игрока
type Client struct {
	Conn *websocket.Conn // WebSocket-соединение
	Send chan []byte     // Канал для отправки сообщений
	Hub  *Hub            // Ссылка на общий Hub
}

// NewClient создаёт нового игрока
func NewClient(conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		Conn: conn,
		Send: make(chan []byte, 256),
		Hub:  hub,
	}
}

// ReadMessages читает сообщения от клиента
func (c *Client) ReadMessages() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("server: ошибка чтения сообщения: %v", err)
			break
		}
		log.Printf("server: получено сообщение: %s", message)
		c.Hub.Broadcast <- message // Рассылаем сообщение всем клиентам
	}
}

// WriteMessages отправляет сообщения клиенту
func (c *Client) WriteMessages() {
	defer c.Conn.Close()
	for message := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("server: ошибка отправки сообщения: %v", err)
			break
		}
	}
}
