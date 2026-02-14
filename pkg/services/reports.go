// reports.go provides financial report generation including Cash Flow,
// Profit & Loss, Balance Sheet, and Payment Calendar.
package services

import (
	"sort"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// ReportService represents report service
type ReportService struct {
	ServiceUsingDB
}

// Initialize a report service singleton instance
var (
	Reports = &ReportService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
	}
)

const (
	// cashFlowBaseQuery retrieves transaction amounts grouped by category and activity type
	cashFlowBaseQuery = `SELECT t.category_id, tc.name as category_name, tc.activity_type, t.type, SUM(t.amount) as total_amount
		FROM "transaction" t
		JOIN transaction_category tc ON t.category_id = tc.category_id AND tc.uid = t.uid
		WHERE t.uid = ? AND t.deleted = 0 AND t.planned = 0
		AND t.transaction_time >= ? AND t.transaction_time < ?
		AND t.type IN (2, 3)`

	// pnlBaseQuery retrieves transaction amounts grouped by cost type
	pnlBaseQuery = `SELECT tc.cost_type, t.type, SUM(t.amount) as total_amount
		FROM "transaction" t
		JOIN transaction_category tc ON t.category_id = tc.category_id AND tc.uid = t.uid
		WHERE t.uid = ? AND t.deleted = 0 AND t.planned = 0
		AND t.transaction_time >= ? AND t.transaction_time < ?
		AND t.type IN (2, 3)`

	// cfoFilterClause is appended when filtering by CFO
	cfoFilterClause = " AND t.cfo_id = ?"

	// maxReportRangeSeconds limits report queries to 10 years
	maxReportRangeSeconds = 10 * 365 * 24 * 60 * 60
)

// matchesCfo returns true if cfoId filter is not set or entity belongs to the specified CFO
func matchesCfo(filterCfoId int64, entityCfoId int64) bool {
	return filterCfoId <= 0 || entityCfoId == filterCfoId
}

// validateTimeRange checks that startTime < endTime and the range does not exceed maxReportRangeSeconds
func validateTimeRange(startTime int64, endTime int64) error {
	if startTime >= endTime {
		return errs.ErrReportStartTimeAfterEndTime
	}
	if endTime-startTime > maxReportRangeSeconds {
		return errs.ErrReportTimeRangeTooLong
	}
	return nil
}

// transactionRow is a helper struct for SQL query results
type transactionRow struct {
	CategoryId   int64  `xorm:"category_id"`
	CategoryName string `xorm:"category_name"`
	ActivityType int32  `xorm:"activity_type"`
	CostType     int32  `xorm:"cost_type"`
	Type         int32  `xorm:"type"`
	Amount       int64  `xorm:"total_amount"`
}

