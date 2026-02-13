package services

import (
	"fmt"
	"strings"

	"xorm.io/builder"
	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

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

func (s *TransactionService) buildTransactionQueryCondition(uid int64, maxTransactionTime int64, minTransactionTime int64, transactionDbType models.TransactionDbType, categoryIds []int64, accountIds []int64, tagFilters []*models.TransactionTagFilter, amountFilter string, keyword string, noDuplicated bool) (string, []any) {
	condition := "uid=? AND deleted=?"
	conditionParams := make([]any, 0, 16)
	conditionParams = append(conditionParams, uid)
	conditionParams = append(conditionParams, false)

	if maxTransactionTime > 0 {
		condition = condition + " AND transaction_time<=?"
		conditionParams = append(conditionParams, maxTransactionTime)
	}

	if minTransactionTime > 0 {
		condition = condition + " AND transaction_time>=?"
		conditionParams = append(conditionParams, minTransactionTime)
	}

	var accountIdsCondition strings.Builder
	accountIdConditionParams := make([]any, 0, len(accountIds))

	for i := 0; i < len(accountIds); i++ {
		if i > 0 {
			accountIdsCondition.WriteString(",")
		}

		accountIdsCondition.WriteString("?")
		accountIdConditionParams = append(accountIdConditionParams, accountIds[i])
	}

	if models.TRANSACTION_DB_TYPE_MODIFY_BALANCE <= transactionDbType && transactionDbType <= models.TRANSACTION_DB_TYPE_EXPENSE {
		condition = condition + " AND type=?"
		conditionParams = append(conditionParams, transactionDbType)
	} else if transactionDbType == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transactionDbType == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		if len(accountIds) == 0 {
			condition = condition + " AND type=?"
			conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_OUT)
		} else if len(accountIds) == 1 {
			condition = condition + " AND (type=? OR type=?)"
			conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_OUT)
			conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_IN)
		} else { // len(accountsIds) > 1
			condition = condition + " AND (type=? OR (type=? AND related_account_id NOT IN (" + accountIdsCondition.String() + ")))"
			conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_OUT)
			conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_IN)
			conditionParams = append(conditionParams, accountIdConditionParams...)
		}
	} else {
		if noDuplicated {
			if len(accountIds) == 0 {
				condition = condition + " AND (type=? OR type=? OR type=? OR type=?)"
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_MODIFY_BALANCE)
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_INCOME)
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_EXPENSE)
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_OUT)
			} else if len(accountIds) == 1 {
				// Do Nothing
			} else { // len(accountsIds) > 1
				condition = condition + " AND (type=? OR type=? OR type=? OR type=? OR (type=? AND related_account_id NOT IN (" + accountIdsCondition.String() + ")))"
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_MODIFY_BALANCE)
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_INCOME)
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_EXPENSE)
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_OUT)
				conditionParams = append(conditionParams, models.TRANSACTION_DB_TYPE_TRANSFER_IN)
				conditionParams = append(conditionParams, accountIdConditionParams...)
			}
		}
	}

	if len(categoryIds) > 0 {
		var conditions strings.Builder

		for i := 0; i < len(categoryIds); i++ {
			if i > 0 {
				conditions.WriteString(",")
			}

			conditions.WriteString("?")
			conditionParams = append(conditionParams, categoryIds[i])
		}

		if conditions.Len() > 1 {
			condition = condition + " AND category_id IN (" + conditions.String() + ")"
		} else {
			condition = condition + " AND category_id = " + conditions.String()
		}
	}

	if len(accountIds) > 0 {
		if accountIdsCondition.Len() > 1 {
			condition = condition + " AND account_id IN (" + accountIdsCondition.String() + ")"
		} else {
			condition = condition + " AND account_id = " + accountIdsCondition.String()
		}

		conditionParams = append(conditionParams, accountIdConditionParams...)
	}

	if amountFilter != "" {
		amountFilterItems := strings.Split(amountFilter, ":")

		if len(amountFilterItems) == 2 && amountFilterItems[0] == "gt" {
			value, err := utils.StringToInt64(amountFilterItems[1])

			if err == nil {
				condition = condition + " AND amount > ?"
				conditionParams = append(conditionParams, value)
			}
		} else if len(amountFilterItems) == 2 && amountFilterItems[0] == "lt" {
			value, err := utils.StringToInt64(amountFilterItems[1])

			if err == nil {
				condition = condition + " AND amount < ?"
				conditionParams = append(conditionParams, value)
			}
		} else if len(amountFilterItems) == 2 && amountFilterItems[0] == "eq" {
			value, err := utils.StringToInt64(amountFilterItems[1])

			if err == nil {
				condition = condition + " AND amount = ?"
				conditionParams = append(conditionParams, value)
			}
		} else if len(amountFilterItems) == 2 && amountFilterItems[0] == "ne" {
			value, err := utils.StringToInt64(amountFilterItems[1])

			if err == nil {
				condition = condition + " AND amount <> ?"
				conditionParams = append(conditionParams, value)
			}
		} else if len(amountFilterItems) == 3 && amountFilterItems[0] == "bt" {
			value1, err := utils.StringToInt64(amountFilterItems[1])
			value2, err := utils.StringToInt64(amountFilterItems[2])

			if err == nil {
				condition = condition + " AND amount >= ? AND amount <= ?"
				conditionParams = append(conditionParams, value1)
				conditionParams = append(conditionParams, value2)
			}
		} else if len(amountFilterItems) == 3 && amountFilterItems[0] == "nb" {
			value1, err := utils.StringToInt64(amountFilterItems[1])
			value2, err := utils.StringToInt64(amountFilterItems[2])

			if err == nil {
				condition = condition + " AND (amount < ? OR amount > ?)"
				conditionParams = append(conditionParams, value1)
				conditionParams = append(conditionParams, value2)
			}
		}
	}

	if keyword != "" {
		condition = condition + " AND comment LIKE ?"
		conditionParams = append(conditionParams, "%%"+keyword+"%%")
	}

	return condition, conditionParams
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

		if tagFilter.Type == models.TRANSACTION_TAG_FILTER_HAS_ANY || tagFilter.Type == models.TRANSACTION_TAG_FILTER_HAS_ALL {
			sess.And(builder.Or(builder.In("transaction_id", subQuery), builder.In("related_id", subQuery)))
		} else if tagFilter.Type == models.TRANSACTION_TAG_FILTER_NOT_HAS_ANY || tagFilter.Type == models.TRANSACTION_TAG_FILTER_NOT_HAS_ALL {
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
		if updateCols[i] == "account_id" {
			relatedUpdateCols[i] = "related_account_id"
		} else if updateCols[i] == "related_account_id" {
			relatedUpdateCols[i] = "account_id"
		} else if updateCols[i] == "amount" {
			relatedUpdateCols[i] = "related_account_amount"
		} else if updateCols[i] == "related_account_amount" {
			relatedUpdateCols[i] = "amount"
		} else {
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
