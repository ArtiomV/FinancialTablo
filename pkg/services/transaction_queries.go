package services

import (
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// GetTotalTransactionCountByUid returns total transaction count of user
func (s *TransactionService) GetTotalTransactionCountByUid(c core.Context, uid int64) (int64, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	count, err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).Count(&models.Transaction{})

	return count, err
}

// GetAllTransactions returns all transactions
func (s *TransactionService) GetAllTransactions(c core.Context, uid int64, pageCount int32, noDuplicated bool) ([]*models.Transaction, error) {
	maxTransactionTime := utils.GetMaxTransactionTimeFromUnixTime(time.Now().Unix())
	var allTransactions []*models.Transaction

	for maxTransactionTime > 0 {
		transactions, err := s.GetAllTransactionsByMaxTime(c, uid, maxTransactionTime, pageCount, noDuplicated)

		if err != nil {
			return nil, err
		}

		allTransactions = append(allTransactions, transactions...)

		if len(transactions) < int(pageCount) {
			maxTransactionTime = 0
			break
		}

		maxTransactionTime = transactions[len(transactions)-1].TransactionTime - 1
	}

	return allTransactions, nil
}

// GetAllTransactionsByMaxTime returns all transactions before given time
func (s *TransactionService) GetAllTransactionsByMaxTime(c core.Context, uid int64, maxTransactionTime int64, count int32, noDuplicated bool) ([]*models.Transaction, error) {
	return s.GetTransactionsByMaxTime(c, uid, maxTransactionTime, 0, 0, nil, nil, nil, false, "", "", 1, count, false, noDuplicated)
}

// GetAllSpecifiedTransactions returns all transactions that match given conditions
func (s *TransactionService) GetAllSpecifiedTransactions(c core.Context, uid int64, maxTransactionTime int64, minTransactionTime int64, transactionType models.TransactionType, categoryIds []int64, accountIds []int64, tagFilters []*models.TransactionTagFilter, noTags bool, amountFilter string, keyword string, pageCount int32, noDuplicated bool) ([]*models.Transaction, error) {
	if maxTransactionTime <= 0 {
		maxTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(time.Now().Unix())
	}

	var allTransactions []*models.Transaction

	for maxTransactionTime > 0 {
		transactions, err := s.GetTransactionsByMaxTime(c, uid, maxTransactionTime, minTransactionTime, transactionType, categoryIds, accountIds, tagFilters, noTags, amountFilter, keyword, 1, pageCount, false, noDuplicated)

		if err != nil {
			return nil, err
		}

		allTransactions = append(allTransactions, transactions...)

		if len(transactions) < int(pageCount) {
			maxTransactionTime = 0
			break
		}

		maxTransactionTime = transactions[len(transactions)-1].TransactionTime - 1
	}

	return allTransactions, nil
}

