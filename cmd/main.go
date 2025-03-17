package main

import (
	"fmt"
	"log"
	"os"
	"tetris/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// Настройка логгера для вывода в stderr
	log.SetOutput(os.Stderr)
	// Установка префикса для логов
	log.SetPrefix("main: ")

	log.Println("Запуск игры Tetris") // Логируем запуск игры

	gameInstance := game.NewGame()
	// Обработка ошибки, которую может вернуть ebiten.RunGame
	if err := ebiten.RunGame(gameInstance); err != nil {
		// Логируем ошибку
		log.Printf("Ошибка при запуске игры: %v", err)
		// Выводим ошибку в stderr с помощью fmt.Fprintf
		fmt.Fprintf(os.Stderr, "Ошибка при запуске игры: %v\n", err)
		os.Exit(1) // Завершаем программу с ненулевым кодом возврата
	}
	log.Println("Игра Tetris завершена")
}
