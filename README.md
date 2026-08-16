# Ledgerly

A personal finance management API service written in pure Go standard library.

## Features

- **Accounts** — checking, savings, credit, and investment accounts with balance tracking
- **Categories** — income and expense categories with hierarchical support
- **Transactions** — credit/debit transactions linked to accounts and categories
- **Transfers** — inter-account fund transfers with balance adjustment
- **Budgets** — monthly and yearly spending limits per category
- **Reports** — monthly financial summary with top expenses and account changes

## Architecture

```
cmd/ledgerly/      entry point
internal/httpapi/  HTTP handlers and routing
internal/service/  business logic layer
internal/store/    in-memory storage with JSON persistence
internal/model/    domain entities and filters
internal/auth/     bearer token authentication
internal/middleware/ request middleware (logging, recovery, rate limiting, timeout)
internal/config/   environment-based configuration
internal/logger/   structured key-value logger
internal/validator/ request validation
```

## API Endpoints

### Accounts
- `GET  /api/v1/accounts`          — list accounts
- `POST /api/v1/accounts`          — create account
- `GET  /api/v1/accounts/{id}`     — get account
- `PUT  /api/v1/accounts/{id}`     — update account
- `DELETE /api/v1/accounts/{id}`   — delete account
- `GET  /api/v1/accounts/{id}/transfers` — list transfers for account

### Categories
- `GET  /api/v1/categories`        — list categories
- `POST /api/v1/categories`        — create category
- `GET  /api/v1/categories/{id}`   — get category
- `PUT  /api/v1/categories/{id}`   — update category
- `DELETE /api/v1/categories/{id}` — delete category

### Transactions
- `GET  /api/v1/transactions`      — list transactions
- `POST /api/v1/transactions`      — create transaction
- `GET  /api/v1/transactions/{id}` — get transaction
- `DELETE /api/v1/transactions/{id}` — delete transaction

### Transfers
- `POST /api/v1/transfers`         — transfer between accounts

### Budgets
- `GET  /api/v1/budgets`           — list budgets
- `POST /api/v1/budgets`           — create budget
- `GET  /api/v1/budgets/{id}`      — get budget
- `PUT  /api/v1/budgets/{id}`      — update budget
- `DELETE /api/v1/budgets/{id}`    — delete budget

### Reports
- `GET /api/v1/reports/monthly?year=2026&month=1` — monthly report

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `LEDGERLY_ADDR` | `:8080` | HTTP server address |
| `LEDGERLY_DATA_DIR` | _(none)_ | Directory for JSON persistence |
| `LEDGERLY_MAX_BODY` | `4096` | Max request body size |
| `LEDGERLY_TIMEOUT` | `10s` | Request timeout |
| `LEDGERLY_RATE_LIMIT` | `100` | Requests per window |
| `LEDGERLY_RATE_WINDOW` | `1m` | Rate limit window |
| `LEDGERLY_AUTH_TOKEN` | _(none)_ | User ID for bootstrap auth token |

## Quick Start

```bash
# Run with in-memory storage
go run ./cmd/ledgerly

# Run with persistent storage
LEDGERLY_DATA_DIR=./data LEDGERLY_AUTH_TOKEN=admin go run ./cmd/ledgerly

# Run tests
go test ./...
```

## Design Notes

- Pure Go standard library — no external dependencies
- Atomic JSON persistence via temp-file + rename
- Deep copy at store boundary for thread-safe reads
- Bearer token authentication via constant-time comparison
- Sliding-window rate limiter per client IP
