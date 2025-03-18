package server

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Настройки WebSocket-соединения
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket обрабатывает входящие WebSocket-подключения
func HandleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("server: ошибка при обновлении соединения: %v", err)
		return
	}

	client := NewClient(conn, hub)
	hub.Register <- client

	go client.ReadMessages()
	go client.WriteMessages()
}
