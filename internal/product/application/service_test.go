package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/angelpro17/Ecommerce.git/internal/product/domain"
)

type mockRepo struct {
	products map[int64]*domain.Product
	nextID   int64
	err      error
}

func newMockRepo() *mockRepo {
	return &mockRepo{products: make(map[int64]*domain.Product), nextID: 1}
}

func (m *mockRepo) Create(_ context.Context, p *domain.Product) error {
	if m.err != nil {
		return m.err
	}
	now := time.Now()
	p.ID, p.CreatedAt, p.UpdatedAt = m.nextID, now, now
	m.products[p.ID] = p
	m.nextID++
	return nil
}

func (m *mockRepo) GetByID(_ context.Context, id int64) (*domain.Product, error) {
	if m.err != nil {
		return nil, m.err
	}
	p, ok := m.products[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return p, nil
}

func (m *mockRepo) GetAll(_ context.Context, _, _ int) ([]domain.Product, int, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	var list []domain.Product
	for _, p := range m.products {
		list = append(list, *p)
	}
	return list, len(list), nil
}

func (m *mockRepo) Update(_ context.Context, p *domain.Product) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.products[p.ID]; !ok {
		return domain.ErrNotFound
	}
	m.products[p.ID] = p
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id int64) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.products[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.products, id)
	return nil
}

func (m *mockRepo) UpdateStock(_ context.Context, id int64, qty int) error {
	if m.err != nil {
		return m.err
	}
	p, ok := m.products[id]
	if !ok {
		return domain.ErrNotFound
	}
	if p.Stock+qty < 0 {
		return errors.New("insufficient stock")
	}
	p.Stock += qty
	return nil
}

// Tests

func TestCreate(t *testing.T) {
	tests := []struct {
		name    string
		pName   string
		price   float64
		stock   int
		wantErr error
	}{
		{"valid", "Laptop", 999.99, 10, nil},
		{"empty name", "", 99.99, 10, domain.ErrInvalidName},
		{"zero price", "X", 0, 10, domain.ErrInvalidPrice},
		{"negative price", "X", -10, 10, domain.ErrInvalidPrice},
		{"negative stock", "X", 99.99, -1, domain.ErrInvalidStock},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewProductService(newMockRepo())
			p, err := svc.Create(context.Background(), tt.pName, "desc", tt.price, tt.stock, "")
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("want %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if p.Name != tt.pName {
					t.Errorf("want name %s, got %s", tt.pName, p.Name)
				}
			}
		})
	}
}

func TestCreate_RepoError(t *testing.T) {
	repo := newMockRepo()
	repo.err = errors.New("db error")
	svc := NewProductService(repo)
	if _, err := svc.Create(context.Background(), "Test", "d", 9.99, 1, ""); err == nil {
		t.Error("expected error")
	}
}

func TestGetByID(t *testing.T) {
	repo := newMockRepo()
	svc := NewProductService(repo)
	created, _ := svc.Create(context.Background(), "Test", "d", 9.99, 1, "")

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{"found", created.ID, false},
		{"not found", 999, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetByID(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestGetAll(t *testing.T) {
	repo := newMockRepo()
	svc := NewProductService(repo)
	for i := 0; i < 3; i++ {
		svc.Create(context.Background(), "P", "d", 10, 5, "")
	}

	tests := []struct {
		name             string
		page, pageSize   int
	}{
		{"defaults", 0, 0},
		{"normal", 1, 10},
		{"large size", 1, 200},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			products, total, err := svc.GetAll(context.Background(), tt.page, tt.pageSize)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if products == nil {
				t.Error("expected products")
			}
			if total != 3 {
				t.Errorf("want total 3, got %d", total)
			}
		})
	}
}

func TestGetAll_RepoError(t *testing.T) {
	repo := newMockRepo()
	repo.err = errors.New("db error")
	svc := NewProductService(repo)
	if _, _, err := svc.GetAll(context.Background(), 1, 10); err == nil {
		t.Error("expected error")
	}
}

func TestUpdate(t *testing.T) {
	repo := newMockRepo()
	svc := NewProductService(repo)
	created, _ := svc.Create(context.Background(), "Original", "d", 50, 10, "")

	newName := "Updated"
	newPrice := 75.0
	newStock := 20
	newDesc := "New"
	empty := ""
	badPrice := -10.0
	badStock := -5
	newURL := "http://img.jpg"

	tests := []struct {
		name    string
		id      int64
		n, d    *string
		p       *float64
		s       *int
		u       *string
		wantErr error
	}{
		{"name", created.ID, &newName, nil, nil, nil, nil, nil},
		{"price", created.ID, nil, nil, &newPrice, nil, nil, nil},
		{"stock", created.ID, nil, nil, nil, &newStock, nil, nil},
		{"desc", created.ID, nil, &newDesc, nil, nil, nil, nil},
		{"url", created.ID, nil, nil, nil, nil, &newURL, nil},
		{"empty name", created.ID, &empty, nil, nil, nil, nil, domain.ErrInvalidName},
		{"bad price", created.ID, nil, nil, &badPrice, nil, nil, domain.ErrInvalidPrice},
		{"bad stock", created.ID, nil, nil, nil, &badStock, nil, domain.ErrInvalidStock},
		{"not found", 999, &newName, nil, nil, nil, nil, domain.ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Update(context.Background(), tt.id, tt.n, tt.d, tt.p, tt.s, tt.u)
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

func TestUpdate_RepoError(t *testing.T) {
	repo := newMockRepo()
	svc := NewProductService(repo)
	created, _ := svc.Create(context.Background(), "T", "d", 50, 10, "")
	repo.err = errors.New("update error")
	n := "Updated"
	if _, err := svc.Update(context.Background(), created.ID, &n, nil, nil, nil, nil); err == nil {
		t.Error("expected error")
	}
}

func TestDelete(t *testing.T) {
	repo := newMockRepo()
	svc := NewProductService(repo)
	created, _ := svc.Create(context.Background(), "Del", "d", 50, 10, "")

	tests := []struct {
		name    string
		id      int64
		wantErr bool
	}{
		{"found", created.ID, false},
		{"not found", 999, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Delete(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}