// GetAllTransactionsInOneAccountWithAccountBalanceByMaxTime returns account statement within time range
func (s *TransactionService) GetAllTransactionsInOneAccountWithAccountBalanceByMaxTime(c core.Context, uid int64, pageCount int32, maxTransactionTime int64, minTransactionTime int64, accountId int64, accountCategory models.AccountCategory) ([]*models.TransactionWithAccountBalance, int64, int64, int64, int64, error) {
	if maxTransactionTime <= 0 {
		maxTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(time.Now().Unix())
	}

	var allTransactions []*models.Transaction

	for maxTransactionTime > 0 {
		transactions, err := s.GetTransactionsByMaxTime(c, uid, maxTransactionTime, 0, 0, nil, []int64{accountId}, nil, false, "", "", 1, pageCount, false, true)

		if err != nil {
			return nil, 0, 0, 0, 0, err
		}

		allTransactions = append(allTransactions, transactions...)

		if len(transactions) < int(pageCount) {
			maxTransactionTime = 0
			break
		}

		maxTransactionTime = transactions[len(transactions)-1].TransactionTime - 1
	}

	allTransactionsAndAccountBalance := make([]*models.TransactionWithAccountBalance, 0, len(allTransactions))

	if len(allTransactions) < 1 {
		return allTransactionsAndAccountBalance, 0, 0, 0, 0, nil
	}

	totalInflows := int64(0)
	totalOutflows := int64(0)
	openingBalance := int64(0)
	accumulatedBalance := int64(0)
	lastAccumulatedBalance := int64(0)

	for i := len(allTransactions) - 1; i >= 0; i-- {
		transaction := allTransactions[i]

		switch transaction.Type {
		case models.TRANSACTION_DB_TYPE_MODIFY_BALANCE:
			accumulatedBalance = accumulatedBalance + transaction.RelatedAccountAmount
		case models.TRANSACTION_DB_TYPE_INCOME:
			accumulatedBalance = accumulatedBalance + transaction.Amount
		case models.TRANSACTION_DB_TYPE_EXPENSE:
			accumulatedBalance = accumulatedBalance - transaction.Amount
		case models.TRANSACTION_DB_TYPE_TRANSFER_OUT:
			accumulatedBalance = accumulatedBalance - transaction.Amount
		case models.TRANSACTION_DB_TYPE_TRANSFER_IN:
			accumulatedBalance = accumulatedBalance + transaction.Amount
		default:
			log.Errorf(c, "[transactions.GetAllTransactionsInOneAccountWithAccountBalanceByMaxTime] transaction type (%d) is invalid (id:%d)", transaction.TransactionId, transaction.Type)
			return nil, 0, 0, 0, 0, errs.ErrTransactionTypeInvalid
		}

		if transaction.TransactionTime < minTransactionTime {
			openingBalance = accumulatedBalance
			lastAccumulatedBalance = accumulatedBalance
			continue
		}

		switch transaction.Type {
		case models.TRANSACTION_DB_TYPE_MODIFY_BALANCE:
			if accountCategory.IsAsset() {
				totalInflows = totalInflows + transaction.RelatedAccountAmount
			} else if accountCategory.IsLiability() {
				totalOutflows = totalOutflows - transaction.RelatedAccountAmount
			}
		case models.TRANSACTION_DB_TYPE_INCOME:
			totalInflows = totalInflows + transaction.Amount
		case models.TRANSACTION_DB_TYPE_EXPENSE:
			totalOutflows = totalOutflows + transaction.Amount
		case models.TRANSACTION_DB_TYPE_TRANSFER_OUT:
			totalOutflows = totalOutflows + transaction.Amount
		case models.TRANSACTION_DB_TYPE_TRANSFER_IN:
			totalInflows = totalInflows + transaction.Amount
		}

		transactionsAndAccountBalance := &models.TransactionWithAccountBalance{
			Transaction:           transaction,
			AccountOpeningBalance: lastAccumulatedBalance,
			AccountClosingBalance: accumulatedBalance,
		}

		lastAccumulatedBalance = accumulatedBalance
		allTransactionsAndAccountBalance = append(allTransactionsAndAccountBalance, transactionsAndAccountBalance)
	}

	return allTransactionsAndAccountBalance, totalInflows, totalOutflows, openingBalance, accumulatedBalance, nil
}

