// interfaces.go defines service layer interfaces for dependency injection.
package services

import (
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// TransactionReader provides read-only access to transactions
type TransactionReader interface {
	GetTotalTransactionCountByUid(c core.Context, uid int64) (int64, error)
	GetAllTransactions(c core.Context, uid int64, pageCount int32, noDuplicated bool) ([]*models.Transaction, error)
	GetAllTransactionsByMaxTime(c core.Context, uid int64, maxTransactionTime int64, count int32, noDuplicated bool) ([]*models.Transaction, error)
	GetAllSpecifiedTransactions(c core.Context, params *models.TransactionQueryParams, pageCount int32) ([]*models.Transaction, error)
	GetAllTransactionsInOneAccountWithAccountBalanceByMaxTime(c core.Context, uid int64, pageCount int32, maxTransactionTime int64, minTransactionTime int64, accountId int64, accountCategory models.AccountCategory) ([]*models.TransactionWithAccountBalance, *models.AccountBalanceResult, error)
	GetAllAccountsDailyOpeningAndClosingBalance(c core.Context, uid int64, maxTransactionTime int64, minTransactionTime int64, clientTimezone *time.Location) (map[int32][]*models.TransactionWithAccountBalance, error)
	GetTransactionsByMaxTime(c core.Context, params *models.TransactionQueryParams) ([]*models.Transaction, error)
	GetTransactionsInMonthByPage(c core.Context, uid int64, year int32, month int32, params *models.TransactionQueryParams) ([]*models.Transaction, error)
	GetTransactionByTransactionId(c core.Context, uid int64, transactionId int64) (*models.Transaction, error)
	GetAllTransactionCount(c core.Context, uid int64) (int64, error)
	GetTransactionCount(c core.Context, params *models.TransactionQueryParams) (int64, error)
	GetTransactionMapByList(transactions []*models.Transaction) map[int64]*models.Transaction
	GetTransactionIds(transactions []*models.Transaction) []int64
	GetRelatedTransferTransaction(originalTransaction *models.Transaction) *models.Transaction
}

// TransactionWriter provides write access to transactions
type TransactionWriter interface {
	CreateTransaction(c core.Context, transaction *models.Transaction, tagIds []int64, pictureIds []int64) error
	ModifyTransaction(c core.Context, transaction *models.Transaction, currentTagIdsCount int, addTagIds []int64, removeTagIds []int64, addPictureIds []int64, removePictureIds []int64) error
	DeleteTransaction(c core.Context, uid int64, transactionId int64) error
	DeleteAllTransactions(c core.Context, uid int64, deleteAccount bool) error
	DeleteAllTransactionsOfAccount(c core.Context, uid int64, accountId int64, pageCount int32) error
	MoveAllTransactionsBetweenAccounts(c core.Context, uid int64, fromAccountId int64, toAccountId int64) error
	BatchCreateTransactions(c core.Context, uid int64, transactions []*models.Transaction, allTagIds map[int][]int64, processHandler core.TaskProcessUpdateHandler) error
	SetTransactionPlanned(c core.Context, uid int64, transactionId int64, planned bool) error
	SetTransactionSourceTemplateId(c core.Context, uid int64, transactionId int64, templateId int64) error
	ConfirmPlannedTransaction(c core.Context, uid int64, transactionId int64, clientTimezone *time.Location) (*models.Transaction, error)
	ModifyAllFuturePlannedTransactions(c core.Context, uid int64, transactionId int64, modifyReq *models.TransactionModifyAllFutureRequest) (int64, error)
	DeleteAllFuturePlannedTransactions(c core.Context, uid int64, transactionId int64) (int64, error)
}

// TransactionStatisticsProvider provides transaction statistics operations
type TransactionStatisticsProvider interface {
	GetAccountsTotalIncomeAndExpense(c core.Context, uid int64, startUnixTime int64, endUnixTime int64, excludeAccountIds []int64, excludeCategoryIds []int64, clientTimezone *time.Location, useTransactionTimezone bool) (map[int64]int64, map[int64]int64, error)
	GetAccountsAndCategoriesTotalInflowAndOutflow(c core.Context, uid int64, startUnixTime int64, endUnixTime int64, tagFilters []*models.TransactionTagFilter, noTags bool, keyword string, clientTimezone *time.Location, useTransactionTimezone bool) ([]*models.Transaction, error)
	GetAccountsAndCategoriesMonthlyInflowAndOutflow(c core.Context, uid int64, startYear int32, startMonth int32, endYear int32, endMonth int32, tagFilters []*models.TransactionTagFilter, noTags bool, keyword string, clientTimezone *time.Location, useTransactionTimezone bool) (map[int32][]*models.Transaction, error)
}

// TransactionScheduler provides transaction scheduling operations
type TransactionScheduler interface {
	GeneratePlannedTransactions(c core.Context, baseTransaction *models.Transaction, tagIds []int64, frequencyType models.TransactionScheduleFrequencyType, frequency string, templateId int64) (int, error)
	CreateScheduledTransactions(c core.Context, currentUnixTime int64, interval time.Duration) error
}

// AccountReader provides read-only access to accounts
type AccountReader interface {
	GetTotalAccountCountByUid(c core.Context, uid int64) (int64, error)
	GetAllAccountsByUid(c core.Context, uid int64) ([]*models.Account, error)
	GetAccountByAccountId(c core.Context, uid int64, accountId int64) (*models.Account, error)
	GetAccountAndSubAccountsByAccountId(c core.Context, uid int64, accountId int64) ([]*models.Account, error)
	GetSubAccountsByAccountId(c core.Context, uid int64, accountId int64) ([]*models.Account, error)
	GetSubAccountsByAccountIds(c core.Context, uid int64, accountIds []int64) ([]*models.Account, error)
	GetAccountsByAccountIds(c core.Context, uid int64, accountIds []int64) (map[int64]*models.Account, error)
	GetMaxDisplayOrder(c core.Context, uid int64, category models.AccountCategory) (int32, error)
	GetMaxSubAccountDisplayOrder(c core.Context, uid int64, category models.AccountCategory, parentAccountId int64) (int32, error)
	GetAccountMapByList(accounts []*models.Account) map[int64]*models.Account
	GetVisibleAccountNameMapByList(accounts []*models.Account) map[string]*models.Account
	GetAccountNames(accounts []*models.Account) []string
	GetAccountOrSubAccountIds(c core.Context, accountIds string, uid int64) ([]int64, error)
	GetAccountOrSubAccountIdsByAccountName(accounts []*models.Account, accountName string) []int64
}

// AccountWriter provides write access to accounts
type AccountWriter interface {
	CreateAccounts(c core.Context, mainAccount *models.Account, mainAccountBalanceTime int64, childrenAccounts []*models.Account, childrenAccountBalanceTimes []int64, clientTimezone *time.Location) error
	ModifyAccounts(c core.Context, mainAccount *models.Account, updateAccounts []*models.Account, addSubAccounts []*models.Account, addSubAccountBalanceTimes []int64, removeSubAccountIds []int64, clientTimezone *time.Location) error
	HideAccount(c core.Context, uid int64, ids []int64, hidden bool) error
	ModifyAccountDisplayOrders(c core.Context, uid int64, accounts []*models.Account) error
	DeleteAccount(c core.Context, uid int64, accountId int64) error
	DeleteSubAccount(c core.Context, uid int64, accountId int64) error
}

// AssetProvider provides access to fixed assets
type AssetProvider interface {
	GetAllAssetsByUid(c core.Context, uid int64) ([]*models.Asset, error)
	GetAssetByAssetId(c core.Context, uid int64, assetId int64) (*models.Asset, error)
	GetMaxDisplayOrder(c core.Context, uid int64) (int32, error)
	CreateAsset(c core.Context, asset *models.Asset) error
	ModifyAsset(c core.Context, asset *models.Asset, nameChanged bool) error
	HideAsset(c core.Context, uid int64, ids []int64, hidden bool) error
	ModifyAssetDisplayOrders(c core.Context, uid int64, assets []*models.Asset) error
	DeleteAsset(c core.Context, uid int64, assetId int64) error
	ExistsAssetName(c core.Context, uid int64, name string) (bool, error)
}

// TaxRecordProvider provides access to tax records
type TaxRecordProvider interface {
	GetAllTaxRecordsByUid(c core.Context, uid int64) ([]*models.TaxRecord, error)
	GetTaxRecordByTaxId(c core.Context, uid int64, taxId int64) (*models.TaxRecord, error)
	CreateTaxRecord(c core.Context, record *models.TaxRecord) error
	ModifyTaxRecord(c core.Context, record *models.TaxRecord) error
	DeleteTaxRecord(c core.Context, uid int64, taxId int64) error
}

// InvestorDealProvider provides access to investor deals
type InvestorDealProvider interface {
	GetAllDealsByUid(c core.Context, uid int64) ([]*models.InvestorDeal, error)
	GetDealByDealId(c core.Context, uid int64, dealId int64) (*models.InvestorDeal, error)
	CreateDeal(c core.Context, deal *models.InvestorDeal) error
	ModifyDeal(c core.Context, deal *models.InvestorDeal) error
	DeleteDeal(c core.Context, uid int64, dealId int64) error
}

// InvestorPaymentProvider provides access to investor payments
type InvestorPaymentProvider interface {
	GetAllPaymentsByDealId(c core.Context, uid int64, dealId int64) ([]*models.InvestorPayment, error)
	GetAllPaymentsByDealIds(c core.Context, uid int64, dealIds []int64) (map[int64][]*models.InvestorPayment, error)
	GetAllPaymentsByUid(c core.Context, uid int64) ([]*models.InvestorPayment, error)
	GetPaymentByPaymentId(c core.Context, uid int64, paymentId int64) (*models.InvestorPayment, error)
	CreatePayment(c core.Context, payment *models.InvestorPayment) error
	ModifyPayment(c core.Context, payment *models.InvestorPayment) error
	DeletePayment(c core.Context, uid int64, paymentId int64) error
}

// ObligationProvider provides access to obligations
type ObligationProvider interface {
	GetAllObligationsByUid(c core.Context, uid int64) ([]*models.Obligation, error)
	GetObligationByObligationId(c core.Context, uid int64, obligationId int64) (*models.Obligation, error)
	CreateObligation(c core.Context, obligation *models.Obligation) error
	ModifyObligation(c core.Context, obligation *models.Obligation) error
	DeleteObligation(c core.Context, uid int64, obligationId int64) error
}

// CFOProvider provides access to CFO entities
type CFOProvider interface {
	GetAllCFOsByUid(c core.Context, uid int64) ([]*models.CFO, error)
	GetCFOByCFOId(c core.Context, uid int64, cfoId int64) (*models.CFO, error)
	GetMaxDisplayOrder(c core.Context, uid int64) (int32, error)
	CreateCFO(c core.Context, cfo *models.CFO) error
	ModifyCFO(c core.Context, cfo *models.CFO, nameChanged bool) error
	HideCFO(c core.Context, uid int64, ids []int64, hidden bool) error
	ModifyCFODisplayOrders(c core.Context, uid int64, cfos []*models.CFO) error
	DeleteCFO(c core.Context, uid int64, cfoId int64) error
	ExistsCFOName(c core.Context, uid int64, name string) (bool, error)
}

// BudgetProvider provides access to budgets
type BudgetProvider interface {
	GetBudgetsByYearMonth(c core.Context, uid int64, year int32, month int32, cfoId int64) ([]*models.Budget, error)
	SaveBudgets(c core.Context, uid int64, year int32, month int32, cfoId int64, items []*models.BudgetItemRequest) error
	GetFactAmountsByYearMonth(c core.Context, uid int64, startTime int64, endTime int64, cfoId int64) (map[int64]int64, error)
}

// ReportProvider provides access to financial reports
type ReportProvider interface {
	GetCashFlow(c core.Context, uid int64, cfoId int64, startTime int64, endTime int64) (*models.CashFlowResponse, error)
	GetPnL(c core.Context, uid int64, cfoId int64, startTime int64, endTime int64) (*models.PnLResponse, error)
	GetBalance(c core.Context, uid int64, cfoId int64) (*models.BalanceResponse, error)
	GetPaymentCalendar(c core.Context, uid int64, startTime int64, endTime int64) (*models.PaymentCalendarResponse, error)
}

// Compile-time interface compliance checks
var (
	_ TransactionReader             = (*TransactionService)(nil)
	_ TransactionWriter             = (*TransactionService)(nil)
	_ TransactionStatisticsProvider = (*TransactionService)(nil)
	_ TransactionScheduler          = (*TransactionService)(nil)
	_ AccountReader                 = (*AccountService)(nil)
	_ AccountWriter                 = (*AccountService)(nil)
	_ AssetProvider                 = (*AssetService)(nil)
	_ TaxRecordProvider             = (*TaxRecordService)(nil)
	_ InvestorDealProvider          = (*InvestorDealService)(nil)
	_ InvestorPaymentProvider       = (*InvestorPaymentService)(nil)
	_ ObligationProvider            = (*ObligationService)(nil)
	_ CFOProvider                   = (*CFOService)(nil)
	_ BudgetProvider                = (*BudgetService)(nil)
	_ ReportProvider                = (*ReportService)(nil)
)
