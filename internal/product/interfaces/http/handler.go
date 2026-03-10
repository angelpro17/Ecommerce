package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/angelpro17/Ecommerce.git/internal/product/domain"
	response "github.com/angelpro17/Ecommerce.git/internal/platform/http"
	"github.com/go-chi/chi/v5"
)

// DTOs de request/response para productos.

type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	ImageURL    string  `json:"image_url"`
}

type UpdateProductRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Stock       *int     `json:"stock,omitempty"`
	ImageURL    *string  `json:"image_url,omitempty"`
}

type ProductResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	ImageURL    string  `json:"image_url,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type ProductsListResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

func toResponse(p *domain.Product) ProductResponse {
	return ProductResponse{
		ID: p.ID, Name: p.Name, Description: p.Description,
		Price: p.Price, Stock: p.Stock, ImageURL: p.ImageURL,
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// Handler gestiona las peticiones HTTP de productos.
type Handler struct {
	service domain.Service
}

func NewHandler(service domain.Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registra las rutas de productos.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/products", func(r chi.Router) {
		r.Get("/", h.GetAll)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	p, err := h.service.Create(r.Context(), req.Name, req.Description, req.Price, req.Stock, req.ImageURL)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidName),
			errors.Is(err, domain.ErrInvalidPrice),
			errors.Is(err, domain.ErrInvalidStock):
			response.WriteError(w, http.StatusBadRequest, err.Error())
		default:
			response.WriteError(w, http.StatusInternalServerError, "Failed to create product")
		}
		return
	}
	response.WriteJSON(w, http.StatusCreated, toResponse(p))
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	p, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.WriteError(w, http.StatusNotFound, "Product not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to get product")
		return
	}
	response.WriteJSON(w, http.StatusOK, toResponse(p))
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	products, total, err := h.service.GetAll(r.Context(), page, pageSize)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "Failed to get products")
		return
	}

	items := make([]ProductResponse, len(products))
	for i, p := range products {
		items[i] = toResponse(&p)
	}
	response.WriteJSON(w, http.StatusOK, ProductsListResponse{
		Products: items, Total: total, Page: page, PageSize: pageSize,
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var req UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	p, err := h.service.Update(r.Context(), id, req.Name, req.Description, req.Price, req.Stock, req.ImageURL)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.WriteError(w, http.StatusNotFound, "Product not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to update product")
		return
	}
	response.WriteJSON(w, http.StatusOK, toResponse(p))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			response.WriteError(w, http.StatusNotFound, "Product not found")
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "Failed to delete product")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
