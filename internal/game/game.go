package game

import (
	"image/color"
	"log"
	"tetris/internal/field"
	"tetris/internal/figure"
	"tetris/internal/models"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

const (
	emptyCellColorValue    = 200
	occupiedCellColorValue = 0
	figureColorValue       = 255
)

// Переменные для цветов
var (
	emptyCellColor    = color.RGBA{emptyCellColorValue, emptyCellColorValue, emptyCellColorValue, 255}    // Серый (пустая клетка)
	occupiedCellColor = color.RGBA{occupiedCellColorValue, occupiedCellColorValue, figureColorValue, 255} // Синий (занятая клетка)
	figureColor       = color.RGBA{figureColorValue, occupiedCellColorValue, occupiedCellColorValue, 255} // Красный цвет
	textColor         = color.RGBA{255, 255, 255, 255}
)

// Game управляет игрой
type Game struct {
	Field        *field.Field
	Figure       *models.Figure
	LastDrop     time.Time
	DropInterval time.Duration
	GameOver     bool // Флаг окончания игры
	//Переменные для сдвига
	LastHorizontalMove     time.Time     // Время последнего горизонтального сдвига
	HorizontalMoveInterval time.Duration // Интервал между горизонтальными сдвигами
	HorizontalMoveDelay    time.Duration // Задержка перед началом повторных сдвигов
	MovingHorizontally     bool          // Движется ли фигура влево/вправо
	HorizontalDirection    int           // Направление сдвига (0 - нет, -1 - влево, 1 - вправо)
	//Переменные для поворота
	LastRotate     time.Time     // Время последнего поворота
	RotateInterval time.Duration // Интервал между поворотами
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
	}
	g.Figure = figure.NewFigure(g.Field)
	return g
}

// Update обновляет игру (каждый кадр)
func (g *Game) Update() error {
	if g.GameOver {
		return nil
	}

	// Обработка горизонтальных перемещений
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.moveHorizontally(-1)
	} else if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.moveHorizontally(1)
	} else {
		g.MovingHorizontally = false
		g.HorizontalDirection = 0
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
		figure.MoveDown(g.Figure, g.Field)
	}

	// Автоматическое падение фигуры по таймеру
	if time.Since(g.LastDrop) > g.DropInterval {
		if !g.IsFigureCollidingAfterMove() {
			figure.MoveDown(g.Figure, g.Field) // Фигура двигается вниз
		} else {
			// Фигура столкнулась с дном или другой фигурой -> фиксируем её
			g.FixFigure()
			g.ClearFullRows()

			// Создаем новую фигуру
			g.Figure = figure.NewFigure(g.Field)

			// Если новая фигура сразу сталкивается, значит, конец игры
			if g.IsFigureColliding() {
				g.GameOver = true
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
			figure.MoveLeft(g.Figure, g.Field)
		} else if direction == 1 {
			figure.MoveRight(g.Figure, g.Field)
		}
		g.LastHorizontalMove = time.Now()
		g.MovingHorizontally = true
		g.HorizontalDirection = direction
	} else if time.Since(g.LastHorizontalMove) > g.HorizontalMoveDelay {
		if direction == -1 {
			figure.MoveLeft(g.Figure, g.Field)
		} else if direction == 1 {
			figure.MoveRight(g.Figure, g.Field)
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
	log.Printf("фигура %s зафиксирована", g.Figure.Shape)
}

// ClearFullRows удаляет полностью заполненные ряды
func (g *Game) ClearFullRows() {
	for y := 0; y < field.Rows; y++ {
		if g.Field.IsRowFull(y) {
			g.Field.ClearRow(y)
		}
	}
	log.Println("проверены и очищены заполненные ряды")
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
	if !g.GameOver {
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
	} else {
		// Отрисовка Game Over
		gameOverText := "Game Over"

		text.Draw(screen, gameOverText, basicfont.Face7x13, field.ScreenWidth/2-50, field.ScreenHeight/2, textColor)
	}
}

// Layout задает размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return field.ScreenWidth, field.ScreenHeight
}

// IsFigureCollidingAfterMove проверяет, будет ли столкновение после перемещения на dx, dy
func (g *Game) IsFigureCollidingAfterMove() bool {
	return figure.IsFigureCollidingAfterMove(g.Figure, g.Field, 0, 1)
}
