package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// newTestReportServiceWithDB creates a ReportService backed by an in-memory SQLite database.
// Dependent providers (assets, taxes, deals, payments) use mock implementations from
// reports_integration_test.go. Options allow overriding individual providers.
func newTestReportServiceWithDB(t *testing.T, opts ...func(*ReportService)) (*ReportService, *testDB) {
	t.Helper()
	tdb := newTestDB(t)

	svc := &ReportService{
		ServiceUsingDB: ServiceUsingDB{container: tdb.container},
		assets:         &mockAssetProvider{},
		taxes:          &mockTaxRecordProvider{},
		deals:          &mockInvestorDealProvider{},
		payments:       &mockInvestorPaymentProvider{},
	}

	for _, opt := range opts {
		opt(svc)
	}

	return svc, tdb
}

// seedResult holds category IDs created by seedTransactionData
type seedResult struct {
	incomeCategoryId  int64
	cogsCategoryId    int64
	opexCategoryId    int64
	finexCategoryId   int64
	investCategoryId  int64
	financeCategoryId int64
}

// seedTransactionData inserts categories and transactions into the test DB.
// Categories use specific activity_type and cost_type values that the SQL queries group by.
// Transactions include income and expense types, plus deleted and planned rows that
// must be excluded by the queries.
func seedTransactionData(t *testing.T, tdb *testDB, uid int64, baseTime int64) *seedResult {
	t.Helper()

	// Insert categories into transaction_category table.
	// The SQL queries JOIN on t.category_id = tc.category_id AND tc.uid = t.uid.
	categories := []*models.TransactionCategory{
		{
			CategoryId:   100,
			Uid:          uid,
			Deleted:      false,
			Type:         models.CATEGORY_TYPE_INCOME,
			Name:         "Sales Revenue",
			ActivityType: int32(models.ACTIVITY_TYPE_OPERATING),
			CostType:     0,
		},
		{
			CategoryId:   201,
			Uid:          uid,
			Deleted:      false,
			Type:         models.CATEGORY_TYPE_EXPENSE,
			Name:         "Cost of Materials",
			ActivityType: int32(models.ACTIVITY_TYPE_OPERATING),
			CostType:     int32(models.COST_TYPE_COGS),
		},
		{
			CategoryId:   202,
			Uid:          uid,
			Deleted:      false,
			Type:         models.CATEGORY_TYPE_EXPENSE,
			Name:         "Rent & Utilities",
			ActivityType: int32(models.ACTIVITY_TYPE_OPERATING),
			CostType:     int32(models.COST_TYPE_OPERATIONAL),
		},
		{
			CategoryId:   203,
			Uid:          uid,
			Deleted:      false,
			Type:         models.CATEGORY_TYPE_EXPENSE,
			Name:         "Bank Interest",
			ActivityType: int32(models.ACTIVITY_TYPE_OPERATING),
			CostType:     int32(models.COST_TYPE_FINANCIAL),
		},
		{
			CategoryId:   300,
			Uid:          uid,
			Deleted:      false,
			Type:         models.CATEGORY_TYPE_INCOME,
			Name:         "Asset Sale",
			ActivityType: int32(models.ACTIVITY_TYPE_INVESTING),
			CostType:     0,
		},
		{
			CategoryId:   400,
			Uid:          uid,
			Deleted:      false,
			Type:         models.CATEGORY_TYPE_INCOME,
			Name:         "Loan Received",
			ActivityType: int32(models.ACTIVITY_TYPE_FINANCING),
			CostType:     0,
		},
	}

	for _, cat := range categories {
		_, err := tdb.engine.Insert(cat)
		if err != nil {
			t.Fatalf("failed to insert category %q: %v", cat.Name, err)
		}
	}

	// Insert transactions. The SQL queries filter by:
	//   t.uid = ? AND t.deleted = 0 AND t.planned = 0
	//   AND t.transaction_time >= ? AND t.transaction_time < ?
	//   AND t.type IN (2, 3)  [INCOME, EXPENSE]
	transactions := []*models.Transaction{
		// Income transactions (type=2)
		{
			TransactionId:   1001,
			Uid:             uid,
			Deleted:         false,
			Type:            models.TRANSACTION_DB_TYPE_INCOME,
			CategoryId:      100, // Sales Revenue
			AccountId:       1,
			TransactionTime: baseTime + 100,
			Amount:          500000, // 5000.00 in cents
			Planned:         false,
		},
		{
			TransactionId:   1002,
			Uid:             uid,
			Deleted:         false,
			Type:            models.TRANSACTION_DB_TYPE_INCOME,
			CategoryId:      100, // Sales Revenue
			AccountId:       1,
			TransactionTime: baseTime + 200,
			Amount:          300000, // 3000.00
			Planned:         false,
		},
		// Expense transactions (type=3)
		{
			TransactionId:   2001,
			Uid:             uid,
			Deleted:         false,
			Type:            models.TRANSACTION_DB_TYPE_EXPENSE,
			CategoryId:      201, // COGS
			AccountId:       1,
			TransactionTime: baseTime + 150,
			Amount:          200000, // 2000.00
			Planned:         false,
		},
		{
			TransactionId:   2002,
			Uid:             uid,
			Deleted:         false,
			Type:            models.TRANSACTION_DB_TYPE_EXPENSE,
			CategoryId:      202, // Operating Expense
			AccountId:       1,
			TransactionTime: baseTime + 250,
			Amount:          100000, // 1000.00
			Planned:         false,
		},
		{
			TransactionId:   2003,
			Uid:             uid,
			Deleted:         false,
			Type:            models.TRANSACTION_DB_TYPE_EXPENSE,
			CategoryId:      203, // Financial Expense
			AccountId:       1,
			TransactionTime: baseTime + 300,
			Amount:          50000, // 500.00
			Planned:         false,
		},
		// Investing income
		{
			TransactionId:   3001,
			Uid:             uid,
			Deleted:         false,
			Type:            models.TRANSACTION_DB_TYPE_INCOME,
			CategoryId:      300, // Asset Sale (investing)
			AccountId:       1,
			TransactionTime: baseTime + 350,
			Amount:          150000, // 1500.00
			Planned:         false,
		},
		// Financing income
		{
			TransactionId:   4001,
			Uid:             uid,
			Deleted:         false,
			Type:            models.TRANSACTION_DB_TYPE_INCOME,
			CategoryId:      400, // Loan Received (financing)
			AccountId:       1,
			TransactionTime: baseTime + 400,
			Amount:          1000000, // 10000.00
			Planned:         false,
		},
		// EXCLUDED: deleted=true
		{
			TransactionId:   9001,
			Uid:             uid,
			Deleted:         true,
			Type:            models.TRANSACTION_DB_TYPE_INCOME,
			CategoryId:      100,
			AccountId:       1,
			TransactionTime: baseTime + 500,
			Amount:          999999,
			Planned:         false,
		},
		// EXCLUDED: planned=true
		{
			TransactionId:   9002,
			Uid:             uid,
			Deleted:         false,
			Type:            models.TRANSACTION_DB_TYPE_INCOME,
			CategoryId:      100,
			AccountId:       1,
			TransactionTime: baseTime + 600,
			Amount:          888888,
			Planned:         true,
		},
	}

	for _, txn := range transactions {
		_, err := tdb.engine.Insert(txn)
		if err != nil {
			t.Fatalf("failed to insert transaction %d: %v", txn.TransactionId, err)
		}
	}

	return &seedResult{
		incomeCategoryId:  100,
		cogsCategoryId:    201,
		opexCategoryId:    202,
		finexCategoryId:   203,
		investCategoryId:  300,
		financeCategoryId: 400,
	}
}

