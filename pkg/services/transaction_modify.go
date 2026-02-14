// transaction_modify.go implements transaction modification operations.
package services

import (
	"fmt"
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// SetTransactionPlanned updates the planned flag of a transaction
func (s *TransactionService) SetTransactionPlanned(c core.Context, uid int64, transactionId int64, planned bool) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	updateModel := &models.Transaction{Planned: planned}
	updatedRows, err := s.UserDataDB(uid).NewSession(c).ID(transactionId).Cols("planned").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

	if err != nil {
		return err
	} else if updatedRows < 1 {
		return errs.ErrTransactionNotFound
	}

	return nil
}

// SetTransactionSourceTemplateId updates the source template id of a transaction
func (s *TransactionService) SetTransactionSourceTemplateId(c core.Context, uid int64, transactionId int64, templateId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	updateModel := &models.Transaction{SourceTemplateId: templateId}
	updatedRows, err := s.UserDataDB(uid).NewSession(c).ID(transactionId).Cols("source_template_id").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

	if err != nil {
		return err
	} else if updatedRows < 1 {
		return errs.ErrTransactionNotFound
	}

	return nil
}

// ModifyTransaction saves an existed transaction to database
func (s *TransactionService) ModifyTransaction(c core.Context, transaction *models.Transaction, currentTagIdsCount int, addTagIds []int64, removeTagIds []int64, addPictureIds []int64, removePictureIds []int64) error {
	if transaction.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	needTagIndexUuidCount := uint16(len(addTagIds))
	tagIndexUuids := s.GenerateUuids(uuid.UUID_TYPE_TAG_INDEX, needTagIndexUuidCount)

	if len(tagIndexUuids) < int(needTagIndexUuidCount) {
		return errs.ErrSystemIsBusy
	}

	updateCols := make([]string, 0, 16)

	now := time.Now().Unix()

	transaction.TransactionTime = utils.GetMinTransactionTimeFromUnixTime(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime))
	transaction.UpdatedUnixTime = now
	updateCols = append(updateCols, "updated_unix_time")

	addTagIds = utils.ToUniqueInt64Slice(addTagIds)
	removeTagIds = utils.ToUniqueInt64Slice(removeTagIds)

	transactionTagIndexes := make([]*models.TransactionTagIndex, len(addTagIds))

	for i := 0; i < len(addTagIds); i++ {
		transactionTagIndexes[i] = &models.TransactionTagIndex{
			TagIndexId:      tagIndexUuids[i],
			Uid:             transaction.Uid,
			Deleted:         false,
			TagId:           addTagIds[i],
			TransactionId:   transaction.TransactionId,
			CreatedUnixTime: now,
			UpdatedUnixTime: now,
		}
	}

	err := s.UserDataDB(transaction.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		// Get and verify current transaction
		oldTransaction := &models.Transaction{}
		has, err := sess.ID(transaction.TransactionId).Where("uid=? AND deleted=?", transaction.Uid, false).Get(oldTransaction)

		if err != nil {
			log.Errorf(c, "[transactions.ModifyTransaction] failed to get current transaction, because %s", err.Error())
			return err
		} else if !has {
			return errs.ErrTransactionNotFound
		}

		transaction.Type = oldTransaction.Type

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			transaction.RelatedId = oldTransaction.RelatedId
		}

		// Check whether account id is valid
		err = s.isAccountIdValid(transaction)

		if err != nil {
			return err
		}

		// Get and verify source and destination account (if necessary)
		sourceAccount, destinationAccount, err := s.getAccountModels(sess, transaction)

		if err != nil {
			log.Errorf(c, "[transactions.ModifyTransaction] failed to get account, because %s", err.Error())
			return err
		}

		if sourceAccount.Hidden || (destinationAccount != nil && destinationAccount.Hidden) {
			return errs.ErrCannotModifyTransactionInHiddenAccount
		}

		if sourceAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS || (destinationAccount != nil && destinationAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS) {
			return errs.ErrCannotModifyTransactionInParentAccount
		}

		if (transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN) &&
			sourceAccount.Currency == destinationAccount.Currency && transaction.Amount != transaction.RelatedAccountAmount {
			return errs.ErrTransactionSourceAndDestinationAmountNotEqual
		}

		if (transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN) &&
			(transaction.Amount < 0 || transaction.RelatedAccountAmount < 0) {
			return errs.ErrTransferTransactionAmountCannotBeLessThanZero
		}

		oldSourceAccount, oldDestinationAccount, err := s.getOldAccountModels(sess, transaction, oldTransaction, sourceAccount, destinationAccount)

		if err != nil {
			log.Errorf(c, "[transactions.ModifyTransaction] failed to get old account, because %s", err.Error())
			return err
		}

		if oldSourceAccount.Hidden || (oldDestinationAccount != nil && oldDestinationAccount.Hidden) {
			return errs.ErrCannotAddTransactionToHiddenAccount
		}

		// Append modified columns and verify
		if transaction.CategoryId != oldTransaction.CategoryId {
			// Get and verify category
			err = s.isCategoryValid(sess, transaction)

			if err != nil {
				return err
			}

			updateCols = append(updateCols, "category_id")
		}

		modifyTransactionTime := false

		if utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime) != utils.GetUnixTimeFromTransactionTime(oldTransaction.TransactionTime) {
			if oldTransaction.Type == models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
				return errs.ErrBalanceModificationTransactionCannotModifyTime
			}

			sameSecondLatestTransaction := &models.Transaction{}
			minTransactionTime := utils.GetMinTransactionTimeFromUnixTime(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime))
			maxTransactionTime := utils.GetMaxTransactionTimeFromUnixTime(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime))

			has, err = sess.Where("uid=? AND deleted=? AND transaction_time>=? AND transaction_time<=?", transaction.Uid, false, minTransactionTime, maxTransactionTime).OrderBy("transaction_time desc").Limit(1).Get(sameSecondLatestTransaction)

			if err != nil {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to get transaction time, because %s", err.Error())
				return err
			}

			if has && sameSecondLatestTransaction.TransactionTime < maxTransactionTime-1 {
				transaction.TransactionTime = sameSecondLatestTransaction.TransactionTime + 1
			} else if has && sameSecondLatestTransaction.TransactionTime == maxTransactionTime-1 {
				return errs.ErrTooMuchTransactionInOneSecond
			}

			updateCols = append(updateCols, "transaction_time")
			modifyTransactionTime = true
		}

		if transaction.TimezoneUtcOffset != oldTransaction.TimezoneUtcOffset {
			updateCols = append(updateCols, "timezone_utc_offset")
		}

		if transaction.AccountId != oldTransaction.AccountId {
			updateCols = append(updateCols, "account_id")
		}

		if transaction.Amount != oldTransaction.Amount {
			if oldTransaction.Type == models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
				transaction.RelatedAccountAmount = oldTransaction.RelatedAccountAmount + transaction.Amount - oldTransaction.Amount
				updateCols = append(updateCols, "related_account_amount")
			}

			updateCols = append(updateCols, "amount")
		}

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			if transaction.RelatedAccountId != oldTransaction.RelatedAccountId {
				updateCols = append(updateCols, "related_account_id")
			}

			if transaction.RelatedAccountAmount != oldTransaction.RelatedAccountAmount {
				updateCols = append(updateCols, "related_account_amount")
			}
		}

		if transaction.HideAmount != oldTransaction.HideAmount {
			updateCols = append(updateCols, "hide_amount")
		}

		if transaction.CounterpartyId != oldTransaction.CounterpartyId {
			updateCols = append(updateCols, "counterparty_id")
		}

		if transaction.Comment != oldTransaction.Comment {
			updateCols = append(updateCols, "comment")
		}

		if transaction.GeoLongitude != oldTransaction.GeoLongitude {
			updateCols = append(updateCols, "geo_longitude")
		}

		if transaction.GeoLatitude != oldTransaction.GeoLatitude {
			updateCols = append(updateCols, "geo_latitude")
		}

		// Get and verify tags
		err = s.isTagsValid(sess, transaction, transactionTagIndexes, addTagIds)

		if err != nil {
			return err
		}

		// Get and verify pictures
		err = s.isPicturesValid(sess, transaction, addPictureIds)

		if err != nil {
			return err
		}

		// Not allow to add transaction before balance modification transaction
		if transaction.Type != models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
			otherTransactionExists := false

			if destinationAccount != nil && sourceAccount.AccountId != destinationAccount.AccountId {
				otherTransactionExists, err = sess.Cols("uid", "deleted", "account_id").Where("uid=? AND deleted=? AND type=? AND (account_id=? OR account_id=?) AND transaction_time>=?", transaction.Uid, false, models.TRANSACTION_DB_TYPE_MODIFY_BALANCE, sourceAccount.AccountId, destinationAccount.AccountId, transaction.TransactionTime).Limit(1).Exist(&models.Transaction{})
			} else {
				otherTransactionExists, err = sess.Cols("uid", "deleted", "account_id").Where("uid=? AND deleted=? AND type=? AND account_id=? AND transaction_time>=?", transaction.Uid, false, models.TRANSACTION_DB_TYPE_MODIFY_BALANCE, sourceAccount.AccountId, transaction.TransactionTime).Limit(1).Exist(&models.Transaction{})
			}

			if err != nil {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to get whether other transactions exist, because %s", err.Error())
				return err
			} else if otherTransactionExists {
				return errs.ErrCannotAddTransactionBeforeBalanceModificationTransaction
			}
		}

		// Update transaction row
		updatedRows, err := sess.ID(transaction.TransactionId).Cols(updateCols...).Where("uid=? AND deleted=?", transaction.Uid, false).Update(transaction)

		if err != nil {
			log.Errorf(c, "[transactions.ModifyTransaction] failed to update transaction, because %s", err.Error())
			return err
		} else if updatedRows < 1 {
			return errs.ErrTransactionNotFound
		}

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			relatedTransaction := s.GetRelatedTransferTransaction(transaction)

			if utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime) != utils.GetUnixTimeFromTransactionTime(relatedTransaction.TransactionTime) {
				return errs.ErrTooMuchTransactionInOneSecond
			}

			relatedUpdateCols := s.getRelatedUpdateColumns(updateCols)
			updatedRows, err := sess.ID(relatedTransaction.TransactionId).Cols(relatedUpdateCols...).Where("uid=? AND deleted=?", relatedTransaction.Uid, false).Update(relatedTransaction)

			if err != nil {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update related transaction, because %s", err.Error())
				return err
			} else if updatedRows < 1 {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update related transaction")
				return errs.ErrDatabaseOperationFailed
			}
		}

		// Update transaction tag index
		if len(removeTagIds) > 0 {
			tagIndexUpdateModel := &models.TransactionTagIndex{
				Deleted:         true,
				DeletedUnixTime: now,
			}

			deletedRows, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=? AND transaction_id=?", transaction.Uid, false, transaction.TransactionId).In("tag_id", removeTagIds).Update(tagIndexUpdateModel)

			if err != nil {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to remove old transaction tag index, because %s", err.Error())
				return err
			} else if deletedRows < 1 {
				return errs.ErrTransactionTagNotFound
			}
		}

		if len(transactionTagIndexes) > 0 {
			for i := 0; i < len(transactionTagIndexes); i++ {
				transactionTagIndex := transactionTagIndexes[i]
				transactionTagIndex.TransactionTime = transaction.TransactionTime

				_, err := sess.Insert(transactionTagIndex)

				if err != nil {
					log.Errorf(c, "[transactions.ModifyTransaction] failed to add new transaction tag index, because %s", err.Error())
					return err
				}
			}
		} else if len(transactionTagIndexes) == 0 && currentTagIdsCount > 0 && modifyTransactionTime {
			tagIndexUpdateModel := &models.TransactionTagIndex{
				TransactionTime: transaction.TransactionTime,
			}

			_, err := sess.Where("uid=? AND deleted=? AND transaction_id=?", transaction.Uid, false, transaction.TransactionId).Update(tagIndexUpdateModel)

			if err != nil {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update transaction tag index, because %s", err.Error())
				return err
			}
		}

		// Update transaction picture
		if len(removePictureIds) > 0 {
			pictureUpdateModel := &models.TransactionPictureInfo{
				Deleted:         true,
				DeletedUnixTime: now,
			}

			deletedRows, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=? AND transaction_id=?", transaction.Uid, false, transaction.TransactionId).In("picture_id", removePictureIds).Update(pictureUpdateModel)

			if err != nil {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to remove old transaction picture info, because %s", err.Error())
				return err
			} else if deletedRows < 1 {
				return errs.ErrTransactionPictureNotFound
			}
		}

		if len(addPictureIds) > 0 {
			pictureUpdateModel := &models.TransactionPictureInfo{
				TransactionId:   transaction.TransactionId,
				UpdatedUnixTime: now,
			}

			_, err = sess.Cols("transaction_id", "updated_unix_time").Where("uid=? AND deleted=? AND transaction_id=?", transaction.Uid, false, models.TransactionPictureNewPictureTransactionId).In("picture_id", addPictureIds).Update(pictureUpdateModel)

			if err != nil {
				log.Errorf(c, "[transactions.ModifyTransaction] failed to update new transaction picture info, because %s", err.Error())
				return err
			}
		}

		// Update account table (skip balance update for planned/future transactions)
		switch oldTransaction.Type {
		case models.TRANSACTION_DB_TYPE_MODIFY_BALANCE:
			if transaction.AccountId != oldTransaction.AccountId {
				return errs.ErrBalanceModificationTransactionCannotChangeAccountId
			}

			if !oldTransaction.Planned && transaction.Amount != oldTransaction.Amount && transaction.RelatedAccountAmount != oldTransaction.RelatedAccountAmount {
				sourceAccount.UpdatedUnixTime = time.Now().Unix()
				updatedRows, err := sess.ID(sourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)+(%d)", oldTransaction.RelatedAccountAmount, transaction.RelatedAccountAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", sourceAccount.Uid, false).Update(sourceAccount)

				if err != nil {
					log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
					return err
				} else if updatedRows < 1 {
					log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
					return errs.ErrDatabaseOperationFailed
				}
			}
		case models.TRANSACTION_DB_TYPE_INCOME:
			if !oldTransaction.Planned {
				var oldAccountNewAmount int64 = 0
				var newAccountNewAmount int64 = 0

				if transaction.AccountId == oldTransaction.AccountId {
					oldAccountNewAmount = transaction.Amount
				} else {
					newAccountNewAmount = transaction.Amount
				}

				if oldAccountNewAmount != oldTransaction.Amount {
					oldSourceAccount.UpdatedUnixTime = time.Now().Unix()
					updatedRows, err := sess.ID(oldSourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)+(%d)", oldTransaction.Amount, oldAccountNewAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", oldSourceAccount.Uid, false).Update(oldSourceAccount)

					if err != nil {
						log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
						return err
					} else if updatedRows < 1 {
						log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
						return errs.ErrDatabaseOperationFailed
					}
				}

				if newAccountNewAmount != 0 {
					sourceAccount.UpdatedUnixTime = time.Now().Unix()
					updatedRows, err := sess.ID(sourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", newAccountNewAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", sourceAccount.Uid, false).Update(sourceAccount)

					if err != nil {
						log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
						return err
					} else if updatedRows < 1 {
						log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
						return errs.ErrDatabaseOperationFailed
					}
				}
			}
		case models.TRANSACTION_DB_TYPE_EXPENSE:
			if !oldTransaction.Planned {
				var oldAccountNewAmount int64 = 0
				var newAccountNewAmount int64 = 0

				if transaction.AccountId == oldTransaction.AccountId {
					oldAccountNewAmount = transaction.Amount
				} else {
					newAccountNewAmount = transaction.Amount
				}

				if oldAccountNewAmount != oldTransaction.Amount {
					oldSourceAccount.UpdatedUnixTime = time.Now().Unix()
					updatedRows, err := sess.ID(oldSourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)-(%d)", oldTransaction.Amount, oldAccountNewAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", oldSourceAccount.Uid, false).Update(oldSourceAccount)

					if err != nil {
						log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
						return err
					} else if updatedRows < 1 {
						log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
						return errs.ErrDatabaseOperationFailed
					}
				}

				if newAccountNewAmount != 0 {
					sourceAccount.UpdatedUnixTime = time.Now().Unix()
					updatedRows, err := sess.ID(sourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", newAccountNewAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", sourceAccount.Uid, false).Update(sourceAccount)

					if err != nil {
						log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
						return err
					} else if updatedRows < 1 {
						log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
						return errs.ErrDatabaseOperationFailed
					}
				}
			}
		case models.TRANSACTION_DB_TYPE_TRANSFER_OUT:
			if !oldTransaction.Planned {
				var oldSourceAccountNewAmount int64 = 0
				var newSourceAccountNewAmount int64 = 0

				if transaction.AccountId == oldTransaction.AccountId {
					oldSourceAccountNewAmount = transaction.Amount
				} else {
					newSourceAccountNewAmount = transaction.Amount
				}

				if oldSourceAccountNewAmount != oldTransaction.Amount {
					oldSourceAccount.UpdatedUnixTime = time.Now().Unix()
					updatedRows, err := sess.ID(oldSourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)-(%d)", oldTransaction.Amount, oldSourceAccountNewAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", oldSourceAccount.Uid, false).Update(oldSourceAccount)

					if err != nil {
						log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
						return err
					} else if updatedRows < 1 {
						log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
						return errs.ErrDatabaseOperationFailed
					}
				}

				if newSourceAccountNewAmount != 0 {
					sourceAccount.UpdatedUnixTime = time.Now().Unix()
					updatedRows, err := sess.ID(sourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", newSourceAccountNewAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", sourceAccount.Uid, false).Update(sourceAccount)

					if err != nil {
						log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
						return err
					} else if updatedRows < 1 {
						log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
						return errs.ErrDatabaseOperationFailed
					}
				}

				var oldDestinationAccountNewAmount int64 = 0
				var newDestinationAccountNewAmount int64 = 0

				if transaction.RelatedAccountId == oldTransaction.RelatedAccountId {
					oldDestinationAccountNewAmount = transaction.RelatedAccountAmount
				} else {
					newDestinationAccountNewAmount = transaction.RelatedAccountAmount
				}

				if oldDestinationAccountNewAmount != oldTransaction.RelatedAccountAmount {
					oldDestinationAccount.UpdatedUnixTime = time.Now().Unix()
					updatedRows, err := sess.ID(oldDestinationAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)+(%d)", oldTransaction.RelatedAccountAmount, oldDestinationAccountNewAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", oldDestinationAccount.Uid, false).Update(oldDestinationAccount)

					if err != nil {
						log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
						return err
					} else if updatedRows < 1 {
						log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
						return errs.ErrDatabaseOperationFailed
					}
				}

				if newDestinationAccountNewAmount != 0 {
					destinationAccount.UpdatedUnixTime = time.Now().Unix()
					updatedRows, err := sess.ID(destinationAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", newDestinationAccountNewAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", destinationAccount.Uid, false).Update(destinationAccount)

					if err != nil {
						log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance, because %s", err.Error())
						return err
					} else if updatedRows < 1 {
						log.Errorf(c, "[transactions.ModifyTransaction] failed to update account balance")
						return errs.ErrDatabaseOperationFailed
					}
				}
			}
		case models.TRANSACTION_DB_TYPE_TRANSFER_IN:
			return errs.ErrTransactionTypeInvalid
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *TransactionService) MoveAllTransactionsBetweenAccounts(c core.Context, uid int64, fromAccountId int64, toAccountId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if fromAccountId <= 0 || toAccountId <= 0 {
		return errs.ErrAccountIdInvalid
	}

	if fromAccountId == toAccountId {
		return errs.ErrCannotMoveTransactionToSameAccount
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		// get and verify from and to account
		fromAccount := &models.Account{}
		has, err := sess.ID(fromAccountId).Where("uid=? AND deleted=?", uid, false).Get(fromAccount)

		if err != nil {
			return err
		} else if !has {
			return errs.ErrAccountNotFound
		}

		toAccount := &models.Account{}
		has, err = sess.ID(toAccountId).Where("uid=? AND deleted=?", uid, false).Get(toAccount)

		if err != nil {
			return err
		} else if !has {
			return errs.ErrAccountNotFound
		}

		if fromAccount.Hidden || toAccount.Hidden {
			return errs.ErrCannotMoveTransactionFromOrToHiddenAccount
		}

		if fromAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS || toAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
			return errs.ErrCannotMoveTransactionFromOrToParentAccount
		}

		if fromAccount.Currency != toAccount.Currency {
			return errs.ErrCannotMoveTransactionBetweenAccountsWithDifferentCurrencies
		}

		// combine balance modification transaction
		var balanceModificationTransactions []*models.Transaction
		err = sess.Where("uid=? AND deleted=? AND type=? AND (account_id=? OR account_id=?)", uid, false, models.TRANSACTION_DB_TYPE_MODIFY_BALANCE, fromAccountId, toAccountId).Find(&balanceModificationTransactions)

		if err != nil {
			return err
		}

		if len(balanceModificationTransactions) > 2 {
			log.Errorf(c, "[transactions.MoveAllTransactionsBetweenAccounts] user \"uid:%d\" has more than 2 balance modification transactions in account \"id:%d\" and account \"id:%d\", cannot combine balance modification transaction", uid, fromAccountId, toAccountId)
			return errs.ErrOperationFailed
		} else if len(balanceModificationTransactions) == 2 && balanceModificationTransactions[0].AccountId != balanceModificationTransactions[1].AccountId {
			// if two balance modification transactions exist, merge the amounts into the earlier one and delete the later transaction
			var earlierTransaction *models.Transaction
			var laterTransaction *models.Transaction

			if balanceModificationTransactions[0].TransactionTime < balanceModificationTransactions[1].TransactionTime {
				earlierTransaction = balanceModificationTransactions[0]
				laterTransaction = balanceModificationTransactions[1]
			} else {
				earlierTransaction = balanceModificationTransactions[1]
				laterTransaction = balanceModificationTransactions[0]
			}

			earlierTransaction.Amount += laterTransaction.Amount
			earlierTransaction.RelatedAccountAmount += laterTransaction.RelatedAccountAmount
			earlierTransaction.UpdatedUnixTime = time.Now().Unix()

			updatedRows, err := sess.ID(earlierTransaction.TransactionId).Cols("amount", "related_account_amount", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(earlierTransaction)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				log.Errorf(c, "[transactions.MoveAllTransactionsBetweenAccounts] failed to update earlier balance modification transaction")
				return errs.ErrDatabaseOperationFailed
			}

			laterTransaction.Deleted = true
			laterTransaction.DeletedUnixTime = time.Now().Unix()

			deletedRows, err := sess.ID(laterTransaction.TransactionId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(laterTransaction)

			if err != nil {
				return err
			} else if deletedRows < 1 {
				log.Errorf(c, "[transactions.MoveAllTransactionsBetweenAccounts] failed to delete later balance modification transaction")
				return errs.ErrDatabaseOperationFailed
			}

			log.Infof(c, "[transactions.MoveAllTransactionsBetweenAccounts] user \"uid:%d\" has combined two balance modification transactions \"id:%d\" and \"id:%d\", retained transaction is \"id:%d\"", uid, earlierTransaction.TransactionId, laterTransaction.TransactionId, earlierTransaction.TransactionId)
		} else if len(balanceModificationTransactions) == 1 {
			// when merging a new balance modification transaction, if its date is later than the account's earliest transaction, update the balance modification transaction time accordingly
			anotherAccountId := int64(0)

			switch balanceModificationTransactions[0].AccountId {
			case fromAccountId:
				anotherAccountId = toAccountId
			case toAccountId:
				anotherAccountId = fromAccountId
			default:
				log.Errorf(c, "[transactions.MoveAllTransactionsBetweenAccounts] user \"uid:%d\" has a balance modification transaction \"id:%d\" which account id is neither \"%d\" nor \"%d\"", uid, balanceModificationTransactions[0].TransactionId, fromAccountId, toAccountId)
				return errs.ErrOperationFailed
			}

			earliestTransaction := &models.Transaction{}
			has, err := sess.Where("uid=? AND deleted=? AND account_id=?", uid, false, anotherAccountId).OrderBy("transaction_time asc").Limit(1).Get(earliestTransaction)

			if err != nil {
				return err
			} else if has && balanceModificationTransactions[0].TransactionTime > earliestTransaction.TransactionTime {
				balanceModificationTransaction := balanceModificationTransactions[0]
				balanceModificationTransaction.TransactionTime = utils.GetMinTransactionTimeFromUnixTime(utils.GetUnixTimeFromTransactionTime(earliestTransaction.TransactionTime) - 1)
				balanceModificationTransaction.UpdatedUnixTime = time.Now().Unix()

				if balanceModificationTransaction.TransactionTime < 0 {
					balanceModificationTransaction.TransactionTime = 0
				}

				updatedRows, err := sess.ID(balanceModificationTransaction.TransactionId).Cols("transaction_time", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(balanceModificationTransaction)

				if err != nil {
					return err
				} else if updatedRows < 1 {
					log.Errorf(c, "[transactions.MoveAllTransactionsBetweenAccounts] failed to update balance modification transaction time")
					return errs.ErrDatabaseOperationFailed
				}

				log.Infof(c, "[transactions.MoveAllTransactionsBetweenAccounts] user \"uid:%d\" has updated balance modification transaction \"id:%d\" time to %d, because earliest transaction time in account \"id:%d\" is %d", uid, balanceModificationTransaction.TransactionId, balanceModificationTransaction.TransactionTime, toAccountId, earliestTransaction.TransactionTime)
			}
		}

		// update all transactions of from account
		updateModel := &models.Transaction{
			AccountId:       toAccountId,
			UpdatedUnixTime: time.Now().Unix(),
		}

		updatedRows, err := sess.Cols("account_id", "updated_unix_time").Where("uid=? AND deleted=? AND account_id=?", uid, false, fromAccountId).Update(updateModel)

		if err != nil {
			return err
		}

		if updatedRows > 0 {
			log.Infof(c, "[transactions.MoveAllTransactionsBetweenAccounts] user \"uid:%d\" has moved %d transactions from account \"id:%d\" to account \"id:%d\"", uid, updatedRows, fromAccountId, toAccountId)
		}

		// update all related transactions of from account
		updateRelatedModel := &models.Transaction{
			RelatedAccountId: toAccountId,
			UpdatedUnixTime:  time.Now().Unix(),
		}

		relatedUpdatedRows, err := sess.Cols("related_account_id", "updated_unix_time").Where("uid=? AND deleted=? AND related_account_id=?", uid, false, fromAccountId).Update(updateRelatedModel)

		if err != nil {
			return err
		}

		if updatedRows > 0 {
			log.Infof(c, "[transactions.MoveAllTransactionsBetweenAccounts] user \"uid:%d\" has moved %d related transactions from account \"id:%d\" to account \"id:%d\"", uid, relatedUpdatedRows, fromAccountId, toAccountId)
		}

		// delete all transfer transactions which related account id and account id are both
		deletedModel := &models.Transaction{
			Deleted:         true,
			DeletedUnixTime: time.Now().Unix(),
		}

		deletedRows, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=? AND (type=? OR type=?) AND account_id=? AND related_account_id=?", uid, false, models.TRANSACTION_DB_TYPE_TRANSFER_OUT, models.TRANSACTION_DB_TYPE_TRANSFER_IN, toAccountId, toAccountId).Update(deletedModel)

		if err != nil {
			return err
		}

		if deletedRows > 0 {
			log.Infof(c, "[transactions.MoveAllTransactionsBetweenAccounts] user \"uid:%d\" has deleted %d transactions which account id and related account id are both \"%d\"", uid, deletedRows, toAccountId)
		}

		// update account balance
		if fromAccount.Balance != 0 {
			toAccount.UpdatedUnixTime = time.Now().Unix()
			updatedRows, err := sess.ID(toAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", fromAccount.Balance)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(toAccount)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				log.Errorf(c, "[transactions.MoveAllTransactionsBetweenAccounts] failed to update to account balance")
				return errs.ErrDatabaseOperationFailed
			}

			log.Infof(c, "[transactions.MoveAllTransactionsBetweenAccounts] user \"uid:%d\" has updated account \"id:%d\" balance from %d to %d", uid, toAccountId, toAccount.Balance, toAccount.Balance+fromAccount.Balance)

			fromAccount.Balance = 0
			fromAccount.UpdatedUnixTime = time.Now().Unix()
			updatedRows, err = sess.ID(fromAccount.AccountId).Cols("balance", "updated_unix_time").Where("uid=? AND deleted=?", fromAccount.Uid, false).Update(fromAccount)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				log.Errorf(c, "[transactions.MoveAllTransactionsBetweenAccounts] failed to update from account balance")
				return errs.ErrDatabaseOperationFailed
			}
		}

		return nil
	})
}

