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
	if len(hub.Clients) >= 2 {
		http.Error(w, "Сервер переполнен", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("server: ошибка при обновлении соединения: %v", err)
		return
	}

	client := hub.AddClient(conn)

	// Если это второй игрок, стартуем игру
	if len(hub.Clients) == 2 {
		hub.Broadcast <- []byte(`{"type":"start"}`)
	}

	go client.ReadMessages()
	go client.WriteMessages()
}
