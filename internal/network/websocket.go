package network

import (
	"log"
	"tetris/internal/interfaces" // Импортируем интерфейс GameInterface

	"github.com/gorilla/websocket"
)

type GameMessage struct {
	Type   string `json:"type"`
	Player int    `json:"player"`
	Data   struct {
		Direction string `json:"direction,omitempty"`
		Lines     int    `json:"lines,omitempty"`
	} `json:"data"`
}

// WebSocketClient управляет соединением с сервером
type WebSocketClient struct {
	Conn           *websocket.Conn
	Send           chan []byte
	PlayerID       int
	GameOver       bool
	GameResult     string
	MessageHandler *MessageHandler
	Game           interfaces.GameInterface
}

// NewWebSocketClient создаёт новое WebSocket-соединение
func NewWebSocketClient(url string, game interfaces.GameInterface) (*WebSocketClient, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	client := &WebSocketClient{
		Conn:           conn,
		Send:           make(chan []byte, 256),
		MessageHandler: nil,
		Game:           game, // Передаём игру
	}

	client.MessageHandler = NewMessageHandler(client)

	go client.readMessages()
	go client.writeMessages()

	return client, nil
}

func (c *WebSocketClient) readMessages() {
	defer c.Conn.Close()
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("network: ошибка чтения сообщения: %v", err)
			break
		}
		c.MessageHandler.HandleMessage(message)
	}
}

// writeMessages отправляет сообщения серверу
func (c *WebSocketClient) writeMessages() {
	defer c.Conn.Close()
	for message := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("network: ошибка отправки сообщения: %v", err)
			break
		}
	}
}

// SendMessage отправляет сообщение серверу
func (c *WebSocketClient) SendMessage(message []byte) {
	c.Send <- message
}
