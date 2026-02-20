// transaction_queries.go implements transaction listing, filtering, and pagination.
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
	return s.fetchAllTransactionPages(c, &models.TransactionQueryParams{
		Uid:                uid,
		MaxTransactionTime: utils.GetMaxTransactionTimeFromUnixTime(time.Now().Unix()),
		NoDuplicated:       noDuplicated,
	}, pageCount)
}

// GetAllTransactionsByMaxTime returns all transactions before given time
func (s *TransactionService) GetAllTransactionsByMaxTime(c core.Context, uid int64, maxTransactionTime int64, count int32, noDuplicated bool) ([]*models.Transaction, error) {
	return s.GetTransactionsByMaxTime(c, &models.TransactionQueryParams{
		Uid:                uid,
		MaxTransactionTime: maxTransactionTime,
		Page:               1,
		Count:              count,
		NoDuplicated:       noDuplicated,
	})
}

// GetAllSpecifiedTransactions returns all transactions that match given conditions
func (s *TransactionService) GetAllSpecifiedTransactions(c core.Context, params *models.TransactionQueryParams, pageCount int32) ([]*models.Transaction, error) {
	if params.MaxTransactionTime <= 0 {
		fetchParams := *params
		fetchParams.MaxTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(time.Now().Unix())
		return s.fetchAllTransactionPages(c, &fetchParams, pageCount)
	}

	return s.fetchAllTransactionPages(c, params, pageCount)
}

// GetAllTransactionsInOneAccountWithAccountBalanceByMaxTime returns account statement within time range
func (s *TransactionService) GetAllTransactionsInOneAccountWithAccountBalanceByMaxTime(c core.Context, uid int64, pageCount int32, maxTransactionTime int64, minTransactionTime int64, accountId int64, accountCategory models.AccountCategory) ([]*models.TransactionWithAccountBalance, *models.AccountBalanceResult, error) {
	if maxTransactionTime <= 0 {
		maxTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(time.Now().Unix())
	}

	allTransactions, err := s.fetchAllTransactionPages(c, &models.TransactionQueryParams{
		Uid:                uid,
		MaxTransactionTime: maxTransactionTime,
		AccountIds:         []int64{accountId},
		NoDuplicated:       true,
	}, pageCount)

	if err != nil {
		return nil, nil, err
	}

	allTransactionsAndAccountBalance := make([]*models.TransactionWithAccountBalance, 0, len(allTransactions))
	result := &models.AccountBalanceResult{}

	if len(allTransactions) < 1 {
		return allTransactionsAndAccountBalance, result, nil
	}

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
			return nil, nil, errs.ErrTransactionTypeInvalid
		}

		if transaction.TransactionTime < minTransactionTime {
			result.OpeningBalance = accumulatedBalance
			lastAccumulatedBalance = accumulatedBalance
			continue
		}

		switch transaction.Type {
		case models.TRANSACTION_DB_TYPE_MODIFY_BALANCE:
			if accountCategory.IsAsset() {
				result.TotalInflows = result.TotalInflows + transaction.RelatedAccountAmount
			} else if accountCategory.IsLiability() {
				result.TotalOutflows = result.TotalOutflows - transaction.RelatedAccountAmount
			}
		case models.TRANSACTION_DB_TYPE_INCOME:
			result.TotalInflows = result.TotalInflows + transaction.Amount
		case models.TRANSACTION_DB_TYPE_EXPENSE:
			result.TotalOutflows = result.TotalOutflows + transaction.Amount
		case models.TRANSACTION_DB_TYPE_TRANSFER_OUT:
			result.TotalOutflows = result.TotalOutflows + transaction.Amount
		case models.TRANSACTION_DB_TYPE_TRANSFER_IN:
			result.TotalInflows = result.TotalInflows + transaction.Amount
		}

		transactionsAndAccountBalance := &models.TransactionWithAccountBalance{
			Transaction:           transaction,
			AccountOpeningBalance: lastAccumulatedBalance,
			AccountClosingBalance: accumulatedBalance,
		}

		lastAccumulatedBalance = accumulatedBalance
		allTransactionsAndAccountBalance = append(allTransactionsAndAccountBalance, transactionsAndAccountBalance)
	}

	result.ClosingBalance = accumulatedBalance

	return allTransactionsAndAccountBalance, result, nil
}

