package httpapi

import (
	"net/http"
	"strconv"

	"github.com/example/ledgerly/internal/model"
	"github.com/example/ledgerly/internal/service"
	"github.com/example/ledgerly/internal/validator"
)

// TransactionHandler handles transaction HTTP requests.
type TransactionHandler struct {
	txSvc       *service.TransactionService
	transferSvc *service.TransferService
	reportSvc   *service.ReportService
	maxBody     int64
}

// NewTransactionHandler creates a new TransactionHandler.
func NewTransactionHandler(txSvc *service.TransactionService, transferSvc *service.TransferService, reportSvc *service.ReportService, maxBody int64) *TransactionHandler {
	return &TransactionHandler{txSvc: txSvc, transferSvc: transferSvc, reportSvc: reportSvc, maxBody: maxBody}
}

// Create handles POST /api/v1/transactions.
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req validator.CreateTransactionRequest
	if err := decodeJSON(r, &req, h.maxBody); err != nil {
		writeError(w, model.ErrInvalidInput)
		return
	}
	if ve := req.Validate(); len(ve) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "validation failed", "details": ve})
		return
	}
	t := &model.Transaction{
		AccountID:   req.AccountID,
		CategoryID:  req.CategoryID,
		Amount:      req.Amount,
		Type:        model.TransactionType(req.Type),
		Description: req.Description,
		Tags:        req.Tags,
		Date:        parseDate(req.Date),
	}
	created, err := h.txSvc.Create(r.Context(), t)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// List handles GET /api/v1/transactions.
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := model.TransactionFilter{
		AccountID:  q.Get("account_id"),
		CategoryID: q.Get("category_id"),
		Type:       model.TransactionType(""),
	}
	if v := q.Get("start_date"); v != "" {
		t := parseDate(v)
		f.StartDate = &t
	}
	if v := q.Get("end_date"); v != "" {
		t := parseDate(v)
		f.EndDate = &t
	}
	list, err := h.txSvc.List(r.Context(), f)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// Get handles GET /api/v1/transactions/{id}.
func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	t, err := h.txSvc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

// Delete handles DELETE /api/v1/transactions/{id}.
func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := pathSegment(r.URL.Path, 4)
	if err := h.txSvc.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Transfer handles POST /api/v1/transfers.
func (h *TransactionHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	var req validator.CreateTransferRequest
	if err := decodeJSON(r, &req, h.maxBody); err != nil {
		writeError(w, model.ErrInvalidInput)
		return
	}
	if ve := req.Validate(); len(ve) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "validation failed", "details": ve})
		return
	}
	t := &model.Transfer{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
		Description:   req.Description,
		Date:          parseDate(req.Date),
	}
	created, err := h.transferSvc.Transfer(r.Context(), t)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// ListTransfers handles GET /api/v1/accounts/{accountID}/transfers.
func (h *TransactionHandler) ListTransfers(w http.ResponseWriter, r *http.Request) {
	accountID := pathSegment(r.URL.Path, 4)
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	list, err := h.transferSvc.ListByAccount(r.Context(), accountID, limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// MonthlyReport handles GET /api/v1/reports/monthly.
func (h *TransactionHandler) MonthlyReport(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	year, _ := strconv.Atoi(q.Get("year"))
	month, _ := strconv.Atoi(q.Get("month"))
	if year == 0 {
		year = 2026
	}
	if month == 0 {
		month = 1
	}
	report, err := h.reportSvc.MonthlyReport(r.Context(), year, month)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}
