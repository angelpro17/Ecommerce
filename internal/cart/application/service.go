package application

import (
	"context"
	"errors"

	"github.com/angelpro17/Ecommerce.git/internal/cart/domain"
)

// CartService implementa los casos de uso del carrito.
type CartService struct {
	repo    domain.Repository
	prodSvc domain.ProductReader
}

func NewCartService(repo domain.Repository, prodSvc domain.ProductReader) *CartService {
	return &CartService{repo: repo, prodSvc: prodSvc}
}

func (s *CartService) GetOrCreateCart(ctx context.Context, userID string) (*domain.Cart, error) {
	cart, err := s.repo.GetActiveByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrCartNotFound) {
			return s.repo.Create(ctx, userID)
		}
		return nil, err
	}
	return cart, nil
}

func (s *CartService) AddItem(ctx context.Context, userID string, productID int64, qty int) (*domain.CartSummary, error) {
	if qty <= 0 {
		return nil, domain.ErrInvalidQuantity
	}

	price, stock, err := s.prodSvc.GetByID(ctx, productID)
	if err != nil {
		return nil, domain.ErrProductNotFound
	}
	if stock < qty {
		return nil, domain.ErrInsufficientStock
	}

	cart, err := s.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	item := &domain.CartItem{CartID: cart.ID, ProductID: productID, Quantity: qty, Price: price}
	if err := s.repo.AddItem(ctx, item); err != nil {
		return nil, err
	}
	return s.GetSummary(ctx, userID)
}

func (s *CartService) UpdateItem(ctx context.Context, userID string, productID int64, qty int) (*domain.CartSummary, error) {
	if qty <= 0 {
		return nil, domain.ErrInvalidQuantity
	}

	_, stock, err := s.prodSvc.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if stock < qty {
		return nil, domain.ErrInsufficientStock
	}

	cart, err := s.repo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateItemQuantity(ctx, cart.ID, productID, qty); err != nil {
		return nil, err
	}
	return s.GetSummary(ctx, userID)
}

func (s *CartService) RemoveItem(ctx context.Context, userID string, productID int64) (*domain.CartSummary, error) {
	cart, err := s.repo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.RemoveItem(ctx, cart.ID, productID); err != nil {
		return nil, err
	}
	return s.GetSummary(ctx, userID)
}

func (s *CartService) GetSummary(ctx context.Context, userID string) (*domain.CartSummary, error) {
	cart, err := s.repo.GetActiveByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrCartNotFound) {
			return &domain.CartSummary{UserID: userID, Items: []domain.CartItem{}, Status: "empty"}, nil
		}
		return nil, err
	}

	items, err := s.repo.GetItems(ctx, cart.ID)
	if err != nil {
		return nil, err
	}

	var subtotal float64
	var count int
	for _, it := range items {
		subtotal += it.Subtotal
		count += it.Quantity
	}
	tax := subtotal * domain.TaxRate

	return &domain.CartSummary{
		CartID: cart.ID, UserID: userID, Items: items,
		ItemsCount: count, Subtotal: subtotal, Tax: tax,
		Total: subtotal + tax, Status: cart.Status,
	}, nil
}

func (s *CartService) ClearCart(ctx context.Context, userID string) error {
	cart, err := s.repo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}
	return s.repo.ClearCart(ctx, cart.ID)
}

func (s *CartService) Checkout(ctx context.Context, userID string) (*domain.CartSummary, error) {
	summary, err := s.GetSummary(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(summary.Items) == 0 {
		return nil, domain.ErrEmptyCart
	}

	for _, it := range summary.Items {
		if err := s.prodSvc.UpdateStock(ctx, it.ProductID, -it.Quantity); err != nil {
			return nil, err
		}
	}

	if err := s.repo.UpdateStatus(ctx, summary.CartID, "completed"); err != nil {
		return nil, err
	}
	summary.Status = "completed"
	return summary, nil
}
