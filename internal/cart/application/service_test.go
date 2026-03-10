package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/angelpro17/Ecommerce.git/internal/cart/domain"
)

// Mock del repositorio de carrito.

type mockCartRepo struct {
	carts   map[int64]*domain.Cart
	items   map[int64][]domain.CartItem
	nextID  int64
	err     error
	itemErr error
}

func newMockCartRepo() *mockCartRepo {
	return &mockCartRepo{carts: make(map[int64]*domain.Cart), items: make(map[int64][]domain.CartItem), nextID: 1}
}

func (m *mockCartRepo) Create(_ context.Context, userID string) (*domain.Cart, error) {
	if m.err != nil {
		return nil, m.err
	}
	now := time.Now()
	c := &domain.Cart{ID: m.nextID, UserID: userID, Status: "active", CreatedAt: now, UpdatedAt: now}
	m.carts[c.ID] = c
	m.nextID++
	return c, nil
}

func (m *mockCartRepo) GetByID(_ context.Context, id int64) (*domain.Cart, error) {
	if m.err != nil {
		return nil, m.err
	}
	c, ok := m.carts[id]
	if !ok {
		return nil, domain.ErrCartNotFound
	}
	return c, nil
}

func (m *mockCartRepo) GetActiveByUserID(_ context.Context, userID string) (*domain.Cart, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, c := range m.carts {
		if c.UserID == userID && c.Status == "active" {
			return c, nil
		}
	}
	return nil, domain.ErrCartNotFound
}

func (m *mockCartRepo) UpdateStatus(_ context.Context, id int64, status string) error {
	if m.err != nil {
		return m.err
	}
	c, ok := m.carts[id]
	if !ok {
		return domain.ErrCartNotFound
	}
	c.Status = status
	return nil
}

func (m *mockCartRepo) AddItem(_ context.Context, item *domain.CartItem) error {
	if m.itemErr != nil {
		return m.itemErr
	}
	item.ID = m.nextID
	m.nextID++
	m.items[item.CartID] = append(m.items[item.CartID], *item)
	return nil
}

func (m *mockCartRepo) UpdateItemQuantity(_ context.Context, cartID, productID int64, qty int) error {
	if m.itemErr != nil {
		return m.itemErr
	}
	for i, it := range m.items[cartID] {
		if it.ProductID == productID {
			m.items[cartID][i].Quantity = qty
			return nil
		}
	}
	return domain.ErrItemNotFound
}

func (m *mockCartRepo) RemoveItem(_ context.Context, cartID, productID int64) error {
	if m.itemErr != nil {
		return m.itemErr
	}
	items := m.items[cartID]
	for i, it := range items {
		if it.ProductID == productID {
			m.items[cartID] = append(items[:i], items[i+1:]...)
			return nil
		}
	}
	return domain.ErrItemNotFound
}

func (m *mockCartRepo) GetItems(_ context.Context, cartID int64) ([]domain.CartItem, error) {
	if m.itemErr != nil {
		return nil, m.itemErr
	}
	items := m.items[cartID]
	for i := range items {
		items[i].Subtotal = items[i].Price * float64(items[i].Quantity)
	}
	return items, nil
}

func (m *mockCartRepo) ClearCart(_ context.Context, cartID int64) error {
	if m.itemErr != nil {
		return m.itemErr
	}
	m.items[cartID] = []domain.CartItem{}
	return nil
}

// Mock del lector de productos.

type mockProductReader struct {
	products map[int64]struct{ price float64; stock int }
	err      error
}

func newMockProductReader() *mockProductReader {
	return &mockProductReader{products: make(map[int64]struct{ price float64; stock int })}
}

func (m *mockProductReader) addProduct(id int64, price float64, stock int) {
	m.products[id] = struct{ price float64; stock int }{price, stock}
}

func (m *mockProductReader) GetByID(_ context.Context, id int64) (float64, int, error) {
	if m.err != nil {
		return 0, 0, m.err
	}
	p, ok := m.products[id]
	if !ok {
		return 0, 0, domain.ErrProductNotFound
	}
	return p.price, p.stock, nil
}

