// Package model defines the domain entities for the ledgerly finance service.
package model

import (
	"strings"
	"time"
)

// AccountType represents the type of account.
type AccountType string

const (
	AccountTypeChecking   AccountType = "checking"
	AccountTypeSavings    AccountType = "savings"
	AccountTypeCredit     AccountType = "credit"
	AccountTypeInvestment AccountType = "investment"
)

// Account represents a financial account.
type Account struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Type      AccountType `json:"type"`
	Currency  string      `json:"currency"`
	Balance   float64     `json:"balance"`
	IsActive  bool        `json:"is_active"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// Clone returns a deep copy of the account.
func (a *Account) Clone() *Account {
	if a == nil {
		return nil
	}
	cp := *a
	return &cp
}

// Normalize trims whitespace and sets default currency.
func (a *Account) Normalize() {
	a.Name = strings.TrimSpace(a.Name)
	if a.Currency == "" {
		a.Currency = "USD"
	}
}

// Credit adds a positive amount to the balance.
func (a *Account) Credit(amount float64) {
	a.Balance += amount
}

// Debit subtracts an amount from the balance.
func (a *Account) Debit(amount float64) {
	a.Balance -= amount
}

// CategoryType represents income or expense.
type CategoryType string

const (
	CategoryTypeIncome  CategoryType = "income"
	CategoryTypeExpense CategoryType = "expense"
)

// Category represents a transaction category.
type Category struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Type      CategoryType `json:"type"`
	Icon      string       `json:"icon,omitempty"`
	Color     string       `json:"color,omitempty"`
	ParentID  *string      `json:"parent_id,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// Clone returns a deep copy of the category.
func (c *Category) Clone() *Category {
	if c == nil {
		return nil
	}
	cp := *c
	if c.ParentID != nil {
		pid := *c.ParentID
		cp.ParentID = &pid
	}
	return &cp
}

// Normalize trims whitespace.
func (c *Category) Normalize() {
	c.Name = strings.TrimSpace(c.Name)
	c.Icon = strings.TrimSpace(c.Icon)
	c.Color = strings.TrimSpace(c.Color)
}

// TransactionType represents credit or debit.
type TransactionType string

const (
	TransactionTypeCredit TransactionType = "credit"
	TransactionTypeDebit  TransactionType = "debit"
)

// Transaction represents a financial transaction.
type Transaction struct {
	ID          string          `json:"id"`
	AccountID   string          `json:"account_id"`
	CategoryID  string          `json:"category_id"`
	Amount      float64         `json:"amount"`
	Type        TransactionType `json:"type"`
	Description string          `json:"description,omitempty"`
	Date        time.Time       `json:"date"`
	Tags        []string        `json:"tags,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// Clone returns a deep copy of the transaction.
func (t *Transaction) Clone() *Transaction {
	if t == nil {
		return nil
	}
	cp := *t
	if len(t.Tags) > 0 {
		cp.Tags = make([]string, len(t.Tags))
		copy(cp.Tags, t.Tags)
	}
	return &cp
}

// Normalize trims whitespace.
func (t *Transaction) Normalize() {
	t.Description = strings.TrimSpace(t.Description)
}

// NetAmount returns the signed amount (positive for credit, negative for debit).
func (t *Transaction) NetAmount() float64 {
	if t.Type == TransactionTypeCredit {
		return t.Amount
	}
	return -t.Amount
}

// Transfer represents a transfer between accounts.
type Transfer struct {
	ID            string    `json:"id"`
	FromAccountID string    `json:"from_account_id"`
	ToAccountID   string    `json:"to_account_id"`
	Amount        float64   `json:"amount"`
	Description   string    `json:"description,omitempty"`
	Date          time.Time `json:"date"`
	CreatedAt     time.Time `json:"created_at"`
}

// Clone returns a deep copy of the transfer.
func (t *Transfer) Clone() *Transfer {
	if t == nil {
		return nil
	}
	cp := *t
	return &cp
}

// Normalize trims whitespace.
func (t *Transfer) Normalize() {
	t.Description = strings.TrimSpace(t.Description)
}

// BudgetPeriod represents the budget period.
type BudgetPeriod string

const (
	BudgetPeriodMonthly BudgetPeriod = "monthly"
	BudgetPeriodYearly  BudgetPeriod = "yearly"
)

// Budget represents a spending budget for a category.
type Budget struct {
	ID         string       `json:"id"`
	CategoryID string       `json:"category_id"`
	Amount     float64      `json:"amount"`
	Period     BudgetPeriod `json:"period"`
	StartDate  time.Time    `json:"start_date"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

// Clone returns a deep copy of the budget.
func (b *Budget) Clone() *Budget {
	if b == nil {
		return nil
	}
	return b
}

// MonthlyReport represents a monthly financial summary.
type MonthlyReport struct {
	Year           int                `json:"year"`
	Month          int                `json:"month"`
	TotalIncome    float64            `json:"total_income"`
	TotalExpense   float64            `json:"total_expense"`
	NetChange      float64            `json:"net_change"`
	ByCategory     map[string]float64 `json:"by_category"`
	TopExpenses    []CategorySpending `json:"top_expenses"`
	AccountChanges map[string]float64 `json:"account_changes"`
}

// CategorySpending represents spending in a category.
type CategorySpending struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Amount       float64 `json:"amount"`
}

// AccountFilter constrains account listing queries.
type AccountFilter struct {
	Type     AccountType
	IsActive *bool
	Query    string
}

// Matches reports whether the account satisfies the filter.
func (f *AccountFilter) Matches(a *Account) bool {
	if f.Type != "" && a.Type != f.Type {
		return false
	}
	if f.IsActive != nil && a.IsActive != *f.IsActive {
		return false
	}
	if f.Query != "" {
		q := strings.ToLower(f.Query)
		return strings.Contains(strings.ToLower(a.Name), q)
	}
	return true
}

// CategoryFilter constrains category listing queries.
type CategoryFilter struct {
	Type     CategoryType
	ParentID *bool
	Query    string
}

// Matches reports whether the category satisfies the filter.
func (f *CategoryFilter) Matches(c *Category) bool {
	if f.Type != "" && c.Type != f.Type {
		return false
	}
	if f.ParentID != nil {
		hasParent := c.ParentID != nil
		if *f.ParentID && !hasParent {
			return false
		}
		if !*f.ParentID && hasParent {
			return false
		}
	}
	if f.Query != "" {
		q := strings.ToLower(f.Query)
		return strings.Contains(strings.ToLower(c.Name), q)
	}
	return true
}

// TransactionFilter constrains transaction listing queries.
type TransactionFilter struct {
	AccountID  string
	CategoryID string
	StartDate  *time.Time
	EndDate    *time.Time
	Type       TransactionType
	Tags       []string
}

// Matches reports whether the transaction satisfies the filter.
func (f *TransactionFilter) Matches(t *Transaction) bool {
	if f.AccountID != "" && t.AccountID != f.AccountID {
		return false
	}
	if f.CategoryID != "" && t.CategoryID != f.CategoryID {
		return false
	}
	if f.StartDate != nil && t.Date.Before(*f.StartDate) {
		return false
	}
	if f.EndDate != nil && t.Date.After(*f.EndDate) {
		return false
	}
	if f.Type != "" && t.Type != f.Type {
		return false
	}
	return true
}
