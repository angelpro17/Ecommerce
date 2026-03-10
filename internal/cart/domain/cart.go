package domain

import (
	"errors"
	"time"
)

// Tasa de impuesto aplicada al total del carrito.
const TaxRate = 0.16

// Cart entidad de dominio del carrito.
type Cart struct {
	ID        int64
	UserID    string
	Status    string
	Items     []CartItem
	Total     float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CartItem representa un item dentro del carrito.
type CartItem struct {
	ID        int64
	CartID    int64
	ProductID int64
	Name      string
	Quantity  int
	Price     float64
	Subtotal  float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CartSummary resumen calculado del carrito.
type CartSummary struct {
	CartID     int64
	UserID     string
	Items      []CartItem
	ItemsCount int
	Subtotal   float64
	Tax        float64
	Total      float64
	Status     string
}

// Errores de dominio.
var (
	ErrCartNotFound      = errors.New("cart not found")
	ErrItemNotFound      = errors.New("item not found in cart")
	ErrInvalidQuantity   = errors.New("quantity must be greater than 0")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrProductNotFound   = errors.New("product not found")
	ErrEmptyCart         = errors.New("cart is empty")
)
