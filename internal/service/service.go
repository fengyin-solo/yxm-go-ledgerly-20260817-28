package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"time"

	"github.com/example/ledgerly/internal/model"
	"github.com/example/ledgerly/internal/store"
)

// AccountService handles account business logic.
type AccountService struct {
	store store.AccountStore
	now   func() time.Time
}

// NewAccountService creates a new AccountService.
func NewAccountService(store store.AccountStore) *AccountService {
	return &AccountService{store: store, now: time.Now}
}

// Create creates a new account.
func (s *AccountService) Create(ctx context.Context, a *model.Account) (*model.Account, error) {
	a.Normalize()
	if a.Name == "" {
		return nil, fmt.Errorf("%w: name is required", model.ErrInvalidInput)
	}
	if a.Type == "" {
		a.Type = model.AccountTypeChecking
	}
	if a.ID == "" {
		a.ID = newID()
	}
	a.Balance = math.Round(a.Balance*100) / 100
	a.IsActive = true
	a.CreatedAt = s.now().UTC()
	a.UpdatedAt = a.CreatedAt
	if err := s.store.Create(ctx, a); err != nil {
		return nil, err
	}
	return a.Clone(), nil
}

// GetByID returns an account by ID.
func (s *AccountService) GetByID(ctx context.Context, id string) (*model.Account, error) {
	return s.store.GetByID(ctx, id)
}

// Update updates an existing account.
func (s *AccountService) Update(ctx context.Context, id string, req *UpdateAccountRequest) (*model.Account, error) {
	a, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	req.Apply(a)
	a.Normalize()
	if a.Name == "" {
		return nil, fmt.Errorf("%w: name is required", model.ErrInvalidInput)
	}
	a.UpdatedAt = s.now().UTC()
	if err := s.store.Update(ctx, a); err != nil {
		return nil, err
	}
	return a.Clone(), nil
}

// Delete deletes an account.
func (s *AccountService) Delete(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}

// List returns accounts matching the filter.
func (s *AccountService) List(ctx context.Context, f model.AccountFilter) ([]*model.Account, error) {
	return s.store.List(ctx, f)
}

// AdjustBalance updates the account balance.
func (s *AccountService) AdjustBalance(ctx context.Context, id string, delta float64) error {
	a, err := s.store.GetByID(ctx, id)
	if err != nil {
		return err
	}
	a.Balance = math.Round((a.Balance+delta)*100) / 100
	a.UpdatedAt = s.now().UTC()
	return s.store.Update(ctx, a)
}

// UpdateAccountRequest describes fields that can be updated on an account.
type UpdateAccountRequest struct {
	Name     *string  `json:"name"`
	Currency *string  `json:"currency"`
	IsActive *bool    `json:"is_active"`
	Balance  *float64 `json:"balance"`
}

func (r *UpdateAccountRequest) Apply(a *model.Account) {
	if r.Name != nil {
		a.Name = *r.Name
	}
	if r.Currency != nil {
		a.Currency = *r.Currency
	}
	if r.IsActive != nil {
		a.IsActive = *r.IsActive
	}
	if r.Balance != nil {
		a.Balance = math.Round(*r.Balance*100) / 100
	}
}

// CategoryService handles category business logic.
type CategoryService struct {
	store store.CategoryStore
	now   func() time.Time
}

// NewCategoryService creates a new CategoryService.
func NewCategoryService(store store.CategoryStore) *CategoryService {
	return &CategoryService{store: store, now: time.Now}
}

// Create creates a new category.
func (s *CategoryService) Create(ctx context.Context, c *model.Category) (*model.Category, error) {
	c.Normalize()
	if c.Name == "" {
		return nil, fmt.Errorf("%w: name is required", model.ErrInvalidInput)
	}
	if c.Type == "" {
		return nil, fmt.Errorf("%w: type is required", model.ErrInvalidInput)
	}
	if c.ID == "" {
		c.ID = newID()
	}
	c.CreatedAt = s.now().UTC()
	c.UpdatedAt = c.CreatedAt
	if err := s.store.Create(ctx, c); err != nil {
		return nil, err
	}
	return c.Clone(), nil
}

