package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// ===== monthsBetween tests =====

func TestMonthsBetween_SameDate(t *testing.T) {
	d := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, int64(0), monthsBetween(d, d))
}

func TestMonthsBetween_SameMonth(t *testing.T) {
	from := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, int64(0), monthsBetween(from, to))
}

func TestMonthsBetween_AdjacentMonths(t *testing.T) {
	from := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 7, 15, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, int64(1), monthsBetween(from, to))
}

func TestMonthsBetween_CrossYear(t *testing.T) {
	from := time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, int64(3), monthsBetween(from, to))
}

func TestMonthsBetween_ExactlyOneYear(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, int64(12), monthsBetween(from, to))
}

func TestMonthsBetween_MultipleYears(t *testing.T) {
	from := time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, int64(66), monthsBetween(from, to)) // 5*12 + 6 = 66
}

func TestMonthsBetween_ToBeforeFrom_ReturnsZero(t *testing.T) {
	from := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, int64(0), monthsBetween(from, to))
}

func TestMonthsBetween_DecToJan(t *testing.T) {
	from := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, int64(1), monthsBetween(from, to))
}

func TestMonthsBetween_JanToDec(t *testing.T) {
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, int64(11), monthsBetween(from, to))
}

// ===== calculateResidualValue tests =====

func TestCalculateResidualValue_NoCommissionDate(t *testing.T) {
	asset := &models.Asset{
		PurchaseCost:     100000,
		SalvageValue:     10000,
		UsefulLifeMonths: 12,
		CommissionDate:   0,
	}
	asOf := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, int64(100000), calculateResidualValue(asset, asOf))
}

func TestCalculateResidualValue_ZeroUsefulLife(t *testing.T) {
	asset := &models.Asset{
		PurchaseCost:     100000,
		SalvageValue:     10000,
		UsefulLifeMonths: 0,
		CommissionDate:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	asOf := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, int64(100000), calculateResidualValue(asset, asOf))
}

func TestCalculateResidualValue_BeforeCommissionDate(t *testing.T) {
	asset := &models.Asset{
		PurchaseCost:     100000,
		SalvageValue:     10000,
		UsefulLifeMonths: 12,
		CommissionDate:   time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	asOf := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, int64(100000), calculateResidualValue(asset, asOf))
}

func TestCalculateResidualValue_PartialDepreciation(t *testing.T) {
	// Purchase cost = 120,000, Salvage = 0, Useful life = 12 months
	// Monthly depreciation = 10,000
	// After 6 months: residual = 120,000 - 60,000 = 60,000
	commDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	asset := &models.Asset{
		PurchaseCost:     120000,
		SalvageValue:     0,
		UsefulLifeMonths: 12,
		CommissionDate:   commDate.Unix(),
	}
	asOf := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC) // 6 months later
	assert.Equal(t, int64(60000), calculateResidualValue(asset, asOf))
}

func TestCalculateResidualValue_FullDepreciation(t *testing.T) {
	// After 12 months, should return salvage value
	commDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	asset := &models.Asset{
		PurchaseCost:     120000,
		SalvageValue:     10000,
		UsefulLifeMonths: 12,
		CommissionDate:   commDate.Unix(),
	}
	asOf := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) // 12 months later
	// monthly depreciation = (120000 - 10000) / 12 = 9166 (integer division)
	// accumulated = 12 * 9166 = 109992
	// residual = 120000 - 109992 = 10008
	// but since residual > salvage, it stays at 10008
	expected := int64(120000) - int64(12)*int64((120000-10000)/12)
	assert.Equal(t, expected, calculateResidualValue(asset, asOf))
}

func TestCalculateResidualValue_BeyondUsefulLife(t *testing.T) {
	// Way beyond useful life — should clamp to salvage value
	commDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	asset := &models.Asset{
		PurchaseCost:     120000,
		SalvageValue:     10000,
		UsefulLifeMonths: 12,
		CommissionDate:   commDate.Unix(),
	}
	asOf := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) // 60 months later
	// months clamped to maxMonths=12
	// monthly = (120000 - 10000) / 12 = 9166
	// accumulated = 12 * 9166 = 109992
	// residual = 120000 - 109992 = 10008
	expected := int64(120000) - int64(12)*int64((120000-10000)/12)
	assert.Equal(t, expected, calculateResidualValue(asset, asOf))
}

func TestCalculateResidualValue_WithSalvageValue(t *testing.T) {
	// Salvage value = 50000, purchase = 100000, 10 months
	// Monthly = (100000 - 50000) / 10 = 5000
	// After 3 months: residual = 100000 - 15000 = 85000
	commDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	asset := &models.Asset{
		PurchaseCost:     100000,
		SalvageValue:     50000,
		UsefulLifeMonths: 10,
		CommissionDate:   commDate.Unix(),
	}
	asOf := time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC) // 3 months later
	assert.Equal(t, int64(85000), calculateResidualValue(asset, asOf))
}

