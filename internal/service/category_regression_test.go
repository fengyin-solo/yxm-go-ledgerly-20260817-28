package service

import (
	"context"
	"testing"
	"time"

	"github.com/example/ledgerly/internal/logger"
	"github.com/example/ledgerly/internal/model"
	"github.com/example/ledgerly/internal/store"
)

func TestCategoryHierarchyFilteringAndIsolation(t *testing.T) {
	ctx := context.Background()
	catStore, _ := store.NewMemoryCategoryStore("", logger.Default(), time.Hour)
	svc := NewCategoryService(catStore)
	parent, _ := svc.Create(ctx, &model.Category{Name: "Food", Type: model.CategoryTypeExpense})
	child, _ := svc.Create(ctx, &model.Category{Name: "Dining", Type: model.CategoryTypeExpense, ParentID: &parent.ID})
	onlyChildren := true
	children, _ := svc.List(ctx, model.CategoryFilter{ParentID: &onlyChildren})
	if len(children) != 1 || children[0].ID != child.ID {
		t.Fatalf("child filter returned wrong categories: %#v", children)
	}
	got, _ := svc.GetByID(ctx, child.ID)
	*got.ParentID = "mutated"
	again, _ := svc.GetByID(ctx, child.ID)
	if again.ParentID == nil || *again.ParentID != parent.ID {
		t.Fatalf("category parent id was not isolated from caller mutation")
	}
}