// GetByID returns a category by ID.
func (s *CategoryService) GetByID(ctx context.Context, id string) (*model.Category, error) {
	return s.store.GetByID(ctx, id)
}

// Update updates an existing category.
func (s *CategoryService) Update(ctx context.Context, id string, req *UpdateCategoryRequest) (*model.Category, error) {
	c, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	req.Apply(c)
	c.Normalize()
	if c.Name == "" {
		return nil, fmt.Errorf("%w: name is required", model.ErrInvalidInput)
	}
	c.UpdatedAt = s.now().UTC()
	if err := s.store.Update(ctx, c); err != nil {
		return nil, err
	}
	return c.Clone(), nil
}

// Delete deletes a category.
func (s *CategoryService) Delete(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}

// List returns categories matching the filter.
func (s *CategoryService) List(ctx context.Context, f model.CategoryFilter) ([]*model.Category, error) {
	return s.store.List(ctx, f)
}

// UpdateCategoryRequest describes fields that can be updated on a category.
type UpdateCategoryRequest struct {
	Name     *string `json:"name"`
	Icon     *string `json:"icon"`
	Color    *string `json:"color"`
	ParentID *string `json:"parent_id"`
}

func (r *UpdateCategoryRequest) Apply(c *model.Category) {
	if r.Name != nil {
		c.Name = *r.Name
	}
	if r.Icon != nil {
		c.Icon = *r.Icon
	}
	if r.Color != nil {
		c.Color = *r.Color
	}
	if r.ParentID != nil {
		c.ParentID = r.ParentID
	}
}

// TransactionService handles transaction business logic.
type TransactionService struct {
	store        store.TransactionStore
	accountStore store.AccountStore
	now          func() time.Time
}

// NewTransactionService creates a new TransactionService.
func NewTransactionService(store store.TransactionStore, accountStore store.AccountStore) *TransactionService {
	return &TransactionService{store: store, accountStore: accountStore, now: time.Now}
}

// Create creates a new transaction and adjusts the account balance.
func (s *TransactionService) Create(ctx context.Context, t *model.Transaction) (*model.Transaction, error) {
	t.Normalize()
	if t.Amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be positive", model.ErrInvalidInput)
	}
	if t.AccountID == "" {
		return nil, fmt.Errorf("%w: account_id is required", model.ErrInvalidInput)
	}
	if t.CategoryID == "" {
		return nil, fmt.Errorf("%w: category_id is required", model.ErrInvalidInput)
	}
	// Check account exists
	account, err := s.accountStore.GetByID(ctx, t.AccountID)
	if err != nil {
		return nil, err
	}
	if !account.IsActive {
		return nil, fmt.Errorf("%w: account is inactive", model.ErrConflict)
	}
	if t.ID == "" {
		t.ID = newID()
	}
	t.Amount = math.Round(t.Amount*100) / 100
	if t.Date.IsZero() {
		t.Date = s.now().UTC()
	}
	t.CreatedAt = s.now().UTC()
	// Adjust account balance
	if t.Type == model.TransactionTypeCredit {
		account.Credit(t.Amount)
	} else {
		account.Debit(t.Amount)
	}
	if err := s.accountStore.Update(ctx, account); err != nil {
		return nil, err
	}
	if err := s.store.Create(ctx, t); err != nil {
		return nil, err
	}
	return t.Clone(), nil
}

// GetByID returns a transaction by ID.
func (s *TransactionService) GetByID(ctx context.Context, id string) (*model.Transaction, error) {
	return s.store.GetByID(ctx, id)
}

// List returns transactions matching the filter.
func (s *TransactionService) List(ctx context.Context, f model.TransactionFilter) ([]*model.Transaction, error) {
	return s.store.List(ctx, f)
}

// Delete deletes a transaction and reverses its effect on the account balance.
func (s *TransactionService) Delete(ctx context.Context, id string) error {
	t, err := s.store.GetByID(ctx, id)
	if err != nil {
		return err
	}
	// Reverse the balance change
	account, err := s.accountStore.GetByID(ctx, t.AccountID)
	if err != nil {
		return err
	}
	if t.Type == model.TransactionTypeCredit {
		account.Debit(t.Amount)
	} else {
		account.Credit(t.Amount)
	}
	if err := s.accountStore.Update(ctx, account); err != nil {
		return err
	}
	return s.store.Delete(ctx, id)
}

