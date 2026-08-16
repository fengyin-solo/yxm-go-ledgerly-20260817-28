package service

import (
	"context"
	"testing"
	"time"

	"github.com/example/ledgerly/internal/logger"
	"github.com/example/ledgerly/internal/model"
	"github.com/example/ledgerly/internal/store"
)

func TestTransactionFilterHonorsInclusiveDatesTypeAndAccount(t *testing.T) {
	ctx := context.Background()
	txStore, _ := store.NewMemoryTransactionStore("", logger.Default(), time.Hour)
	accStore, _ := store.NewMemoryAccountStore("", logger.Default(), time.Hour)
	_ = accStore.Create(ctx, &model.Account{ID: "acc-a", Name: "A", Balance: 100, IsActive: true})
	_ = accStore.Create(ctx, &model.Account{ID: "acc-b", Name: "B", Balance: 100, IsActive: true})
	svc := NewTransactionService(txStore, accStore)
	_, _ = svc.Create(ctx, &model.Transaction{AccountID: "acc-a", CategoryID: "food", Amount: 10, Type: model.TransactionTypeDebit, Date: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)})
	_, _ = svc.Create(ctx, &model.Transaction{AccountID: "acc-a", CategoryID: "food", Amount: 5, Type: model.TransactionTypeCredit, Date: time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)})
	_, _ = svc.Create(ctx, &model.Transaction{AccountID: "acc-b", CategoryID: "food", Amount: 7, Type: model.TransactionTypeDebit, Date: time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)})
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	got, err := svc.List(ctx, model.TransactionFilter{AccountID: "acc-a", Type: model.TransactionTypeDebit, StartDate: &start, EndDate: &end})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].AccountID != "acc-a" || got[0].Type != model.TransactionTypeDebit {
		t.Fatalf("filter returned wrong transactions: %#v", got)
	}
}
