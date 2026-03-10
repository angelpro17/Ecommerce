package postgres

import (
	"context"

	"github.com/angelpro17/Ecommerce.git/internal/product/domain"
)

// ProductReaderAdapter adapta product.Repository al puerto segregado cart.ProductReader.
type ProductReaderAdapter struct {
	repo domain.Repository
}

func NewProductReaderAdapter(repo domain.Repository) *ProductReaderAdapter {
	return &ProductReaderAdapter{repo: repo}
}

func (a *ProductReaderAdapter) GetByID(ctx context.Context, id int64) (float64, int, error) {
	p, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return 0, 0, err
	}
	return p.Price, p.Stock, nil
}

func (a *ProductReaderAdapter) UpdateStock(ctx context.Context, id int64, qty int) error {
	return a.repo.UpdateStock(ctx, id, qty)
}