// TestReportService_GetCashFlow_WithDB verifies cash flow calculation against real SQL queries.
func TestReportService_GetCashFlow_WithDB(t *testing.T) {
	svc, tdb := newTestReportServiceWithDB(t)
	defer tdb.close()

	uid := int64(1)
	baseTime := int64(1700000000)
	seedTransactionData(t, tdb, uid, baseTime)

	startTime := baseTime
	endTime := baseTime + 10000

	result, err := svc.GetCashFlow(nil, uid, 0, startTime, endTime)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	// We expect 3 activities: Operating, Investing, Financing
	assert.Equal(t, 3, len(result.Activities))

	// Operating activity (activity_type=1):
	//   Income: 500000 + 300000 = 800000 (Sales Revenue)
	//   Expense: 200000 (COGS) + 100000 (OpEx) + 50000 (FinEx) = 350000
	//   Net: 800000 - 350000 = 450000
	operating := result.Activities[0]
	assert.Equal(t, int32(models.ACTIVITY_TYPE_OPERATING), operating.ActivityType)
	assert.Equal(t, int64(800000), operating.TotalIncome)
	assert.Equal(t, int64(350000), operating.TotalExpense)
	assert.Equal(t, int64(450000), operating.TotalNet)

	// Investing activity (activity_type=2):
	//   Income: 150000 (Asset Sale)
	//   Expense: 0
	//   Net: 150000
	investing := result.Activities[1]
	assert.Equal(t, int32(models.ACTIVITY_TYPE_INVESTING), investing.ActivityType)
	assert.Equal(t, int64(150000), investing.TotalIncome)
	assert.Equal(t, int64(0), investing.TotalExpense)
	assert.Equal(t, int64(150000), investing.TotalNet)

	// Financing activity (activity_type=3):
	//   Income: 1000000 (Loan Received)
	//   Expense: 0
	//   Net: 1000000
	financing := result.Activities[2]
	assert.Equal(t, int32(models.ACTIVITY_TYPE_FINANCING), financing.ActivityType)
	assert.Equal(t, int64(1000000), financing.TotalIncome)
	assert.Equal(t, int64(0), financing.TotalExpense)
	assert.Equal(t, int64(1000000), financing.TotalNet)

	// Total net: 450000 + 150000 + 1000000 = 1600000
	assert.Equal(t, int64(1600000), result.TotalNet)
}

