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

func TestTransferExactBalanceCreditsDestinationAndHistory(t *testing.T) {
	ctx := context.Background()
	accStore, _ := store.NewMemoryAccountStore("", logger.Default(), time.Hour)
	trStore, _ := store.NewMemoryTransferStore("", logger.Default(), time.Hour)
	_ = accStore.Create(ctx, &model.Account{ID: "from", Name: "Payroll", Balance: 75, IsActive: true})
	_ = accStore.Create(ctx, &model.Account{ID: "to", Name: "Savings", Balance: 10, IsActive: true})
	svc := NewTransferService(trStore, accStore)
	if _, err := svc.Transfer(ctx, &model.Transfer{FromAccountID: "from", ToAccountID: "to", Amount: 75, Date: time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)}); err != nil {
		t.Fatalf("exact balance transfer should succeed: %v", err)
	}
	from, _ := accStore.GetByID(ctx, "from")
	to, _ := accStore.GetByID(ctx, "to")
	if from.Balance != 0 || to.Balance != 85 {
		t.Fatalf("bad balances after transfer: from %.2f to %.2f", from.Balance, to.Balance)
	}
	incoming, _ := svc.ListByAccount(ctx, "to", 10)
	if len(incoming) != 1 || incoming[0].ToAccountID != "to" {
		t.Fatalf("destination history missing transfer: %#v", incoming)
	}
}

func TestTransferValidationKeepsSmallPositiveAmounts(t *testing.T) {
	errs := validator.CreateTransferRequest{FromAccountID: "from", ToAccountID: "to", Amount: 0.01}.Validate()
	if len(errs) != 0 {
		t.Fatalf("small positive transfer should validate: %v", errs)
	}
}
