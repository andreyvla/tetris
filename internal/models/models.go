package models

import "log"

// Shape представляет собой тип фигуры
type Shape int

const (
	ShapeI Shape = iota // ShapeI - Фигура "палочка" (I)
	ShapeO              // ShapeO - Фигура "квадрат" (O)
	ShapeL              // ShapeL - Фигура "L" (L)
	ShapeJ              // ShapeJ - Фигура "зеркальная L" (J)
	ShapeT              // ShapeT - Фигура "T" (T)
	ShapeS              // ShapeS - Фигура "S" (S)
	ShapeZ              // ShapeZ - Фигура "Z" (Z)
)

// Figure представляет собой фигуру
type Figure struct {
	Shape Shape      // Тип фигуры (одна из констант Shape)
	Cells [4][4]bool // Матрица 4x4 для хранения формы фигуры
	X, Y  int        // Координаты фигуры (левый верхний угол)
}

// String возвращает строковое представление типа фигуры для логов
func (s Shape) String() string {
	switch s {
	case ShapeI:
		return "ShapeI"
	case ShapeO:
		return "ShapeO"
	case ShapeL:
		return "ShapeL"
	case ShapeJ:
		return "ShapeJ"
	case ShapeT:
		return "ShapeT"
	case ShapeS:
		return "ShapeS"
	case ShapeZ:
		return "ShapeZ"
	default:
		log.Printf("неизвестная фигура: %d", s)
		return "UnknownShape"
	}
}
