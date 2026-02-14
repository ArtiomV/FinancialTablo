// transaction_create.go implements transaction creation operations.
package services

import (
	"fmt"
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// CreateTransaction saves a new transaction to database
func (s *TransactionService) CreateTransaction(c core.Context, transaction *models.Transaction, tagIds []int64, pictureIds []int64) error {
	if transaction.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	// Check whether account id is valid
	err := s.isAccountIdValid(transaction)

	if err != nil {
		return err
	}

	now := time.Now().Unix()

	needTransactionUuidCount := 1

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		needTransactionUuidCount = 2
	}

	transactionUuids := s.GenerateUuids(uuid.UUID_TYPE_TRANSACTION, uint16(needTransactionUuidCount))

	if len(transactionUuids) < needTransactionUuidCount {
		return errs.ErrSystemIsBusy
	}

	tagIds = utils.ToUniqueInt64Slice(tagIds)
	needTagIndexUuidCount := uint16(len(tagIds))
	tagIndexUuids := s.GenerateUuids(uuid.UUID_TYPE_TAG_INDEX, needTagIndexUuidCount)

	if len(tagIndexUuids) < int(needTagIndexUuidCount) {
		return errs.ErrSystemIsBusy
	}

	transaction.TransactionId = transactionUuids[0]

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		transaction.RelatedId = transactionUuids[1]
	}

	transaction.TransactionTime = utils.GetMinTransactionTimeFromUnixTime(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime))

	transaction.CreatedUnixTime = now
	transaction.UpdatedUnixTime = now

	transactionTagIndexes := make([]*models.TransactionTagIndex, len(tagIds))

	for i := 0; i < len(tagIds); i++ {
		transactionTagIndexes[i] = &models.TransactionTagIndex{
			TagIndexId:      tagIndexUuids[i],
			Uid:             transaction.Uid,
			Deleted:         false,
			TagId:           tagIds[i],
			TransactionId:   transaction.TransactionId,
			CreatedUnixTime: now,
			UpdatedUnixTime: now,
		}
	}

	pictureUpdateModel := &models.TransactionPictureInfo{
		TransactionId:   transaction.TransactionId,
		UpdatedUnixTime: now,
	}

	userDataDb := s.UserDataDB(transaction.Uid)

	return userDataDb.DoTransaction(c, func(sess *xorm.Session) error {
		return s.doCreateTransaction(c, userDataDb, sess, transaction, transactionTagIndexes, tagIds, pictureIds, pictureUpdateModel)
	})
}

