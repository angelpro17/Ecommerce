package domain

import "context"

// Repository puerto de persistencia de productos.
type Repository interface {
	Create(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, id int64) (*Product, error)
	GetAll(ctx context.Context, page, pageSize int) ([]Product, int, error)
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id int64) error
	UpdateStock(ctx context.Context, id int64, quantity int) error
}

// Service puerto de casos de uso de productos.
type Service interface {
	Create(ctx context.Context, name, description string, price float64, stock int, imageURL string) (*Product, error)
	GetByID(ctx context.Context, id int64) (*Product, error)
	GetAll(ctx context.Context, page, pageSize int) ([]Product, int, error)
	Update(ctx context.Context, id int64, name, description *string, price *float64, stock *int, imageURL *string) (*Product, error)
	Delete(ctx context.Context, id int64) error
}
