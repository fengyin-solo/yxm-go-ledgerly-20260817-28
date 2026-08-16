package service

import (
	"context"
	"testing"
	"time"

	"github.com/example/ledgerly/internal/logger"
	"github.com/example/ledgerly/internal/model"
	"github.com/example/ledgerly/internal/store"
	"github.com/example/ledgerly/internal/validator"
)

func TestTransactionDeleteRestoresBalanceAndRemovesLedgerEntry(t *testing.T) {
	ctx := context.Background()
	accStore, _ := store.NewMemoryAccountStore("", logger.Default(), time.Hour)
	txStore, _ := store.NewMemoryTransactionStore("", logger.Default(), time.Hour)
	if err := accStore.Create(ctx, &model.Account{ID: "acc-checking", Name: "Checking", Type: model.AccountTypeChecking, Balance: 100, IsActive: true}); err != nil {
		t.Fatal(err)
	}
	svc := NewTransactionService(txStore, accStore)
	tx, err := svc.Create(ctx, &model.Transaction{AccountID: "acc-checking", CategoryID: "cat-food", Amount: 25, Type: model.TransactionTypeDebit, Date: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	if err := svc.Delete(ctx, tx.ID); err != nil {
		t.Fatalf("delete transaction: %v", err)
	}
	acc, _ := accStore.GetByID(ctx, "acc-checking")
	if acc.Balance != 100 {
		t.Fatalf("delete should restore original balance, got %.2f", acc.Balance)
	}
	list, err := svc.List(ctx, model.TransactionFilter{AccountID: "acc-checking"})
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("deleted transaction still appears in list: %d", len(list))
	}
}

func TestMonthlyReportTreatsDebitsAsExpenses(t *testing.T) {
	ctx := context.Background()
	accStore, _ := store.NewMemoryAccountStore("", logger.Default(), time.Hour)
	catStore, _ := store.NewMemoryCategoryStore("", logger.Default(), time.Hour)
	txStore, _ := store.NewMemoryTransactionStore("", logger.Default(), time.Hour)
	_ = accStore.Create(ctx, &model.Account{ID: "acc", Name: "Checking", Type: model.AccountTypeChecking, Balance: 200, IsActive: true})
	_ = catStore.Create(ctx, &model.Category{ID: "groceries", Name: "Groceries", Type: model.CategoryTypeExpense})
	txSvc := NewTransactionService(txStore, accStore)
	if _, err := txSvc.Create(ctx, &model.Transaction{AccountID: "acc", CategoryID: "groceries", Amount: 42.25, Type: model.TransactionTypeDebit, Date: time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)}); err != nil {
		t.Fatal(err)
	}
	report, err := NewReportService(txStore, catStore, accStore).MonthlyReport(ctx, 2026, 8)
	if err != nil {
		t.Fatal(err)
	}
	if report.TotalExpense != 42.25 || report.NetChange != -42.25 || report.AccountChanges["acc"] != -42.25 {
		t.Fatalf("bad debit rollup: expense=%.2f net=%.2f account=%.2f", report.TotalExpense, report.NetChange, report.AccountChanges["acc"])
	}
}

func TestTransactionRequestRejectsZeroAmount(t *testing.T) {
	errs := validator.CreateTransactionRequest{AccountID: "acc", CategoryID: "cat", Type: "debit", Amount: 0}.Validate()
	if len(errs) == 0 {
		t.Fatal("zero amount transaction should be rejected")
	}
}