func (s *TransactionService) doCreateTransaction(c core.Context, database *datastore.Database, sess *xorm.Session, transaction *models.Transaction, transactionTagIndexes []*models.TransactionTagIndex, tagIds []int64, pictureIds []int64, pictureUpdateModel *models.TransactionPictureInfo) error {
	// Get and verify source and destination account
	sourceAccount, destinationAccount, err := s.getAccountModels(sess, transaction)

	if err != nil {
		return err
	}

	if sourceAccount.Hidden || (destinationAccount != nil && destinationAccount.Hidden) {
		return errs.ErrCannotAddTransactionToHiddenAccount
	}

	if sourceAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS || (destinationAccount != nil && destinationAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS) {
		return errs.ErrCannotAddTransactionToParentAccount
	}

	if (transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN) &&
		sourceAccount.Currency == destinationAccount.Currency && transaction.Amount != transaction.RelatedAccountAmount {
		return errs.ErrTransactionSourceAndDestinationAmountNotEqual
	}

	if (transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN) &&
		(transaction.Amount < 0 || transaction.RelatedAccountAmount < 0) {
		return errs.ErrTransferTransactionAmountCannotBeLessThanZero
	}

	// Get and verify category
	err = s.isCategoryValid(sess, transaction)

	if err != nil {
		return err
	}

	// Get and verify tags
	err = s.isTagsValid(sess, transaction, transactionTagIndexes, tagIds)

	if err != nil {
		return err
	}

	// Get and verify pictures
	err = s.isPicturesValid(sess, transaction, pictureIds)

	if err != nil {
		return err
	}

	// Verify balance modification transaction and calculate real amount
	if transaction.Type == models.TRANSACTION_DB_TYPE_MODIFY_BALANCE {
		otherTransactionExists, err := sess.Cols("uid", "deleted", "account_id").Where("uid=? AND deleted=? AND account_id=?", transaction.Uid, false, sourceAccount.AccountId).Limit(1).Exist(&models.Transaction{})

		if err != nil {
			log.Errorf(c, "[transactions.doCreateTransaction] failed to get whether other transactions exist, because %s", err.Error())
			return err
		} else if otherTransactionExists {
			return errs.ErrBalanceModificationTransactionCannotAddWhenNotEmpty
		}

		transaction.RelatedAccountId = transaction.AccountId
		transaction.RelatedAccountAmount = transaction.Amount - sourceAccount.Balance
	} else { // Not allow to add transaction before balance modification transaction
		otherTransactionExists := false

		if destinationAccount != nil && sourceAccount.AccountId != destinationAccount.AccountId {
			otherTransactionExists, err = sess.Cols("uid", "deleted", "account_id").Where("uid=? AND deleted=? AND type=? AND (account_id=? OR account_id=?) AND transaction_time>=?", transaction.Uid, false, models.TRANSACTION_DB_TYPE_MODIFY_BALANCE, sourceAccount.AccountId, destinationAccount.AccountId, transaction.TransactionTime).Limit(1).Exist(&models.Transaction{})
		} else {
			otherTransactionExists, err = sess.Cols("uid", "deleted", "account_id").Where("uid=? AND deleted=? AND type=? AND account_id=? AND transaction_time>=?", transaction.Uid, false, models.TRANSACTION_DB_TYPE_MODIFY_BALANCE, sourceAccount.AccountId, transaction.TransactionTime).Limit(1).Exist(&models.Transaction{})
		}

		if err != nil {
			log.Errorf(c, "[transactions.doCreateTransaction] failed to get whether other transactions exist, because %s", err.Error())
			return err
		} else if otherTransactionExists {
			return errs.ErrCannotAddTransactionBeforeBalanceModificationTransaction
		}
	}

	// Insert transaction row
	var relatedTransaction *models.Transaction

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		relatedTransaction = s.GetRelatedTransferTransaction(transaction)
	}

	insertTransactionSavePointName := "insert_transaction"
	err = database.SetSavePoint(sess, insertTransactionSavePointName)

	if err != nil {
		log.Errorf(c, "[transactions.doCreateTransaction] failed to set save point \"%s\", because %s", insertTransactionSavePointName, err.Error())
		return err
	}

	createdRows, err := sess.Insert(transaction)

	if err != nil || createdRows < 1 { // maybe another transaction has same time
		if err != nil {
			log.Warnf(c, "[transactions.doCreateTransaction] cannot create transaction, because %s, regenerate transaction time value", err.Error())
		} else {
			log.Warnf(c, "[transactions.doCreateTransaction] cannot create transaction, regenerate transaction time value")
		}

		err = database.RollbackToSavePoint(sess, insertTransactionSavePointName)

		if err != nil {
			log.Errorf(c, "[transactions.doCreateTransaction] failed to rollback to save point \"%s\", because %s", insertTransactionSavePointName, err.Error())
			return err
		}

		sameSecondLatestTransaction := &models.Transaction{}
		minTransactionTime := utils.GetMinTransactionTimeFromUnixTime(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime))
		maxTransactionTime := utils.GetMaxTransactionTimeFromUnixTime(utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime))

		has, err := sess.Where("uid=? AND transaction_time>=? AND transaction_time<=?", transaction.Uid, minTransactionTime, maxTransactionTime).OrderBy("transaction_time desc").Limit(1).Get(sameSecondLatestTransaction)

		if err != nil {
			log.Errorf(c, "[transactions.doCreateTransaction] failed to get transaction time, because %s", err.Error())
			return err
		} else if !has {
			log.Errorf(c, "[transactions.doCreateTransaction] it should have transactions in %d - %d, but result is empty", minTransactionTime, maxTransactionTime)
			return errs.ErrDatabaseOperationFailed
		} else if sameSecondLatestTransaction.TransactionTime == maxTransactionTime-1 {
			return errs.ErrTooMuchTransactionInOneSecond
		}

		transaction.TransactionTime = sameSecondLatestTransaction.TransactionTime + 1
		createdRows, err := sess.Insert(transaction)

		if err != nil {
			log.Errorf(c, "[transactions.doCreateTransaction] failed to add transaction again, because %s", err.Error())
			return err
		} else if createdRows < 1 {
			log.Errorf(c, "[transactions.doCreateTransaction] failed to add transaction again")
			return errs.ErrDatabaseOperationFailed
		}
	}

	if relatedTransaction != nil {
		relatedTransaction.TransactionTime = transaction.TransactionTime + 1

		if utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime) != utils.GetUnixTimeFromTransactionTime(relatedTransaction.TransactionTime) {
			return errs.ErrTooMuchTransactionInOneSecond
		}

		createdRows, err := sess.Insert(relatedTransaction)

		if err != nil {
			log.Errorf(c, "[transactions.doCreateTransaction] failed to add related transaction, because %s", err.Error())
			return err
		} else if createdRows < 1 {
			log.Errorf(c, "[transactions.doCreateTransaction] failed to add related transaction")
			return errs.ErrDatabaseOperationFailed
		}
	}

	err = nil

	// Insert transaction tag index
	if len(transactionTagIndexes) > 0 {
		for i := 0; i < len(transactionTagIndexes); i++ {
			transactionTagIndex := transactionTagIndexes[i]
			transactionTagIndex.TransactionTime = transaction.TransactionTime

			_, err := sess.Insert(transactionTagIndex)

			if err != nil {
				log.Errorf(c, "[transactions.doCreateTransaction] failed to add transaction tag index, because %s", err.Error())
				return err
			}
		}
	}

	// Update transaction picture
	if len(pictureIds) > 0 {
		_, err = sess.Cols("transaction_id", "updated_unix_time").Where("uid=? AND deleted=? AND transaction_id=?", transaction.Uid, false, models.TransactionPictureNewPictureTransactionId).In("picture_id", pictureIds).Update(pictureUpdateModel)

		if err != nil {
			log.Errorf(c, "[transactions.doCreateTransaction] failed to update transaction picture info, because %s", err.Error())
			return err
		}
	}

	// Update account table (skip balance update for planned/future transactions)
	if !transaction.Planned {
		switch transaction.Type {
		case models.TRANSACTION_DB_TYPE_MODIFY_BALANCE:
			if transaction.RelatedAccountAmount != 0 {
				sourceAccount.UpdatedUnixTime = time.Now().Unix()
				updatedRows, err := sess.ID(sourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", transaction.RelatedAccountAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", sourceAccount.Uid, false).Update(sourceAccount)

				if err != nil {
					log.Errorf(c, "[transactions.doCreateTransaction] failed to update account balance, because %s", err.Error())
					return err
				} else if updatedRows < 1 {
					log.Errorf(c, "[transactions.doCreateTransaction] failed to update account balance")
					return errs.ErrDatabaseOperationFailed
				}
			}
		case models.TRANSACTION_DB_TYPE_INCOME:
			if transaction.Amount != 0 {
				sourceAccount.UpdatedUnixTime = time.Now().Unix()
				updatedRows, err := sess.ID(sourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", transaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", sourceAccount.Uid, false).Update(sourceAccount)

				if err != nil {
					log.Errorf(c, "[transactions.doCreateTransaction] failed to update account balance, because %s", err.Error())
					return err
				} else if updatedRows < 1 {
					log.Errorf(c, "[transactions.doCreateTransaction] failed to update account balance")
					return errs.ErrDatabaseOperationFailed
				}
			}
		case models.TRANSACTION_DB_TYPE_EXPENSE:
			if transaction.Amount != 0 {
				sourceAccount.UpdatedUnixTime = time.Now().Unix()
				updatedRows, err := sess.ID(sourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", transaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", sourceAccount.Uid, false).Update(sourceAccount)

				if err != nil {
					log.Errorf(c, "[transactions.doCreateTransaction] failed to update account balance, because %s", err.Error())
					return err
				} else if updatedRows < 1 {
					log.Errorf(c, "[transactions.doCreateTransaction] failed to update account balance")
					return errs.ErrDatabaseOperationFailed
				}
			}
		case models.TRANSACTION_DB_TYPE_TRANSFER_OUT:
			if transaction.Amount != 0 {
				sourceAccount.UpdatedUnixTime = time.Now().Unix()
				updatedSourceRows, err := sess.ID(sourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", transaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", sourceAccount.Uid, false).Update(sourceAccount)

				if err != nil {
					log.Errorf(c, "[transactions.doCreateTransaction] failed to update account balance, because %s", err.Error())
					return err
				} else if updatedSourceRows < 1 {
					log.Errorf(c, "[transactions.doCreateTransaction] failed to update account balance")
					return errs.ErrDatabaseOperationFailed
				}
			}

			if transaction.RelatedAccountAmount != 0 {
				destinationAccount.UpdatedUnixTime = time.Now().Unix()
				updatedDestinationRows, err := sess.ID(destinationAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", transaction.RelatedAccountAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", destinationAccount.Uid, false).Update(destinationAccount)

				if err != nil {
					log.Errorf(c, "[transactions.doCreateTransaction] failed to update account balance, because %s", err.Error())
					return err
				} else if updatedDestinationRows < 1 {
					log.Errorf(c, "[transactions.doCreateTransaction] failed to update account balance")
					return errs.ErrDatabaseOperationFailed
				}
			}
		case models.TRANSACTION_DB_TYPE_TRANSFER_IN:
			return errs.ErrTransactionTypeInvalid
		}
	}

	return err
}
