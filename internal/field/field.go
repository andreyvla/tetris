package field

import "log"

const (
	CellSize     = 32                      // CellSize - Размер одной клетки в пикселях
	ScreenWidth  = 320                     // ScreenWidth - Ширина игрового экрана в пикселях
	ScreenHeight = 480                     // ScreenHeight - Высота игрового экрана в пикселях
	Rows         = ScreenHeight / CellSize // Rows - Количество строк на игровом поле
	Cols         = ScreenWidth / CellSize  // Cols - Количество столбцов на игровом поле
)

// FieldCells представляет тип данных для клеток игрового поля
type FieldCells [Rows][Cols]bool

// Field представляет игровое поле
type Field struct {
	Cells FieldCells // Cells - false — пусто, true — занято
}

// NewField создает новое игровое поле
func NewField() *Field {
	f := &Field{}
	for i := range f.Cells {
		for j := range f.Cells[i] {
			f.Cells[i][j] = false // Явно инициализируем все клетки как пустые
		}
	}
	log.Printf("создано новое поле размером %dx%d", Cols, Rows)
	return f
}

// IsOccupied проверяет, занята ли клетка
func (f *Field) IsOccupied(x, y int) bool {
	// Проверка на выход за границы поля
	if x < 0 || x >= Cols || y < 0 || y >= Rows {
		log.Printf("попытка доступа за границы поля: x=%d, y=%d", x, y)
		return true // Считаем, что за границей поле всегда занято
	}
	return f.Cells[y][x]
}

// SetOccupied помечает клетку как занятую
func (f *Field) SetOccupied(x, y int) {
	// Проверка на выход за границы поля
	if x < 0 || x >= Cols || y < 0 || y >= Rows {
		log.Printf("попытка установить занятую клетку за границей поля: x=%d, y=%d", x, y)
		return
	}
	f.Cells[y][x] = true
	log.Printf("установлена занятая клетка: x=%d, y=%d", x, y)
}

// IsRowFull проверяет, заполнен ли ряд полностью
func (f *Field) IsRowFull(y int) bool {
	// Проверка на выход за границы поля
	if y < 0 || y >= Rows {
		log.Printf("попытка проверить заполненность строки за границей поля: y=%d", y)
		return false
	}
	for x := 0; x < Cols; x++ {
		if !f.Cells[y][x] {
			return false
		}
	}
	return true
}

// ClearRow удаляет заполненный ряд и сдвигает все сверху вниз
func (f *Field) ClearRow(y int) {
	// Проверка на выход за границы поля
	if y < 0 || y >= Rows {
		log.Printf("попытка удалить строку за границей поля: y=%d", y)
		return
	}
	log.Printf("удалена заполненная строка: y=%d", y)
	// Сдвигаем все строки сверху вниз
	for row := y; row > 0; row-- {
		for x := 0; x < Cols; x++ {
			f.Cells[row][x] = f.Cells[row-1][x]
		}
	}
	// Очищаем верхнюю строку
	for x := range Cols {
		f.Cells[0][x] = false
	}
}
