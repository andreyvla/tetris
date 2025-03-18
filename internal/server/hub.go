package server

import (
	"log"
	"sync"
)

// Hub управляет всеми подключениями игроков
type Hub struct {
	Clients    map[*Client]bool // Все активные клиенты
	Broadcast  chan []byte      // Канал для рассылки сообщений
	Register   chan *Client     // Канал для новых подключений
	Unregister chan *Client     // Канал для отключений
	mu         sync.Mutex       // Для потокобезопасности
}

// NewHub создаёт новый центр управления подключениями
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run запускает обработку событий WebSocket
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
			log.Println("server: новый игрок подключен")

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Println("server: игрок отключен")

		case message := <-h.Broadcast:
			h.mu.Lock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}