// TestReportService_GetPnL_WithDB verifies P&L calculation against real SQL queries.
func TestReportService_GetPnL_WithDB(t *testing.T) {
	svc, tdb := newTestReportServiceWithDB(t)
	defer tdb.close()

	uid := int64(1)
	baseTime := int64(1700000000)
	seedTransactionData(t, tdb, uid, baseTime)

	startTime := baseTime
	endTime := baseTime + 10000

	result, err := svc.GetPnL(nil, uid, 0, startTime, endTime)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	// Revenue: all income transactions = 500000 + 300000 + 150000 + 1000000 = 1950000
	assert.Equal(t, int64(1950000), result.Revenue)

	// COGS: expense with cost_type=1 (COGS) = 200000
	assert.Equal(t, int64(200000), result.CostOfGoods)

	// GrossProfit = Revenue - COGS = 1950000 - 200000 = 1750000
	assert.Equal(t, int64(1750000), result.GrossProfit)

	// OperatingExpense: expense with cost_type=2 (operational) = 100000
	assert.Equal(t, int64(100000), result.OperatingExpense)

	// Depreciation: 0 (no assets in mock)
	assert.Equal(t, int64(0), result.Depreciation)

	// OperatingProfit = GrossProfit - OperatingExpense - Depreciation = 1750000 - 100000 = 1650000
	assert.Equal(t, int64(1650000), result.OperatingProfit)

	// FinancialExpense: expense with cost_type=3 (financial) = 50000
	assert.Equal(t, int64(50000), result.FinancialExpense)

	// TaxExpense: 0 (no tax records in mock)
	assert.Equal(t, int64(0), result.TaxExpense)

	// NetProfit = OperatingProfit - FinancialExpense - TaxExpense = 1650000 - 50000 = 1600000
	assert.Equal(t, int64(1600000), result.NetProfit)
}

