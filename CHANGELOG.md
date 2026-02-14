# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased] — Round 3

### Features
- Executable migration runner with version tracking (`pkg/migrations/`)
- Report label constants for future i18n support

### Refactoring
- Integrate `useTransactionList` composable into desktop ListPage.vue (-103 lines)
- Extract 14 hardcoded report labels into `models` constants

### Tests
- In-memory SQLite test helper for service integration tests (`testutil_db_test.go`)
- CRUD integration tests for 8 services: Assets, CFOs, Locations, Obligations, TaxRecords, InvestorDeals, InvestorPayments, Budgets (49 tests)
- Integration tests for ReportService (GetCashFlow, GetPnL, GetBalance, GetPaymentCalendar) with real SQL (13 tests)

---

## [Round 2] — Previous Refactoring

### Features
- Add `Currency` field to TaxRecord model, removing hardcoded "RUB" in payment calendar
- Add `gtfield=StartTime` validation to `ReportRequest.EndTime` for server-side date range enforcement

### Refactoring
- Split `transaction_modify.go` (969 lines) into three focused files: `transaction_modify.go`, `transaction_planned.go`, `transaction_move.go`
- Replace magic numbers in SQL queries with `fmt.Sprintf` + model constants (`buildCashFlowQuery()`, `buildPnlQuery()`)
- Convert 8 API handler files to use service interfaces for dependency injection (reports, assets, cfos, budgets, investor_deals, locations, obligations, tax_records)
- Extract `useTransactionList` composable from `ListPage.vue`, reducing component by ~416 lines
- Add `LocationProvider` interface with compile-time check to `interfaces.go`

### Tests
- Add edge-case unit tests for report helpers (`calculateResidualValue`, `validateTimeRange`, `matchesCfo`, `monthsBetween`)
- Add constructor and singleton tests for API handlers (reports, assets)
- Add 43 unit tests for dengioperacii transaction row parser (income/expense/transfer parsing, edge cases)
- Add 13 unit tests for `useConfirmAction` composable (confirm/cancel flows, error handling, processed errors)
- Add 30 unit tests for `usePeriodNavigation` composable (all navigation modes, boundary wrapping, guard clauses)
- Add validation tests for `ReportRequest` binding tags (`gtfield`, `required`, `min`)

### Documentation
- Add SQL migration definitions for all 10 new tables (`pkg/migrations/round2_new_tables.sql`)

---

## [Round 1] — Previous Refactoring

### Refactoring
- Replace magic numbers with named `ActivityType` and `CostType` constants in reports service
- Fix HTTP status codes: `NotFound` errors now return 404, `AlreadyExists`/`InUse` errors return 409
- Log and surface warnings instead of silently swallowing errors in report methods
- Extract `matchesCfo` helper to eliminate repeated CFO filter pattern
- Extract raw SQL queries into named constants (`cashFlowBaseQuery`, `pnlBaseQuery`, `cfoFilterClause`)

### Documentation
- Add comprehensive godoc comments with business formulas to all report methods

### Previous Release (984edd5)
- Add Cash Flow (ОДДС), P&L (ОПиУ), Balance Sheet, and Payment Calendar reports
- Add CFO (Center of Financial Responsibility) entity and filtering
- Add fixed assets with straight-line depreciation
- Add obligations (receivables/payables) tracking
- Add tax records management
- Add investor deals and payment tracking
- Add budget management with plan/fact comparison
- Add counterparty management
- Add planned (scheduled) transactions that do not affect balances
- Add transaction split functionality
- Split large service and handler files into focused modules
- Extract magic numbers, introduce struct-key maps, switch statements
- Add unit tests for frontend utilities and transaction helpers
- Add godoc comments to transaction service types