// TransferService handles transfer business logic.
type TransferService struct {
	store        store.TransferStore
	accountStore store.AccountStore
	now          func() time.Time
}

// NewTransferService creates a new TransferService.
func NewTransferService(store store.TransferStore, accountStore store.AccountStore) *TransferService {
	return &TransferService{store: store, accountStore: accountStore, now: time.Now}
}

// Transfer performs a transfer between two accounts.
func (s *TransferService) Transfer(ctx context.Context, t *model.Transfer) (*model.Transfer, error) {
	t.Normalize()
	if t.Amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be positive", model.ErrInvalidInput)
	}
	if t.FromAccountID == "" || t.ToAccountID == "" {
		return nil, fmt.Errorf("%w: from_account_id and to_account_id are required", model.ErrInvalidInput)
	}
	if t.FromAccountID == t.ToAccountID {
		return nil, fmt.Errorf("%w: cannot transfer to the same account", model.ErrConflict)
	}
	from, err := s.accountStore.GetByID(ctx, t.FromAccountID)
	if err != nil {
		return nil, err
	}
	if !from.IsActive {
		return nil, fmt.Errorf("%w: source account is inactive", model.ErrConflict)
	}
	to, err := s.accountStore.GetByID(ctx, t.ToAccountID)
	if err != nil {
		return nil, err
	}
	if !to.IsActive {
		return nil, fmt.Errorf("%w: destination account is inactive", model.ErrConflict)
	}
	if from.Balance <= t.Amount {
		return nil, fmt.Errorf("%w: insufficient funds", model.ErrInvalidInput)
	}
	t.Amount = math.Round(t.Amount*100) / 100
	if t.ID == "" {
		t.ID = newID()
	}
	if t.Date.IsZero() {
		t.Date = s.now().UTC()
	}
	t.CreatedAt = s.now().UTC()
	from.Debit(t.Amount)
	_ = to
	if err := s.accountStore.Update(ctx, from); err != nil {
		return nil, err
	}
	if err := s.accountStore.Update(ctx, to); err != nil {
		return nil, err
	}
	if err := s.store.Create(ctx, t); err != nil {
		return nil, err
	}
	return t.Clone(), nil
}

// ListByAccount returns transfers for an account.
func (s *TransferService) ListByAccount(ctx context.Context, accountID string, limit int) ([]*model.Transfer, error) {
	return s.store.ListByAccount(ctx, accountID, limit)
}

// BudgetService handles budget business logic.
type BudgetService struct {
	store         store.BudgetStore
	categoryStore store.CategoryStore
	now           func() time.Time
}

// NewBudgetService creates a new BudgetService.
func NewBudgetService(store store.BudgetStore, categoryStore store.CategoryStore) *BudgetService {
	return &BudgetService{store: store, categoryStore: categoryStore, now: time.Now}
}

// Create creates a new budget.
func (s *BudgetService) Create(ctx context.Context, b *model.Budget) (*model.Budget, error) {
	if b.CategoryID == "" {
		return nil, fmt.Errorf("%w: category_id is required", model.ErrInvalidInput)
	}
	if b.Amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be positive", model.ErrInvalidInput)
	}
	if b.Period == "" {
		b.Period = model.BudgetPeriodMonthly
	}
	if _, err := s.categoryStore.GetByID(ctx, b.CategoryID); err != nil {
		return nil, err
	}
	if b.ID == "" {
		b.ID = newID()
	}
	b.Amount = math.Round(b.Amount*100) / 100
	b.CreatedAt = s.now().UTC()
	b.UpdatedAt = b.CreatedAt
	if err := s.store.Create(ctx, b); err != nil {
		return nil, err
	}
	return b.Clone(), nil
}

// GetByID returns a budget by ID.
func (s *BudgetService) GetByID(ctx context.Context, id string) (*model.Budget, error) {
	return s.store.GetByID(ctx, id)
}

