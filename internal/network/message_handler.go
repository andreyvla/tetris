package network

import (
	"encoding/json"
	"log"
)

type Message struct {
	Type   string      `json:"type"`
	Player int         `json:"player,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

type MessageHandler struct {
	Client *WebSocketClient
}

func NewMessageHandler(client *WebSocketClient) *MessageHandler {
	return &MessageHandler{Client: client}
}

func (h *MessageHandler) HandleMessage(msgBytes []byte) {
	var msg Message
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		log.Printf("network: ошибка парсинга JSON: %v", err)
		return
	}

	switch msg.Type {
	case "init":
		h.Client.PlayerID = msg.Player
		log.Printf("network: назначен playerID = %d", msg.Player)

	case "start":
		h.Client.Game.SetGameStarted(true)
		log.Println("🎮 Игра началась! 🎮")

	case "game_over":
		h.Client.Game.SetGameOver(true)

		dataMap, ok := msg.Data.(map[string]interface{})
		if !ok {
			log.Println("network: ошибка парсинга game_over")
			return
		}

		status, ok := dataMap["status"].(string)
		if !ok {
			log.Println("network: неверный формат данных в game_over")
			return
		}

		if status == "win" {
			log.Println("🎉 Вы победили! 🎉")
		} else if status == "lose" {
			log.Println("😢 Вы проиграли... 😢")
		}

	}
}
