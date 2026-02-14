# Architecture

## Overview

ezbookkeeping is a full-stack personal/business bookkeeping application with a Go backend and Vue 3 frontend. It supports dual-app rendering (desktop and mobile), 20+ import/export formats, multi-currency, planned transactions, budgets, asset management, investor deals, and financial reports (Cash Flow, P&L, Balance Sheet, Payment Calendar).

## Technology Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.25+, Gin (HTTP), xorm (ORM) |
| Database | SQLite (default), MySQL, PostgreSQL |
| Frontend | Vue 3, Vuetify 3, TypeScript, Pinia |
| Build | Vite (frontend), `go build` (backend) |
| Tests | Go `testing` + testify (backend), Jest + ts-jest (frontend) |

## Directory Structure

```
ezbookkeeping-main/
├── ezbookkeeping.go          # Application entry point
├── cmd/                      # CLI commands (server, cron, etc.)
├── conf/                     # Configuration templates
├── pkg/                      # Go backend packages
│   ├── api/                  # HTTP handlers (Gin controllers)
│   ├── auth/                 # Authentication & authorization
│   ├── converters/           # 20+ import/export format converters
│   ├── core/                 # Core types (Context, Handler, etc.)
│   ├── cron/                 # Scheduled task runner
│   ├── datastore/            # Database connection management
│   ├── errs/                 # Error code definitions
│   ├── exchangerates/        # Exchange rate providers
│   ├── llm/                  # LLM integration (receipt recognition)
│   ├── middlewares/          # HTTP middleware (auth, CORS, etc.)
│   ├── models/               # Data models & request/response types
│   ├── services/             # Business logic layer
│   ├── settings/             # Application settings
│   ├── utils/                # Utility functions
│   └── validators/           # Custom request validators
├── src/                      # Vue 3 frontend
│   ├── components/           # Reusable UI components
│   │   ├── base/             # Shared composable logic (Base.ts)
│   │   ├── common/           # Cross-platform components
│   │   ├── desktop/          # Desktop-only components
│   │   └── mobile/           # Mobile-only components
│   ├── composables/          # Vue composables (extracted logic)
│   ├── consts/               # Constants (currencies, icons, etc.)
│   ├── core/                 # Core enums & type classes
│   ├── lib/                  # Utility libraries
│   │   └── transaction/      # Transaction-specific utilities
│   ├── locales/              # i18n translations
│   ├── models/               # Frontend data models
│   ├── router/               # Vue Router configuration
│   ├── stores/               # Pinia state management (23 stores)
│   └── views/                # Page views
│       ├── base/             # Shared page logic (Base.ts)
│       ├── desktop/          # Desktop pages
│       └── mobile/           # Mobile pages
├── dist/                     # Built frontend assets
├── templates/                # Go templates
└── testdata/                 # Test fixtures
```

## Backend Architecture

### Three-Layer Pattern

```
HTTP Request → pkg/api/ (handlers) → pkg/services/ (business logic) → pkg/models/ + xorm (ORM)
```

