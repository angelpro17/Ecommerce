package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/angelpro17/Ecommerce.git/internal/cart/domain"
	response "github.com/angelpro17/Ecommerce.git/internal/platform/http"
	"github.com/go-chi/chi/v5"
)

// DTOs de request/response para el carrito.

type AddItemRequest struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type UpdateItemRequest struct {
	Quantity int `json:"quantity"`
}

type CartItemResponse struct {
	ID        int64   `json:"id"`
	CartID    int64   `json:"cart_id"`
	ProductID int64   `json:"product_id"`
	Name      string  `json:"name,omitempty"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Subtotal  float64 `json:"subtotal"`
}

type CartSummaryResponse struct {
	CartID     int64              `json:"cart_id"`
	UserID     string             `json:"user_id"`
	Items      []CartItemResponse `json:"items"`
	ItemsCount int                `json:"items_count"`
	Subtotal   float64            `json:"subtotal"`
	Tax        float64            `json:"tax"`
	Total      float64            `json:"total"`
	Status     string             `json:"status"`
}

func toSummaryResponse(s *domain.CartSummary) CartSummaryResponse {
	items := make([]CartItemResponse, len(s.Items))
	for i, it := range s.Items {
		items[i] = CartItemResponse{
			ID: it.ID, CartID: it.CartID, ProductID: it.ProductID,
			Name: it.Name, Quantity: it.Quantity, Price: it.Price, Subtotal: it.Subtotal,
		}
	}
	return CartSummaryResponse{
		CartID: s.CartID, UserID: s.UserID, Items: items,
		ItemsCount: s.ItemsCount, Subtotal: s.Subtotal,
		Tax: s.Tax, Total: s.Total, Status: s.Status,
	}
}

// Handler gestiona las peticiones HTTP del carrito.
type Handler struct {
	service domain.Service
}

func NewHandler(service domain.Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registra las rutas del carrito.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/cart", func(r chi.Router) {
		r.Get("/", h.GetSummary)
		r.Post("/items", h.AddItem)
		r.Put("/items/{productId}", h.UpdateItem)
		r.Delete("/items/{productId}", h.RemoveItem)
		r.Delete("/", h.ClearCart)
		r.Post("/checkout", h.Checkout)
	})
}

func getUserID(r *http.Request) string {
	if id := r.Header.Get("X-User-ID"); id != "" {
		return id
	}
	return "anonymous"
}

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	var req AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	summary, err := h.service.AddItem(r.Context(), getUserID(r), req.ProductID, req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidQuantity):
			response.WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, domain.ErrProductNotFound):
			response.WriteError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, domain.ErrInsufficientStock):
			response.WriteError(w, http.StatusConflict, err.Error())
		default:
			response.WriteError(w, http.StatusInternalServerError, "Failed to add item")
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, toSummaryResponse(summary))
}

func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(chi.URLParam(r, "productId"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var req UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	summary, err := h.service.UpdateItem(r.Context(), getUserID(r), productID, req.Quantity)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrItemNotFound):
			response.WriteError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, domain.ErrInvalidQuantity):
			response.WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, domain.ErrInsufficientStock):
			response.WriteError(w, http.StatusConflict, err.Error())
		default:
			response.WriteError(w, http.StatusInternalServerError, "Failed to update item")
		}
		return
	}
	response.WriteJSON(w, http.StatusOK, toSummaryResponse(summary))
}

func (h *Handler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseInt(chi.URLParam(r, "productId"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	summary, err := h.service.RemoveItem(r.Context(), getUserID(r), productID)
	if err != nil {
		if errors.Is(err, domain.ErrItemNotFound) {
			response.WriteError(w, http.StatusNotFound, err.Error())
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to remove item")
		return
	}
	response.WriteJSON(w, http.StatusOK, toSummaryResponse(summary))
}

func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.service.GetSummary(r.Context(), getUserID(r))
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "Failed to get cart")
		return
	}
	response.WriteJSON(w, http.StatusOK, toSummaryResponse(summary))
}

func (h *Handler) ClearCart(w http.ResponseWriter, r *http.Request) {
	if err := h.service.ClearCart(r.Context(), getUserID(r)); err != nil {
		if errors.Is(err, domain.ErrCartNotFound) {
			response.WriteError(w, http.StatusNotFound, "Cart not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to clear cart")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	summary, err := h.service.Checkout(r.Context(), getUserID(r))
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Checkout successful",
		"order":   toSummaryResponse(summary),
	})
}
