package figure

import (
	"log"
	"math/rand"
	"tetris/internal/field"
	"tetris/internal/models"
)

const (
	figureWidth  = 4
	figureHeight = 4
)

// NewFigure создает новую случайную фигуру
func NewFigure(fld *field.Field) *models.Figure {
	shape := models.Shape(rand.Intn(7)) // Выбираем случайную фигуру
	fig := &models.Figure{
		Shape: shape,
		X:     field.Cols/2 - figureWidth/2, // Центрируем по горизонтали
		Y:     0,                            // Фигура всегда появляется вверху
	}
	SetShape(fig, shape) // Устанавливаем форму
	log.Printf("создана новая фигура: %s", fig.Shape)
	return fig
}

// SetShape задает матрицу для фигуры
func SetShape(f *models.Figure, shape models.Shape) {
	switch shape {
	case models.ShapeI:
		f.Cells = [4][4]bool{
			{false, false, false, false},
			{true, true, true, true},
			{false, false, false, false},
			{false, false, false, false},
		}
	case models.ShapeO:
		f.Cells = [4][4]bool{
			{false, true, true, false},
			{false, true, true, false},
			{false, false, false, false},
			{false, false, false, false},
		}
	case models.ShapeL:
		f.Cells = [4][4]bool{
			{false, false, true, false},
			{true, true, true, false},
			{false, false, false, false},
			{false, false, false, false},
		}
	case models.ShapeJ:
		f.Cells = [4][4]bool{
			{true, false, false, false},
			{true, true, true, false},
			{false, false, false, false},
			{false, false, false, false},
		}
	case models.ShapeT:
		f.Cells = [4][4]bool{
			{false, true, false, false},
			{true, true, true, false},
			{false, false, false, false},
			{false, false, false, false},
		}
	case models.ShapeS:
		f.Cells = [4][4]bool{
			{false, true, true, false},
			{true, true, false, false},
			{false, false, false, false},
			{false, false, false, false},
		}
	case models.ShapeZ:
		f.Cells = [4][4]bool{
			{true, true, false, false},
			{false, true, true, false},
			{false, false, false, false},
			{false, false, false, false},
		}
	default:
		log.Printf("неизвестная фигура: %s", shape.String())
	}
}

// MoveLeft перемещает фигуру влево (если возможно)
func MoveLeft(f *models.Figure, fld *field.Field) {
	if !IsFigureCollidingAfterMove(f, fld, -1, 0) {
		f.X--
		log.Printf("фигура %s сдвинута влево", f.Shape)
	}
}

// MoveRight перемещает фигуру вправо (если возможно)
func MoveRight(f *models.Figure, fld *field.Field) {
	if !IsFigureCollidingAfterMove(f, fld, 1, 0) {
		f.X++
		log.Printf("фигура %s сдвинута вправо", f.Shape)
	}
}

// MoveDown перемещает фигуру вниз (если возможно)
func MoveDown(f *models.Figure, fld *field.Field) {
	if !IsFigureCollidingAfterMove(f, fld, 0, 1) {
		f.Y++
		log.Printf("фигура %s сдвинута вниз", f.Shape)
	}
}

// Rotate поворачивает фигуру
func Rotate(f *models.Figure, fld *field.Field) {
	// Создаем временную матрицу для хранения повернутых клеток
	var rotatedCells [4][4]bool

	// Поворачиваем фигуру на 90 градусов по часовой стрелке
	for i := range figureHeight {
		for j := range figureWidth {
			rotatedCells[j][figureHeight-1-i] = f.Cells[i][j]
		}
	}

	// Создаем временную фигуру, чтобы проверить столкновения
	tempFigure := &models.Figure{
		Shape: f.Shape,
		Cells: rotatedCells,
		X:     f.X,
		Y:     f.Y,
	}

	// Проверяем, не будет ли столкновения после поворота
	if !IsFigureCollidingAfterMove(tempFigure, fld, 0, 0) {
		// Если столкновения нет, применяем поворот
		f.Cells = rotatedCells
		log.Printf("фигура %s повернута", f.Shape)
	} else {
		log.Printf("поворот фигуры %s невозможен: есть столкновение", f.Shape)
	}
}

// IsFigureCollidingAfterMove проверяет, будет ли столкновение после перемещения на dx, dy
func IsFigureCollidingAfterMove(fig *models.Figure, fld *field.Field, dx, dy int) bool {
	for row := 0; row < figureHeight; row++ {
		for col := 0; col < figureWidth; col++ {
			if fig.Cells[row][col] {
				x := fig.X + col + dx
				y := fig.Y + row + dy
				if y >= field.Rows || x < 0 || x >= field.Cols || fld.IsOccupied(x, y) {
					return true
				}
			}
		}
	}
	return false
}