// ConfirmPlannedTransaction confirms a planned transaction by setting Planned=false and updating the date to now
func (s *TransactionService) ConfirmPlannedTransaction(c core.Context, uid int64, transactionId int64, clientTimezone *time.Location) (*models.Transaction, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	var transaction *models.Transaction

	err := s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		transaction = &models.Transaction{}
		has, err := sess.ID(transactionId).Where("uid=? AND deleted=?", uid, false).Get(transaction)

		if err != nil {
			return err
		} else if !has {
			return errs.ErrTransactionNotFound
		}

		if !transaction.Planned {
			return errs.ErrNothingWillBeUpdated
		}

		newTransactionTime := utils.GetMinTransactionTimeFromUnixTime(now)
		transaction.Planned = false
		transaction.TransactionTime = newTransactionTime
		transaction.UpdatedUnixTime = now

		_, err = sess.ID(transactionId).Where("uid=? AND deleted=?", uid, false).Cols("planned", "transaction_time", "updated_unix_time").Update(transaction)

		if err != nil {
			return err
		}

		// If there is a related transfer transaction, update it too
		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			relatedTransaction := &models.Transaction{
				Planned:         false,
				TransactionTime: newTransactionTime,
				UpdatedUnixTime: now,
			}
			_, err = sess.ID(transaction.RelatedId).Where("uid=? AND deleted=?", uid, false).Cols("planned", "transaction_time", "updated_unix_time").Update(relatedTransaction)
			if err != nil {
				return err
			}
		}

		// Apply balance changes now that the planned transaction is confirmed
		accountUpdateTime := time.Now().Unix()

		switch transaction.Type {
		case models.TRANSACTION_DB_TYPE_MODIFY_BALANCE:
			if transaction.RelatedAccountAmount != 0 {
				sourceAccount := &models.Account{UpdatedUnixTime: accountUpdateTime}
				updatedRows, err := sess.ID(transaction.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", transaction.RelatedAccountAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(sourceAccount)
				if err != nil {
					return err
				} else if updatedRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
		case models.TRANSACTION_DB_TYPE_INCOME:
			if transaction.Amount != 0 {
				sourceAccount := &models.Account{UpdatedUnixTime: accountUpdateTime}
				updatedRows, err := sess.ID(transaction.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", transaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(sourceAccount)
				if err != nil {
					return err
				} else if updatedRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
		case models.TRANSACTION_DB_TYPE_EXPENSE:
			if transaction.Amount != 0 {
				sourceAccount := &models.Account{UpdatedUnixTime: accountUpdateTime}
				updatedRows, err := sess.ID(transaction.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", transaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(sourceAccount)
				if err != nil {
					return err
				} else if updatedRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
		case models.TRANSACTION_DB_TYPE_TRANSFER_OUT:
			if transaction.Amount != 0 {
				sourceAccount := &models.Account{UpdatedUnixTime: accountUpdateTime}
				updatedSourceRows, err := sess.ID(transaction.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", transaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(sourceAccount)
				if err != nil {
					return err
				} else if updatedSourceRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
			if transaction.RelatedAccountAmount != 0 {
				destinationAccount := &models.Account{UpdatedUnixTime: accountUpdateTime}
				updatedDestRows, err := sess.ID(transaction.RelatedAccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", transaction.RelatedAccountAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(destinationAccount)
				if err != nil {
					return err
				} else if updatedDestRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// ModifyAllFuturePlannedTransactions modifies all future planned transactions with the same source template
func (s *TransactionService) ModifyAllFuturePlannedTransactions(c core.Context, uid int64, transactionId int64, modifyReq *models.TransactionModifyAllFutureRequest) (int64, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()
	var affectedCount int64

	err := s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		// Get the source transaction to find the SourceTemplateId
		sourceTransaction := &models.Transaction{}
		has, err := sess.ID(transactionId).Where("uid=? AND deleted=?", uid, false).Get(sourceTransaction)

		if err != nil {
			return err
		} else if !has {
			return errs.ErrTransactionNotFound
		}

		log.Infof(c, "[transactions.ModifyAllFuturePlannedTransactions] sourceTransaction id=%d planned=%v sourceTemplateId=%d transactionTime=%d amount=%d",
			sourceTransaction.TransactionId, sourceTransaction.Planned, sourceTransaction.SourceTemplateId, sourceTransaction.TransactionTime, sourceTransaction.Amount)

		if sourceTransaction.SourceTemplateId == 0 {
			// No source template — nothing to bulk-update
			log.Infof(c, "[transactions.ModifyAllFuturePlannedTransactions] sourceTemplateId=0, skipping bulk update")
			affectedCount = 0
			return nil
		}

		// Find all planned transactions with the same template and date >= this transaction's date
		updateCols := make([]string, 0, 8)
		updateTransaction := &models.Transaction{
			UpdatedUnixTime: now,
		}
		updateCols = append(updateCols, "updated_unix_time")

		if modifyReq.SourceAmount != 0 {
			updateTransaction.Amount = modifyReq.SourceAmount
			updateCols = append(updateCols, "amount")
		}

		if modifyReq.CategoryId != 0 {
			updateTransaction.CategoryId = modifyReq.CategoryId
			updateCols = append(updateCols, "category_id")
		}

		if modifyReq.SourceAccountId != 0 {
			updateTransaction.AccountId = modifyReq.SourceAccountId
			updateCols = append(updateCols, "account_id")
		}

		if modifyReq.DestinationAccountId != 0 {
			updateTransaction.RelatedAccountId = modifyReq.DestinationAccountId
			updateCols = append(updateCols, "related_account_id")
		}

		if modifyReq.DestinationAmount != 0 {
			updateTransaction.RelatedAccountAmount = modifyReq.DestinationAmount
			updateCols = append(updateCols, "related_account_amount")
		}

		updateTransaction.HideAmount = modifyReq.HideAmount
		updateCols = append(updateCols, "hide_amount")

		updateTransaction.CounterpartyId = modifyReq.CounterpartyId
		updateCols = append(updateCols, "counterparty_id")

		updateTransaction.Comment = modifyReq.Comment
		updateCols = append(updateCols, "comment")

		log.Infof(c, "[transactions.ModifyAllFuturePlannedTransactions] updating where: uid=%d, planned=true, source_template_id=%d, transaction_time>=%d, updateCols=%v",
			uid, sourceTransaction.SourceTemplateId, sourceTransaction.TransactionTime, updateCols)

		affectedCount, err = sess.Where("uid=? AND deleted=? AND planned=? AND source_template_id=? AND transaction_time>=?",
			uid, false, true, sourceTransaction.SourceTemplateId, sourceTransaction.TransactionTime).
			Cols(updateCols...).Update(updateTransaction)

		log.Infof(c, "[transactions.ModifyAllFuturePlannedTransactions] update result: affectedCount=%d, err=%v", affectedCount, err)

		return err
	})

	if err != nil {
		return 0, err
	}

	return affectedCount, nil
}
