// transaction_statistics.go provides aggregated transaction statistics and trends.
package services

import (
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// GetAccountsTotalIncomeAndExpense returns the every accounts total income and expense amount by specific date range
func (s *TransactionService) GetAccountsTotalIncomeAndExpense(c core.Context, uid int64, startUnixTime int64, endUnixTime int64, excludeAccountIds []int64, excludeCategoryIds []int64, clientTimezone *time.Location, useTransactionTimezone bool) (map[int64]int64, map[int64]int64, error) {
	if uid <= 0 {
		return nil, nil, errs.ErrUserIdInvalid
	}

	startLocalDateTime := utils.FormatUnixTimeToNumericLocalDateTime(startUnixTime, clientTimezone)
	endLocalDateTime := utils.FormatUnixTimeToNumericLocalDateTime(endUnixTime, clientTimezone)

	startUnixTime = utils.GetMinUnixTimeWithSameLocalDateTime(startUnixTime, utils.GetTimezoneOffsetMinutes(startUnixTime, clientTimezone))
	endUnixTime = utils.GetMaxUnixTimeWithSameLocalDateTime(endUnixTime, utils.GetTimezoneOffsetMinutes(endUnixTime, clientTimezone))

	startTransactionTime := utils.GetMinTransactionTimeFromUnixTime(startUnixTime)
	endTransactionTime := utils.GetMaxTransactionTimeFromUnixTime(endUnixTime)

	condition := "uid=? AND deleted=? AND (type=? OR type=?)"
	conditionParams := make([]any, 0, 4+len(excludeAccountIds)+len(excludeCategoryIds))
	conditionParams = append(conditionParams, uid)
	conditionParams = append(conditionParams, false)
	conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_INCOME)
	conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_EXPENSE)

	if len(excludeAccountIds) > 0 {
		var accountIdsCondition strings.Builder
		accountIdConditionParams := make([]any, 0, len(excludeAccountIds))

		for i := 0; i < len(excludeAccountIds); i++ {
			if i > 0 {
				accountIdsCondition.WriteString(",")
			}

			accountIdsCondition.WriteString("?")
			accountIdConditionParams = append(accountIdConditionParams, excludeAccountIds[i])
		}

		condition = condition + " AND account_id NOT IN (" + accountIdsCondition.String() + ")"
		conditionParams = append(conditionParams, accountIdConditionParams...)
	}

	if len(excludeCategoryIds) > 0 {
		var categoryIdsCondition strings.Builder
		categoryIdConditionParams := make([]any, 0, len(excludeCategoryIds))

		for i := 0; i < len(excludeCategoryIds); i++ {
			if i > 0 {
				categoryIdsCondition.WriteString(",")
			}

			categoryIdsCondition.WriteString("?")
			categoryIdConditionParams = append(categoryIdConditionParams, excludeCategoryIds[i])
		}

		condition = condition + " AND category_id NOT IN (" + categoryIdsCondition.String() + ")"
		conditionParams = append(conditionParams, categoryIdConditionParams...)
	}

	condition = condition + " AND transaction_time>=? AND transaction_time<=?"

	minTransactionTime := startTransactionTime
	maxTransactionTime := endTransactionTime
	var allTransactions []*models.Transaction

	for maxTransactionTime > 0 {
		var transactions []*models.Transaction

		finalConditionParams := make([]any, 0, 6)
		finalConditionParams = append(finalConditionParams, conditionParams...)
		finalConditionParams = append(finalConditionParams, minTransactionTime)
		finalConditionParams = append(finalConditionParams, maxTransactionTime)

		err := s.UserDataDB(uid).NewSession(c).Select("type, account_id, transaction_time, timezone_utc_offset, amount").Where(condition, finalConditionParams...).Limit(pageCountForLoadTransactionAmounts, 0).OrderBy("transaction_time desc").Find(&transactions)

		if err != nil {
			return nil, nil, err
		}

		allTransactions = append(allTransactions, transactions...)

		if len(transactions) < pageCountForLoadTransactionAmounts {
			maxTransactionTime = 0
			break
		}

		maxTransactionTime = transactions[len(transactions)-1].TransactionTime - 1
	}

	incomeAmounts := make(map[int64]int64)
	expenseAmounts := make(map[int64]int64)

	for i := 0; i < len(allTransactions); i++ {
		transaction := allTransactions[i]
		timeZone := clientTimezone

		if useTransactionTimezone {
			timeZone = time.FixedZone("Transaction Timezone", int(transaction.TimezoneUtcOffset)*60)
		}

		localDateTime := utils.FormatUnixTimeToNumericLocalDateTime(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime), timeZone)

		if localDateTime < startLocalDateTime || localDateTime > endLocalDateTime {
			continue
		}

		var amountsMap map[int64]int64

		switch transaction.Type {
		case models.TRANSACTION_DB_TYPE_INCOME:
			amountsMap = incomeAmounts
		case models.TRANSACTION_DB_TYPE_EXPENSE:
			amountsMap = expenseAmounts
		}

		totalAmounts, exists := amountsMap[transaction.AccountId]

		if !exists {
			totalAmounts = 0
		}

		totalAmounts += transaction.Amount
		amountsMap[transaction.AccountId] = totalAmounts
	}

	return incomeAmounts, expenseAmounts, nil
}