// TestReportService_GetBalance_WithDB verifies balance sheet calculation using real DB
// for accounts and obligations.
func TestReportService_GetBalance_WithDB(t *testing.T) {
	svc, tdb := newTestReportServiceWithDB(t)
	defer tdb.close()

	uid := int64(1)

	// Insert accounts: asset accounts and a liability account
	accounts := []*models.Account{
		{
			AccountId: 1,
			Uid:       uid,
			Deleted:   false,
			Category:  models.ACCOUNT_CATEGORY_CHECKING_ACCOUNT, // asset
			Type:      models.ACCOUNT_TYPE_SINGLE_ACCOUNT,
			Name:      "Main Checking",
			Balance:   500000,
			Currency:  "RUB",
			Color:     "000000",
		},
		{
			AccountId: 2,
			Uid:       uid,
			Deleted:   false,
			Category:  models.ACCOUNT_CATEGORY_CREDIT_CARD, // liability
			Type:      models.ACCOUNT_TYPE_SINGLE_ACCOUNT,
			Name:      "Credit Card",
			Balance:   150000,
			Currency:  "RUB",
			Color:     "000000",
		},
		{
			AccountId: 3,
			Uid:       uid,
			Deleted:   false,
			Category:  models.ACCOUNT_CATEGORY_SAVINGS_ACCOUNT, // asset
			Type:      models.ACCOUNT_TYPE_SINGLE_ACCOUNT,
			Name:      "Savings",
			Balance:   300000,
			Currency:  "RUB",
			Color:     "000000",
		},
	}

	for _, acc := range accounts {
		_, err := tdb.engine.Insert(acc)
		if err != nil {
			t.Fatalf("failed to insert account %q: %v", acc.Name, err)
		}
	}

	// Insert obligations
	obligations := []*models.Obligation{
		{
			ObligationId:   1,
			Uid:            uid,
			Deleted:        false,
			ObligationType: models.OBLIGATION_TYPE_RECEIVABLE,
			Amount:         200000,
			PaidAmount:     50000,
		},
		{
			ObligationId:   2,
			Uid:            uid,
			Deleted:        false,
			ObligationType: models.OBLIGATION_TYPE_PAYABLE,
			Amount:         100000,
			PaidAmount:     0,
		},
	}

	for _, o := range obligations {
		_, err := tdb.engine.Insert(o)
		if err != nil {
			t.Fatalf("failed to insert obligation: %v", err)
		}
	}

	result, err := svc.GetBalance(nil, uid, 0)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	// Assets: Cash & Bank = 500000 + 300000 = 800000, Receivables = 200000 - 50000 = 150000
	// Total Assets = 800000 + 150000 = 950000
	assert.Equal(t, int64(950000), result.TotalAssets)

	// Liabilities: Credit Card = 150000, Payables = 100000
	// Total Liabilities = 150000 + 100000 = 250000
	assert.Equal(t, int64(250000), result.TotalLiability)

	// Equity = Assets - Liabilities = 950000 - 250000 = 700000
	assert.Equal(t, int64(700000), result.Equity)

	// Verify individual asset lines
	assert.True(t, len(result.AssetLines) >= 2)
	foundCash := false
	foundReceivables := false
	for _, line := range result.AssetLines {
		if line.Label == "Cash & Bank Accounts" {
			assert.Equal(t, int64(800000), line.Amount)
			foundCash = true
		}
		if line.Label == "Receivables" {
			assert.Equal(t, int64(150000), line.Amount)
			foundReceivables = true
		}
	}
	assert.True(t, foundCash, "expected Cash & Bank Accounts line")
	assert.True(t, foundReceivables, "expected Receivables line")

	// Verify individual liability lines
	foundCreditCard := false
	foundPayables := false
	for _, line := range result.LiabilityLines {
		if line.Label == "Credit Cards & Debts" {
			assert.Equal(t, int64(150000), line.Amount)
			foundCreditCard = true
		}
		if line.Label == "Payables" {
			assert.Equal(t, int64(100000), line.Amount)
			foundPayables = true
		}
	}
	assert.True(t, foundCreditCard, "expected Credit Cards & Debts line")
	assert.True(t, foundPayables, "expected Payables line")
}

// TestReportService_GetPnL_EmptyData verifies that P&L returns zero results
// when no transactions exist.
func TestReportService_GetPnL_EmptyData(t *testing.T) {
	svc, tdb := newTestReportServiceWithDB(t)
	defer tdb.close()

	uid := int64(1)
	startTime := int64(1700000000)
	endTime := startTime + 10000

	result, err := svc.GetPnL(nil, uid, 0, startTime, endTime)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	assert.Equal(t, int64(0), result.Revenue)
	assert.Equal(t, int64(0), result.CostOfGoods)
	assert.Equal(t, int64(0), result.GrossProfit)
	assert.Equal(t, int64(0), result.OperatingExpense)
	assert.Equal(t, int64(0), result.Depreciation)
	assert.Equal(t, int64(0), result.OperatingProfit)
	assert.Equal(t, int64(0), result.FinancialExpense)
	assert.Equal(t, int64(0), result.TaxExpense)
	assert.Equal(t, int64(0), result.NetProfit)
}

