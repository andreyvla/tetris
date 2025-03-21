package server

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
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
			log.Printf("server: Игрок %d подключился", client.PlayerID)
			h.mu.Unlock()

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
				if client.GameOver {
					continue // Не отправляем сообщения проигравшему игроку
				}
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

// GameOver обрабатывает завершение игры
func (h *Hub) GameOver(loserID int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	winnerID := 1
	if loserID == 1 {
		winnerID = 2
	}

	log.Printf("server: Игрок %d проиграл, победил игрок %d", loserID, winnerID)

	// Останавливаем игру у обоих игроков
	for client := range h.Clients {
		client.GameOver = true

		status := "win"
		if client.PlayerID == loserID {
			status = "lose"
		}

		message := map[string]interface{}{
			"type":   "game_over",
			"status": status,
		}
		msg, _ := json.Marshal(message)
		client.Send <- msg
	}
}

func (h *Hub) RestartGame() {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Println("server: Перезапуск игры для всех игроков")

	for client := range h.Clients {
		client.GameOver = false // Сбрасываем флаг завершения игры

		// Отправляем игроку команду на рестарт
		restartMessage := map[string]interface{}{
			"type": "restart",
		}
		msg, _ := json.Marshal(restartMessage)
		client.Send <- msg
	}
}

func (h *Hub) AddClient(conn *websocket.Conn) *Client {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Определяем playerID (1 для первого игрока, 2 для второго)
	playerID := 1
	if len(h.Clients) > 0 {
		playerID = 2
	}

	client := NewClient(conn, h, playerID)
	h.Clients[client] = true
	h.Register <- client
	log.Printf("server: Игрок %d подключен", playerID)

	// Отправляем клиенту его playerID
	h.SendInitMessage(client)

	// Если оба игрока подключены, отправляем сигнал старта
	if len(h.Clients) == 2 {
		startMessage := `{"type":"start"}`
		for c := range h.Clients {
			c.Send <- []byte(startMessage)
		}
		log.Println("server: Оба игрока подключены, игра начинается!")
	}

	return client
}

func (h *Hub) SendInitMessage(client *Client) {
	message := map[string]interface{}{
		"type":   "init",
		"player": client.PlayerID,
	}
	msg, _ := json.Marshal(message)
	client.Send <- msg
}
