// transaction_helpers.go contains pure helper functions for transaction processing:
// account ID validation, related column mapping, and query condition building.
package services

import (
	"fmt"
	"strings"

	"xorm.io/builder"
	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// fetchAllTransactionPages fetches all transactions by iterating through pages using the given query params.
// It updates MaxTransactionTime on each iteration to paginate through all results.
func (s *TransactionService) fetchAllTransactionPages(c core.Context, params *models.TransactionQueryParams, pageCount int32) ([]*models.Transaction, error) {
	maxTransactionTime := params.MaxTransactionTime
	var allTransactions []*models.Transaction

	for maxTransactionTime > 0 {
		pageParams := *params
		pageParams.MaxTransactionTime = maxTransactionTime
		pageParams.Page = 1
		pageParams.Count = pageCount

		transactions, err := s.GetTransactionsByMaxTime(c, &pageParams)

		if err != nil {
			return nil, err
		}

		allTransactions = append(allTransactions, transactions...)

		if len(transactions) < int(pageCount) {
			break
		}

		maxTransactionTime = transactions[len(transactions)-1].TransactionTime - 1
	}

	return allTransactions, nil
}

// GetRelatedTransferTransaction returns the related transaction for transfer transaction
func (s *TransactionService) GetRelatedTransferTransaction(originalTransaction *models.Transaction) *models.Transaction {
	var relatedType models.TransactionDbType
	var relatedTransactionTime int64

	switch originalTransaction.Type {
	case models.TRANSACTION_DB_TYPE_TRANSFER_OUT:
		relatedType = models.TRANSACTION_DB_TYPE_TRANSFER_IN
		relatedTransactionTime = originalTransaction.TransactionTime + 1
	case models.TRANSACTION_DB_TYPE_TRANSFER_IN:
		relatedType = models.TRANSACTION_DB_TYPE_TRANSFER_OUT
		relatedTransactionTime = originalTransaction.TransactionTime - 1
	default:
		return nil
	}

	relatedTransaction := &models.Transaction{
		TransactionId:        originalTransaction.RelatedId,
		Uid:                  originalTransaction.Uid,
		Deleted:              originalTransaction.Deleted,
		Type:                 relatedType,
		CategoryId:           originalTransaction.CategoryId,
		TransactionTime:      relatedTransactionTime,
		TimezoneUtcOffset:    originalTransaction.TimezoneUtcOffset,
		AccountId:            originalTransaction.RelatedAccountId,
		Amount:               originalTransaction.RelatedAccountAmount,
		RelatedId:            originalTransaction.TransactionId,
		RelatedAccountId:     originalTransaction.AccountId,
		RelatedAccountAmount: originalTransaction.Amount,
		Comment:              originalTransaction.Comment,
		GeoLongitude:         originalTransaction.GeoLongitude,
		GeoLatitude:          originalTransaction.GeoLatitude,
		CreatedIp:            originalTransaction.CreatedIp,
		CreatedUnixTime:      originalTransaction.CreatedUnixTime,
		UpdatedUnixTime:      originalTransaction.UpdatedUnixTime,
		DeletedUnixTime:      originalTransaction.DeletedUnixTime,
	}

	return relatedTransaction
}

func (s *TransactionService) buildTransactionQueryCondition(params *models.TransactionQueryParams) builder.Cond {
	uid := params.Uid
	maxTransactionTime := params.MaxTransactionTime
	minTransactionTime := params.MinTransactionTime
	categoryIds := params.CategoryIds
	accountIds := params.AccountIds
	amountFilter := params.AmountFilter
	keyword := params.Keyword
	noDuplicated := params.NoDuplicated

	var transactionDbType models.TransactionDbType = 0

	if params.TransactionType > 0 {
		var err error
		transactionDbType, err = params.TransactionType.ToTransactionDbType()

		if err != nil {
			transactionDbType = 0
		}
	}

	cond := builder.And(builder.Eq{"uid": uid}, builder.Eq{"deleted": false})

	if maxTransactionTime > 0 {
		cond = cond.And(builder.Lte{"transaction_time": maxTransactionTime})
	}

	if minTransactionTime > 0 {
		cond = cond.And(builder.Gte{"transaction_time": minTransactionTime})
	}

	// Type filter
	switch {
	case models.TRANSACTION_DB_TYPE_MODIFY_BALANCE <= transactionDbType && transactionDbType <= models.TRANSACTION_DB_TYPE_EXPENSE:
		cond = cond.And(builder.Eq{"type": transactionDbType})
	case transactionDbType == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transactionDbType == models.TRANSACTION_DB_TYPE_TRANSFER_IN:
		switch {
		case len(accountIds) == 0:
			cond = cond.And(builder.Eq{"type": models.TRANSACTION_DB_TYPE_TRANSFER_OUT})
		case len(accountIds) == 1:
			cond = cond.And(builder.Or(
				builder.Eq{"type": models.TRANSACTION_DB_TYPE_TRANSFER_OUT},
				builder.Eq{"type": models.TRANSACTION_DB_TYPE_TRANSFER_IN},
			))
		default: // len(accountIds) > 1
			accountIdValues := make([]any, len(accountIds))
			for i, id := range accountIds {
				accountIdValues[i] = id
			}
			cond = cond.And(builder.Or(
				builder.Eq{"type": models.TRANSACTION_DB_TYPE_TRANSFER_OUT},
				builder.And(
					builder.Eq{"type": models.TRANSACTION_DB_TYPE_TRANSFER_IN},
					builder.NotIn("related_account_id", accountIdValues...),
				),
			))
		}
	default:
		if noDuplicated {
			switch {
			case len(accountIds) == 0:
				cond = cond.And(builder.In("type",
					models.TRANSACTION_DB_TYPE_MODIFY_BALANCE,
					models.TRANSACTION_DB_TYPE_INCOME,
					models.TRANSACTION_DB_TYPE_EXPENSE,
					models.TRANSACTION_DB_TYPE_TRANSFER_OUT,
				))
			case len(accountIds) == 1:
				// Do Nothing
			default: // len(accountIds) > 1
				accountIdValues := make([]any, len(accountIds))
				for i, id := range accountIds {
					accountIdValues[i] = id
				}
				cond = cond.And(builder.Or(
					builder.In("type",
						models.TRANSACTION_DB_TYPE_MODIFY_BALANCE,
						models.TRANSACTION_DB_TYPE_INCOME,
						models.TRANSACTION_DB_TYPE_EXPENSE,
						models.TRANSACTION_DB_TYPE_TRANSFER_OUT,
					),
					builder.And(
						builder.Eq{"type": models.TRANSACTION_DB_TYPE_TRANSFER_IN},
						builder.NotIn("related_account_id", accountIdValues...),
					),
				))
			}
		}
	}

	// Category filter
	if len(categoryIds) > 0 {
		categoryIdValues := make([]any, len(categoryIds))
		for i, id := range categoryIds {
			categoryIdValues[i] = id
		}
		cond = cond.And(builder.In("category_id", categoryIdValues...))
	}

	// Account filter
	if len(accountIds) > 0 {
		accountIdValues := make([]any, len(accountIds))
		for i, id := range accountIds {
			accountIdValues[i] = id
		}
		cond = cond.And(builder.In("account_id", accountIdValues...))
	}

	// Amount filter
	if amountFilter != "" {
		amountFilterItems := strings.Split(amountFilter, ":")

		if len(amountFilterItems) >= 2 {
			switch amountFilterItems[0] {
			case "gt":
				if value, err := utils.StringToInt64(amountFilterItems[1]); err == nil {
					cond = cond.And(builder.Gt{"amount": value})
				}
			case "lt":
				if value, err := utils.StringToInt64(amountFilterItems[1]); err == nil {
					cond = cond.And(builder.Lt{"amount": value})
				}
			case "eq":
				if value, err := utils.StringToInt64(amountFilterItems[1]); err == nil {
					cond = cond.And(builder.Eq{"amount": value})
				}
			case "ne":
				if value, err := utils.StringToInt64(amountFilterItems[1]); err == nil {
					cond = cond.And(builder.Neq{"amount": value})
				}
			case "bt":
				if len(amountFilterItems) == 3 {
					value1, err := utils.StringToInt64(amountFilterItems[1])
					value2, err2 := utils.StringToInt64(amountFilterItems[2])

					if err == nil && err2 == nil {
						cond = cond.And(builder.Gte{"amount": value1}, builder.Lte{"amount": value2})
					}
				}
			case "nb":
				if len(amountFilterItems) == 3 {
					value1, err := utils.StringToInt64(amountFilterItems[1])
					value2, err2 := utils.StringToInt64(amountFilterItems[2])

					if err == nil && err2 == nil {
						cond = cond.And(builder.Or(builder.Lt{"amount": value1}, builder.Gt{"amount": value2}))
					}
				}
			}
		}
	}

	// Keyword filter
	if keyword != "" {
		cond = cond.And(builder.Like{"comment", "%" + keyword + "%"})
	}

	// Counterparty filter
	if params.CounterpartyId > 0 {
		cond = cond.And(builder.Eq{"counterparty_id": params.CounterpartyId})
	}

	return cond
}

func (s *TransactionService) appendFilterTagIdsConditionToQuery(sess *xorm.Session, uid int64, maxTransactionTime int64, minTransactionTime int64, tagFilters []*models.TransactionTagFilter, noTags bool) *xorm.Session {
	if noTags {
		subQueryCondition := builder.And(builder.Eq{"uid": uid}, builder.Eq{"deleted": false})

		if maxTransactionTime > 0 {
			subQueryCondition = subQueryCondition.And(builder.Lte{"transaction_time": maxTransactionTime})
		}

		if minTransactionTime > 0 {
			subQueryCondition = subQueryCondition.And(builder.Gte{"transaction_time": minTransactionTime})
		}

		subQuery := builder.Select("transaction_id").From("transaction_tag_index").Where(subQueryCondition)
		sess.NotIn("transaction_id", subQuery).NotIn("related_id", subQuery)
		return sess
	}

	if len(tagFilters) < 1 {
		return sess
	}

	for i := 0; i < len(tagFilters); i++ {
		tagFilter := tagFilters[i]
		subQueryCondition := builder.And(builder.Eq{"uid": uid}, builder.Eq{"deleted": false})

		if maxTransactionTime > 0 {
			subQueryCondition = subQueryCondition.And(builder.Lte{"transaction_time": maxTransactionTime})
		}

		if minTransactionTime > 0 {
			subQueryCondition = subQueryCondition.And(builder.Gte{"transaction_time": minTransactionTime})
		}

		subQueryCondition = subQueryCondition.And(builder.In("tag_id", tagFilter.TagIds))
		subQuery := builder.Select("transaction_id").From("transaction_tag_index").Where(subQueryCondition)

		if tagFilter.Type == models.TRANSACTION_TAG_FILTER_HAS_ALL || tagFilter.Type == models.TRANSACTION_TAG_FILTER_NOT_HAS_ALL {
			subQuery = subQuery.GroupBy("transaction_id").Having(fmt.Sprintf("COUNT(DISTINCT tag_id) >= %d", len(tagFilter.TagIds)))
		}

		switch tagFilter.Type {
		case models.TRANSACTION_TAG_FILTER_HAS_ANY, models.TRANSACTION_TAG_FILTER_HAS_ALL:
			sess.And(builder.Or(builder.In("transaction_id", subQuery), builder.In("related_id", subQuery)))
		case models.TRANSACTION_TAG_FILTER_NOT_HAS_ANY, models.TRANSACTION_TAG_FILTER_NOT_HAS_ALL:
			sess.NotIn("transaction_id", subQuery).NotIn("related_id", subQuery)
		}
	}

	return sess
}

func (s *TransactionService) isAccountIdValid(transaction *models.Transaction) error {
	switch transaction.Type {
	case models.TRANSACTION_DB_TYPE_MODIFY_BALANCE:
		if transaction.RelatedAccountId != 0 && transaction.RelatedAccountId != transaction.AccountId {
			return errs.ErrTransactionDestinationAccountCannotBeSet
		}
	case models.TRANSACTION_DB_TYPE_INCOME,
		models.TRANSACTION_DB_TYPE_EXPENSE:
		if transaction.RelatedAccountId != 0 {
			return errs.ErrTransactionDestinationAccountCannotBeSet
		} else if transaction.RelatedAccountAmount != 0 {
			return errs.ErrTransactionDestinationAmountCannotBeSet
		}
	case models.TRANSACTION_DB_TYPE_TRANSFER_OUT:
		if transaction.AccountId == transaction.RelatedAccountId {
			return errs.ErrTransactionSourceAndDestinationIdCannotBeEqual
		}
	case models.TRANSACTION_DB_TYPE_TRANSFER_IN:
		return errs.ErrTransactionTypeInvalid
	default:
		return errs.ErrTransactionTypeInvalid
	}

	return nil
}

func (s *TransactionService) getAccountModels(sess *xorm.Session, transaction *models.Transaction) (sourceAccount *models.Account, destinationAccount *models.Account, err error) {
	sourceAccount = &models.Account{}
	destinationAccount = &models.Account{}

	has, err := sess.ID(transaction.AccountId).Where("uid=? AND deleted=?", transaction.Uid, false).Get(sourceAccount)

	if err != nil {
		return nil, nil, err
	} else if !has {
		return nil, nil, errs.ErrSourceAccountNotFound
	}

	// check whether the related account is valid
	switch transaction.Type {
	case models.TRANSACTION_DB_TYPE_MODIFY_BALANCE:
		if transaction.RelatedAccountId != 0 && transaction.RelatedAccountId != transaction.AccountId {
			return nil, nil, errs.ErrAccountIdInvalid
		} else {
			destinationAccount = sourceAccount
		}
	case models.TRANSACTION_DB_TYPE_INCOME, models.TRANSACTION_DB_TYPE_EXPENSE:
		if transaction.RelatedAccountId != 0 {
			return nil, nil, errs.ErrAccountIdInvalid
		}

		destinationAccount = nil
	case models.TRANSACTION_DB_TYPE_TRANSFER_OUT, models.TRANSACTION_DB_TYPE_TRANSFER_IN:
		if transaction.RelatedAccountId <= 0 {
			return nil, nil, errs.ErrAccountIdInvalid
		} else {
			has, err = sess.ID(transaction.RelatedAccountId).Where("uid=? AND deleted=?", transaction.Uid, false).Get(destinationAccount)

			if err != nil {
				return nil, nil, err
			} else if !has {
				return nil, nil, errs.ErrDestinationAccountNotFound
			}
		}
	}

	// check whether the parent accounts are valid
	if sourceAccount.ParentAccountId > 0 && destinationAccount != nil && sourceAccount.ParentAccountId != destinationAccount.ParentAccountId && destinationAccount.ParentAccountId > 0 {
		var accounts []*models.Account
		err := sess.Where("uid=? AND deleted=? and (account_id=? or account_id=?)", transaction.Uid, false, sourceAccount.ParentAccountId, destinationAccount.ParentAccountId).Find(&accounts)

		if err != nil {
			return nil, nil, err
		}

		if len(accounts) < 2 {
			return nil, nil, errs.ErrAccountNotFound
		}

		for i := 0; i < len(accounts); i++ {
			account := accounts[i]

			if account.Hidden {
				return nil, nil, errs.ErrCannotUseHiddenAccount
			}
		}
	} else if sourceAccount.ParentAccountId > 0 && (destinationAccount == nil || sourceAccount.ParentAccountId == destinationAccount.ParentAccountId || destinationAccount.ParentAccountId == 0) {
		sourceParentAccount := &models.Account{}
		has, err = sess.ID(sourceAccount.ParentAccountId).Where("uid=? AND deleted=?", transaction.Uid, false).Get(sourceParentAccount)

		if err != nil {
			return nil, nil, err
		} else if !has {
			return nil, nil, errs.ErrSourceAccountNotFound
		}

		if sourceParentAccount.Hidden {
			return nil, nil, errs.ErrCannotUseHiddenAccount
		}
	} else if sourceAccount.ParentAccountId == 0 && destinationAccount != nil && destinationAccount.ParentAccountId > 0 {
		destinationParentAccount := &models.Account{}
		has, err = sess.ID(destinationAccount.ParentAccountId).Where("uid=? AND deleted=?", transaction.Uid, false).Get(destinationParentAccount)

		if err != nil {
			return nil, nil, err
		} else if !has {
			return nil, nil, errs.ErrDestinationAccountNotFound
		}

		if destinationParentAccount.Hidden {
			return nil, nil, errs.ErrCannotUseHiddenAccount
		}
	}

	return sourceAccount, destinationAccount, nil
}

func (s *TransactionService) getOldAccountModels(sess *xorm.Session, transaction *models.Transaction, oldTransaction *models.Transaction, sourceAccount *models.Account, destinationAccount *models.Account) (oldSourceAccount *models.Account, oldDestinationAccount *models.Account, err error) {
	oldSourceAccount = &models.Account{}
	oldDestinationAccount = &models.Account{}

	if transaction.AccountId == oldTransaction.AccountId {
		oldSourceAccount = sourceAccount
	} else {
		has, err := sess.ID(oldTransaction.AccountId).Where("uid=? AND deleted=?", transaction.Uid, false).Get(oldSourceAccount)

		if err != nil {
			return nil, nil, err
		} else if !has {
			return nil, nil, errs.ErrSourceAccountNotFound
		}
	}

	if transaction.RelatedAccountId == oldTransaction.RelatedAccountId {
		oldDestinationAccount = destinationAccount
	} else {
		has, err := sess.ID(oldTransaction.RelatedAccountId).Where("uid=? AND deleted=?", transaction.Uid, false).Get(oldDestinationAccount)

		if err != nil {
			return nil, nil, err
		} else if !has {
			return nil, nil, errs.ErrDestinationAccountNotFound
		}
	}
	return oldSourceAccount, oldDestinationAccount, nil
}

func (s *TransactionService) getRelatedUpdateColumns(updateCols []string) []string {
	relatedUpdateCols := make([]string, len(updateCols))

	for i := 0; i < len(updateCols); i++ {
		switch updateCols[i] {
		case "account_id":
			relatedUpdateCols[i] = "related_account_id"
		case "related_account_id":
			relatedUpdateCols[i] = "account_id"
		case "amount":
			relatedUpdateCols[i] = "related_account_amount"
		case "related_account_amount":
			relatedUpdateCols[i] = "amount"
		default:
			relatedUpdateCols[i] = updateCols[i]
		}
	}

	return relatedUpdateCols
}

func (s *TransactionService) isCategoryValid(sess *xorm.Session, transaction *models.Transaction) error {
	switch transaction.Type {
	case models.TRANSACTION_DB_TYPE_MODIFY_BALANCE:
		if transaction.CategoryId != 0 {
			return errs.ErrBalanceModificationTransactionCannotSetCategory
		}
	default:
		category := &models.TransactionCategory{}
		has, err := sess.ID(transaction.CategoryId).Where("uid=? AND deleted=?", transaction.Uid, false).Get(category)

		if err != nil {
			return err
		} else if !has {
			return errs.ErrTransactionCategoryNotFound
		}

		if category.Hidden {
			return errs.ErrCannotUseHiddenTransactionCategory
		}

		if (transaction.Type == models.TRANSACTION_DB_TYPE_INCOME && category.Type != models.CATEGORY_TYPE_INCOME) ||
			(transaction.Type == models.TRANSACTION_DB_TYPE_EXPENSE && category.Type != models.CATEGORY_TYPE_EXPENSE) ||
			((transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN) && category.Type != models.CATEGORY_TYPE_TRANSFER) {
			return errs.ErrTransactionCategoryTypeInvalid
		}
	}

	return nil
}

func (s *TransactionService) isTagsValid(sess *xorm.Session, transaction *models.Transaction, transactionTagIndexes []*models.TransactionTagIndex, tagIds []int64) error {
	if len(transactionTagIndexes) > 0 {
		var tags []*models.TransactionTag
		err := sess.Where("uid=? AND deleted=?", transaction.Uid, false).In("tag_id", tagIds).Find(&tags)

		if err != nil {
			return err
		}

		tagMap := make(map[int64]*models.TransactionTag)

		for i := 0; i < len(tags); i++ {
			if tags[i].Hidden {
				return errs.ErrCannotUseHiddenTransactionTag
			}

			tagMap[tags[i].TagId] = tags[i]
		}

		for i := 0; i < len(transactionTagIndexes); i++ {
			if _, exists := tagMap[transactionTagIndexes[i].TagId]; !exists {
				return errs.ErrTransactionTagNotFound
			}
		}
	}

	return nil
}

func (s *TransactionService) isPicturesValid(sess *xorm.Session, transaction *models.Transaction, pictureIds []int64) error {
	if len(pictureIds) > 0 {
		var pictureInfos []*models.TransactionPictureInfo
		err := sess.Where("uid=? AND deleted=?", transaction.Uid, false).In("picture_id", pictureIds).Find(&pictureInfos)

		if err != nil {
			return err
		}

		pictureInfoMap := make(map[int64]*models.TransactionPictureInfo)

		for i := 0; i < len(pictureInfos); i++ {
			if pictureInfos[i].TransactionId != models.TransactionPictureNewPictureTransactionId && pictureInfos[i].TransactionId != transaction.TransactionId {
				return errs.ErrTransactionPictureIdInvalid
			}

			pictureInfoMap[pictureInfos[i].PictureId] = pictureInfos[i]
		}

		for i := 0; i < len(pictureIds); i++ {
			if _, exists := pictureInfoMap[pictureIds[i]]; !exists {
				return errs.ErrTransactionPictureNotFound
			}
		}
	}

	return nil
}
