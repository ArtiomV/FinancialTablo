# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

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
