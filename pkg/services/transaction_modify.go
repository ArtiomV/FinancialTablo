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
