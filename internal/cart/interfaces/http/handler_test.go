package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/angelpro17/Ecommerce.git/internal/cart/domain"
	"github.com/go-chi/chi/v5"
)

type mockCartService struct {
	summary     *domain.CartSummary
	shouldError bool
	errorType   string
}

func newMockCartService() *mockCartService {
	return &mockCartService{
		summary: &domain.CartSummary{
			CartID: 1, UserID: "user1", Items: []domain.CartItem{},
			ItemsCount: 0, Subtotal: 0, Tax: 0, Total: 0, Status: "active",
		},
	}
}

func (m *mockCartService) setError(t string) { m.shouldError = true; m.errorType = t }

func (m *mockCartService) GetOrCreateCart(_ context.Context, _ string) (*domain.Cart, error) {
	if m.shouldError {
		return nil, domain.ErrCartNotFound
	}
	return &domain.Cart{ID: 1, UserID: "user1", Status: "active"}, nil
}

func (m *mockCartService) AddItem(_ context.Context, _ string, _ int64, _ int) (*domain.CartSummary, error) {
	if m.shouldError {
		switch m.errorType {
		case "quantity":
			return nil, domain.ErrInvalidQuantity
		case "product":
			return nil, domain.ErrProductNotFound
		case "stock":
			return nil, domain.ErrInsufficientStock
		default:
			return nil, domain.ErrCartNotFound
		}
	}
	return m.summary, nil
}

func (m *mockCartService) UpdateItem(_ context.Context, _ string, _ int64, _ int) (*domain.CartSummary, error) {
	if m.shouldError {
		switch m.errorType {
		case "item":
			return nil, domain.ErrItemNotFound
		case "quantity":
			return nil, domain.ErrInvalidQuantity
		case "stock":
			return nil, domain.ErrInsufficientStock
		default:
			return nil, domain.ErrCartNotFound
		}
	}
	return m.summary, nil
}

func (m *mockCartService) RemoveItem(_ context.Context, _ string, _ int64) (*domain.CartSummary, error) {
	if m.shouldError {
		switch m.errorType {
		case "item":
			return nil, domain.ErrItemNotFound
		default:
			return nil, errors.New("internal error")
		}
	}
	return m.summary, nil
}

func (m *mockCartService) GetSummary(_ context.Context, _ string) (*domain.CartSummary, error) {
	if m.shouldError {
		return nil, domain.ErrCartNotFound
	}
	return m.summary, nil
}

func (m *mockCartService) ClearCart(_ context.Context, _ string) error {
	if m.shouldError {
		switch m.errorType {
		case "cart":
			return domain.ErrCartNotFound
		default:
			return errors.New("internal error")
		}
	}
	return nil
}

func (m *mockCartService) Checkout(_ context.Context, _ string) (*domain.CartSummary, error) {
	if m.shouldError {
		return nil, domain.ErrCartNotFound
	}
	m.summary.Status = "completed"
	return m.summary, nil
}

// Tests

