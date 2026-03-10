package domain

import "context"

// Repository puerto de persistencia del carrito.
type Repository interface {
	Create(ctx context.Context, userID string) (*Cart, error)
	GetByID(ctx context.Context, id int64) (*Cart, error)
	GetActiveByUserID(ctx context.Context, userID string) (*Cart, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	AddItem(ctx context.Context, item *CartItem) error
	UpdateItemQuantity(ctx context.Context, cartID, productID int64, quantity int) error
	RemoveItem(ctx context.Context, cartID, productID int64) error
	GetItems(ctx context.Context, cartID int64) ([]CartItem, error)
	ClearCart(ctx context.Context, cartID int64) error
}

// ProductReader puerto segregado para leer datos de producto desde el carrito (ISP).
type ProductReader interface {
	GetByID(ctx context.Context, id int64) (price float64, stock int, err error)
	UpdateStock(ctx context.Context, id int64, quantity int) error
}

// Service puerto de casos de uso del carrito.
type Service interface {
	GetOrCreateCart(ctx context.Context, userID string) (*Cart, error)
	AddItem(ctx context.Context, userID string, productID int64, quantity int) (*CartSummary, error)
	UpdateItem(ctx context.Context, userID string, productID int64, quantity int) (*CartSummary, error)
	RemoveItem(ctx context.Context, userID string, productID int64) (*CartSummary, error)
	GetSummary(ctx context.Context, userID string) (*CartSummary, error)
	ClearCart(ctx context.Context, userID string) error
	Checkout(ctx context.Context, userID string) (*CartSummary, error)
}