// GetCashFlow returns a Cash Flow Statement (ОДДС / Statement of Cash Flows).
// Groups all non-transfer transactions by category activity_type:
//   - Operating (activity_type=1): day-to-day business transactions
//   - Investing (activity_type=2): asset purchases/sales, long-term investments
//   - Financing (activity_type=3): loans, investor contributions, debt payments
//
// Only confirmed (planned=false) income and expense transactions are included.
// Transfers between accounts are excluded.
// Optionally filtered by CFO (Center of Financial Responsibility).
func (s *ReportService) GetCashFlow(c core.Context, uid int64, cfoId int64, startTime int64, endTime int64) (*models.CashFlowResponse, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}
	if err := validateTimeRange(startTime, endTime); err != nil {
		return nil, err
	}

	var rows []*transactionRow

	query := cashFlowBaseQuery
	args := []interface{}{uid, startTime, endTime}

	if cfoId > 0 {
		query += cfoFilterClause
		args = append(args, cfoId)
	}

	query += " GROUP BY t.category_id, tc.name, tc.activity_type, t.type"

	err := s.UserDataDB(uid).NewSession(c).SQL(query, args...).Find(&rows)

	if err != nil {
		return nil, err
	}

	// Group by activity type
	activityMap := map[int32]*models.CashFlowActivity{
		int32(models.ACTIVITY_TYPE_OPERATING): {ActivityType: int32(models.ACTIVITY_TYPE_OPERATING), ActivityName: "Operating", Lines: []*models.CashFlowActivityLine{}},
		int32(models.ACTIVITY_TYPE_INVESTING): {ActivityType: int32(models.ACTIVITY_TYPE_INVESTING), ActivityName: "Investing", Lines: []*models.CashFlowActivityLine{}},
		int32(models.ACTIVITY_TYPE_FINANCING): {ActivityType: int32(models.ACTIVITY_TYPE_FINANCING), ActivityName: "Financing", Lines: []*models.CashFlowActivityLine{}},
	}

	// Track per-category aggregation
	type catKey struct {
		activityType int32
		categoryId   int64
	}
	catAgg := map[catKey]*models.CashFlowActivityLine{}

	for _, row := range rows {
		at := row.ActivityType
		if at < int32(models.ACTIVITY_TYPE_OPERATING) || at > int32(models.ACTIVITY_TYPE_FINANCING) {
			at = int32(models.ACTIVITY_TYPE_OPERATING)
		}

		key := catKey{activityType: at, categoryId: row.CategoryId}
		line, exists := catAgg[key]
		if !exists {
			line = &models.CashFlowActivityLine{
				CategoryId:   row.CategoryId,
				CategoryName: row.CategoryName,
			}
			catAgg[key] = line
		}

		if row.Type == int32(models.TRANSACTION_DB_TYPE_INCOME) {
			line.Income += row.Amount
		} else if row.Type == int32(models.TRANSACTION_DB_TYPE_EXPENSE) {
			line.Expense += row.Amount
		}
	}

	for key, line := range catAgg {
		line.Net = line.Income - line.Expense
		activity := activityMap[key.activityType]
		activity.Lines = append(activity.Lines, line)
		activity.TotalIncome += line.Income
		activity.TotalExpense += line.Expense
		activity.TotalNet += line.Net
	}

	activities := []*models.CashFlowActivity{activityMap[int32(models.ACTIVITY_TYPE_OPERATING)], activityMap[int32(models.ACTIVITY_TYPE_INVESTING)], activityMap[int32(models.ACTIVITY_TYPE_FINANCING)]}
	totalNet := int64(0)

	for _, a := range activities {
		totalNet += a.TotalNet
	}

	return &models.CashFlowResponse{
		Activities: activities,
		TotalNet:   totalNet,
	}, nil
}