func TestHandler_AddItem(t *testing.T) {
	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
		errType    string
		userID     string
	}{
		{"valid", AddItemRequest{ProductID: 1, Quantity: 2}, http.StatusOK, "", "user1"},
		{"anonymous", AddItemRequest{ProductID: 1, Quantity: 2}, http.StatusOK, "", ""},
		{"bad json", "invalid", http.StatusBadRequest, "", ""},
		{"quantity", AddItemRequest{ProductID: 1, Quantity: 2}, http.StatusBadRequest, "quantity", ""},
		{"product", AddItemRequest{ProductID: 1, Quantity: 2}, http.StatusNotFound, "product", ""},
		{"stock", AddItemRequest{ProductID: 1, Quantity: 2}, http.StatusConflict, "stock", ""},
		{"internal", AddItemRequest{ProductID: 1, Quantity: 2}, http.StatusInternalServerError, "other", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newMockCartService()
			if tt.errType != "" {
				svc.setError(tt.errType)
			}
			h := NewHandler(svc)
			body := marshalBody(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/cart/items", bytes.NewReader(body))
			if tt.userID != "" {
				req.Header.Set("X-User-ID", tt.userID)
			}
			w := httptest.NewRecorder()
			h.AddItem(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("want %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandler_UpdateItem(t *testing.T) {
	tests := []struct {
		name       string
		productID  string
		body       interface{}
		wantStatus int
		errType    string
	}{
		{"valid", "1", UpdateItemRequest{Quantity: 3}, http.StatusOK, ""},
		{"bad id", "abc", UpdateItemRequest{Quantity: 3}, http.StatusBadRequest, ""},
		{"bad json", "1", "invalid", http.StatusBadRequest, ""},
		{"item", "1", UpdateItemRequest{Quantity: 3}, http.StatusNotFound, "item"},
		{"quantity", "1", UpdateItemRequest{Quantity: 3}, http.StatusBadRequest, "quantity"},
		{"stock", "1", UpdateItemRequest{Quantity: 3}, http.StatusConflict, "stock"},
		{"internal", "1", UpdateItemRequest{Quantity: 3}, http.StatusInternalServerError, "other"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newMockCartService()
			if tt.errType != "" {
				svc.setError(tt.errType)
			}
			h := NewHandler(svc)
			r := chi.NewRouter()
			r.Put("/cart/items/{productId}", h.UpdateItem)
			body := marshalBody(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/cart/items/"+tt.productID, bytes.NewReader(body))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("want %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandler_RemoveItem(t *testing.T) {
	tests := []struct {
		name, productID, errType string
		wantStatus               int
	}{
		{"valid", "1", "", http.StatusOK},
		{"bad id", "abc", "", http.StatusBadRequest},
		{"not found", "1", "item", http.StatusNotFound},
		{"internal", "1", "other", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newMockCartService()
			if tt.errType != "" {
				svc.setError(tt.errType)
			}
			h := NewHandler(svc)
			r := chi.NewRouter()
			r.Delete("/cart/items/{productId}", h.RemoveItem)
			req := httptest.NewRequest(http.MethodDelete, "/cart/items/"+tt.productID, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("want %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandler_GetSummary(t *testing.T) {
	tests := []struct {
		name, errType string
		wantStatus    int
	}{
		{"success", "", http.StatusOK},
		{"error", "other", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newMockCartService()
			if tt.errType != "" {
				svc.setError(tt.errType)
			}
			h := NewHandler(svc)
			req := httptest.NewRequest(http.MethodGet, "/cart", nil)
			w := httptest.NewRecorder()
			h.GetSummary(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("want %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandler_ClearCart(t *testing.T) {
	tests := []struct {
		name, errType string
		wantStatus    int
	}{
		{"success", "", http.StatusNoContent},
		{"not found", "cart", http.StatusNotFound},
		{"internal", "other", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newMockCartService()
			if tt.errType != "" {
				svc.setError(tt.errType)
			}
			h := NewHandler(svc)
			req := httptest.NewRequest(http.MethodDelete, "/cart", nil)
			w := httptest.NewRecorder()
			h.ClearCart(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("want %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandler_Checkout(t *testing.T) {
	tests := []struct {
		name, errType string
		wantStatus    int
	}{
		{"success", "", http.StatusOK},
		{"error", "other", http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newMockCartService()
			if tt.errType != "" {
				svc.setError(tt.errType)
			}
			h := NewHandler(svc)
			req := httptest.NewRequest(http.MethodPost, "/cart/checkout", nil)
			w := httptest.NewRecorder()
			h.Checkout(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("want %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHandler_RegisterRoutes(t *testing.T) {
	svc := newMockCartService()
	h := NewHandler(svc)
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	routes := []struct{ method, path string }{
		{http.MethodGet, "/cart"},
		{http.MethodPost, "/cart/items"},
		{http.MethodPut, "/cart/items/1"},
		{http.MethodDelete, "/cart/items/1"},
		{http.MethodDelete, "/cart"},
		{http.MethodPost, "/cart/checkout"},
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

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name, headerID, expected string
	}{
		{"with header", "user123", "user123"},
		{"without header", "", "anonymous"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.headerID != "" {
				req.Header.Set("X-User-ID", tt.headerID)
			}
			if result := getUserID(req); result != tt.expected {
				t.Errorf("want %s, got %s", tt.expected, result)
			}
		})
	}
}

// marshalBody serializa el body del request; si es string lo devuelve como bytes.
func marshalBody(body interface{}) []byte {
	if s, ok := body.(string); ok {
		return []byte(s)
	}
	b, _ := json.Marshal(body)
	return b
}