// GetAllAccountsDailyOpeningAndClosingBalance returns daily opening and closing balance of all accounts within time range
func (s *TransactionService) GetAllAccountsDailyOpeningAndClosingBalance(c core.Context, uid int64, maxTransactionTime int64, minTransactionTime int64, clientTimezone *time.Location) (map[int32][]*models.TransactionWithAccountBalance, error) {
	if maxTransactionTime <= 0 {
		maxTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(time.Now().Unix())
	}

	allTransactions, err := s.fetchAllTransactionPages(c, &models.TransactionQueryParams{
		Uid:                uid,
		MaxTransactionTime: maxTransactionTime,
	}, int32(pageCountForLoadTransactionAmounts))

	if err != nil {
		return nil, err
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
func (s *TransactionService) GetTransactionsByMaxTime(c core.Context, params *models.TransactionQueryParams) ([]*models.Transaction, error) {
	if params.Uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	page := params.Page
	count := params.Count

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

	if params.NeedOneMoreItem {
		actualCount++
	}

	cond := s.buildTransactionQueryCondition(params)
	sess := s.UserDataDB(params.Uid).NewSession(c).Where(cond)
	sess = s.appendFilterTagIdsConditionToQuery(sess, params.Uid, params.MaxTransactionTime, params.MinTransactionTime, params.TagFilters, params.NoTags)

	err := sess.Limit(int(actualCount), int(count*(page-1))).OrderBy("transaction_time desc").Find(&transactions)

	return transactions, err
}

// GetTransactionsInMonthByPage returns all transactions in given year and month
func (s *TransactionService) GetTransactionsInMonthByPage(c core.Context, uid int64, year int32, month int32, params *models.TransactionQueryParams) ([]*models.Transaction, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	minTransactionTime, maxTransactionTime, err := utils.GetTransactionTimeRangeByYearMonth(year, month)

	if err != nil {
		return nil, errs.ErrSystemError
	}

	// Create a copy of params with month-specific time range and noDuplicated=true
	monthParams := &models.TransactionQueryParams{
		Uid:                uid,
		MaxTransactionTime: maxTransactionTime,
		MinTransactionTime: minTransactionTime,
		TransactionType:    params.TransactionType,
		CategoryIds:        params.CategoryIds,
		AccountIds:         params.AccountIds,
		TagFilters:         params.TagFilters,
		NoTags:             params.NoTags,
		AmountFilter:       params.AmountFilter,
		Keyword:            params.Keyword,
		CounterpartyId:     params.CounterpartyId,
		NoDuplicated:       true,
	}

	var transactions []*models.Transaction

	cond := s.buildTransactionQueryCondition(monthParams)
	sess := s.UserDataDB(uid).NewSession(c).Where(cond)
	sess = s.appendFilterTagIdsConditionToQuery(sess, uid, maxTransactionTime, minTransactionTime, params.TagFilters, params.NoTags)

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
	return s.GetTransactionCount(c, &models.TransactionQueryParams{
		Uid: uid,
	})
}

// GetTransactionCount returns count of transactions
func (s *TransactionService) GetTransactionCount(c core.Context, params *models.TransactionQueryParams) (int64, error) {
	if params.Uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	// GetTransactionCount always uses noDuplicated=true
	countParams := &models.TransactionQueryParams{
		Uid:                params.Uid,
		MaxTransactionTime: params.MaxTransactionTime,
		MinTransactionTime: params.MinTransactionTime,
		TransactionType:    params.TransactionType,
		CategoryIds:        params.CategoryIds,
		AccountIds:         params.AccountIds,
		TagFilters:         params.TagFilters,
		NoTags:             params.NoTags,
		AmountFilter:       params.AmountFilter,
		Keyword:            params.Keyword,
		CounterpartyId:     params.CounterpartyId,
		NoDuplicated:       true,
	}

	cond := s.buildTransactionQueryCondition(countParams)
	sess := s.UserDataDB(params.Uid).NewSession(c).Where(cond)
	sess = s.appendFilterTagIdsConditionToQuery(sess, params.Uid, params.MaxTransactionTime, params.MinTransactionTime, params.TagFilters, params.NoTags)

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