func (m *mockProductReader) UpdateStock(_ context.Context, id int64, qty int) error {
	if m.err != nil {
		return m.err
	}
	p, ok := m.products[id]
	if !ok {
		return domain.ErrProductNotFound
	}
	if p.stock+qty < 0 {
		return errors.New("insufficient stock")
	}
	p.stock += qty
	m.products[id] = p
	return nil
}

// Tests

func setup() (*CartService, *mockCartRepo, *mockProductReader) {
	cr := newMockCartRepo()
	pr := newMockProductReader()
	pr.addProduct(1, 99.99, 10)
	return NewCartService(cr, pr), cr, pr
}

func TestGetOrCreateCart(t *testing.T) {
	svc, _, _ := setup()
	c1, err := svc.GetOrCreateCart(context.Background(), "user1")
	if err != nil || c1 == nil {
		t.Fatalf("unexpected: err=%v, cart=%v", err, c1)
	}
	c2, err := svc.GetOrCreateCart(context.Background(), "user1")
	if err != nil || c2.ID != c1.ID {
		t.Error("should return existing cart")
	}
}

func TestGetOrCreateCart_Error(t *testing.T) {
	svc, cr, _ := setup()
	cr.err = errors.New("db error")
	if _, err := svc.GetOrCreateCart(context.Background(), "user1"); err == nil {
		t.Error("expected error")
	}
}

func TestAddItem(t *testing.T) {
	tests := []struct {
		name    string
		prodID  int64
		qty     int
		wantErr error
	}{
		{"valid", 1, 2, nil},
		{"zero qty", 1, 0, domain.ErrInvalidQuantity},
		{"negative qty", 1, -1, domain.ErrInvalidQuantity},
		{"not found", 999, 1, domain.ErrProductNotFound},
		{"no stock", 1, 100, domain.ErrInsufficientStock},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, _ := setup()
			_, err := svc.AddItem(context.Background(), "user1", tt.prodID, tt.qty)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("want %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Errorf("unexpected: %v", err)
			}
		})
	}
}

func TestAddItem_CartError(t *testing.T) {
	svc, cr, _ := setup()
	cr.err = errors.New("cart error")
	if _, err := svc.AddItem(context.Background(), "user1", 1, 1); err == nil {
		t.Error("expected error")
	}
}

func TestAddItem_ItemError(t *testing.T) {
	svc, cr, _ := setup()
	svc.GetOrCreateCart(context.Background(), "user1")
	cr.itemErr = errors.New("item error")
	if _, err := svc.AddItem(context.Background(), "user1", 1, 1); err == nil {
		t.Error("expected error")
	}
}

func TestUpdateItem(t *testing.T) {
	svc, _, _ := setup()
	svc.AddItem(context.Background(), "user1", 1, 1)

	tests := []struct {
		name    string
		qty     int
		wantErr error
	}{
		{"valid", 3, nil},
		{"zero qty", 0, domain.ErrInvalidQuantity},
		{"no stock", 100, domain.ErrInsufficientStock},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.UpdateItem(context.Background(), "user1", 1, tt.qty)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("want %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Errorf("unexpected: %v", err)
			}
		})
	}
}

func TestUpdateItem_CartError(t *testing.T) {
	svc, cr, _ := setup()
	svc.AddItem(context.Background(), "user1", 1, 1)
	cr.err = errors.New("cart error")
	if _, err := svc.UpdateItem(context.Background(), "user1", 1, 2); err == nil {
		t.Error("expected error")
	}
}

func TestUpdateItem_ProductError(t *testing.T) {
	svc, _, pr := setup()
	svc.AddItem(context.Background(), "user1", 1, 1)
	pr.err = errors.New("product error")
	if _, err := svc.UpdateItem(context.Background(), "user1", 1, 2); err == nil {
		t.Error("expected error")
	}
}

func TestUpdateItem_UpdateError(t *testing.T) {
	svc, cr, _ := setup()
	svc.AddItem(context.Background(), "user1", 1, 1)
	cr.itemErr = errors.New("update error")
	if _, err := svc.UpdateItem(context.Background(), "user1", 1, 2); err == nil {
		t.Error("expected error")
	}
}

