package main

import (
	"log"
	"tetris/internal/game"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	gameInstance := game.NewGame()
	if err := ebiten.RunGame(gameInstance); err != nil {
		log.Fatal(err)
	}
}