// TestReportService_GetCashFlow_WithCfoFilter verifies that CFO filtering works in cash flow.
func TestReportService_GetCashFlow_WithCfoFilter(t *testing.T) {
	svc, tdb := newTestReportServiceWithDB(t)
	defer tdb.close()

	uid := int64(1)
	baseTime := int64(1700000000)
	cfoId := int64(42)

	// Insert categories
	categories := []*models.TransactionCategory{
		{
			CategoryId:   500,
			Uid:          uid,
			Deleted:      false,
			Type:         models.CATEGORY_TYPE_INCOME,
			Name:         "CFO Sales",
			ActivityType: int32(models.ACTIVITY_TYPE_OPERATING),
			CostType:     0,
		},
		{
			CategoryId:   501,
			Uid:          uid,
			Deleted:      false,
			Type:         models.CATEGORY_TYPE_EXPENSE,
			Name:         "CFO Expense",
			ActivityType: int32(models.ACTIVITY_TYPE_OPERATING),
			CostType:     int32(models.COST_TYPE_OPERATIONAL),
		},
	}

	for _, cat := range categories {
		_, err := tdb.engine.Insert(cat)
		if err != nil {
			t.Fatalf("failed to insert category: %v", err)
		}
	}

	// Insert transactions: some with cfo_id=42, some with cfo_id=0 (no CFO)
	transactions := []*models.Transaction{
		{
			TransactionId:   5001,
			Uid:             uid,
			Deleted:         false,
			Type:            models.TRANSACTION_DB_TYPE_INCOME,
			CategoryId:      500,
			AccountId:       1,
			TransactionTime: baseTime + 100,
			Amount:          100000,
			Planned:         false,
			CfoId:           cfoId,
		},
		{
			TransactionId:   5002,
			Uid:             uid,
			Deleted:         false,
			Type:            models.TRANSACTION_DB_TYPE_INCOME,
			CategoryId:      500,
			AccountId:       1,
			TransactionTime: baseTime + 200,
			Amount:          200000,
			Planned:         false,
			CfoId:           0,
		},
		{
			TransactionId:   5003,
			Uid:             uid,
			Deleted:         false,
			Type:            models.TRANSACTION_DB_TYPE_EXPENSE,
			CategoryId:      501,
			AccountId:       1,
			TransactionTime: baseTime + 300,
			Amount:          30000,
			Planned:         false,
			CfoId:           cfoId,
		},
	}

	for _, txn := range transactions {
		_, err := tdb.engine.Insert(txn)
		if err != nil {
			t.Fatalf("failed to insert transaction: %v", err)
		}
	}

	startTime := baseTime
	endTime := baseTime + 10000

	// With CFO filter: only transactions with cfo_id=42
	result, err := svc.GetCashFlow(nil, uid, cfoId, startTime, endTime)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	// Operating: Income=100000 (only txn 5001), Expense=30000 (txn 5003)
	operating := result.Activities[0]
	assert.Equal(t, int64(100000), operating.TotalIncome)
	assert.Equal(t, int64(30000), operating.TotalExpense)
	assert.Equal(t, int64(70000), operating.TotalNet)

	// Without CFO filter: all transactions
	resultAll, err := svc.GetCashFlow(nil, uid, 0, startTime, endTime)
	assert.Nil(t, err)
	operatingAll := resultAll.Activities[0]
	// Income: 100000 + 200000 = 300000, Expense: 30000
	assert.Equal(t, int64(300000), operatingAll.TotalIncome)
	assert.Equal(t, int64(30000), operatingAll.TotalExpense)
	assert.Equal(t, int64(270000), operatingAll.TotalNet)
}

