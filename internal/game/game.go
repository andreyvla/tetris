package game

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"tetris/internal/field"
	"tetris/internal/figure"
	"tetris/internal/models"
	"tetris/internal/network"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

const (
	emptyCellColorValue    = 200
	occupiedCellColorValue = 0
	figureColorValue       = 255
	oneLineScore           = 100
	twoLineScore           = 300
	threeLineScore         = 700
	fourLineScore          = 1500
	//Score board
	scoreBoardWidth  = 150
	scoreBoardHeight = 50
	//Game over
	gameOverRectWidth  = 200
	gameOverRectHeight = 100
	gameOverRectX      = (field.ScreenWidth - gameOverRectWidth) / 2
	gameOverRectY      = (field.ScreenHeight - gameOverRectHeight) / 2
	//Расположение табло
	scoreBoardX = field.ScreenWidth + 10
	scoreBoardY = 10
	//Pause Rect
	pauseRectWidth  = scoreBoardWidth
	pauseRectHeight = 30
	pauseRectX      = scoreBoardX
	pauseRectY      = scoreBoardY + scoreBoardHeight + 10
)

// Переменные для цветов
var (
	emptyCellColor    = color.RGBA{emptyCellColorValue, emptyCellColorValue, emptyCellColorValue, 255}    // Серый (пустая клетка)
	occupiedCellColor = color.RGBA{occupiedCellColorValue, occupiedCellColorValue, figureColorValue, 255} // Синий (занятая клетка)
	figureColor       = color.RGBA{figureColorValue, occupiedCellColorValue, occupiedCellColorValue, 255} // Красный цвет
	textColor         = color.RGBA{0, 0, 0, 255}                                                          // Черный цвет
	scoreBoardColor   = color.RGBA{200, 200, 200, 255}                                                    // Серый цвет для рамки поля со счетом
	gameOverRectColor = color.RGBA{100, 100, 100, 255}
	pauseRectColor    = color.RGBA{200, 200, 200, 255}
)

// Game управляет игрой
type Game struct {
	Field        *field.Field
	Figure       *models.Figure
	WebSocket    *network.WebSocketClient
	LastDrop     time.Time
	DropInterval time.Duration
	GameStarted  bool
	GameOver     bool
	GameResult   string
	//Переменные для сдвига
	LastHorizontalMove     time.Time     // Время последнего горизонтального сдвига
	HorizontalMoveInterval time.Duration // Интервал между горизонтальными сдвигами
	HorizontalMoveDelay    time.Duration // Задержка перед началом повторных сдвигов
	MovingHorizontally     bool          // Движется ли фигура влево/вправо
	HorizontalDirection    int           // Направление сдвига (0 - нет, -1 - влево, 1 - вправо)
	//Переменные для поворота
	LastRotate     time.Time     // Время последнего поворота
	RotateInterval time.Duration // Интервал между поворотами
	//Счет
	Score    int       // Текущий счет
	fontFace font.Face // Шрифт
	//Пауза
	Paused        bool          //На паузе ли игра?
	LastPause     time.Time     // Время последнего переключения паузы
	PauseInterval time.Duration // Интервал между переключениями
}

func (g *Game) SetGameStarted(started bool) {
	g.GameStarted = started
}

func (g *Game) SetGameOver(over bool) {
	g.GameOver = over
}

// NewGame создает новую игру
func NewGame() *Game {
	g := &Game{
		Field:                  field.NewField(),
		LastDrop:               time.Now(),
		DropInterval:           time.Second / 2, // Фигура падает раз в 0.5 секунды
		GameOver:               false,
		HorizontalMoveInterval: time.Millisecond * 50,  // Интервал между повторными сдвигами
		HorizontalMoveDelay:    time.Millisecond * 250, // Задержка перед повторными сдвигами
		MovingHorizontally:     false,
		HorizontalDirection:    0,
		LastRotate:             time.Now(),
		RotateInterval:         time.Millisecond * 200, // Интервал между поворотами
		Score:                  0,                      // Изначальный счет - 0
		fontFace:               basicfont.Face7x13,
		Paused:                 false,
		LastPause:              time.Now(),
		PauseInterval:          time.Millisecond * 200, //Интервал между паузами
	}
	g.Figure = figure.NewFigure(g.Field)
	wsClient, err := network.NewWebSocketClient("ws://localhost:8080/ws", g)
	if err != nil {
		log.Fatalf("Ошибка подключения к WebSocket: %v", err)
	}
	g.WebSocket = wsClient

	return g
}
func (g *Game) MoveFigure(direction string) {
	switch direction {
	case "left":
		figure.MoveLeft(g.Figure, g.Field, g.WebSocket, g.WebSocket.PlayerID)
	case "right":
		figure.MoveRight(g.Figure, g.Field, g.WebSocket, g.WebSocket.PlayerID)
	case "down":
		figure.MoveDown(g.Figure, g.Field, g.WebSocket, g.WebSocket.PlayerID)
	}
}