// GetAccountsAndCategoriesTotalInflowAndOutflow returns the every accounts and categories total inflows and outflows amount by specific date range
func (s *TransactionService) GetAccountsAndCategoriesTotalInflowAndOutflow(c core.Context, uid int64, startUnixTime int64, endUnixTime int64, tagFilters []*models.TransactionTagFilter, noTags bool, keyword string, clientTimezone *time.Location, useTransactionTimezone bool) ([]*models.Transaction, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var startLocalDateTime, endLocalDateTime, startTransactionTime, endTransactionTime int64

	if startUnixTime > 0 {
		startLocalDateTime = utils.FormatUnixTimeToNumericLocalDateTime(startUnixTime, clientTimezone)
		startUnixTime = utils.GetMinUnixTimeWithSameLocalDateTime(startUnixTime, utils.GetTimezoneOffsetMinutes(startUnixTime, clientTimezone))
		startTransactionTime = utils.GetMinTransactionTimeFromUnixTime(startUnixTime)
	}

	if endUnixTime > 0 {
		endLocalDateTime = utils.FormatUnixTimeToNumericLocalDateTime(endUnixTime, clientTimezone)
		endUnixTime = utils.GetMaxUnixTimeWithSameLocalDateTime(endUnixTime, utils.GetTimezoneOffsetMinutes(endUnixTime, clientTimezone))
		endTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(endUnixTime)
	}

	condition := "uid=? AND deleted=? AND (type=? OR type=? OR type=? OR type=?)"
	conditionParams := make([]any, 0, 6)
	conditionParams = append(conditionParams, uid)
	conditionParams = append(conditionParams, false)
	conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_INCOME)
	conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_EXPENSE)
	conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_OUT)
	conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_IN)

	minTransactionTime := startTransactionTime
	maxTransactionTime := endTransactionTime
	var allTransactions []*models.Transaction

	for maxTransactionTime >= 0 {
		var transactions []*models.Transaction

		finalCondition := condition
		finalConditionParams := make([]any, 0, 6)
		finalConditionParams = append(finalConditionParams, conditionParams...)

		if minTransactionTime > 0 {
			finalCondition = finalCondition + " AND transaction_time>=?"
			finalConditionParams = append(finalConditionParams, minTransactionTime)
		}

		if maxTransactionTime > 0 {
			finalCondition = finalCondition + " AND transaction_time<=?"
			finalConditionParams = append(finalConditionParams, maxTransactionTime)
		}

		if keyword != "" {
			finalCondition = finalCondition + " AND comment LIKE ?"
			finalConditionParams = append(finalConditionParams, "%%"+keyword+"%%")
		}

		sess := s.UserDataDB(uid).NewSession(c).Select("type, category_id, account_id, related_account_id, transaction_time, timezone_utc_offset, amount").Where(finalCondition, finalConditionParams...)
		sess = s.appendFilterTagIdsConditionToQuery(sess, uid, maxTransactionTime, minTransactionTime, tagFilters, noTags)

		err := sess.Limit(pageCountForLoadTransactionAmounts, 0).OrderBy("transaction_time desc").Find(&transactions)

		if err != nil {
			return nil, err
		}

		allTransactions = append(allTransactions, transactions...)

		if len(transactions) < pageCountForLoadTransactionAmounts {
			maxTransactionTime = -1
			break
		}

		maxTransactionTime = transactions[len(transactions)-1].TransactionTime - 1
	}

	transactionTotalAmountsMap := make(map[monthCategoryAccountKey]*models.Transaction)

	for i := 0; i < len(allTransactions); i++ {
		transaction := allTransactions[i]
		timeZone := clientTimezone

		if useTransactionTimezone {
			timeZone = time.FixedZone("Transaction Timezone", int(transaction.TimezoneUtcOffset)*60)
		}

		localDateTime := utils.FormatUnixTimeToNumericLocalDateTime(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime), timeZone)

		if (startLocalDateTime > 0 && localDateTime < startLocalDateTime) || (endLocalDateTime > 0 && localDateTime > endLocalDateTime) {
			continue
		}

		key := monthCategoryAccountKey{
			CategoryId: transaction.CategoryId,
			AccountId:  transaction.AccountId,
		}

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			key.RelatedAccountId = transaction.RelatedAccountId
			key.Type = transaction.Type
		}

		totalAmounts, exists := transactionTotalAmountsMap[key]

		if !exists {
			totalAmounts = &models.Transaction{
				Type:             transaction.Type,
				CategoryId:       transaction.CategoryId,
				AccountId:        transaction.AccountId,
				RelatedAccountId: transaction.RelatedAccountId,
				Amount:           0,
			}

			transactionTotalAmountsMap[key] = totalAmounts
		}

		totalAmounts.Amount += transaction.Amount
	}

	transactionTotalAmounts := make([]*models.Transaction, 0, len(transactionTotalAmountsMap))

	for _, totalAmounts := range transactionTotalAmountsMap {
		transactionTotalAmounts = append(transactionTotalAmounts, totalAmounts)
	}

	return transactionTotalAmounts, nil
}