// TestReportService_GetCashFlow_ExcludesDeletedAndPlanned confirms that deleted and planned
// transactions are excluded from cash flow results.
func TestReportService_GetCashFlow_ExcludesDeletedAndPlanned(t *testing.T) {
	svc, tdb := newTestReportServiceWithDB(t)
	defer tdb.close()

	uid := int64(1)
	baseTime := int64(1700000000)
	seedTransactionData(t, tdb, uid, baseTime)

	startTime := baseTime
	endTime := baseTime + 10000

	result, err := svc.GetCashFlow(nil, uid, 0, startTime, endTime)
	assert.Nil(t, err)

	// Total does NOT include deleted (999999) or planned (888888) amounts
	// Valid income: 800000 + 150000 + 1000000 = 1950000
	// Valid expense: 200000 + 100000 + 50000 = 350000
	// Net: 1950000 - 350000 = 1600000
	assert.Equal(t, int64(1600000), result.TotalNet)
}

// TestReportService_GetCashFlow_TimeRange verifies that only transactions within the
// specified time window are included.
func TestReportService_GetCashFlow_TimeRange(t *testing.T) {
	svc, tdb := newTestReportServiceWithDB(t)
	defer tdb.close()

	uid := int64(1)
	baseTime := int64(1700000000)
	seedTransactionData(t, tdb, uid, baseTime)

	// Narrow time window: only covers the first income transaction (baseTime+100)
	startTime := baseTime + 50
	endTime := baseTime + 120

	result, err := svc.GetCashFlow(nil, uid, 0, startTime, endTime)
	assert.Nil(t, err)

	// Only the first Sales Revenue income (500000) falls in this window
	operating := result.Activities[0]
	assert.Equal(t, int64(500000), operating.TotalIncome)
	assert.Equal(t, int64(0), operating.TotalExpense)
	assert.Equal(t, int64(500000), result.TotalNet)
}

// TestReportService_GetPnL_WithDepreciation verifies that depreciation from assets
// is included in the P&L calculation.
func TestReportService_GetPnL_WithDepreciation(t *testing.T) {
	now := time.Now()
	commDate := now.AddDate(0, -6, 0)

	mockAssets := &mockAssetProvider{
		assets: []*models.Asset{
			{
				AssetId:          1,
				Uid:              1,
				PurchaseCost:     1200000, // 12000.00
				SalvageValue:     0,
				UsefulLifeMonths: 12,
				CommissionDate:   commDate.Unix(),
			},
		},
	}

	svc, tdb := newTestReportServiceWithDB(t, func(s *ReportService) {
		s.assets = mockAssets
	})
	defer tdb.close()

	uid := int64(1)
	startTime := commDate.Unix()
	endTime := now.Unix() + 1

	result, err := svc.GetPnL(nil, uid, 0, startTime, endTime)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	// Monthly depreciation = (1200000 - 0) / 12 = 100000 per month
	// For 6 months: 6 * 100000 = 600000
	assert.Equal(t, int64(600000), result.Depreciation)
	// OperatingProfit = 0 - 0 - 600000 = -600000
	assert.Equal(t, int64(-600000), result.OperatingProfit)
}

// TestReportService_GetPnL_WithTaxRecords verifies that tax records within the period
// are included in P&L tax expense.
func TestReportService_GetPnL_WithTaxRecords(t *testing.T) {
	baseTime := int64(1700000000)

	mockTaxes := &mockTaxRecordProvider{
		records: []*models.TaxRecord{
			{
				TaxId:     1,
				Uid:       1,
				TaxAmount: 75000,
				DueDate:   baseTime + 500, // within the query window
			},
			{
				TaxId:     2,
				Uid:       1,
				TaxAmount: 25000,
				DueDate:   baseTime + 99999, // outside the query window
			},
		},
	}

	svc, tdb := newTestReportServiceWithDB(t, func(s *ReportService) {
		s.taxes = mockTaxes
	})
	defer tdb.close()

	uid := int64(1)
	startTime := baseTime
	endTime := baseTime + 10000

	seedTransactionData(t, tdb, uid, baseTime)

	result, err := svc.GetPnL(nil, uid, 0, startTime, endTime)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	// Only the first tax record (75000) should be included (due date within window)
	assert.Equal(t, int64(75000), result.TaxExpense)

	// NetProfit = OperatingProfit - FinancialExpense - TaxExpense
	// OperatingProfit = GrossProfit - OpEx - Depr = (1950000 - 200000) - 100000 - 0 = 1650000
	// NetProfit = 1650000 - 50000 - 75000 = 1525000
	assert.Equal(t, int64(1525000), result.NetProfit)
}