// Update обновляет игру (каждый кадр)
func (g *Game) Update() error {
	if !g.GameStarted {
		return nil // Ждём, пока сервер отправит "start"
	}

	// Проверяем, нажата ли клавиша "P" и не прошло ли еще достаточно времени с момента последнего переключения паузы
	if ebiten.IsKeyPressed(ebiten.KeyP) && time.Since(g.LastPause) > g.PauseInterval {
		g.Paused = !g.Paused
		g.LastPause = time.Now()
	}

	if g.GameOver && ebiten.IsKeyPressed(ebiten.KeyR) {
		g.RestartGame()
		return nil
	}
	if g.GameOver || g.Paused {
		return nil
	}

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.MoveFigure("left")
	} else if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.MoveFigure("right")
	} else if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.MoveFigure("down")
	}

	//Если есть направление, но кнопки не нажаты, значит, надо продолжать двигать
	if g.HorizontalDirection != 0 {
		if time.Since(g.LastHorizontalMove) > g.HorizontalMoveInterval {
			g.moveHorizontally(g.HorizontalDirection)
			g.LastHorizontalMove = time.Now()
		}
	}

	// Поворот
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		if time.Since(g.LastRotate) > g.RotateInterval {
			figure.Rotate(g.Figure, g.Field)
			g.LastRotate = time.Now()
		}
	}

	// Ускорение падения вниз при нажатии
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		figure.MoveDown(g.Figure, g.Field, g.WebSocket, g.WebSocket.PlayerID)
	}

	// Автоматическое падение фигуры по таймеру
	if time.Since(g.LastDrop) > g.DropInterval {
		if !g.IsFigureCollidingAfterMove() {
			figure.MoveDown(g.Figure, g.Field, g.WebSocket, g.WebSocket.PlayerID) // Фигура двигается вниз
		} else {
			// Фигура столкнулась с дном или другой фигурой -> фиксируем её
			g.FixFigure()
			g.ClearFullRows()

			// Создаем новую фигуру
			g.Figure = figure.NewFigure(g.Field)

			// Если новая фигура сразу сталкивается, значит, конец игры
			if g.IsFigureColliding() {
				g.GameOver = true
				winner := 3 - g.WebSocket.PlayerID
				message := fmt.Sprintf(`{"type":"game_over","winner":%d}`, winner)
				g.WebSocket.SendMessage([]byte(message))

				log.Printf("game: Игрок %d проиграл, победил игрок %d", g.WebSocket.PlayerID, winner)

				// Сообщаем клиенту о завершении игры
				g.EndGame(winner)
				return nil
			}
		}
		g.LastDrop = time.Now()
	}

	return nil
}

// moveHorizontally перемещает фигуру по горизонтали в заданном направлении
func (g *Game) moveHorizontally(direction int) {
	if !g.MovingHorizontally {
		if direction == -1 {
			figure.MoveLeft(g.Figure, g.Field, g.WebSocket, g.WebSocket.PlayerID)
		} else if direction == 1 {
			figure.MoveRight(g.Figure, g.Field, g.WebSocket, g.WebSocket.PlayerID)
		}
		g.LastHorizontalMove = time.Now()
		g.MovingHorizontally = true
		g.HorizontalDirection = direction
	} else if time.Since(g.LastHorizontalMove) > g.HorizontalMoveDelay {
		if direction == -1 {
			figure.MoveLeft(g.Figure, g.Field, g.WebSocket, g.WebSocket.PlayerID)
		} else if direction == 1 {
			figure.MoveRight(g.Figure, g.Field, g.WebSocket, g.WebSocket.PlayerID)
		}
		g.LastHorizontalMove = time.Now()
	}
}

// FixFigure фиксирует фигуру в поле
func (g *Game) FixFigure() {
	for row := 0; row < 4; row++ {
		for col := 0; col < 4; col++ {
			if g.Figure.Cells[row][col] {
				x := g.Figure.X + col
				y := g.Figure.Y + row
				g.Field.SetOccupied(x, y) // Фиксируем все клетки фигуры
			}
		}
	}
}

// ClearFullRows удаляет полностью заполненные ряды
func (g *Game) ClearFullRows() {
	var rowsCleared int
	for y := 0; y < field.Rows; y++ {
		if g.Field.IsRowFull(y) {
			g.Field.ClearRow(y)
			rowsCleared++
		}
	}
	switch rowsCleared {
	case 1:
		g.Score += oneLineScore
	case 2:
		g.Score += twoLineScore
	case 3:
		g.Score += threeLineScore
	case 4:
		g.Score += fourLineScore
	}
}

