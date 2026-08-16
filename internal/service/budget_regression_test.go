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

func TestBudgetRequiresExistingCategoryAndKeepsPeriodUpdatesIsolated(t *testing.T) {
	ctx := context.Background()
	catStore, _ := store.NewMemoryCategoryStore("", logger.Default(), time.Hour)
	budgetStore, _ := store.NewMemoryBudgetStore("", logger.Default(), time.Hour)
	svc := NewBudgetService(budgetStore, catStore)
	if _, err := svc.Create(ctx, &model.Budget{CategoryID: "missing", Amount: 100, Period: model.BudgetPeriodMonthly}); err == nil {
		t.Fatal("budget for missing category should fail")
	}
	_ = catStore.Create(ctx, &model.Category{ID: "groceries", Name: "Groceries", Type: model.CategoryTypeExpense})
	b, err := svc.Create(ctx, &model.Budget{CategoryID: "groceries", Amount: 1200.129, Period: model.BudgetPeriodYearly})
	if err != nil {
		t.Fatalf("create budget: %v", err)
	}
	monthly := model.BudgetPeriodMonthly
	if _, err := svc.Update(ctx, b.ID, &UpdateBudgetRequest{Period: &monthly}); err != nil {
		t.Fatalf("update budget: %v", err)
	}
	got, _ := svc.GetByID(ctx, b.ID)
	got.Amount = -99
	again, _ := svc.GetByID(ctx, b.ID)
	if again.Period != model.BudgetPeriodMonthly || again.Amount != 1200.13 {
		t.Fatalf("budget update or cloning broke: period=%s amount=%.2f", again.Period, again.Amount)
	}
}

func TestBudgetRequestAcceptsYearlyPeriod(t *testing.T) {
	errs := validator.CreateBudgetRequest{CategoryID: "groceries", Amount: 10, Period: "yearly"}.Validate()
	if len(errs) != 0 {
		t.Fatalf("yearly budget should validate: %v", errs)
	}
}
