package validator

import (
	"strings"

	"github.com/example/ledgerly/internal/model"
)

// CreateAccountRequest is the payload for creating an account.
type CreateAccountRequest struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}

// Validate checks the request and returns any errors.
func (r CreateAccountRequest) Validate() (ve model.ValidationErrors) {
	if r.Name = trimSpace(r.Name); r.Name == "" {
		ve = append(ve, model.ValidationError{Field: "name", Message: "required"})
	}
	if r.Type != "" && r.Type != "checking" && r.Type != "savings" && r.Type != "credit" && r.Type != "investment" {
		ve = append(ve, model.ValidationError{Field: "type", Message: "must be checking, savings, credit, or investment"})
	}
	if r.Balance < 0 {
		ve = append(ve, model.ValidationError{Field: "balance", Message: "must be non-negative"})
	}
	return ve
}

// CreateCategoryRequest is the payload for creating a category.
type CreateCategoryRequest struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Icon     string  `json:"icon,omitempty"`
	Color    string  `json:"color,omitempty"`
	ParentID *string `json:"parent_id,omitempty"`
}

// Validate checks the request and returns any errors.
func (r CreateCategoryRequest) Validate() (ve model.ValidationErrors) {
	if r.Name = trimSpace(r.Name); r.Name == "" {
		ve = append(ve, model.ValidationError{Field: "name", Message: "required"})
	}
	if r.Type != "income" && r.Type != "expense" {
		ve = append(ve, model.ValidationError{Field: "type", Message: "must be income or expense"})
	}
	return ve
}

// CreateTransactionRequest is the payload for creating a transaction.
type CreateTransactionRequest struct {
	AccountID   string   `json:"account_id"`
	CategoryID  string   `json:"category_id"`
	Amount      float64  `json:"amount"`
	Type        string   `json:"type"`
	Date        string   `json:"date,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// Validate checks the request and returns any errors.
func (r CreateTransactionRequest) Validate() (ve model.ValidationErrors) {
	if r.AccountID = trimSpace(r.AccountID); r.AccountID == "" {
		ve = append(ve, model.ValidationError{Field: "account_id", Message: "required"})
	}
	if r.CategoryID = trimSpace(r.CategoryID); r.CategoryID == "" {
		ve = append(ve, model.ValidationError{Field: "category_id", Message: "required"})
	}
	if r.Amount < 0 {
		ve = append(ve, model.ValidationError{Field: "amount", Message: "must be positive"})
	}
	if r.Type != "credit" && r.Type != "debit" {
		ve = append(ve, model.ValidationError{Field: "type", Message: "must be credit or debit"})
	}
	return ve
}

// CreateTransferRequest is the payload for creating a transfer.
type CreateTransferRequest struct {
	FromAccountID string  `json:"from_account_id"`
	ToAccountID   string  `json:"to_account_id"`
	Amount        float64 `json:"amount"`
	Date          string  `json:"date,omitempty"`
	Description   string  `json:"description,omitempty"`
}

// Validate checks the request and returns any errors.
func (r CreateTransferRequest) Validate() (ve model.ValidationErrors) {
	if r.FromAccountID = trimSpace(r.FromAccountID); r.FromAccountID == "" {
		ve = append(ve, model.ValidationError{Field: "from_account_id", Message: "required"})
	}
	if r.ToAccountID = trimSpace(r.ToAccountID); r.ToAccountID == "" {
		ve = append(ve, model.ValidationError{Field: "to_account_id", Message: "required"})
	}
	if r.Amount <= 0 {
		ve = append(ve, model.ValidationError{Field: "amount", Message: "must be positive"})
	}
	return ve
}

// CreateBudgetRequest is the payload for creating a budget.
type CreateBudgetRequest struct {
	CategoryID string  `json:"category_id"`
	Amount     float64 `json:"amount"`
	Period     string  `json:"period,omitempty"`
	StartDate  string  `json:"start_date,omitempty"`
}

// Validate checks the request and returns any errors.
func (r CreateBudgetRequest) Validate() (ve model.ValidationErrors) {
	if r.CategoryID = trimSpace(r.CategoryID); r.CategoryID == "" {
		ve = append(ve, model.ValidationError{Field: "category_id", Message: "required"})
	}
	if r.Amount <= 0 {
		ve = append(ve, model.ValidationError{Field: "amount", Message: "must be positive"})
	}
	if r.Period != "" && r.Period != "monthly" && r.Period != "yearly" {
		ve = append(ve, model.ValidationError{Field: "period", Message: "must be monthly or yearly"})
	}
	return ve
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}