// GetAccountsAndCategoriesMonthlyInflowAndOutflow returns the every accounts monthly inflows and outflows amount by specific date range
func (s *TransactionService) GetAccountsAndCategoriesMonthlyInflowAndOutflow(c core.Context, uid int64, startYear int32, startMonth int32, endYear int32, endMonth int32, tagFilters []*models.TransactionTagFilter, noTags bool, keyword string, clientTimezone *time.Location, useTransactionTimezone bool) (map[int32][]*models.Transaction, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var startTransactionTime, endTransactionTime int64
	var err error

	if startYear > 0 && startMonth > 0 {
		startTransactionTime, _, err = utils.GetTransactionTimeRangeByYearMonth(startYear, startMonth)

		if err != nil {
			return nil, errs.ErrSystemError
		}
	}

	if endYear > 0 && endMonth > 0 {
		_, endTransactionTime, err = utils.GetTransactionTimeRangeByYearMonth(endYear, endMonth)

		if err != nil {
			return nil, errs.ErrSystemError
		}
	}

	condition := "uid=? AND deleted=? AND (type=? OR type=? OR type=? OR type=?)"
	conditionParams := make([]any, 0, 6)
	conditionParams = append(conditionParams, uid)
	conditionParams = append(conditionParams, false)
	conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_INCOME)
	conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_EXPENSE)
	conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_OUT)
	conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_IN)

	minTransactionTime := startTransactionTime
	maxTransactionTime := endTransactionTime
	var allTransactions []*models.Transaction

	for maxTransactionTime >= 0 {
		var transactions []*models.Transaction

		finalCondition := condition
		finalConditionParams := make([]any, 0, 6)
		finalConditionParams = append(finalConditionParams, conditionParams...)

		if minTransactionTime > 0 {
			finalCondition = finalCondition + " AND transaction_time>=?"
			finalConditionParams = append(finalConditionParams, minTransactionTime)
		}

		if maxTransactionTime > 0 {
			finalCondition = finalCondition + " AND transaction_time<=?"
			finalConditionParams = append(finalConditionParams, maxTransactionTime)
		}

		if keyword != "" {
			finalCondition = finalCondition + " AND comment LIKE ?"
			finalConditionParams = append(finalConditionParams, "%%"+keyword+"%%")
		}

		sess := s.UserDataDB(uid).NewSession(c).Select("type, category_id, account_id, related_account_id, transaction_time, timezone_utc_offset, amount").Where(finalCondition, finalConditionParams...)
		sess = s.appendFilterTagIdsConditionToQuery(sess, uid, maxTransactionTime, minTransactionTime, tagFilters, noTags)

		err := sess.Limit(pageCountForLoadTransactionAmounts, 0).OrderBy("transaction_time desc").Find(&transactions)

		if err != nil {
			return nil, err
		}

		allTransactions = append(allTransactions, transactions...)

		if len(transactions) < pageCountForLoadTransactionAmounts {
			maxTransactionTime = -1
			break
		}

		maxTransactionTime = transactions[len(transactions)-1].TransactionTime - 1
	}

	startYearMonth := startYear*100 + startMonth
	endYearMonth := endYear*100 + endMonth
	transactionsMonthlyAmountsMap := make(map[monthCategoryAccountKey]*models.Transaction)
	transactionsMonthlyAmounts := make(map[int32][]*models.Transaction)

	for i := 0; i < len(allTransactions); i++ {
		transaction := allTransactions[i]
		timeZone := clientTimezone

		if useTransactionTimezone {
			timeZone = time.FixedZone("Transaction Timezone", int(transaction.TimezoneUtcOffset)*60)
		}

		yearMonth := utils.FormatUnixTimeToNumericYearMonth(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime), timeZone)

		if (startYearMonth > 0 && yearMonth < startYearMonth) || (endYearMonth > 0 && yearMonth > endYearMonth) {
			continue
		}

		key := monthCategoryAccountKey{
			YearMonth:  yearMonth,
			CategoryId: transaction.CategoryId,
			AccountId:  transaction.AccountId,
		}

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			key.RelatedAccountId = transaction.RelatedAccountId
			key.Type = transaction.Type
		}

		transactionAmounts, exists := transactionsMonthlyAmountsMap[key]

		if !exists {
			transactionAmounts = &models.Transaction{
				Type:             transaction.Type,
				CategoryId:       transaction.CategoryId,
				AccountId:        transaction.AccountId,
				RelatedAccountId: transaction.RelatedAccountId,
				Amount:           0,
			}

			transactionsMonthlyAmountsMap[key] = transactionAmounts
		}

		transactionAmounts.Amount += transaction.Amount
	}

	for key, transaction := range transactionsMonthlyAmountsMap {
		monthlyAmounts, exists := transactionsMonthlyAmounts[key.YearMonth]

		if !exists {
			monthlyAmounts = make([]*models.Transaction, 0, 0)
		}

		monthlyAmounts = append(monthlyAmounts, transaction)
		transactionsMonthlyAmounts[key.YearMonth] = monthlyAmounts
	}

	return transactionsMonthlyAmounts, nil
}
