package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/angelpro17/Ecommerce.git/internal/product/domain"
	"github.com/go-chi/chi/v5"
)

type mockService struct {
	products    map[int64]*domain.Product
	nextID      int64
	shouldError bool
	errorType   string
}

func newMockService() *mockService {
	return &mockService{products: make(map[int64]*domain.Product), nextID: 1}
}

func (m *mockService) setError(t string) { m.shouldError = true; m.errorType = t }

func (m *mockService) Create(_ context.Context, name, desc string, price float64, stock int, url string) (*domain.Product, error) {
	if m.shouldError {
		switch m.errorType {
		case "name":
			return nil, domain.ErrInvalidName
		case "price":
			return nil, domain.ErrInvalidPrice
		case "stock":
			return nil, domain.ErrInvalidStock
		default:
			return nil, errors.New("internal error")
		}
	}
	p := &domain.Product{ID: m.nextID, Name: name, Description: desc, Price: price, Stock: stock, ImageURL: url}
	m.products[p.ID] = p
	m.nextID++
	return p, nil
}

func (m *mockService) GetByID(_ context.Context, id int64) (*domain.Product, error) {
	if m.shouldError {
		if m.errorType == "internal" {
			return nil, errors.New("internal error")
		}
		return nil, domain.ErrNotFound
	}
	p, ok := m.products[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return p, nil
}

func (m *mockService) GetAll(_ context.Context, _, _ int) ([]domain.Product, int, error) {
	if m.shouldError {
		return nil, 0, errors.New("internal error")
	}
	var list []domain.Product
	for _, p := range m.products {
		list = append(list, *p)
	}
	return list, len(list), nil
}

func (m *mockService) Update(_ context.Context, id int64, name, desc *string, _ *float64, _ *int, _ *string) (*domain.Product, error) {
	if m.shouldError {
		if m.errorType == "internal" {
			return nil, errors.New("internal error")
		}
		return nil, domain.ErrNotFound
	}
	p, ok := m.products[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if name != nil {
		p.Name = *name
	}
	return p, nil
}

func (m *mockService) Delete(_ context.Context, id int64) error {
	if m.shouldError {
		if m.errorType == "internal" {
			return errors.New("internal error")
		}
		return domain.ErrNotFound
	}
	if _, ok := m.products[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.products, id)
	return nil
}

// Tests

func TestHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
		errType    string
	}{
		{"valid", CreateProductRequest{Name: "Test", Price: 9.99, Stock: 10}, http.StatusCreated, ""},
		{"invalid json", "bad", http.StatusBadRequest, ""},
		{"name err", CreateProductRequest{Name: "T", Price: 9.99, Stock: 10}, http.StatusBadRequest, "name"},
		{"price err", CreateProductRequest{Name: "T", Price: 9.99, Stock: 10}, http.StatusBadRequest, "price"},
		{"stock err", CreateProductRequest{Name: "T", Price: 9.99, Stock: 10}, http.StatusBadRequest, "stock"},
		{"internal", CreateProductRequest{Name: "T", Price: 9.99, Stock: 10}, http.StatusInternalServerError, "other"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newMockService()
			if tt.errType != "" {
				svc.setError(tt.errType)
			}
			h := NewHandler(svc)
			body := marshalBody(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
			w := httptest.NewRecorder()
			h.Create(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("want %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandler_GetByID(t *testing.T) {
	svc := newMockService()
	svc.Create(context.Background(), "Test", "d", 9.99, 10, "")
	h := NewHandler(svc)

	tests := []struct {
		name, id   string
		wantStatus int
	}{
		{"found", "1", http.StatusOK},
		{"bad id", "abc", http.StatusBadRequest},
		{"not found", "999", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Get("/products/{id}", h.GetByID)
			req := httptest.NewRequest(http.MethodGet, "/products/"+tt.id, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("want %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandler_GetByID_Internal(t *testing.T) {
	svc := newMockService()
	svc.Create(context.Background(), "T", "d", 9.99, 10, "")
	svc.setError("internal")
	h := NewHandler(svc)
	r := chi.NewRouter()
	r.Get("/products/{id}", h.GetByID)
	req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", w.Code)
	}
}

func TestHandler_GetAll(t *testing.T) {
	svc := newMockService()
	h := NewHandler(svc)

	tests := []struct {
		name, query string
	}{
		{"no params", ""},
		{"with page", "?page=1&page_size=10"},
		{"defaults", "?page=-1&page_size=0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/products"+tt.query, nil)
			w := httptest.NewRecorder()
			h.GetAll(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("want 200, got %d", w.Code)
			}
		})
	}
}

func TestHandler_GetAll_Error(t *testing.T) {
	svc := newMockService()
	svc.setError("other")
	h := NewHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	w := httptest.NewRecorder()
	h.GetAll(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", w.Code)
	}
}

func TestHandler_Update(t *testing.T) {
	svc := newMockService()
	svc.Create(context.Background(), "Test", "d", 9.99, 10, "")
	h := NewHandler(svc)
	newName := "Updated"

	tests := []struct {
		name, id   string
		body       interface{}
		wantStatus int
	}{
		{"valid", "1", UpdateProductRequest{Name: &newName}, http.StatusOK},
		{"bad id", "abc", UpdateProductRequest{Name: &newName}, http.StatusBadRequest},
		{"bad body", "1", "invalid", http.StatusBadRequest},
		{"not found", "999", UpdateProductRequest{Name: &newName}, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Put("/products/{id}", h.Update)
			body := marshalBody(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/products/"+tt.id, bytes.NewReader(body))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("want %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandler_Update_Internal(t *testing.T) {
	svc := newMockService()
	svc.Create(context.Background(), "T", "d", 9.99, 10, "")
	svc.setError("internal")
	h := NewHandler(svc)
	r := chi.NewRouter()
	r.Put("/products/{id}", h.Update)
	n := "U"
	body, _ := json.Marshal(UpdateProductRequest{Name: &n})
	req := httptest.NewRequest(http.MethodPut, "/products/1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", w.Code)
	}
}

func TestHandler_Delete(t *testing.T) {
	tests := []struct {
		name, id   string
		wantStatus int
		setup      func(*mockService)
	}{
		{"valid", "1", http.StatusNoContent, func(m *mockService) {
			m.Create(context.Background(), "T", "d", 9.99, 10, "")
		}},
		{"bad id", "abc", http.StatusBadRequest, func(m *mockService) {}},
		{"not found", "999", http.StatusNotFound, func(m *mockService) {}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newMockService()
			tt.setup(svc)
			h := NewHandler(svc)
			r := chi.NewRouter()
			r.Delete("/products/{id}", h.Delete)
			req := httptest.NewRequest(http.MethodDelete, "/products/"+tt.id, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("want %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandler_Delete_Internal(t *testing.T) {
	svc := newMockService()
	svc.Create(context.Background(), "T", "d", 9.99, 10, "")
	svc.setError("internal")
	h := NewHandler(svc)
	r := chi.NewRouter()
	r.Delete("/products/{id}", h.Delete)
	req := httptest.NewRequest(http.MethodDelete, "/products/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("want 500, got %d", w.Code)
	}
}

func TestHandler_RegisterRoutes(t *testing.T) {
	svc := newMockService()
	svc.Create(context.Background(), "Test", "d", 9.99, 10, "")
	h := NewHandler(svc)
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	routes := []struct{ method, path string }{
		{http.MethodGet, "/products"},
		{http.MethodPost, "/products"},
		{http.MethodGet, "/products/1"},
		{http.MethodPut, "/products/1"},
		{http.MethodDelete, "/products/1"},
	}
	for _, rt := range routes {
		req := httptest.NewRequest(rt.method, rt.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code == http.StatusNotFound {
			t.Errorf("route %s %s not registered", rt.method, rt.path)
		}
	}
}

// marshalBody serializa el body; si es string lo devuelve como bytes.
func marshalBody(body interface{}) []byte {
	if s, ok := body.(string); ok {
		return []byte(s)
	}
	b, _ := json.Marshal(body)
	return b
}