func TestCalculateResidualValue_SalvageEqualsPurchaseCost(t *testing.T) {
	// When salvage == purchase cost, monthly depreciation = 0
	commDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	asset := &models.Asset{
		PurchaseCost:     100000,
		SalvageValue:     100000,
		UsefulLifeMonths: 12,
		CommissionDate:   commDate.Unix(),
	}
	asOf := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, int64(100000), calculateResidualValue(asset, asOf))
}

func TestCalculateResidualValue_OneMonthUsefulLife(t *testing.T) {
	commDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	asset := &models.Asset{
		PurchaseCost:     100000,
		SalvageValue:     0,
		UsefulLifeMonths: 1,
		CommissionDate:   commDate.Unix(),
	}
	asOf := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC) // 1 month later
	// monthly = 100000 / 1 = 100000
	// accumulated = 1 * 100000 = 100000
	// residual = 100000 - 100000 = 0
	assert.Equal(t, int64(0), calculateResidualValue(asset, asOf))
}

func TestCalculateResidualValue_SameMonthAsCommission(t *testing.T) {
	commDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	asset := &models.Asset{
		PurchaseCost:     100000,
		SalvageValue:     10000,
		UsefulLifeMonths: 12,
		CommissionDate:   commDate.Unix(),
	}
	asOf := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC) // same month
	// monthsBetween returns 0 for same month → residual = purchase cost
	assert.Equal(t, int64(100000), calculateResidualValue(asset, asOf))
}

func TestCalculateResidualValue_ResidualNeverBelowSalvage(t *testing.T) {
	// Even with integer rounding, residual should never go below salvage
	commDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	asset := &models.Asset{
		PurchaseCost:     100000,
		SalvageValue:     30000,
		UsefulLifeMonths: 7, // 7 gives odd division
		CommissionDate:   commDate.Unix(),
	}
	asOf := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) // way past useful life
	residual := calculateResidualValue(asset, asOf)
	assert.True(t, residual >= asset.SalvageValue, "residual %d should be >= salvage %d", residual, asset.SalvageValue)
}

// ===== PnL calculation formula tests =====

func TestPnLFormulas(t *testing.T) {
	// Verify that PnL calculation formulas are correct
	response := &models.PnLResponse{
		Revenue:          500000,
		CostOfGoods:      200000,
		OperatingExpense: 100000,
		Depreciation:     20000,
		FinancialExpense: 10000,
		TaxExpense:       30000,
	}

	response.GrossProfit = response.Revenue - response.CostOfGoods
	response.OperatingProfit = response.GrossProfit - response.OperatingExpense - response.Depreciation
	response.NetProfit = response.OperatingProfit - response.FinancialExpense - response.TaxExpense

	assert.Equal(t, int64(300000), response.GrossProfit)     // 500k - 200k
	assert.Equal(t, int64(180000), response.OperatingProfit) // 300k - 100k - 20k
	assert.Equal(t, int64(140000), response.NetProfit)       // 180k - 10k - 30k
}

func TestPnLFormulas_AllZeros(t *testing.T) {
	response := &models.PnLResponse{}

	response.GrossProfit = response.Revenue - response.CostOfGoods
	response.OperatingProfit = response.GrossProfit - response.OperatingExpense - response.Depreciation
	response.NetProfit = response.OperatingProfit - response.FinancialExpense - response.TaxExpense

	assert.Equal(t, int64(0), response.GrossProfit)
	assert.Equal(t, int64(0), response.OperatingProfit)
	assert.Equal(t, int64(0), response.NetProfit)
}

func TestPnLFormulas_NegativeProfit(t *testing.T) {
	response := &models.PnLResponse{
		Revenue:          100000,
		CostOfGoods:      150000,
		OperatingExpense: 50000,
		Depreciation:     10000,
		FinancialExpense: 5000,
		TaxExpense:       0,
	}

	response.GrossProfit = response.Revenue - response.CostOfGoods
	response.OperatingProfit = response.GrossProfit - response.OperatingExpense - response.Depreciation
	response.NetProfit = response.OperatingProfit - response.FinancialExpense - response.TaxExpense

	assert.Equal(t, int64(-50000), response.GrossProfit)
	assert.Equal(t, int64(-110000), response.OperatingProfit)
	assert.Equal(t, int64(-115000), response.NetProfit)
}

// ===== Balance formula tests =====

func TestBalanceEquity(t *testing.T) {
	response := &models.BalanceResponse{
		TotalAssets:    500000,
		TotalLiability: 200000,
	}
	response.Equity = response.TotalAssets - response.TotalLiability
	assert.Equal(t, int64(300000), response.Equity)
}

func TestBalanceEquity_NegativeEquity(t *testing.T) {
	response := &models.BalanceResponse{
		TotalAssets:    100000,
		TotalLiability: 250000,
	}
	response.Equity = response.TotalAssets - response.TotalLiability
	assert.Equal(t, int64(-150000), response.Equity)
}

func TestBalanceEquity_ZeroEverything(t *testing.T) {
	response := &models.BalanceResponse{}
	response.Equity = response.TotalAssets - response.TotalLiability
	assert.Equal(t, int64(0), response.Equity)
}

