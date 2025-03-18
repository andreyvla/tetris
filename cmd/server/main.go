package main

import (
	"log"
	"net/http"
	"tetris/internal/server"
)

func main() {
	log.SetPrefix("server: ") // Префикс логов

	hub := server.NewHub()
	go hub.Run() // Запускаем обработку WebSocket-соединений

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		server.HandleWebSocket(hub, w, r)
	})

	log.Println("Запуск WebSocket-сервера на порту :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
