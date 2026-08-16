package store

import (
	"context"

	"github.com/example/ledgerly/internal/model"
)

// AccountStore manages accounts.
type AccountStore interface {
	Create(ctx context.Context, a *model.Account) error
	GetByID(ctx context.Context, id string) (*model.Account, error)
	Update(ctx context.Context, a *model.Account) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, f model.AccountFilter) ([]*model.Account, error)
}

// CategoryStore manages categories.
type CategoryStore interface {
	Create(ctx context.Context, c *model.Category) error
	GetByID(ctx context.Context, id string) (*model.Category, error)
	Update(ctx context.Context, c *model.Category) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, f model.CategoryFilter) ([]*model.Category, error)
}

// TransactionStore manages transactions.
type TransactionStore interface {
	Create(ctx context.Context, t *model.Transaction) error
	GetByID(ctx context.Context, id string) (*model.Transaction, error)
	List(ctx context.Context, f model.TransactionFilter) ([]*model.Transaction, error)
	Delete(ctx context.Context, id string) error
}

// TransferStore manages transfers.
type TransferStore interface {
	Create(ctx context.Context, t *model.Transfer) error
	ListByAccount(ctx context.Context, accountID string, limit int) ([]*model.Transfer, error)
}

// BudgetStore manages budgets.
type BudgetStore interface {
	Create(ctx context.Context, b *model.Budget) error
	GetByID(ctx context.Context, id string) (*model.Budget, error)
	Update(ctx context.Context, b *model.Budget) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*model.Budget, error)
}