// GetPnL returns a Profit & Loss statement (ОПиУ / Income Statement).
// Formula:
//
//	Revenue (all income transactions)
//	- Cost of Goods Sold (expenses with cost_type=COGS)
//	= Gross Profit
//	- Operating Expenses (expenses with cost_type=operational)
//	- Depreciation (straight-line, calculated from assets)
//	= Operating Profit (EBIT)
//	- Financial Expenses (expenses with cost_type=financial)
//	- Tax Expenses (from tax_record table, matched by period)
//	= Net Profit
func (s *ReportService) GetPnL(c core.Context, uid int64, cfoId int64, startTime int64, endTime int64) (*models.PnLResponse, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}
	if err := validateTimeRange(startTime, endTime); err != nil {
		return nil, err
	}

	var rows []*transactionRow

	query := pnlBaseQuery
	args := []interface{}{uid, startTime, endTime}

	if cfoId > 0 {
		query += cfoFilterClause
		args = append(args, cfoId)
	}

	query += " GROUP BY tc.cost_type, t.type"

	err := s.UserDataDB(uid).NewSession(c).SQL(query, args...).Find(&rows)

	if err != nil {
		return nil, err
	}

	response := &models.PnLResponse{}

	for _, row := range rows {
		if row.Type == int32(models.TRANSACTION_DB_TYPE_INCOME) {
			response.Revenue += row.Amount
		} else if row.Type == int32(models.TRANSACTION_DB_TYPE_EXPENSE) {
			switch models.CostType(row.CostType) {
			case models.COST_TYPE_COGS:
				response.CostOfGoods += row.Amount
			case models.COST_TYPE_OPERATIONAL:
				response.OperatingExpense += row.Amount
			case models.COST_TYPE_FINANCIAL:
				response.FinancialExpense += row.Amount
			default:
				response.OperatingExpense += row.Amount
			}
		}
	}

	// Calculate depreciation from assets
	assets, err := Assets.GetAllAssetsByUid(c, uid)
	if err != nil {
		log.Warnf(c, "[reports.GetPnL] failed to load assets for uid:%d: %s", uid, err.Error())
		response.Warnings = append(response.Warnings, "Failed to load asset data for depreciation calculation")
	} else {
		now := time.Now()
		for _, asset := range assets {
			if asset.CommissionDate <= 0 || asset.UsefulLifeMonths <= 0 {
				continue
			}
			if !matchesCfo(cfoId, asset.CfoId) {
				continue
			}

			commDate := time.Unix(asset.CommissionDate, 0)
			asOfDate := now
			if endTime > 0 {
				asOfDate = time.Unix(endTime, 0)
			}

			// Only count depreciation within the period
			startDate := time.Unix(startTime, 0)
			monthlyDepr := (asset.PurchaseCost - asset.SalvageValue) / int64(asset.UsefulLifeMonths)

			// Months from commission to end of period
			monthsToEnd := monthsBetween(commDate, asOfDate)
			// Months from commission to start of period
			monthsToStart := monthsBetween(commDate, startDate)

			maxMonths := int64(asset.UsefulLifeMonths)
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
			if periodDepr > 0 {
				response.Depreciation += periodDepr
			}
		}
	}

	// Get taxes for the period
	taxRecords, err := TaxRecords.GetAllTaxRecordsByUid(c, uid)
	if err != nil {
		log.Warnf(c, "[reports.GetPnL] failed to load tax records for uid:%d: %s", uid, err.Error())
		response.Warnings = append(response.Warnings, "Failed to load tax records for tax expense calculation")
	} else {
		for _, tr := range taxRecords {
			if !matchesCfo(cfoId, tr.CfoId) {
				continue
			}
			if tr.DueDate >= startTime && tr.DueDate < endTime {
				response.TaxExpense += tr.TaxAmount
			}
		}
	}

	response.GrossProfit = response.Revenue - response.CostOfGoods
	response.OperatingProfit = response.GrossProfit - response.OperatingExpense - response.Depreciation
	response.NetProfit = response.OperatingProfit - response.FinancialExpense - response.TaxExpense

	return response, nil
}

