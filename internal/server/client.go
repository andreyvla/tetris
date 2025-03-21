package server

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gorilla/websocket"
)

// Client представляет подключённого игрока
type Client struct {
	Conn     *websocket.Conn // WebSocket-соединение
	Send     chan []byte     // Канал для отправки сообщений
	Hub      *Hub            // Ссылка на общий Hub
	PlayerID int             // ID игрока (1 или 2)
	GameOver bool            // Флаг завершения игры
}

// NewClient создаёт нового игрока
func NewClient(conn *websocket.Conn, hub *Hub, playerID int) *Client {
	client := &Client{
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      hub,
		PlayerID: playerID,
		GameOver: false,
	}

	// Отправляем игроку его ID сразу при подключении
	initMessage := fmt.Sprintf(`{"type":"init","player":%d}`, client.PlayerID)
	client.Send <- []byte(initMessage)

	// Запускаем обработку ввода в отдельной горутине
	go client.HandleInput()

	return client
}

// ReadMessages читает сообщения от сервера
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

		// Логируем полученные сообщения
		log.Printf("server: получено сообщение: %s", message)

		// Распознаём тип сообщения
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("server: ошибка парсинга JSON: %v", err)
			continue
		}

		// Обработка game_over
		if msg["type"] == "game_over" {
			status, ok := msg["status"].(string)
			if ok {
				c.HandleGameOver(status)
			}
			continue
		}

		// Отправляем сообщение в хаб
		c.Hub.Broadcast <- message
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

// HandleInput обрабатывает ввод игрока
func (c *Client) HandleInput() {
	for {
		if c.GameOver {
			fmt.Println("Нажмите R, чтобы начать заново")

			var input string
			fmt.Scanln(&input)
			if strings.ToUpper(input) == "R" {
				c.RestartGame()
			}
		}
	}
}

// RestartGame отправляет запрос на рестарт игры
func (c *Client) RestartGame() {
	log.Println("client: Перезапуск игры...")
	c.GameOver = false // Сбрасываем флаг окончания игры

	// Отправляем серверу команду на рестарт
	restartMessage := map[string]interface{}{
		"type":   "restart",
		"player": c.PlayerID, // Сообщаем, кто хочет рестарт
	}
	msg, _ := json.Marshal(restartMessage)
	c.Send <- msg
}

func (c *Client) HandleGameOver(status string) {
	if status == "lose" {
		fmt.Println("😢 Вы проиграли... 😢")
		fmt.Println("Нажмите R, чтобы начать заново")
		c.GameOver = true
	} else if status == "win" {
		fmt.Println("🎉 Вы победили! 🎉")
		fmt.Println("Нажмите R, чтобы начать заново")
		c.GameOver = true
	}
}
