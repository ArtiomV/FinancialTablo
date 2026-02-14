// transaction_delete.go implements transaction deletion operations.
package services

import (
	"fmt"
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// DeleteTransaction deletes an existed transaction from database
func (s *TransactionService) DeleteTransaction(c core.Context, uid int64, transactionId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.Transaction{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	tagIndexUpdateModel := &models.TransactionTagIndex{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	pictureUpdateModel := &models.TransactionPictureInfo{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		// Get and verify current transaction
		oldTransaction := &models.Transaction{}
		has, err := sess.ID(transactionId).Where("uid=? AND deleted=?", uid, false).Get(oldTransaction)

		if err != nil {
			return err
		} else if !has {
			return errs.ErrTransactionNotFound
		}

		// Get and verify source and destination account
		sourceAccount, destinationAccount, err := s.getAccountModels(sess, oldTransaction)

		if err != nil {
			return err
		}

		if sourceAccount.Hidden || (destinationAccount != nil && destinationAccount.Hidden) {
			return errs.ErrCannotDeleteTransactionInHiddenAccount
		}

		if sourceAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS || (destinationAccount != nil && destinationAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS) {
			return errs.ErrCannotDeleteTransactionInParentAccount
		}

		// Update transaction row to deleted
		deletedRows, err := sess.ID(oldTransaction.TransactionId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrTransactionNotFound
		}

		if oldTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || oldTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			deletedRows, err = sess.ID(oldTransaction.RelatedId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

			if err != nil {
				return err
			} else if deletedRows < 1 {
				return errs.ErrTransactionNotFound
			}
		}

		// Update transaction tag index
		_, err = sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=? AND transaction_id=?", uid, false, oldTransaction.TransactionId).Update(tagIndexUpdateModel)

		if err != nil {
			return err
		}

		// Update transaction picture
		_, err = sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=? AND transaction_id=?", uid, false, oldTransaction.TransactionId).Update(pictureUpdateModel)

		if err != nil {
			return err
		}

		// Update account table (skip balance update for planned/future transactions)
		if !oldTransaction.Planned {
			switch oldTransaction.Type {
			case models.TRANSACTION_DB_TYPE_MODIFY_BALANCE:
				if oldTransaction.RelatedAccountAmount != 0 {
					sourceAccount.UpdatedUnixTime = time.Now().Unix()
					updatedRows, err := sess.ID(sourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", oldTransaction.RelatedAccountAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", sourceAccount.Uid, false).Update(sourceAccount)

					if err != nil {
						return err
					} else if updatedRows < 1 {
						log.Errorf(c, "[transactions.DeleteTransaction] failed to update account balance")
						return errs.ErrDatabaseOperationFailed
					}
				}
			case models.TRANSACTION_DB_TYPE_INCOME:
				if oldTransaction.Amount != 0 {
					sourceAccount.UpdatedUnixTime = time.Now().Unix()
					updatedRows, err := sess.ID(sourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", oldTransaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", sourceAccount.Uid, false).Update(sourceAccount)

					if err != nil {
						return err
					} else if updatedRows < 1 {
						log.Errorf(c, "[transactions.DeleteTransaction] failed to update account balance")
						return errs.ErrDatabaseOperationFailed
					}
				}
			case models.TRANSACTION_DB_TYPE_EXPENSE:
				if oldTransaction.Amount != 0 {
					sourceAccount.UpdatedUnixTime = time.Now().Unix()
					updatedRows, err := sess.ID(sourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", oldTransaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", sourceAccount.Uid, false).Update(sourceAccount)

					if err != nil {
						return err
					} else if updatedRows < 1 {
						log.Errorf(c, "[transactions.DeleteTransaction] failed to update account balance")
						return errs.ErrDatabaseOperationFailed
					}
				}
			case models.TRANSACTION_DB_TYPE_TRANSFER_OUT:
				if oldTransaction.Amount != 0 {
					sourceAccount.UpdatedUnixTime = time.Now().Unix()
					updatedSourceRows, err := sess.ID(sourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", oldTransaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", sourceAccount.Uid, false).Update(sourceAccount)

					if err != nil {
						return err
					} else if updatedSourceRows < 1 {
						log.Errorf(c, "[transactions.DeleteTransaction] failed to update account balance")
						return errs.ErrDatabaseOperationFailed
					}
				}

				if oldTransaction.RelatedAccountAmount != 0 {
					destinationAccount.UpdatedUnixTime = time.Now().Unix()
					updatedDestinationRows, err := sess.ID(destinationAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", oldTransaction.RelatedAccountAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", destinationAccount.Uid, false).Update(destinationAccount)

					if err != nil {
						return err
					} else if updatedDestinationRows < 1 {
						log.Errorf(c, "[transactions.DeleteTransaction] failed to update related account balance")
						return errs.ErrDatabaseOperationFailed
					}
				}
			case models.TRANSACTION_DB_TYPE_TRANSFER_IN:
				return errs.ErrTransactionTypeInvalid
			}
		}

		return err
	})
}

// DeleteAllTransactions deletes all existed transactions from database
func (s *TransactionService) DeleteAllTransactions(c core.Context, uid int64, deleteAccount bool) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.Transaction{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	tagIndexUpdateModel := &models.TransactionTagIndex{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	pictureUpdateModel := &models.TransactionPictureInfo{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	accountUpdateModel := &models.Account{
		Balance:         0,
		Deleted:         deleteAccount,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		// Update all transactions to deleted
		_, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		}

		// Update all transaction tag index to deleted
		_, err = sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(tagIndexUpdateModel)

		if err != nil {
			return err
		}

		// Update all transaction pictures to deleted
		_, err = sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(pictureUpdateModel)

		if err != nil {
			return err
		}

		// Update all accounts to deleted or set amount to zero
		_, err = sess.Cols("balance", "deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(accountUpdateModel)

		if err != nil {
			return err
		}

		return nil
	})
}

// DeleteAllTransactionsOfAccount deletes all existed transactions of specific account from database
func (s *TransactionService) DeleteAllTransactionsOfAccount(c core.Context, uid int64, accountId int64, pageCount int32) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if accountId <= 0 {
		return errs.ErrAccountIdInvalid
	}

	transactions, err := s.GetAllSpecifiedTransactions(c, &models.TransactionQueryParams{
		Uid:          uid,
		AccountIds:   []int64{accountId},
		NoDuplicated: true,
	}, pageCount)

	if err != nil {
		return err
	}

	if len(transactions) < 1 {
		return nil
	}

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]

		switch transaction.Type {
		case models.TRANSACTION_DB_TYPE_TRANSFER_IN:
			err = s.DeleteTransaction(c, uid, transaction.RelatedId)
		default:
			err = s.DeleteTransaction(c, uid, transaction.TransactionId)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteAllFuturePlannedTransactions deletes all future planned transactions with the same source template
func (s *TransactionService) DeleteAllFuturePlannedTransactions(c core.Context, uid int64, transactionId int64) (int64, error) {
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

		log.Infof(c, "[transactions.DeleteAllFuturePlannedTransactions] sourceTransaction id=%d planned=%v sourceTemplateId=%d transactionTime=%d",
			sourceTransaction.TransactionId, sourceTransaction.Planned, sourceTransaction.SourceTemplateId, sourceTransaction.TransactionTime)

		if sourceTransaction.SourceTemplateId == 0 {
			affectedCount = 0
			return nil
		}

		updateModel := &models.Transaction{
			Deleted:         true,
			DeletedUnixTime: now,
		}

		affectedCount, err = sess.Where("uid=? AND deleted=? AND planned=? AND source_template_id=? AND transaction_time>=?",
			uid, false, true, sourceTransaction.SourceTemplateId, sourceTransaction.TransactionTime).
			Cols("deleted", "deleted_unix_time").Update(updateModel)

		log.Infof(c, "[transactions.DeleteAllFuturePlannedTransactions] delete result: affectedCount=%d, err=%v", affectedCount, err)

		return err
	})

	if err != nil {
		return 0, err
	}

	return affectedCount, nil
}
