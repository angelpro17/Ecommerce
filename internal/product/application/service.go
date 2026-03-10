package application

import (
	"context"

	"github.com/angelpro17/Ecommerce.git/internal/product/domain"
)

// ProductService implementa los casos de uso de productos.
type ProductService struct {
	repo domain.Repository
}

func NewProductService(repo domain.Repository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) Create(ctx context.Context, name, description string, price float64, stock int, imageURL string) (*domain.Product, error) {
	if name == "" {
		return nil, domain.ErrInvalidName
	}
	if price <= 0 {
		return nil, domain.ErrInvalidPrice
	}
	if stock < 0 {
		return nil, domain.ErrInvalidStock
	}

	p := &domain.Product{
		Name: name, Description: description,
		Price: price, Stock: stock, ImageURL: imageURL,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProductService) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) GetAll(ctx context.Context, page, pageSize int) ([]domain.Product, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.GetAll(ctx, page, pageSize)
}

func (s *ProductService) Update(ctx context.Context, id int64, name, description *string, price *float64, stock *int, imageURL *string) (*domain.Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		if *name == "" {
			return nil, domain.ErrInvalidName
		}
		p.Name = *name
	}
	if description != nil {
		p.Description = *description
	}
	if price != nil {
		if *price <= 0 {
			return nil, domain.ErrInvalidPrice
		}
		p.Price = *price
	}
	if stock != nil {
		if *stock < 0 {
			return nil, domain.ErrInvalidStock
		}
		p.Stock = *stock
	}
	if imageURL != nil {
		p.ImageURL = *imageURL
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProductService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