// Update updates an existing budget.
func (s *BudgetService) Update(ctx context.Context, id string, req *UpdateBudgetRequest) (*model.Budget, error) {
	b, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	req.Apply(b)
	if b.Amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be positive", model.ErrInvalidInput)
	}
	b.Amount = math.Round(b.Amount*100) / 100
	b.UpdatedAt = s.now().UTC()
	if err := s.store.Update(ctx, b); err != nil {
		return nil, err
	}
	return b.Clone(), nil
}

// Delete deletes a budget.
func (s *BudgetService) Delete(ctx context.Context, id string) error {
	return s.store.Delete(ctx, id)
}

// List returns all budgets.
func (s *BudgetService) List(ctx context.Context) ([]*model.Budget, error) {
	return s.store.List(ctx)
}

// GetSpending calculates actual spending for a budget in a given period.
func (s *BudgetService) GetSpending(ctx context.Context, categoryID string, period model.BudgetPeriod, startDate time.Time) (float64, error) {
	return 0, nil // simplified; real implementation would sum transactions
}

// UpdateBudgetRequest describes fields that can be updated on a budget.
type UpdateBudgetRequest struct {
	Amount *float64            `json:"amount"`
	Period *model.BudgetPeriod `json:"period"`
}

func (r *UpdateBudgetRequest) Apply(b *model.Budget) {
	if r.Amount != nil {
		b.Amount = math.Round(*r.Amount*100) / 100
	}
	if r.Period != nil {
		b.Period = *r.Period
	}
}

// ReportService generates financial reports.
type ReportService struct {
	transactionStore store.TransactionStore
	categoryStore    store.CategoryStore
	accountStore     store.AccountStore
}

// NewReportService creates a new ReportService.
func NewReportService(txStore store.TransactionStore, catStore store.CategoryStore, accStore store.AccountStore) *ReportService {
	return &ReportService{
		transactionStore: txStore,
		categoryStore:    catStore,
		accountStore:     accStore,
	}
}

// MonthlyReport generates a monthly financial report.
func (s *ReportService) MonthlyReport(ctx context.Context, year, month int) (*model.MonthlyReport, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)
	f := model.TransactionFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}
	txs, err := s.transactionStore.List(ctx, f)
	if err != nil {
		return nil, err
	}
	catMap := make(map[string]string)
	cats, _ := s.categoryStore.List(ctx, model.CategoryFilter{})
	for _, c := range cats {
		catMap[c.ID] = c.Name
	}
	report := &model.MonthlyReport{
		Year:           year,
		Month:          month,
		ByCategory:     make(map[string]float64),
		TopExpenses:    []model.CategorySpending{},
		AccountChanges: make(map[string]float64),
	}
	for _, t := range txs {
		if t.Type == model.TransactionTypeCredit {
			report.TotalIncome += t.Amount
		} else {
			report.TotalExpense += t.Amount
			report.ByCategory[t.CategoryID] += t.Amount
		}
		report.AccountChanges[t.AccountID] += t.NetAmount()
	}
	report.NetChange = report.TotalIncome - report.TotalExpense
	report.TotalIncome = math.Round(report.TotalIncome*100) / 100
	report.TotalExpense = math.Round(report.TotalExpense*100) / 100
	report.NetChange = math.Round(report.NetChange*100) / 100
	// Top expenses
	var expenses []model.CategorySpending
	for catID, amt := range report.ByCategory {
		expenses = append(expenses, model.CategorySpending{
			CategoryID:   catID,
			CategoryName: catMap[catID],
			Amount:       math.Round(amt*100) / 100,
		})
	}
	sortByAmountDesc(expenses)
	if len(expenses) > 10 {
		expenses = expenses[:10]
	}
	report.TopExpenses = expenses
	return report, nil
}

func sortByAmountDesc(s []model.CategorySpending) {
	for i := 0; i < len(s)-1; i++ {
		for j := i + 1; j < len(s); j++ {
			if s[j].Amount > s[i].Amount {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

func newID() string {
	buf := make([]byte, 16)
	rand.Read(buf) // #nosec G404
	return fmt.Sprintf("%x", buf)
}