// GetAllAccountsDailyOpeningAndClosingBalance returns daily opening and closing balance of all accounts within time range
func (s *TransactionService) GetAllAccountsDailyOpeningAndClosingBalance(c core.Context, uid int64, maxTransactionTime int64, minTransactionTime int64, clientTimezone *time.Location) (map[int32][]*models.TransactionWithAccountBalance, error) {
	if maxTransactionTime <= 0 {
		maxTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(time.Now().Unix())
	}

	var allTransactions []*models.Transaction

	for maxTransactionTime > 0 {
		transactions, err := s.GetTransactionsByMaxTime(c, uid, maxTransactionTime, 0, 0, nil, nil, nil, false, "", "", 1, pageCountForLoadTransactionAmounts, false, false)

		if err != nil {
			return nil, err
		}

		allTransactions = append(allTransactions, transactions...)

		if len(transactions) < pageCountForLoadTransactionAmounts {
			maxTransactionTime = 0
			break
		}

		maxTransactionTime = transactions[len(transactions)-1].TransactionTime - 1
	}

	accountDailyLastBalances := make(map[dayAccountKey]*models.TransactionWithAccountBalance)
	accountDailyBalances := make(map[int32][]*models.TransactionWithAccountBalance)

	if len(allTransactions) < 1 {
		return accountDailyBalances, nil
	}

	accumulatedBalances := make(map[int64]int64)
	accumulatedBalancesBeforeStartTime := make(map[int64]int64)

	for i := len(allTransactions) - 1; i >= 0; i-- {
		transaction := allTransactions[i]
		accumulatedBalance := accumulatedBalances[transaction.AccountId]
		lastAccumulatedBalance := accumulatedBalances[transaction.AccountId]

		switch transaction.Type {
		case models.TRANSACTION_DB_TYPE_MODIFY_BALANCE:
			accumulatedBalance = accumulatedBalance + transaction.RelatedAccountAmount
		case models.TRANSACTION_DB_TYPE_INCOME:
			accumulatedBalance = accumulatedBalance + transaction.Amount
		case models.TRANSACTION_DB_TYPE_EXPENSE:
			accumulatedBalance = accumulatedBalance - transaction.Amount
		case models.TRANSACTION_DB_TYPE_TRANSFER_OUT:
			accumulatedBalance = accumulatedBalance - transaction.Amount
		case models.TRANSACTION_DB_TYPE_TRANSFER_IN:
			accumulatedBalance = accumulatedBalance + transaction.Amount
		default:
			log.Errorf(c, "[transactions.GetAllTransactionsWithAccountBalanceByMaxTime] transaction type (%d) is invalid (id:%d)", transaction.TransactionId, transaction.Type)
			return nil, errs.ErrTransactionTypeInvalid
		}

		accumulatedBalances[transaction.AccountId] = accumulatedBalance

		if transaction.TransactionTime < minTransactionTime {
			accumulatedBalancesBeforeStartTime[transaction.AccountId] = accumulatedBalance
			continue
		}

		yearMonthDay := utils.FormatUnixTimeToNumericYearMonthDay(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime), clientTimezone)
		key := dayAccountKey{YearMonthDay: yearMonthDay, AccountId: transaction.AccountId}
		dailyAccountBalance, exists := accountDailyLastBalances[key]

		if exists {
			dailyAccountBalance.AccountClosingBalance = accumulatedBalance
		} else {
			dailyAccountBalance = &models.TransactionWithAccountBalance{
				Transaction: &models.Transaction{
					AccountId: transaction.AccountId,
				},
				AccountOpeningBalance: lastAccumulatedBalance,
				AccountClosingBalance: accumulatedBalance,
			}
			accountDailyLastBalances[key] = dailyAccountBalance
		}
	}

	firstTransactionTime := allTransactions[len(allTransactions)-1].TransactionTime

	if minTransactionTime > firstTransactionTime {
		firstTransactionTime = minTransactionTime
	}

	firstYearMonthDay := utils.FormatUnixTimeToNumericYearMonthDay(utils.GetUnixTimeFromTransactionTime(firstTransactionTime), clientTimezone)

	// fill in the opening balance for accounts that do not have transactions on the first day
	for accountId, accumulatedBalance := range accumulatedBalancesBeforeStartTime {
		if accumulatedBalance == 0 {
			continue
		}

		key := dayAccountKey{YearMonthDay: firstYearMonthDay, AccountId: accountId}

		if _, exists := accountDailyLastBalances[key]; exists {
			continue
		}

		accountDailyLastBalances[key] = &models.TransactionWithAccountBalance{
			Transaction: &models.Transaction{
				AccountId: accountId,
			},
			AccountOpeningBalance: accumulatedBalance,
			AccountClosingBalance: accumulatedBalance,
		}
	}

	for key, transactionWithAccountBalance := range accountDailyLastBalances {
		dailyAccountBalances, exists := accountDailyBalances[key.YearMonthDay]

		if !exists {
			dailyAccountBalances = make([]*models.TransactionWithAccountBalance, 0)
		}

		dailyAccountBalances = append(dailyAccountBalances, transactionWithAccountBalance)
		accountDailyBalances[key.YearMonthDay] = dailyAccountBalances
	}

	return accountDailyBalances, nil
}

// GetTransactionsByMaxTime returns transactions before given time
func (s *TransactionService) GetTransactionsByMaxTime(c core.Context, uid int64, maxTransactionTime int64, minTransactionTime int64, transactionType models.TransactionType, categoryIds []int64, accountIds []int64, tagFilters []*models.TransactionTagFilter, noTags bool, amountFilter string, keyword string, page int32, count int32, needOneMoreItem bool, noDuplicated bool) ([]*models.Transaction, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var err error
	var transactionDbType models.TransactionDbType = 0

	if transactionType > 0 {
		transactionDbType, err = transactionType.ToTransactionDbType()

		if err != nil {
			return nil, err
		}
	}

	if page < 0 {
		return nil, errs.ErrPageIndexInvalid
	} else if page == 0 {
		page = 1
	}

	if count < 1 {
		return nil, errs.ErrPageCountInvalid
	}

	var transactions []*models.Transaction

	actualCount := count

	if needOneMoreItem {
		actualCount++
	}

	condition, conditionParams := s.buildTransactionQueryCondition(uid, maxTransactionTime, minTransactionTime, transactionDbType, categoryIds, accountIds, tagFilters, amountFilter, keyword, noDuplicated)
	sess := s.UserDataDB(uid).NewSession(c).Where(condition, conditionParams...)
	sess = s.appendFilterTagIdsConditionToQuery(sess, uid, maxTransactionTime, minTransactionTime, tagFilters, noTags)

	err = sess.Limit(int(actualCount), int(count*(page-1))).OrderBy("transaction_time desc").Find(&transactions)

	return transactions, err
}

