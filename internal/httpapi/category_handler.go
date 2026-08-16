package httpapi

import (
	"net/http"

	"github.com/example/ledgerly/internal/model"
	"github.com/example/ledgerly/internal/service"
	"github.com/example/ledgerly/internal/validator"
)

// CategoryHandler handles category HTTP requests.
type CategoryHandler struct {
	svc     *service.CategoryService
	maxBody int64
}

// NewCategoryHandler creates a new CategoryHandler.
func NewCategoryHandler(svc *service.CategoryService, maxBody int64) *CategoryHandler {
	return &CategoryHandler{svc: svc, maxBody: maxBody}
}

// List handles GET /api/v1/categories.
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := model.CategoryFilter{}
	if v := q.Get("type"); v != "" {
		f.Type = model.CategoryType(v)
	}
	if v := q.Get("top_level"); v == "true" {
		b := false
		f.ParentID = &b
	}
	f.Query = q.Get("q")
	list, err := h.svc.List(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// Create handles POST /api/v1/categories.
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req validator.CreateCategoryRequest
	if err := decodeJSON(r, &req, h.maxBody); err != nil {
		writeError(w, model.ErrInvalidInput)
		return
	}
	if ve := req.Validate(); len(ve) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "validation failed", "details": ve})
		return
	}
	c := &model.Category{
		Name:     req.Name,
		Type:     model.CategoryType(req.Type),
		Icon:     req.Icon,
		Color:    req.Color,
		ParentID: req.ParentID,
	}
	created, err := h.svc.Create(r.Context(), c)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// Get handles GET /api/v1/categories/{id}.
func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	c, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// Update handles PUT /api/v1/categories/{id}.
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	var req service.UpdateCategoryRequest
	if err := decodeJSON(r, &req, h.maxBody); err != nil {
		writeError(w, model.ErrInvalidInput)
		return
	}
	updated, err := h.svc.Update(r.Context(), id, &req)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// Delete handles DELETE /api/v1/categories/{id}.
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
