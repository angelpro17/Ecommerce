package domain

import (
	"errors"
	"time"
)

// Product entidad de dominio del producto.
type Product struct {
	ID          int64
	Name        string
	Description string
	Price       float64
	Stock       int
	ImageURL    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Errores de dominio.
var (
	ErrNotFound     = errors.New("product not found")
	ErrInvalidName  = errors.New("name is required")
	ErrInvalidPrice = errors.New("price must be greater than 0")
	ErrInvalidStock = errors.New("stock cannot be negative")
)