// GetTransactionsInMonthByPage returns all transactions in given year and month
func (s *TransactionService) GetTransactionsInMonthByPage(c core.Context, uid int64, year int32, month int32, transactionType models.TransactionType, categoryIds []int64, accountIds []int64, tagFilters []*models.TransactionTagFilter, noTags bool, amountFilter string, keyword string) ([]*models.Transaction, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var err error
	var transactionDbType models.TransactionDbType = 0

	if transactionType > 0 {
		transactionDbType, err = transactionType.ToTransactionDbType()

		if err != nil {
			return nil, err
		}
	}

	minTransactionTime, maxTransactionTime, err := utils.GetTransactionTimeRangeByYearMonth(year, month)

	if err != nil {
		return nil, errs.ErrSystemError
	}

	var transactions []*models.Transaction

	condition, conditionParams := s.buildTransactionQueryCondition(uid, maxTransactionTime, minTransactionTime, transactionDbType, categoryIds, accountIds, tagFilters, amountFilter, keyword, true)
	sess := s.UserDataDB(uid).NewSession(c).Where(condition, conditionParams...)
	sess = s.appendFilterTagIdsConditionToQuery(sess, uid, maxTransactionTime, minTransactionTime, tagFilters, noTags)

	err = sess.OrderBy("transaction_time desc").Find(&transactions)

	transactionsInMonth := make([]*models.Transaction, 0, len(transactions))

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]
		transactionUnixTime := utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime)
		transactionTimeZone := time.FixedZone("Transaction Timezone", int(transaction.TimezoneUtcOffset)*60)

		if utils.IsUnixTimeEqualsYearAndMonth(transactionUnixTime, transactionTimeZone, year, month) {
			transactionsInMonth = append(transactionsInMonth, transaction)
		}
	}

	return transactionsInMonth, err
}

// GetTransactionByTransactionId returns a transaction model according to transaction id
func (s *TransactionService) GetTransactionByTransactionId(c core.Context, uid int64, transactionId int64) (*models.Transaction, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if transactionId <= 0 {
		return nil, errs.ErrTransactionIdInvalid
	}

	transaction := &models.Transaction{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(transactionId).Where("uid=? AND deleted=?", uid, false).Get(transaction)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrTransactionNotFound
	}

	return transaction, nil
}

// GetAllTransactionCount returns total count of transactions
func (s *TransactionService) GetAllTransactionCount(c core.Context, uid int64) (int64, error) {
	return s.GetTransactionCount(c, uid, 0, 0, 0, nil, nil, nil, false, "", "")
}

// GetTransactionCount returns count of transactions
func (s *TransactionService) GetTransactionCount(c core.Context, uid int64, maxTransactionTime int64, minTransactionTime int64, transactionType models.TransactionType, categoryIds []int64, accountIds []int64, tagFilters []*models.TransactionTagFilter, noTags bool, amountFilter string, keyword string) (int64, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	var err error
	var transactionDbType models.TransactionDbType = 0

	if transactionType > 0 {
		transactionDbType, err = transactionType.ToTransactionDbType()

		if err != nil {
			return 0, err
		}
	}

	condition, conditionParams := s.buildTransactionQueryCondition(uid, maxTransactionTime, minTransactionTime, transactionDbType, categoryIds, accountIds, tagFilters, amountFilter, keyword, true)
	sess := s.UserDataDB(uid).NewSession(c).Where(condition, conditionParams...)
	sess = s.appendFilterTagIdsConditionToQuery(sess, uid, maxTransactionTime, minTransactionTime, tagFilters, noTags)

	return sess.Count(&models.Transaction{})
}

// GetTransactionMapByList returns a transaction map by a list
func (s *TransactionService) GetTransactionMapByList(transactions []*models.Transaction) map[int64]*models.Transaction {
	transactionMap := make(map[int64]*models.Transaction)

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]
		transactionMap[transaction.TransactionId] = transaction
	}

	return transactionMap
}

// GetTransactionIds returns transaction ids list
func (s *TransactionService) GetTransactionIds(transactions []*models.Transaction) []int64 {
	transactionIds := make([]int64, len(transactions))

	for i := 0; i < len(transactions); i++ {
		transactionIds[i] = transactions[i].TransactionId
	}

	return transactionIds
}
