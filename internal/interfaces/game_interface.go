package interfaces

type GameInterface interface {
	SetGameStarted(started bool)
	SetGameOver(over bool)
}
