package httpapi

import (
	"net/http"

	"github.com/example/ledgerly/internal/model"
	"github.com/example/ledgerly/internal/service"
	"github.com/example/ledgerly/internal/validator"
)

// AccountHandler handles account HTTP requests.
type AccountHandler struct {
	svc     *service.AccountService
	maxBody int64
}

// NewAccountHandler creates a new AccountHandler.
func NewAccountHandler(svc *service.AccountService, maxBody int64) *AccountHandler {
	return &AccountHandler{svc: svc, maxBody: maxBody}
}

// List handles GET /api/v1/accounts.
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := model.AccountFilter{}
	if v := q.Get("type"); v != "" {
		f.Type = model.AccountType(v)
	}
	if v := q.Get("active"); v == "true" {
		b := true
		f.IsActive = &b
	} else if v == "false" {
		b := false
		f.IsActive = &b
	}
	f.Query = q.Get("q")
	list, err := h.svc.List(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// Create handles POST /api/v1/accounts.
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req validator.CreateAccountRequest
	if err := decodeJSON(r, &req, h.maxBody); err != nil {
		writeError(w, model.ErrInvalidInput)
		return
	}
	if ve := req.Validate(); len(ve) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "validation failed", "details": ve})
		return
	}
	a := &model.Account{
		Name:     req.Name,
		Type:     model.AccountType(req.Type),
		Currency: req.Currency,
		Balance:  req.Balance,
	}
	created, err := h.svc.Create(r.Context(), a)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// Get handles GET /api/v1/accounts/{id}.
func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	a, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

// Update handles PUT /api/v1/accounts/{id}.
func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	var req service.UpdateAccountRequest
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

// Delete handles DELETE /api/v1/accounts/{id}.
func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	if err := h.svc.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
