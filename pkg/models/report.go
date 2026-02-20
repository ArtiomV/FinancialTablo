package models

// Balance sheet line labels
const (
	BalanceLabelCashAndBank    = "Cash & Bank Accounts"
	BalanceLabelReceivables    = "Receivables"
	BalanceLabelFixedAssets    = "Fixed Assets"
	BalanceLabelPayables       = "Payables"
	BalanceLabelCreditCards    = "Credit Cards & Debts"
	BalanceLabelTaxLiabilities = "Tax Liabilities"
	BalanceLabelInvestorDebt   = "Investor Debt"
)

// Cash flow activity names
const (
	ActivityNameOperating = "Operating"
	ActivityNameInvesting = "Investing"
	ActivityNameFinancing = "Financing"
)

// Payment calendar item types
const (
	PaymentTypeReceivable = "Receivable"
	PaymentTypePayable    = "Payable"
	PaymentTypeTax        = "Tax"
	PaymentTypePlanned    = "Planned"
)

// ReportRequest represents a report request
type ReportRequest struct {
	CfoId     int64 `form:"cfoId,string"`
	StartTime int64 `form:"startTime" binding:"required,min=1"`
	EndTime   int64 `form:"endTime" binding:"required,min=1,gtfield=StartTime"`
}

// CashFlowActivityLine represents a line in cash flow report
type CashFlowActivityLine struct {
	CategoryId   int64  `json:"categoryId,string"`
	CategoryName string `json:"categoryName"`
	Income       int64  `json:"income"`
	Expense      int64  `json:"expense"`
	Net          int64  `json:"net"`
}

// CashFlowActivity represents an activity section in cash flow report
type CashFlowActivity struct {
	ActivityType int32                   `json:"activityType"`
	ActivityName string                  `json:"activityName"`
	Lines        []*CashFlowActivityLine `json:"lines"`
	TotalIncome  int64                   `json:"totalIncome"`
	TotalExpense int64                   `json:"totalExpense"`
	TotalNet     int64                   `json:"totalNet"`
}

// CashFlowResponse represents the cash flow report response
type CashFlowResponse struct {
	Activities []*CashFlowActivity `json:"activities"`
	TotalNet   int64               `json:"totalNet"`
}

// PnLLine represents a line in P&L report
type PnLLine struct {
	Label  string `json:"label"`
	Amount int64  `json:"amount"`
}

// PnLResponse represents the P&L report response
type PnLResponse struct {
	Revenue          int64      `json:"revenue"`
	CostOfGoods      int64      `json:"costOfGoods"`
	GrossProfit      int64      `json:"grossProfit"`
	OperatingExpense int64      `json:"operatingExpense"`
	Depreciation     int64      `json:"depreciation"`
	OperatingProfit  int64      `json:"operatingProfit"`
	FinancialExpense int64      `json:"financialExpense"`
	TaxExpense       int64      `json:"taxExpense"`
	NetProfit        int64      `json:"netProfit"`
	Details          []*PnLLine `json:"details"`
	Warnings         []string   `json:"warnings,omitempty"`
}

// BalanceSection represents a section in balance sheet
type BalanceLine struct {
	Label  string `json:"label"`
	Amount int64  `json:"amount"`
}

// BalanceResponse represents the balance sheet report response
type BalanceResponse struct {
	AssetLines      []*BalanceLine `json:"assetLines"`
	TotalAssets     int64          `json:"totalAssets"`
	LiabilityLines  []*BalanceLine `json:"liabilityLines"`
	TotalLiability  int64          `json:"totalLiability"`
	Equity          int64          `json:"equity"`
	Warnings        []string       `json:"warnings,omitempty"`
}

// PaymentCalendarItem represents a payment calendar entry
type PaymentCalendarItem struct {
	Date        int64  `json:"date"`
	Type        string `json:"type"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
	Currency    string `json:"currency"`
}

// PaymentCalendarResponse represents the payment calendar response
type PaymentCalendarResponse struct {
	Items    []*PaymentCalendarItem `json:"items"`
	Warnings []string               `json:"warnings,omitempty"`
}

// BalanceReportRequest represents a balance sheet report request (no time range needed)
type BalanceReportRequest struct {
	CfoId int64 `form:"cfoId,string"`
}
