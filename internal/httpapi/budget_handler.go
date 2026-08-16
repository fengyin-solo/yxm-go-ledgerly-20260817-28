package httpapi

import (
	"net/http"

	"github.com/example/ledgerly/internal/model"
	"github.com/example/ledgerly/internal/service"
	"github.com/example/ledgerly/internal/validator"
)

// BudgetHandler handles budget HTTP requests.
type BudgetHandler struct {
	svc     *service.BudgetService
	maxBody int64
}

// NewBudgetHandler creates a new BudgetHandler.
func NewBudgetHandler(svc *service.BudgetService, maxBody int64) *BudgetHandler {
	return &BudgetHandler{svc: svc, maxBody: maxBody}
}

// List handles GET /api/v1/budgets.
func (h *BudgetHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// Create handles POST /api/v1/budgets.
func (h *BudgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req validator.CreateBudgetRequest
	if err := decodeJSON(r, &req, h.maxBody); err != nil {
		writeError(w, model.ErrInvalidInput)
		return
	}
	if ve := req.Validate(); len(ve) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "validation failed", "details": ve})
		return
	}
	b := &model.Budget{
		CategoryID: req.CategoryID,
		Amount:     req.Amount,
		Period:     model.BudgetPeriod(req.Period),
		StartDate:  parseDate(req.StartDate),
	}
	created, err := h.svc.Create(r.Context(), b)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// Get handles GET /api/v1/budgets/{id}.
func (h *BudgetHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	b, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

// Update handles PUT /api/v1/budgets/{id}.
func (h *BudgetHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	var req service.UpdateBudgetRequest
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

// Delete handles DELETE /api/v1/budgets/{id}.
func (h *BudgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