// GetBalance returns a Balance Sheet (Баланс / Statement of Financial Position).
// Structure:
//
//	ASSETS = Cash & Bank Accounts + Receivables + Fixed Assets (residual value)
//	LIABILITIES = Payables + Credit Debts + Tax Liabilities + Investor Debt
//	EQUITY = Total Assets - Total Liabilities
//
// Fixed asset residual values use straight-line depreciation:
//
//	monthly_depreciation = (purchase_cost - salvage_value) / useful_life_months
//	residual = purchase_cost - (months_elapsed * monthly_depreciation)
func (s *ReportService) GetBalance(c core.Context, uid int64, cfoId int64) (*models.BalanceResponse, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	response := &models.BalanceResponse{
		AssetLines:     []*models.BalanceLine{},
		LiabilityLines: []*models.BalanceLine{},
	}

	// 1. Cash in accounts (assets)
	var accounts []*models.Account
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).Find(&accounts)
	if err != nil {
		return nil, err
	}

	cashAssets := int64(0)
	cashLiabilities := int64(0)
	for _, acc := range accounts {
		if acc.Category.IsAsset() {
			cashAssets += acc.Balance
		}
		if acc.Category.IsLiability() {
			cashLiabilities += acc.Balance
		}
	}
	if cashAssets != 0 {
		response.AssetLines = append(response.AssetLines, &models.BalanceLine{Label: "Cash & Bank Accounts", Amount: cashAssets})
	}

	// 2. Receivables (obligation type 1, not fully paid)
	var obligations []*models.Obligation
	err = s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).Find(&obligations)
	if err != nil {
		return nil, err
	}

	receivables := int64(0)
	payables := int64(0)
	for _, o := range obligations {
		if !matchesCfo(cfoId, o.CfoId) {
			continue
		}
		remaining := o.Amount - o.PaidAmount
		if remaining <= 0 {
			continue
		}
		if o.ObligationType == models.OBLIGATION_TYPE_RECEIVABLE {
			receivables += remaining
		} else if o.ObligationType == models.OBLIGATION_TYPE_PAYABLE {
			payables += remaining
		}
	}
	if receivables != 0 {
		response.AssetLines = append(response.AssetLines, &models.BalanceLine{Label: "Receivables", Amount: receivables})
	}

	// 3. Fixed assets (residual values)
	assets, err := Assets.GetAllAssetsByUid(c, uid)
	if err != nil {
		log.Warnf(c, "[reports.GetBalance] failed to load assets for uid:%d: %s", uid, err.Error())
		response.Warnings = append(response.Warnings, "Failed to load asset data for fixed assets calculation")
	} else {
		now := time.Now()
		totalResidual := int64(0)
		for _, asset := range assets {
			if !matchesCfo(cfoId, asset.CfoId) {
				continue
			}
			residual := calculateResidualValue(asset, now)
			totalResidual += residual
		}
		if totalResidual != 0 {
			response.AssetLines = append(response.AssetLines, &models.BalanceLine{Label: "Fixed Assets", Amount: totalResidual})
		}
	}

	// Calculate total assets
	for _, line := range response.AssetLines {
		response.TotalAssets += line.Amount
	}

	// LIABILITIES
	if payables != 0 {
		response.LiabilityLines = append(response.LiabilityLines, &models.BalanceLine{Label: "Payables", Amount: payables})
	}
	if cashLiabilities != 0 {
		response.LiabilityLines = append(response.LiabilityLines, &models.BalanceLine{Label: "Credit Cards & Debts", Amount: cashLiabilities})
	}

	// Tax liabilities (unpaid)
	taxRecords, err := TaxRecords.GetAllTaxRecordsByUid(c, uid)
	if err != nil {
		log.Warnf(c, "[reports.GetBalance] failed to load tax records for uid:%d: %s", uid, err.Error())
		response.Warnings = append(response.Warnings, "Failed to load tax records for tax liabilities calculation")
	} else {
		taxLiability := int64(0)
		for _, tr := range taxRecords {
			if !matchesCfo(cfoId, tr.CfoId) {
				continue
			}
			if tr.Status != models.TAX_STATUS_PAID {
				remaining := tr.TaxAmount - tr.PaidAmount
				if remaining > 0 {
					taxLiability += remaining
				}
			}
		}
		if taxLiability != 0 {
			response.LiabilityLines = append(response.LiabilityLines, &models.BalanceLine{Label: "Tax Liabilities", Amount: taxLiability})
		}
	}

	// Investor debt
	deals, err := InvestorDeals.GetAllDealsByUid(c, uid)
	if err != nil {
		log.Warnf(c, "[reports.GetBalance] failed to load investor deals for uid:%d: %s", uid, err.Error())
		response.Warnings = append(response.Warnings, "Failed to load investor deals for investor debt calculation")
	} else {
		investorDebt := int64(0)
		for _, deal := range deals {
			if !matchesCfo(cfoId, deal.CfoId) {
				continue
			}
			payments, pErr := InvestorPayments.GetAllPaymentsByDealId(c, uid, deal.DealId)
			if pErr != nil {
				continue
			}
			totalPaid := int64(0)
			for _, p := range payments {
				totalPaid += p.Amount
			}
			remaining := deal.TotalToRepay - totalPaid
			if remaining > 0 {
				investorDebt += remaining
			}
		}
		if investorDebt != 0 {
			response.LiabilityLines = append(response.LiabilityLines, &models.BalanceLine{Label: "Investor Debt", Amount: investorDebt})
		}
	}

	// Calculate total liabilities
	for _, line := range response.LiabilityLines {
		response.TotalLiability += line.Amount
	}

	response.Equity = response.TotalAssets - response.TotalLiability

	return response, nil
}

