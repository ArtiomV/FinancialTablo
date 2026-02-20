package api

import (
	"math"
	"sort"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// TransactionCountHandler returns transaction total count of current user
func (a *TransactionsApi) TransactionCountHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionCountReq models.TransactionCountRequest
	err := c.ShouldBindQuery(&transactionCountReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionCountHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	allAccountIds, err := a.accounts.GetAccountOrSubAccountIds(c, transactionCountReq.AccountIds, uid)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionCountHandler] get account error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allCategoryIds, err := a.transactionCategories.GetCategoryOrSubCategoryIds(c, transactionCountReq.CategoryIds, uid)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionCountHandler] get transaction category error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	noTags := transactionCountReq.TagFilter == models.TransactionNoTagFilterValue
	var tagFilters []*models.TransactionTagFilter

	if !noTags {
		tagFilters, err = models.ParseTransactionTagFilter(transactionCountReq.TagFilter)

		if err != nil {
			log.Warnf(c, "[transactions.TransactionCountHandler] parse transaction filters error, because %s", err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	totalCount, err := a.transactions.GetTransactionCount(c, &models.TransactionQueryParams{
		Uid:                uid,
		MaxTransactionTime: transactionCountReq.MaxTime,
		MinTransactionTime: transactionCountReq.MinTime,
		TransactionType:    transactionCountReq.Type,
		CategoryIds:        allCategoryIds,
		AccountIds:         allAccountIds,
		TagFilters:         tagFilters,
		NoTags:             noTags,
		AmountFilter:       transactionCountReq.AmountFilter,
		Keyword:            transactionCountReq.Keyword,
		CounterpartyId:     transactionCountReq.CounterpartyId,
	})

	if err != nil {
		log.Errorf(c, "[transactions.TransactionCountHandler] failed to get transaction count for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	countResp := &models.TransactionCountResponse{
		TotalCount: totalCount,
	}

	return countResp, nil
}

// TransactionListHandler returns transaction list of current user
func (a *TransactionsApi) TransactionListHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionListReq models.TransactionListByMaxTimeRequest
	err := c.ShouldBindQuery(&transactionListReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionListHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionListHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionListHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	allAccountIds, err := a.accounts.GetAccountOrSubAccountIds(c, transactionListReq.AccountIds, uid)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionListHandler] get account error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allCategoryIds, err := a.transactionCategories.GetCategoryOrSubCategoryIds(c, transactionListReq.CategoryIds, uid)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionListHandler] get transaction category error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	noTags := transactionListReq.TagFilter == models.TransactionNoTagFilterValue
	var tagFilters []*models.TransactionTagFilter

	if !noTags {
		tagFilters, err = models.ParseTransactionTagFilter(transactionListReq.TagFilter)

		if err != nil {
			log.Warnf(c, "[transactions.TransactionListHandler] parse transaction tag filters error, because %s", err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	var totalCount int64

	if transactionListReq.WithCount {
		totalCount, err = a.transactions.GetTransactionCount(c, &models.TransactionQueryParams{
			Uid:                uid,
			MaxTransactionTime: transactionListReq.MaxTime,
			MinTransactionTime: transactionListReq.MinTime,
			TransactionType:    transactionListReq.Type,
			CategoryIds:        allCategoryIds,
			AccountIds:         allAccountIds,
			TagFilters:         tagFilters,
			NoTags:             noTags,
			AmountFilter:       transactionListReq.AmountFilter,
			Keyword:            transactionListReq.Keyword,
			CounterpartyId:     transactionListReq.CounterpartyId,
		})

		if err != nil {
			log.Errorf(c, "[transactions.TransactionListHandler] failed to get transaction count for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	transactions, err := a.transactions.GetTransactionsByMaxTime(c, &models.TransactionQueryParams{
		Uid:                uid,
		MaxTransactionTime: transactionListReq.MaxTime,
		MinTransactionTime: transactionListReq.MinTime,
		TransactionType:    transactionListReq.Type,
		CategoryIds:        allCategoryIds,
		AccountIds:         allAccountIds,
		TagFilters:         tagFilters,
		NoTags:             noTags,
		AmountFilter:       transactionListReq.AmountFilter,
		Keyword:            transactionListReq.Keyword,
		CounterpartyId:     transactionListReq.CounterpartyId,
		Page:               transactionListReq.Page,
		Count:              transactionListReq.Count,
		NeedOneMoreItem:    true,
		NoDuplicated:       true,
	})

	if err != nil {
		log.Errorf(c, "[transactions.TransactionListHandler] failed to get transactions earlier than \"%d\" for user \"uid:%d\", because %s", transactionListReq.MaxTime, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	hasMore := false
	var nextTimeSequenceId *int64

	if len(transactions) > int(transactionListReq.Count) {
		hasMore = true
		nextTimeSequenceId = &transactions[transactionListReq.Count].TransactionTime
		transactions = transactions[:transactionListReq.Count]
	}

	transactionResult, err := a.getTransactionResponseListResult(c, user, transactions, clientTimezone, transactionListReq.WithPictures, transactionListReq.TrimAccount, transactionListReq.TrimCategory, transactionListReq.TrimTag)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionListHandler] failed to assemble transaction result for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactionResps := &models.TransactionInfoPageWrapperResponse{
		Items: transactionResult,
	}

	if hasMore {
		transactionResps.NextTimeSequenceId = nextTimeSequenceId
	}

	if transactionListReq.WithCount {
		transactionResps.TotalCount = &totalCount
	}

	return transactionResps, nil
}

// TransactionMonthListHandler returns all transaction list of current user by month
func (a *TransactionsApi) TransactionMonthListHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionListReq models.TransactionListInMonthByPageRequest
	err := c.ShouldBindQuery(&transactionListReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionMonthListHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionMonthListHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionMonthListHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	allAccountIds, err := a.accounts.GetAccountOrSubAccountIds(c, transactionListReq.AccountIds, uid)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionMonthListHandler] get account error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allCategoryIds, err := a.transactionCategories.GetCategoryOrSubCategoryIds(c, transactionListReq.CategoryIds, uid)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionMonthListHandler] get transaction category error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	noTags := transactionListReq.TagFilter == models.TransactionNoTagFilterValue
	var tagFilters []*models.TransactionTagFilter

	if !noTags {
		tagFilters, err = models.ParseTransactionTagFilter(transactionListReq.TagFilter)

		if err != nil {
			log.Warnf(c, "[transactions.TransactionMonthListHandler] parse transaction tag filters error, because %s", err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	transactions, err := a.transactions.GetTransactionsInMonthByPage(c, uid, transactionListReq.Year, transactionListReq.Month, &models.TransactionQueryParams{
		Uid:             uid,
		TransactionType: transactionListReq.Type,
		CategoryIds:     allCategoryIds,
		AccountIds:      allAccountIds,
		TagFilters:      tagFilters,
		NoTags:          noTags,
		AmountFilter:    transactionListReq.AmountFilter,
		Keyword:         transactionListReq.Keyword,
		CounterpartyId:     transactionListReq.CounterpartyId,
	})

	if err != nil {
		log.Errorf(c, "[transactions.TransactionMonthListHandler] failed to get transactions in month \"%d-%d\" for user \"uid:%d\", because %s", transactionListReq.Year, transactionListReq.Month, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactionResult, err := a.getTransactionResponseListResult(c, user, transactions, clientTimezone, transactionListReq.WithPictures, transactionListReq.TrimAccount, transactionListReq.TrimCategory, transactionListReq.TrimTag)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionMonthListHandler] failed to assemble transaction result for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactionResps := &models.TransactionInfoPageWrapperResponse2{
		Items:      transactionResult,
		TotalCount: int64(transactionResult.Len()),
	}

	return transactionResps, nil
}

// TransactionListAllHandler returns all transaction list of current user
func (a *TransactionsApi) TransactionListAllHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionAllListReq models.TransactionAllListRequest
	err := c.ShouldBindQuery(&transactionAllListReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionListAllHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionListAllHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionListAllHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	allAccountIds, err := a.accounts.GetAccountOrSubAccountIds(c, transactionAllListReq.AccountIds, uid)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionListAllHandler] get account error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allCategoryIds, err := a.transactionCategories.GetCategoryOrSubCategoryIds(c, transactionAllListReq.CategoryIds, uid)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionListAllHandler] get transaction category error, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	noTags := transactionAllListReq.TagFilter == models.TransactionNoTagFilterValue
	var tagFilters []*models.TransactionTagFilter

	if !noTags {
		tagFilters, err = models.ParseTransactionTagFilter(transactionAllListReq.TagFilter)

		if err != nil {
			log.Warnf(c, "[transactions.TransactionListAllHandler] parse transaction tag filters error, because %s", err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	maxTransactionTime := int64(math.MaxInt64)
	minTransactionTime := int64(0)

	if transactionAllListReq.EndTime > 0 {
		maxTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(transactionAllListReq.EndTime)
	}

	if transactionAllListReq.StartTime > 0 {
		minTransactionTime = utils.GetMinTransactionTimeFromUnixTime(transactionAllListReq.StartTime)
	}

	allTransactions, err := a.transactions.GetAllSpecifiedTransactions(c, &models.TransactionQueryParams{
		Uid:                uid,
		MaxTransactionTime: maxTransactionTime,
		MinTransactionTime: minTransactionTime,
		TransactionType:    transactionAllListReq.Type,
		CategoryIds:        allCategoryIds,
		AccountIds:         allAccountIds,
		TagFilters:         tagFilters,
		NoTags:             noTags,
		AmountFilter:       transactionAllListReq.AmountFilter,
		Keyword:            transactionAllListReq.Keyword,
		CounterpartyId:     transactionAllListReq.CounterpartyId,
		NoDuplicated:       true,
	}, pageCountForDataExport)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionListAllHandler] failed to get all transactions for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactionResult, err := a.getTransactionResponseListResult(c, user, allTransactions, clientTimezone, transactionAllListReq.WithPictures, transactionAllListReq.TrimAccount, transactionAllListReq.TrimCategory, transactionAllListReq.TrimTag)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionListAllHandler] failed to assemble transaction result for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return transactionResult, nil
}

// TransactionReconciliationStatementHandler returns transaction reconciliation statement list of current user
func (a *TransactionsApi) TransactionReconciliationStatementHandler(c *core.WebContext) (any, *errs.Error) {
	var reconciliationStatementRequest models.TransactionReconciliationStatementRequest
	err := c.ShouldBindQuery(&reconciliationStatementRequest)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionReconciliationStatementHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionReconciliationStatementHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionReconciliationStatementHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	account, err := a.accounts.GetAccountByAccountId(c, uid, reconciliationStatementRequest.AccountId)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionReconciliationStatementHandler] failed to get account \"id:%d\" for user \"uid:%d\", because %s", reconciliationStatementRequest.AccountId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if account.Type != models.ACCOUNT_TYPE_SINGLE_ACCOUNT {
		log.Errorf(c, "[transactions.TransactionReconciliationStatementHandler] account \"id:%d\" for user \"uid:%d\" is not a single account", reconciliationStatementRequest.AccountId, uid)
		return nil, errs.ErrAccountTypeInvalid
	}

	maxTransactionTime := int64(0)

	if reconciliationStatementRequest.EndTime > 0 {
		maxTransactionTime = utils.GetMaxTransactionTimeFromUnixTime(reconciliationStatementRequest.EndTime)
	}

	minTransactionTime := int64(0)

	if reconciliationStatementRequest.StartTime > 0 {
		minTransactionTime = utils.GetMinTransactionTimeFromUnixTime(reconciliationStatementRequest.StartTime)
	}

	transactionsWithAccountBalance, balanceResult, err := a.transactions.GetAllTransactionsInOneAccountWithAccountBalanceByMaxTime(c, uid, pageCountForAccountStatement, maxTransactionTime, minTransactionTime, reconciliationStatementRequest.AccountId, account.Category)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionReconciliationStatementHandler] failed to get transactions from \"%d\" to \"%d\" for user \"uid:%d\", because %s", reconciliationStatementRequest.StartTime, reconciliationStatementRequest.EndTime, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactions := make([]*models.Transaction, len(transactionsWithAccountBalance))
	transactionAccountBalanceMap := make(map[int64]*models.TransactionWithAccountBalance, len(transactionsWithAccountBalance))

	for i := 0; i < len(transactionsWithAccountBalance); i++ {
		transactionWithBalance := transactionsWithAccountBalance[i]
		transactions[i] = transactionWithBalance.Transaction
		transactionAccountBalanceMap[transactionWithBalance.TransactionId] = transactionWithBalance
		transactionAccountBalanceMap[transactionWithBalance.RelatedId] = transactionWithBalance
	}

	transactionResult, err := a.getTransactionResponseListResult(c, user, transactions, clientTimezone, false, true, true, true)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionReconciliationStatementHandler] failed to assemble transaction result for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	responseItems := make([]*models.TransactionReconciliationStatementResponseItem, len(transactionResult))

	for i := 0; i < len(transactionResult); i++ {
		transactionResult := transactionResult[i]
		accountOpeningBalance := int64(0)
		accountClosingBalance := int64(0)

		if transactionWithBalance, exists := transactionAccountBalanceMap[transactionResult.Id]; exists {
			accountOpeningBalance = transactionWithBalance.AccountOpeningBalance
			accountClosingBalance = transactionWithBalance.AccountClosingBalance
		} else {
			log.Warnf(c, "[transactions.TransactionReconciliationStatementHandler] missing account balance for transaction \"id:%d\" of user \"uid:%d\"", transactionResult.Id, uid)
		}

		responseItems[i] = &models.TransactionReconciliationStatementResponseItem{
			TransactionInfoResponse: transactionResult,
			AccountOpeningBalance:   accountOpeningBalance,
			AccountClosingBalance:   accountClosingBalance,
		}
	}

	reconciliationStatementResp := &models.TransactionReconciliationStatementResponse{
		Transactions:   responseItems,
		TotalInflows:   balanceResult.TotalInflows,
		TotalOutflows:  balanceResult.TotalOutflows,
		OpeningBalance: balanceResult.OpeningBalance,
		ClosingBalance: balanceResult.ClosingBalance,
	}

	return reconciliationStatementResp, nil
}

// TransactionGetHandler returns one specific transaction of current user
func (a *TransactionsApi) TransactionGetHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionGetReq models.TransactionGetRequest
	err := c.ShouldBindQuery(&transactionGetReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionGetHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionGetHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	transaction, err := a.transactions.GetTransactionByTransactionId(c, uid, transactionGetReq.Id)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionGetHandler] failed to get transaction \"id:%d\" for user \"uid:%d\", because %s", transactionGetReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		transaction = a.transactions.GetRelatedTransferTransaction(transaction)
	}

	accountIds := make([]int64, 0, 2)
	accountIds = append(accountIds, transaction.AccountId)

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
		accountIds = append(accountIds, transaction.RelatedAccountId)
		accountIds = utils.ToUniqueInt64Slice(accountIds)
	}

	accountMap, err := a.accounts.GetAccountsByAccountIds(c, uid, accountIds)

	if _, exists := accountMap[transaction.AccountId]; !exists {
		log.Warnf(c, "[transactions.TransactionGetHandler] account of transaction \"id:%d\" does not exist for user \"uid:%d\"", transaction.TransactionId, uid)
		return nil, errs.ErrTransactionNotFound
	}

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
		if _, exists := accountMap[transaction.RelatedAccountId]; !exists {
			log.Warnf(c, "[transactions.TransactionGetHandler] related account of transaction \"id:%d\" does not exist for user \"uid:%d\"", transaction.TransactionId, uid)
			return nil, errs.ErrTransactionNotFound
		}
	}

	allTransactionTagIds, err := a.transactionTags.GetAllTagIdsOfTransactions(c, uid, []int64{transaction.TransactionId})

	if err != nil {
		log.Errorf(c, "[transactions.TransactionGetHandler] failed to get transactions tag ids for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	var category *models.TransactionCategory
	var tagMap map[int64]*models.TransactionTag
	var pictureInfos []*models.TransactionPictureInfo

	if !transactionGetReq.TrimCategory {
		category, err = a.transactionCategories.GetCategoryByCategoryId(c, uid, transaction.CategoryId)

		if err != nil {
			log.Errorf(c, "[transactions.TransactionGetHandler] failed to get transactions category for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	if !transactionGetReq.TrimTag {
		tagMap, err = a.transactionTags.GetTagsByTagIds(c, uid, utils.ToUniqueInt64Slice(a.transactionTags.GetTransactionTagIds(allTransactionTagIds)))

		if err != nil {
			log.Errorf(c, "[transactions.TransactionGetHandler] failed to get transactions tags for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	if transactionGetReq.WithPictures && a.CurrentConfig().EnableTransactionPictures {
		pictureInfos, err = a.transactionPictures.GetPictureInfosByTransactionId(c, uid, transaction.TransactionId)

		if err != nil {
			log.Errorf(c, "[transactions.TransactionGetHandler] failed to get transactions pictures for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}
	}

	transactionEditable := transaction.IsEditable(user, clientTimezone, accountMap[transaction.AccountId], accountMap[transaction.RelatedAccountId])
	transactionTagIds := allTransactionTagIds[transaction.TransactionId]
	transactionResp := transaction.ToTransactionInfoResponse(transactionTagIds, transactionEditable)

	if !transactionGetReq.TrimAccount {
		if sourceAccount := accountMap[transaction.AccountId]; sourceAccount != nil {
			transactionResp.SourceAccount = sourceAccount.ToAccountInfoResponse()
		}

		if destinationAccount := accountMap[transaction.RelatedAccountId]; destinationAccount != nil {
			transactionResp.DestinationAccount = destinationAccount.ToAccountInfoResponse()
		}
	}

	if !transactionGetReq.TrimCategory {
		if category != nil {
			transactionResp.Category = category.ToTransactionCategoryInfoResponse()
		}
	}

	if !transactionGetReq.TrimTag {
		transactionResp.Tags = a.getTransactionTagInfoResponses(transactionTagIds, tagMap)
	}

	if transactionGetReq.WithPictures && a.CurrentConfig().EnableTransactionPictures {
		transactionResp.Pictures = a.GetTransactionPictureInfoResponseList(pictureInfos)
	}

	// Load splits for this transaction
	splits, splitErr := a.transactionSplits.GetSplitsByTransactionId(c, uid, transaction.TransactionId)
	if splitErr != nil {
		log.Warnf(c, "[transactions.TransactionGetHandler] failed to get splits for transaction \"id:%d\" for user \"uid:%d\", because %s", transaction.TransactionId, uid, splitErr.Error())
	} else if len(splits) > 0 {
		splitResponses := make([]models.TransactionSplitResponse, len(splits))
		for i, split := range splits {
			splitResponses[i] = models.TransactionSplitResponse{
				CategoryId: split.CategoryId,
				Amount:     split.Amount,
				
				TagIds:     split.GetTagIdStringSlice(),
			}
		}
		transactionResp.Splits = splitResponses
	}

	return transactionResp, nil
}

func (a *TransactionsApi) filterTransactions(c *core.WebContext, uid int64, transactions []*models.Transaction, accountMap map[int64]*models.Account) []*models.Transaction {
	finalTransactions := make([]*models.Transaction, 0, len(transactions))

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]

		if _, exists := accountMap[transaction.AccountId]; !exists {
			log.Warnf(c, "[transactions.filterTransactions] account of transaction \"id:%d\" does not exist for user \"uid:%d\"", transaction.TransactionId, uid)
			continue
		}

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			if _, exists := accountMap[transaction.RelatedAccountId]; !exists {
				log.Warnf(c, "[transactions.filterTransactions] related account of transaction \"id:%d\" does not exist for user \"uid:%d\"", transaction.TransactionId, uid)
				continue
			}
		}

		finalTransactions = append(finalTransactions, transaction)
	}

	return finalTransactions
}

func (a *TransactionsApi) getTransactionTagInfoResponses(tagIds []int64, allTransactionTags map[int64]*models.TransactionTag) []*models.TransactionTagInfoResponse {
	allTags := make([]*models.TransactionTagInfoResponse, 0, len(tagIds))

	for i := 0; i < len(tagIds); i++ {
		tag := allTransactionTags[tagIds[i]]

		if tag == nil {
			continue
		}

		allTags = append(allTags, tag.ToTransactionTagInfoResponse())
	}

	return allTags
}

func (a *TransactionsApi) getTransactionResponseListResult(c *core.WebContext, user *models.User, transactions []*models.Transaction, clientTimezone *time.Location, withPictures bool, trimAccount bool, trimCategory bool, trimTag bool) (models.TransactionInfoResponseSlice, error) {
	uid := user.Uid
	transactionIds := make([]int64, len(transactions))
	accountIds := make([]int64, 0, len(transactions)*2)
	categoryIds := make([]int64, 0, len(transactions))

	for i := 0; i < len(transactions); i++ {
		transactionId := transactions[i].TransactionId

		if transactions[i].Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			transactionId = transactions[i].RelatedId
		}

		transactionIds[i] = transactionId
		accountIds = append(accountIds, transactions[i].AccountId)

		if transactions[i].Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN || transactions[i].Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			accountIds = append(accountIds, transactions[i].RelatedAccountId)
		}

		categoryIds = append(categoryIds, transactions[i].CategoryId)
	}

	allAccounts, err := a.accounts.GetAccountsByAccountIds(c, uid, utils.ToUniqueInt64Slice(accountIds))

	if err != nil {
		log.Errorf(c, "[transactions.getTransactionResponseListResult] failed to get accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, err
	}

	transactions = a.filterTransactions(c, uid, transactions, allAccounts)

	allTransactionTagIds, err := a.transactionTags.GetAllTagIdsOfTransactions(c, uid, transactionIds)

	if err != nil {
		log.Errorf(c, "[transactions.getTransactionResponseListResult] failed to get transactions tag ids for user \"uid:%d\", because %s", uid, err.Error())
		return nil, err
	}

	var categoryMap map[int64]*models.TransactionCategory
	var tagMap map[int64]*models.TransactionTag
	var pictureInfoMap map[int64][]*models.TransactionPictureInfo
	var splitMap map[int64][]*models.TransactionSplit

	if !trimCategory {
		categoryMap, err = a.transactionCategories.GetCategoriesByCategoryIds(c, uid, utils.ToUniqueInt64Slice(categoryIds))

		if err != nil {
			log.Errorf(c, "[transactions.getTransactionResponseListResult] failed to get transactions categories for user \"uid:%d\", because %s", uid, err.Error())
			return nil, err
		}
	}

	if !trimTag {
		tagMap, err = a.transactionTags.GetTagsByTagIds(c, uid, utils.ToUniqueInt64Slice(a.transactionTags.GetTransactionTagIds(allTransactionTagIds)))

		if err != nil {
			log.Errorf(c, "[transactions.getTransactionResponseListResult] failed to get transactions tags for user \"uid:%d\", because %s", uid, err.Error())
			return nil, err
		}
	}

	if withPictures && a.CurrentConfig().EnableTransactionPictures {
		pictureInfoMap, err = a.transactionPictures.GetPictureInfosByTransactionIds(c, uid, utils.ToUniqueInt64Slice(a.transactions.GetTransactionIds(transactions)))

		if err != nil {
			log.Errorf(c, "[transactions.getTransactionResponseListResult] failed to get transactions pictures for user \"uid:%d\", because %s", uid, err.Error())
			return nil, err
		}
	}

	// Load splits for all transactions in batch
	splitMap, err = a.transactionSplits.GetSplitsByTransactionIds(c, uid, utils.ToUniqueInt64Slice(a.transactions.GetTransactionIds(transactions)))
	if err != nil {
		log.Warnf(c, "[transactions.getTransactionResponseListResult] failed to get transaction splits for user \"uid:%d\", because %s", uid, err.Error())
		// Non-fatal: continue without splits
		splitMap = nil
	}

	result := make(models.TransactionInfoResponseSlice, len(transactions))

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			transaction = a.transactions.GetRelatedTransferTransaction(transaction)
		}

		transactionEditable := transaction.IsEditable(user, clientTimezone, allAccounts[transaction.AccountId], allAccounts[transaction.RelatedAccountId])
		transactionTagIds := allTransactionTagIds[transaction.TransactionId]
		result[i] = transaction.ToTransactionInfoResponse(transactionTagIds, transactionEditable)

		if !trimAccount {
			if sourceAccount := allAccounts[transaction.AccountId]; sourceAccount != nil {
				result[i].SourceAccount = sourceAccount.ToAccountInfoResponse()
			}

			if destinationAccount := allAccounts[transaction.RelatedAccountId]; destinationAccount != nil {
				result[i].DestinationAccount = destinationAccount.ToAccountInfoResponse()
			}
		}

		if !trimCategory {
			if category := categoryMap[transaction.CategoryId]; category != nil {
				result[i].Category = category.ToTransactionCategoryInfoResponse()
			}
		}

		if !trimTag {
			result[i].Tags = a.getTransactionTagInfoResponses(transactionTagIds, tagMap)
		}

		if withPictures && a.CurrentConfig().EnableTransactionPictures {
			pictureInfos, exists := pictureInfoMap[transaction.TransactionId]

			if exists {
				result[i].Pictures = a.GetTransactionPictureInfoResponseList(pictureInfos)
			}
		}

		// Attach splits if any
		if splitMap != nil {
			if splits, exists := splitMap[transaction.TransactionId]; exists && len(splits) > 0 {
				splitResponses := make([]models.TransactionSplitResponse, len(splits))
				for j, split := range splits {
					splitResponses[j] = models.TransactionSplitResponse{
						CategoryId: split.CategoryId,
						Amount:     split.Amount,
						TagIds:     split.GetTagIdStringSlice(),
					}
				}
				result[i].Splits = splitResponses
			}
		}
	}

	sort.Sort(result)

	return result, nil
}
