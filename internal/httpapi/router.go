package httpapi

import (
	"net/http"

	"github.com/example/ledgerly/internal/auth"
)

// Router holds all HTTP handlers.
type Router struct {
	accounts     *AccountHandler
	categories   *CategoryHandler
	transactions *TransactionHandler
	budgets      *BudgetHandler
}

// NewRouter creates a new Router.
func NewRouter(accounts *AccountHandler, categories *CategoryHandler, transactions *TransactionHandler, budgets *BudgetHandler) *Router {
	return &Router{
		accounts:     accounts,
		categories:   categories,
		transactions: transactions,
		budgets:      budgets,
	}
}

// Handler returns the top-level HTTP handler with all routes registered.
func (rt *Router) Handler(authenticator *auth.Authenticator) http.Handler {
	mux := http.NewServeMux()

	// Accounts
	mux.HandleFunc("GET /api/v1/accounts", rt.accounts.List)
	mux.HandleFunc("POST /api/v1/accounts", rt.accounts.Create)
	mux.HandleFunc("GET /api/v1/accounts/{id}", rt.accounts.Get)
	mux.HandleFunc("PUT /api/v1/accounts/{id}", rt.accounts.Update)
	mux.HandleFunc("DELETE /api/v1/accounts/{id}", rt.accounts.Delete)
	mux.HandleFunc("GET /api/v1/accounts/{accountID}/transfers", rt.transactions.ListTransfers)

	// Categories
	mux.HandleFunc("GET /api/v1/categories", rt.categories.List)
	mux.HandleFunc("POST /api/v1/categories", rt.categories.Create)
	mux.HandleFunc("GET /api/v1/categories/{id}", rt.categories.Get)
	mux.HandleFunc("PUT /api/v1/categories/{id}", rt.categories.Update)
	mux.HandleFunc("DELETE /api/v1/categories/{id}", rt.categories.Delete)

	// Transactions
	mux.HandleFunc("GET /api/v1/transactions", rt.transactions.List)
	mux.HandleFunc("POST /api/v1/transactions", rt.transactions.Create)
	mux.HandleFunc("GET /api/v1/transactions/{id}", rt.transactions.Get)
	mux.HandleFunc("DELETE /api/v1/transactions/{id}", rt.transactions.Delete)

	// Transfers
	mux.HandleFunc("POST /api/v1/transfers", rt.transactions.Transfer)

	// Budgets
	mux.HandleFunc("GET /api/v1/budgets", rt.budgets.List)
	mux.HandleFunc("POST /api/v1/budgets", rt.budgets.Create)
	mux.HandleFunc("GET /api/v1/budgets/{id}", rt.budgets.Get)
	mux.HandleFunc("PUT /api/v1/budgets/{id}", rt.budgets.Update)
	mux.HandleFunc("DELETE /api/v1/budgets/{id}", rt.budgets.Delete)

	// Reports
	mux.HandleFunc("GET /api/v1/reports/monthly", rt.transactions.MonthlyReport)

	return mux
}