// GetPaymentCalendar returns upcoming payments from three sources:
//  1. Obligations (receivables/payables) with due dates in range
//  2. Tax records with due dates in range
//  3. Planned (unconfirmed) transactions with dates in range
//
// Results are sorted by date ascending.
func (s *ReportService) GetPaymentCalendar(c core.Context, uid int64, startTime int64, endTime int64) (*models.PaymentCalendarResponse, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}
	if err := validateTimeRange(startTime, endTime); err != nil {
		return nil, err
	}

	items := []*models.PaymentCalendarItem{}
	var warnings []string

	// 1. Obligations with due dates in range
	var obligations []*models.Obligation
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=? AND status!=? AND due_date>=? AND due_date<?", uid, false, models.OBLIGATION_STATUS_PAID, startTime, endTime).Find(&obligations)
	if err != nil {
		log.Warnf(c, "[reports.GetPaymentCalendar] failed to load obligations for uid:%d: %s", uid, err.Error())
		warnings = append(warnings, "Failed to load obligations")
	} else {
		for _, o := range obligations {
			typeName := "Receivable"
			if o.ObligationType == models.OBLIGATION_TYPE_PAYABLE {
				typeName = "Payable"
			}
			remaining := o.Amount - o.PaidAmount
			items = append(items, &models.PaymentCalendarItem{
				Date:        o.DueDate,
				Type:        typeName,
				Amount:      remaining,
				Description: o.Comment,
				Currency:    o.Currency,
			})
		}
	}

	// 2. Tax records with due dates in range
	var taxRecords []*models.TaxRecord
	err = s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=? AND status!=? AND due_date>=? AND due_date<?", uid, false, models.TAX_STATUS_PAID, startTime, endTime).Find(&taxRecords)
	if err != nil {
		log.Warnf(c, "[reports.GetPaymentCalendar] failed to load tax records for uid:%d: %s", uid, err.Error())
		warnings = append(warnings, "Failed to load tax records")
	} else {
		for _, tr := range taxRecords {
			remaining := tr.TaxAmount - tr.PaidAmount
			items = append(items, &models.PaymentCalendarItem{
				Date:        tr.DueDate,
				Type:        "Tax",
				Amount:      remaining,
				Description: tr.Comment,
				Currency:    "RUB",
			})
		}
	}

	// 3. Planned transactions in range
	var plannedTransactions []*models.Transaction
	err = s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=? AND planned=? AND transaction_time>=? AND transaction_time<?", uid, false, true, startTime, endTime).Find(&plannedTransactions)
	if err != nil {
		log.Warnf(c, "[reports.GetPaymentCalendar] failed to load planned transactions for uid:%d: %s", uid, err.Error())
		warnings = append(warnings, "Failed to load planned transactions")
	} else {
		for _, t := range plannedTransactions {
			typeName := "Planned"
			items = append(items, &models.PaymentCalendarItem{
				Date:        t.TransactionTime,
				Type:        typeName,
				Amount:      t.Amount,
				Description: t.Comment,
				Currency:    "",
			})
		}
	}

	// Sort by date
	sort.Slice(items, func(i, j int) bool {
		return items[i].Date < items[j].Date
	})

	return &models.PaymentCalendarResponse{
		Items:    items,
		Warnings: warnings,
	}, nil
}

// calculateResidualValue calculates the residual (book) value of a fixed asset
// at a given point in time using straight-line depreciation.
// If the asset has no commission date or zero useful life, returns purchase cost.
// Returns at minimum the salvage value.
func calculateResidualValue(asset *models.Asset, asOf time.Time) int64 {
	if asset.CommissionDate <= 0 || asset.UsefulLifeMonths <= 0 {
		return asset.PurchaseCost
	}

	commDate := time.Unix(asset.CommissionDate, 0)
	months := monthsBetween(commDate, asOf)
	if months <= 0 {
		return asset.PurchaseCost
	}

	monthlyDepr := (asset.PurchaseCost - asset.SalvageValue) / int64(asset.UsefulLifeMonths)
	maxMonths := int64(asset.UsefulLifeMonths)

	if months > maxMonths {
		months = maxMonths
	}

	accumulated := months * monthlyDepr
	residual := asset.PurchaseCost - accumulated
	if residual < asset.SalvageValue {
		residual = asset.SalvageValue
	}

	return residual
}

// monthsBetween calculates the number of whole months between two dates.
// Returns 0 if 'to' is before 'from'.
func monthsBetween(from time.Time, to time.Time) int64 {
	if to.Before(from) {
		return 0
	}

	years := int64(to.Year() - from.Year())
	months := int64(to.Month() - from.Month())

	return years*12 + months
}