// TestReportService_GetCashFlow_InvalidUid_WithDB verifies that invalid uid returns an error.
func TestReportService_GetCashFlow_InvalidUid_WithDB(t *testing.T) {
	svc, tdb := newTestReportServiceWithDB(t)
	defer tdb.close()

	result, err := svc.GetCashFlow(nil, 0, 0, 1000, 2000)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

// TestReportService_GetCashFlow_InvalidTimeRange_WithDB verifies that invalid time ranges
// are rejected.
func TestReportService_GetCashFlow_InvalidTimeRange_WithDB(t *testing.T) {
	svc, tdb := newTestReportServiceWithDB(t)
	defer tdb.close()

	// startTime >= endTime
	_, err := svc.GetCashFlow(nil, 1, 0, 2000, 1000)
	assert.NotNil(t, err)

	// Range too long (> 10 years)
	_, err = svc.GetCashFlow(nil, 1, 0, 1000, 1000+maxReportRangeSeconds+1)
	assert.NotNil(t, err)
}

// TestReportService_GetBalance_EmptyData verifies that balance sheet returns zeros
// when no accounts or obligations exist.
func TestReportService_GetBalance_EmptyData(t *testing.T) {
	svc, tdb := newTestReportServiceWithDB(t)
	defer tdb.close()

	result, err := svc.GetBalance(nil, 1, 0)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(0), result.TotalAssets)
	assert.Equal(t, int64(0), result.TotalLiability)
	assert.Equal(t, int64(0), result.Equity)
}

// TestReportService_GetPaymentCalendar_WithDB verifies payment calendar aggregation
// across obligations, tax records, and planned transactions.
func TestReportService_GetPaymentCalendar_WithDB(t *testing.T) {
	svc, tdb := newTestReportServiceWithDB(t)
	defer tdb.close()

	uid := int64(1)
	baseTime := int64(1700000000)

	// Insert an obligation with due date in range
	obligation := &models.Obligation{
		ObligationId:   10,
		Uid:            uid,
		Deleted:        false,
		ObligationType: models.OBLIGATION_TYPE_PAYABLE,
		Amount:         50000,
		PaidAmount:     10000,
		DueDate:        baseTime + 500,
		Status:         models.OBLIGATION_STATUS_ACTIVE,
		Currency:       "RUB",
		Comment:        "Supplier payment",
	}
	_, err := tdb.engine.Insert(obligation)
	assert.Nil(t, err)

	// Insert a tax record with due date in range
	taxRecord := &models.TaxRecord{
		TaxId:      20,
		Uid:        uid,
		Deleted:    false,
		TaxAmount:  30000,
		PaidAmount: 0,
		DueDate:    baseTime + 600,
		Status:     models.TAX_STATUS_PENDING,
		Currency:   "RUB",
		Comment:    "Q4 tax",
	}
	_, err = tdb.engine.Insert(taxRecord)
	assert.Nil(t, err)

	// Insert a planned transaction in range
	planned := &models.Transaction{
		TransactionId:   30,
		Uid:             uid,
		Deleted:         false,
		Type:            models.TRANSACTION_DB_TYPE_EXPENSE,
		CategoryId:      1,
		AccountId:       1,
		TransactionTime: baseTime + 700,
		Amount:          20000,
		Planned:         true,
		Comment:         "Planned purchase",
	}
	_, err = tdb.engine.Insert(planned)
	assert.Nil(t, err)

	startTime := baseTime
	endTime := baseTime + 10000

	result, err := svc.GetPaymentCalendar(nil, uid, startTime, endTime)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, len(result.Items))

	// Items should be sorted by date ascending
	assert.Equal(t, baseTime+500, result.Items[0].Date)
	assert.Equal(t, "Payable", result.Items[0].Type)
	assert.Equal(t, int64(40000), result.Items[0].Amount) // 50000 - 10000

	assert.Equal(t, baseTime+600, result.Items[1].Date)
	assert.Equal(t, "Tax", result.Items[1].Type)
	assert.Equal(t, int64(30000), result.Items[1].Amount)

	assert.Equal(t, baseTime+700, result.Items[2].Date)
	assert.Equal(t, "Planned", result.Items[2].Type)
	assert.Equal(t, int64(20000), result.Items[2].Amount)
}