**pkg/api/** — HTTP handlers registered via Gin router. Each handler parses request, calls service, formats response. Files are grouped by domain (accounts.go, transactions.go, reports.go, etc.).

**pkg/services/** — Business logic. Singleton instances initialized with DB references. Key services:
- `Transactions` — CRUD, filtering, pagination, planned transactions, scheduling
- `Reports` — Cash Flow, P&L, Balance Sheet, Payment Calendar
- `Assets` — Asset management with linear depreciation calculation
- `InvestorDeals` / `InvestorPayments` — Investor module
- `Budgets` — Budget tracking with plan-fact analysis
- `Obligations` — Receivables and payables management

**pkg/models/** — Data structs mapped to DB tables via xorm tags. Also contains request/response DTOs and model-level validation.

### Key Patterns

- **Query builder**: `buildTransactionQueryCondition()` in `transaction_helpers.go` uses xorm's `builder.Cond` interface to compose SQL conditions via switch statements.
- **Planned transactions**: `Transaction.Planned = true` marks future transactions that don't affect account balances. `ConfirmPlannedTransaction` sets `Planned = false` and applies balance changes.
- **Scheduled transactions**: `CreateScheduledTransactions()` runs via cron, checking templates against time intervals.
- **Depreciation**: `calculateResidualValue()` implements straight-line depreciation: `monthly = (cost - salvage) / useful_life_months`.

### Database Tables (key additions)

| Table | Purpose |
|---|---|
| `transaction` | All transactions (income, expense, transfer) |
| `account` | User accounts with balances |
| `transaction_category` | Categories with `activity_type` (1=operating, 2=investing, 3=financing) and `cost_type` (1=COGS, 2=operational, 3=financial) |
| `cfo` | Centers of Financial Responsibility |
| `location` | Physical locations |
| `asset` | Fixed assets with depreciation tracking |
| `investor_deal` | Investor deals with repayment terms |
| `investor_payment` | Individual investor payments |
| `budget` | Budget entries (plan amounts per category/period) |
| `obligation` | Receivables (type=1) and payables (type=2) |
| `tax_record` | Tax obligations with due dates |

### Reports

| Report | Endpoint | Description |
|---|---|---|
| Cash Flow (ОДДС) | `GET /api/v1/reports/cashflow.json` | Groups by activity_type (operating/investing/financing) |
| P&L (ОПиУ) | `GET /api/v1/reports/pnl.json` | Revenue, COGS, operating/financial expenses, depreciation, taxes |
| Balance Sheet | `GET /api/v1/reports/balance.json` | Assets vs Liabilities, Equity = Assets - Liabilities |
| Payment Calendar | `GET /api/v1/reports/payment_calendar.json` | Obligations + taxes + planned transactions sorted by date |

## Frontend Architecture

### Dual-App Rendering

Two separate entry points: `desktop-main.ts` and `mobile-main.ts`. Each renders its own root component (`DesktopApp.vue` / `MobileApp.vue`) with platform-specific UI but shared business logic.

### Base Pattern

Shared logic lives in `*Base.ts` files:
- `src/views/base/` — Page-level composables (e.g., `TransactionListPageBase.ts`)
- `src/components/base/` — Component-level composables (e.g., `PieChartBase.ts`)

Desktop and mobile views import these bases and add platform-specific templates.

### State Management

23 Pinia stores in `src/stores/`:
- **transaction.ts** (1542 lines) — Transaction CRUD, filters, pagination, import
- **statistics.ts** (1876 lines) — Statistical analysis, charts, trends
- **account.ts** (1100 lines) — Account management
- **explorer.ts** (1180 lines) — Custom data explorer

### Composables (`src/composables/`)

Extracted reusable logic:
- `useInfiniteScroll` — IntersectionObserver-based infinite scroll
- `useConfirmAction` — Confirm dialog → action → success/error pattern
- `usePeriodNavigation` — Period navigation (week/month/quarter/year)

### Extracted Components

Transaction list page decomposed into:
- `TransactionTableRow.vue` — Single transaction row with actions
- `TransactionPeriodToolbar.vue` — Period selector + type filter buttons
- `TransactionFilterPanel.vue` — Advanced filter dropdown
- `TransactionTotalsBar.vue` — Income/expense/balance totals display

### Utility Libraries (`src/lib/transaction/`)

- `filterParams.ts` — URL query string builder from filter state
- `amountCalc.ts` — Currency-aware amount calculations

## Testing

### Backend (Go)

```bash
go test ./pkg/services/ -v     # Service layer tests
go test ./...                   # All tests
```

Key test files:
- `transaction_helpers_test.go` — 46 tests for query builder, validation, column mapping
- `reports_test.go` — 34 tests for depreciation, PnL formulas, balance equity
- `accounts_test.go`, `transaction_tags_test.go`, `transaction_categories_test.go`

### Frontend (Jest)

```bash
npm test                        # All frontend tests
```

Test suites in `__tests__/` directories:
- `lib/transaction/__tests__/filterParams.test.ts` — Filter URL builder tests
- `lib/transaction/__tests__/amountCalc.test.ts` — Amount calculation tests
- `lib/__tests__/common.ts` — Common utility tests
- `lib/__tests__/fiscal_year.ts` — Fiscal year calculation tests
- `lib/calendar/__tests__/chinese_calendar.ts` — Chinese calendar tests (38K+ tests)

## Build & Deploy

### Development

```bash
# Frontend
npm install
npm run dev                     # Vite dev server

# Backend
go run ezbookkeeping.go server  # Start Go server
```

### Production

```bash
# Build frontend
NODE_OPTIONS="--max-old-space-size=3072" npm run build

# Build backend
CGO_ENABLED=1 go build -o ezbookkeeping .

# Docker (using pre-built frontend)
docker build -f Dockerfile.fast -t ezbookkeeping .
docker run -d --name ezbookkeeping -p 80:8080 \
  -v /data:/ezbookkeeping/data \
  -e 'EBK_SERVER_STATIC_ROOT_PATH=public' \
  ezbookkeeping
```

### Key Build Notes

- `CGO_ENABLED=1` is required (go-sqlite3 uses CGO)
- Go 1.25+ required (per go.mod)
- Frontend build needs `--max-old-space-size=3072` for large TypeScript compilation
- Docker needs `.git` directory (git-rev-sync npm package)