// IsFigureColliding проверяет, сталкивается ли фигура
func (g *Game) IsFigureColliding() bool {
	for row := 0; row < 4; row++ {
		for col := 0; col < 4; col++ {
			if g.Figure.Cells[row][col] {
				x := g.Figure.X + col
				y := g.Figure.Y + row
				if y >= field.Rows || g.Field.IsOccupied(x, y) {
					return true
				}
			}
		}
	}
	return false
}

// Draw отрисовывает игру
func (g *Game) Draw(screen *ebiten.Image) {
	// Отрисовка поля
	for y := 0; y < field.Rows; y++ {
		for x := 0; x < field.Cols; x++ {
			c := emptyCellColor // Серый (пустая клетка)
			if g.Field.IsOccupied(x, y) {
				c = occupiedCellColor // Синяя (занятая клетка)
			}
			cell := ebiten.NewImage(field.CellSize-2, field.CellSize-2)
			cell.Fill(c)
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*field.CellSize+1), float64(y*field.CellSize+1))
			screen.DrawImage(cell, op)
		}
	}

	// Отрисовка текущей фигуры
	if !g.GameOver && !g.Paused {
		for row := 0; row < 4; row++ {
			for col := 0; col < 4; col++ {
				if g.Figure.Cells[row][col] {
					fig := ebiten.NewImage(field.CellSize-2, field.CellSize-2)
					fig.Fill(figureColor)
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Translate(float64((g.Figure.X+col)*field.CellSize+1), float64((g.Figure.Y+row)*field.CellSize+1))
					screen.DrawImage(fig, op)
				}
			}
		}
	} else if g.Paused {
		pausedText := "Paused"
		text.Draw(screen, pausedText, g.fontFace, field.ScreenWidth/2-(font.MeasureString(g.fontFace, pausedText).Ceil()/2), field.ScreenHeight/2+g.fontFace.Metrics().Ascent.Ceil()/2, textColor)
	} else {
		// Отрисовка Game Over
		if g.GameOver {
			message := "Game Over"
			if g.GameResult == "win" {
				message = "You Win! 🎉"
			} else if g.GameResult == "lose" {
				message = "You Lose... 😢"
			}

			text.Draw(screen, message, g.fontFace, field.ScreenWidth/2-50, field.ScreenHeight/2, textColor)
			text.Draw(screen, "Press R to restart", g.fontFace, field.ScreenWidth/2-70, field.ScreenHeight/2+20, textColor)
		}
	}
	//Рисуем рамку для счета
	scoreBoard := ebiten.NewImage(scoreBoardWidth, scoreBoardHeight)
	scoreBoard.Fill(scoreBoardColor)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(scoreBoardX), float64(scoreBoardY))
	screen.DrawImage(scoreBoard, op)
	// Отображение очков
	scoreText := fmt.Sprintf("Score: %d", g.Score)
	text.Draw(screen, scoreText, g.fontFace, scoreBoardX+10, scoreBoardY+30, textColor)

	//Рисуем рамку для паузы
	pauseRect := ebiten.NewImage(pauseRectWidth, pauseRectHeight)
	pauseRect.Fill(pauseRectColor)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(pauseRectX), float64(pauseRectY))
	screen.DrawImage(pauseRect, op)

	//Добавляем текст про паузу в прямоугольник
	pauseText := "Press P for pause"
	text.Draw(screen, pauseText, g.fontFace, pauseRectX+pauseRectWidth/2-(font.MeasureString(g.fontFace, pauseText).Ceil()/2), pauseRectY+pauseRectHeight/2+g.fontFace.Metrics().Ascent.Ceil()/2, textColor)

}

// Layout задает размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return field.ScreenWidth + scoreBoardWidth + 10, field.ScreenHeight
}

// IsFigureCollidingAfterMove проверяет, будет ли столкновение после перемещения на dx, dy
func (g *Game) IsFigureCollidingAfterMove() bool {
	return figure.IsFigureCollidingAfterMove(g.Figure, g.Field, 0, 1)
}

// RestartGame сбрасывает игру
func (g *Game) RestartGame() {
	log.Println("game: Перезапуск игры...")

	wsClient, err := network.NewWebSocketClient("ws://localhost:8080/ws", g)
	if err != nil {
		log.Fatalf("Ошибка подключения к WebSocket: %v", err)
	}
	g.WebSocket = wsClient

}

func (g *Game) EndGame(winner int) {
	log.Println("game: Игра завершена!")

	if winner == g.WebSocket.PlayerID {
		log.Println("game: Вы победили! 🎉")
	} else {
		log.Println("game: Вы проиграли... 😢")
	}

	// Устанавливаем флаг GameOver, но НЕ закрываем соединение
	g.GameOver = true
	g.WebSocket.GameOver = true

	// Отправляем клиенту сообщение о завершении игры
	gameOverMessage := map[string]interface{}{
		"type":   "game_over",
		"winner": winner,
	}
	msg, _ := json.Marshal(gameOverMessage)
	g.WebSocket.Send <- msg
}