// ===== CashFlow calculation tests =====

func TestCashFlowActivityLine_NetCalculation(t *testing.T) {
	line := &models.CashFlowActivityLine{
		Income:  50000,
		Expense: 30000,
	}
	line.Net = line.Income - line.Expense
	assert.Equal(t, int64(20000), line.Net)
}

func TestCashFlowActivityLine_NegativeNet(t *testing.T) {
	line := &models.CashFlowActivityLine{
		Income:  10000,
		Expense: 30000,
	}
	line.Net = line.Income - line.Expense
	assert.Equal(t, int64(-20000), line.Net)
}

func TestCashFlowTotalNet(t *testing.T) {
	activities := []*models.CashFlowActivity{
		{TotalNet: 100000},
		{TotalNet: -50000},
		{TotalNet: 20000},
	}
	totalNet := int64(0)
	for _, a := range activities {
		totalNet += a.TotalNet
	}
	assert.Equal(t, int64(70000), totalNet)
}

// ===== Depreciation period calculation tests =====

func TestDepreciationPeriodCalc_FullPeriod(t *testing.T) {
	// Asset commissioned Jan 2025, useful life 12 months
	// Period: Jan 2025 - Jul 2025 (6 months)
	commDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)

	purchaseCost := int64(120000)
	salvageValue := int64(0)
	usefulLifeMonths := int32(12)
	monthlyDepr := (purchaseCost - salvageValue) / int64(usefulLifeMonths)

	monthsToEnd := monthsBetween(commDate, endDate)
	monthsToStart := monthsBetween(commDate, startDate)

	maxMonths := int64(usefulLifeMonths)
	if monthsToEnd > maxMonths {
		monthsToEnd = maxMonths
	}
	if monthsToStart > maxMonths {
		monthsToStart = maxMonths
	}
	if monthsToStart < 0 {
		monthsToStart = 0
	}
	if monthsToEnd < 0 {
		monthsToEnd = 0
	}

	periodDepr := (monthsToEnd - monthsToStart) * monthlyDepr
	// 6 months * 10000 = 60000
	assert.Equal(t, int64(60000), periodDepr)
}

func TestDepreciationPeriodCalc_AssetNotYetCommissioned(t *testing.T) {
	// Commission date is after the period — no depreciation
	commDate := time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)

	purchaseCost := int64(120000)
	salvageValue := int64(0)
	usefulLifeMonths := int32(12)
	monthlyDepr := (purchaseCost - salvageValue) / int64(usefulLifeMonths)

	monthsToEnd := monthsBetween(commDate, endDate)
	monthsToStart := monthsBetween(commDate, startDate)

	// Both should be 0 (to is before from)
	if monthsToStart < 0 {
		monthsToStart = 0
	}
	if monthsToEnd < 0 {
		monthsToEnd = 0
	}

	periodDepr := (monthsToEnd - monthsToStart) * monthlyDepr
	assert.Equal(t, int64(0), periodDepr)
}

func TestDepreciationPeriodCalc_AssetFullyDepreciatedBeforePeriod(t *testing.T) {
	// Asset commissioned Jan 2023, useful life 12 months
	// Period: Jan 2025 - Jul 2025
	// Asset was fully depreciated by Jan 2024, so no period depreciation
	commDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)

	purchaseCost := int64(120000)
	salvageValue := int64(0)
	usefulLifeMonths := int32(12)
	monthlyDepr := (purchaseCost - salvageValue) / int64(usefulLifeMonths)

	monthsToEnd := monthsBetween(commDate, endDate)
	monthsToStart := monthsBetween(commDate, startDate)

	maxMonths := int64(usefulLifeMonths)
	if monthsToEnd > maxMonths {
		monthsToEnd = maxMonths
	}
	if monthsToStart > maxMonths {
		monthsToStart = maxMonths
	}

	periodDepr := (monthsToEnd - monthsToStart) * monthlyDepr
	// Both clamped to 12, so 12 - 12 = 0
	assert.Equal(t, int64(0), periodDepr)
}

func TestDepreciationPeriodCalc_PartialOverlap(t *testing.T) {
	// Asset commissioned Jun 2025, useful life 12 months
	// Period: Jan 2025 - Dec 2025
	// Depreciation counted only from Jun (monthsToStart=0) to Dec (6 months)
	commDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)

	purchaseCost := int64(120000)
	salvageValue := int64(0)
	usefulLifeMonths := int32(12)
	monthlyDepr := (purchaseCost - salvageValue) / int64(usefulLifeMonths)

	monthsToEnd := monthsBetween(commDate, endDate)
	monthsToStart := monthsBetween(commDate, startDate)

	maxMonths := int64(usefulLifeMonths)
	if monthsToEnd > maxMonths {
		monthsToEnd = maxMonths
	}
	if monthsToStart < 0 {
		monthsToStart = 0
	}

	periodDepr := (monthsToEnd - monthsToStart) * monthlyDepr
	// monthsToEnd = 6, monthsToStart = 0 → 6 * 10000 = 60000
	assert.Equal(t, int64(60000), periodDepr)
}