func TestRemoveItem(t *testing.T) {
	svc, _, _ := setup()
	svc.AddItem(context.Background(), "user1", 1, 1)
	if _, err := svc.RemoveItem(context.Background(), "user1", 1); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestRemoveItem_CartError(t *testing.T) {
	svc, cr, _ := setup()
	svc.AddItem(context.Background(), "user1", 1, 1)
	cr.err = errors.New("cart error")
	if _, err := svc.RemoveItem(context.Background(), "user1", 1); err == nil {
		t.Error("expected error")
	}
}

func TestRemoveItem_RemoveError(t *testing.T) {
	svc, cr, _ := setup()
	svc.AddItem(context.Background(), "user1", 1, 1)
	cr.itemErr = errors.New("remove error")
	if _, err := svc.RemoveItem(context.Background(), "user1", 1); err == nil {
		t.Error("expected error")
	}
}

func TestGetSummary(t *testing.T) {
	svc, _, _ := setup()

	// Carrito vacio
	summary, err := svc.GetSummary(context.Background(), "user1")
	if err != nil || summary.Status != "empty" {
		t.Error("expected empty status")
	}

	// Con items
	svc.AddItem(context.Background(), "user1", 1, 2)
	summary, err = svc.GetSummary(context.Background(), "user1")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if summary.Subtotal != 199.98 {
		t.Errorf("want subtotal 199.98, got %f", summary.Subtotal)
	}
	if summary.Tax != 199.98*domain.TaxRate {
		t.Errorf("want tax %f, got %f", 199.98*domain.TaxRate, summary.Tax)
	}
}

func TestGetSummary_Error(t *testing.T) {
	svc, cr, _ := setup()
	svc.AddItem(context.Background(), "user1", 1, 1)
	cr.err = errors.New("get error")
	if _, err := svc.GetSummary(context.Background(), "user1"); err == nil {
		t.Error("expected error")
	}
}

func TestGetSummary_ItemsError(t *testing.T) {
	svc, cr, _ := setup()
	svc.AddItem(context.Background(), "user1", 1, 1)
	cr.itemErr = errors.New("items error")
	if _, err := svc.GetSummary(context.Background(), "user1"); err == nil {
		t.Error("expected error")
	}
}

func TestClearCart(t *testing.T) {
	svc, _, _ := setup()
	svc.AddItem(context.Background(), "user1", 1, 1)
	if err := svc.ClearCart(context.Background(), "user1"); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestClearCart_Error(t *testing.T) {
	svc, cr, _ := setup()
	cr.err = errors.New("cart error")
	if err := svc.ClearCart(context.Background(), "user1"); err == nil {
		t.Error("expected error")
	}
}

func TestCheckout(t *testing.T) {
	svc, _, _ := setup()
	svc.AddItem(context.Background(), "user1", 1, 2)
	summary, err := svc.Checkout(context.Background(), "user1")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if summary.Status != "completed" {
		t.Error("expected completed status")
	}
}

func TestCheckout_EmptyCart(t *testing.T) {
	svc, _, _ := setup()
	svc.GetOrCreateCart(context.Background(), "user1")
	if _, err := svc.Checkout(context.Background(), "user1"); !errors.Is(err, domain.ErrEmptyCart) {
		t.Errorf("want ErrEmptyCart, got %v", err)
	}
}

func TestCheckout_SummaryError(t *testing.T) {
	svc, cr, _ := setup()
	cr.err = domain.ErrCartNotFound
	if _, err := svc.Checkout(context.Background(), "user1"); !errors.Is(err, domain.ErrEmptyCart) {
		t.Errorf("want ErrEmptyCart, got %v", err)
	}
}

func TestCheckout_StockError(t *testing.T) {
	svc, _, pr := setup()
	svc.AddItem(context.Background(), "user1", 1, 1)
	pr.err = errors.New("stock error")
	if _, err := svc.Checkout(context.Background(), "user1"); err == nil {
		t.Error("expected error")
	}
}

func TestCheckout_StatusError(t *testing.T) {
	svc, cr, _ := setup()
	svc.AddItem(context.Background(), "user1", 1, 1)
	cr.err = errors.New("status error")
	if _, err := svc.Checkout(context.Background(), "user1"); err == nil {
		t.Error("expected error")
	}
}
