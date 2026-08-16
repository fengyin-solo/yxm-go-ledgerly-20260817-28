package store

import (
	"context"
	"testing"
	"time"

	"github.com/example/ledgerly/internal/logger"
	"github.com/example/ledgerly/internal/model"
)

func newTestAccountStore(t *testing.T) *MemoryAccountStore {
	t.Helper()
	f := t.TempDir() + "/accounts.json"
	s, err := NewMemoryAccountStore(f, logger.Default(), time.Hour)
	if err != nil {
		t.Fatalf("new account store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func newTestCategoryStore(t *testing.T) *MemoryCategoryStore {
	t.Helper()
	f := t.TempDir() + "/categories.json"
	s, err := NewMemoryCategoryStore(f, logger.Default(), time.Hour)
	if err != nil {
		t.Fatalf("new category store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func newTestTransactionStore(t *testing.T) *MemoryTransactionStore {
	t.Helper()
	f := t.TempDir() + "/transactions.json"
	s, err := NewMemoryTransactionStore(f, logger.Default(), time.Hour)
	if err != nil {
		t.Fatalf("new transaction store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func newTestTransferStore(t *testing.T) *MemoryTransferStore {
	t.Helper()
	f := t.TempDir() + "/transfers.json"
	s, err := NewMemoryTransferStore(f, logger.Default(), time.Hour)
	if err != nil {
		t.Fatalf("new transfer store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func newTestBudgetStore(t *testing.T) *MemoryBudgetStore {
	t.Helper()
	f := t.TempDir() + "/budgets.json"
	s, err := NewMemoryBudgetStore(f, logger.Default(), time.Hour)
	if err != nil {
		t.Fatalf("new budget store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestAccountStoreCRUD(t *testing.T) {
	s := newTestAccountStore(t)
	ctx := context.Background()

	// Create
	a := &model.Account{ID: "acc1", Name: "Main Checking", Type: model.AccountTypeChecking, Balance: 1000.50, IsActive: true}
	if err := s.Create(ctx, a); err != nil {
		t.Fatalf("create account: %v", err)
	}

	// Get
	got, err := s.GetByID(ctx, "acc1")
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if got.Name != "Main Checking" {
		t.Fatalf("unexpected name %s", got.Name)
	}
	if got.Balance != 1000.50 {
		t.Fatalf("unexpected balance %f", got.Balance)
	}

	// Update
	got.Balance = 2000.75
	if err := s.Update(ctx, got); err != nil {
		t.Fatalf("update account: %v", err)
	}
	got2, _ := s.GetByID(ctx, "acc1")
	if got2.Balance != 2000.75 {
		t.Fatalf("expected 2000.75, got %f", got2.Balance)
	}

	// List
	list, _ := s.List(ctx, model.AccountFilter{})
	if len(list) != 1 {
		t.Fatalf("expected 1, got %d", len(list))
	}

	// Delete
	if err := s.Delete(ctx, "acc1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err = s.GetByID(ctx, "acc1")
	if err != model.ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestAccountStoreFilter(t *testing.T) {
	s := newTestAccountStore(t)
	ctx := context.Background()

	s.Create(ctx, &model.Account{ID: "acc1", Name: "Checking", Type: model.AccountTypeChecking, IsActive: true})
	s.Create(ctx, &model.Account{ID: "acc2", Name: "Savings", Type: model.AccountTypeSavings, IsActive: true})
	s.Create(ctx, &model.Account{ID: "acc3", Name: "Old Checking", Type: model.AccountTypeChecking, IsActive: false})

	list, _ := s.List(ctx, model.AccountFilter{Type: model.AccountTypeChecking})
	if len(list) != 2 {
		t.Fatalf("expected 2 checking accounts, got %d", len(list))
	}

	active := true
	list, _ = s.List(ctx, model.AccountFilter{IsActive: &active})
	if len(list) != 2 {
		t.Fatalf("expected 2 active accounts, got %d", len(list))
	}

	list, _ = s.List(ctx, model.AccountFilter{Query: "checking"})
	if len(list) != 2 {
		t.Fatalf("expected 2 matching query, got %d", len(list))
	}
}

func TestCategoryStoreCRUD(t *testing.T) {
	s := newTestCategoryStore(t)
	ctx := context.Background()

	c := &model.Category{ID: "cat1", Name: "Groceries", Type: model.CategoryTypeExpense}
	if err := s.Create(ctx, c); err != nil {
		t.Fatalf("create category: %v", err)
	}

	got, err := s.GetByID(ctx, "cat1")
	if err != nil {
		t.Fatalf("get category: %v", err)
	}
	if got.Name != "Groceries" {
		t.Fatalf("unexpected name")
	}

	got.Name = "Food"
	if err := s.Update(ctx, got); err != nil {
		t.Fatalf("update category: %v", err)
	}

	if err := s.Delete(ctx, "cat1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err = s.GetByID(ctx, "cat1")
	if err != model.ErrNotFound {
		t.Fatalf("expected not found")
	}
}

func TestCategoryStoreFilter(t *testing.T) {
	s := newTestCategoryStore(t)
	ctx := context.Background()

	s.Create(ctx, &model.Category{ID: "cat1", Name: "Salary", Type: model.CategoryTypeIncome})
	s.Create(ctx, &model.Category{ID: "cat2", Name: "Groceries", Type: model.CategoryTypeExpense})
	s.Create(ctx, &model.Category{ID: "cat3", Name: "Rent", Type: model.CategoryTypeExpense})

	list, _ := s.List(ctx, model.CategoryFilter{Type: model.CategoryTypeExpense})
	if len(list) != 2 {
		t.Fatalf("expected 2 expense categories, got %d", len(list))
	}

	list, _ = s.List(ctx, model.CategoryFilter{Query: "gro"})
	if len(list) != 1 {
		t.Fatalf("expected 1 matching query, got %d", len(list))
	}
}

func TestTransactionStoreCRUD(t *testing.T) {
	s := newTestTransactionStore(t)
	ctx := context.Background()

	now := time.Now().UTC()
	tx := &model.Transaction{ID: "tx1", AccountID: "acc1", CategoryID: "cat1", Amount: 50.00, Type: model.TransactionTypeDebit, Date: now}
	if err := s.Create(ctx, tx); err != nil {
		t.Fatalf("create transaction: %v", err)
	}

	got, err := s.GetByID(ctx, "tx1")
	if err != nil {
		t.Fatalf("get transaction: %v", err)
	}
	if got.Amount != 50.00 {
		t.Fatalf("unexpected amount %f", got.Amount)
	}

	if err := s.Delete(ctx, "tx1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err = s.GetByID(ctx, "tx1")
	if err != model.ErrNotFound {
		t.Fatalf("expected not found")
	}
}

func TestTransactionStoreListFilter(t *testing.T) {
	s := newTestTransactionStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	s.Create(ctx, &model.Transaction{ID: "tx1", AccountID: "acc1", CategoryID: "cat1", Amount: 10, Type: model.TransactionTypeDebit, Date: now})
	s.Create(ctx, &model.Transaction{ID: "tx2", AccountID: "acc2", CategoryID: "cat1", Amount: 20, Type: model.TransactionTypeDebit, Date: now})
	s.Create(ctx, &model.Transaction{ID: "tx3", AccountID: "acc1", CategoryID: "cat2", Amount: 30, Type: model.TransactionTypeCredit, Date: now})

	list, _ := s.List(ctx, model.TransactionFilter{AccountID: "acc1"})
	if len(list) != 2 {
		t.Fatalf("expected 2 for acc1, got %d", len(list))
	}

	list, _ = s.List(ctx, model.TransactionFilter{CategoryID: "cat1"})
	if len(list) != 2 {
		t.Fatalf("expected 2 for cat1, got %d", len(list))
	}

	list, _ = s.List(ctx, model.TransactionFilter{Type: model.TransactionTypeCredit})
	if len(list) != 1 {
		t.Fatalf("expected 1 credit, got %d", len(list))
	}
}

func TestTransferStoreCRUD(t *testing.T) {
	s := newTestTransferStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	tr := &model.Transfer{ID: "tr1", FromAccountID: "acc1", ToAccountID: "acc2", Amount: 100.00, Date: now}
	if err := s.Create(ctx, tr); err != nil {
		t.Fatalf("create transfer: %v", err)
	}

	list, _ := s.ListByAccount(ctx, "acc1", 50)
	if len(list) != 1 {
		t.Fatalf("expected 1 transfer for acc1, got %d", len(list))
	}
	list, _ = s.ListByAccount(ctx, "acc2", 50)
	if len(list) != 1 {
		t.Fatalf("expected 1 transfer for acc2, got %d", len(list))
	}
}

func TestTransferStoreListByAccountLimit(t *testing.T) {
	s := newTestTransferStore(t)
	ctx := context.Background()
	now := time.Now().UTC()

	for i := 0; i < 5; i++ {
		s.Create(ctx, &model.Transfer{ID: "tr" + string(rune('a'+i)), FromAccountID: "acc1", ToAccountID: "acc2", Amount: 10, Date: now})
	}

	list, _ := s.ListByAccount(ctx, "acc1", 3)
	if len(list) != 3 {
		t.Fatalf("expected 3 with limit, got %d", len(list))
	}
}

func TestBudgetStoreCRUD(t *testing.T) {
	s := newTestBudgetStore(t)
	ctx := context.Background()

	b := &model.Budget{ID: "bud1", CategoryID: "cat1", Amount: 500.00, Period: model.BudgetPeriodMonthly}
	if err := s.Create(ctx, b); err != nil {
		t.Fatalf("create budget: %v", err)
	}

	got, err := s.GetByID(ctx, "bud1")
	if err != nil {
		t.Fatalf("get budget: %v", err)
	}
	if got.Amount != 500.00 {
		t.Fatalf("unexpected amount %f", got.Amount)
	}

	got.Amount = 600.00
	if err := s.Update(ctx, got); err != nil {
		t.Fatalf("update: %v", err)
	}

	list, _ := s.List(ctx)
	if len(list) != 1 {
		t.Fatalf("expected 1, got %d", len(list))
	}

	if err := s.Delete(ctx, "bud1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err = s.GetByID(ctx, "bud1")
	if err != model.ErrNotFound {
		t.Fatalf("expected not found")
	}
}

func TestPersistenceAndReload(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()

	// Create and save
	accPath := dir + "/accounts.json"
	accStore, err := NewMemoryAccountStore(accPath, logger.Default(), time.Hour)
	if err != nil {
		t.Fatalf("create account store: %v", err)
	}
	accStore.Create(ctx, &model.Account{ID: "acc1", Name: "Reload Test", Type: model.AccountTypeChecking, IsActive: true})
	accStore.Close()

	// Reload
	accStore2, err := NewMemoryAccountStore(accPath, logger.Default(), time.Hour)
	if err != nil {
		t.Fatalf("reload account store: %v", err)
	}
	defer accStore2.Close()

	list, _ := accStore2.List(ctx, model.AccountFilter{})
	if len(list) != 1 || list[0].Name != "Reload Test" {
		t.Fatalf("expected account after reload")
	}
}

func TestMissingResource(t *testing.T) {
	s := newTestAccountStore(t)
	ctx := context.Background()
	_, err := s.GetByID(ctx, "doesnotexist")
	if err != model.ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestDuplicateCreate(t *testing.T) {
	s := newTestAccountStore(t)
	ctx := context.Background()
	s.Create(ctx, &model.Account{ID: "dup1", Name: "Test", Type: model.AccountTypeChecking, IsActive: true})
	err := s.Create(ctx, &model.Account{ID: "dup1", Name: "Test 2", Type: model.AccountTypeChecking, IsActive: true})
	if err != model.ErrAlreadyExists {
		t.Fatalf("expected already exists, got %v", err)
	}
}

func TestUpdateNonexistent(t *testing.T) {
	s := newTestAccountStore(t)
	ctx := context.Background()
	err := s.Update(ctx, &model.Account{ID: "nope", Name: "Test", Type: model.AccountTypeChecking, IsActive: true})
	if err != model.ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestDeleteNonexistent(t *testing.T) {
	s := newTestAccountStore(t)
	ctx := context.Background()
	err := s.Delete(ctx, "nope")
	if err != model.ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}
