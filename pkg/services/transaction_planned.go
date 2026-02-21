// transaction_planned.go implements planned transaction operations.
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
)

// SetTransactionPlanned updates the planned flag of a transaction (and its transfer pair if applicable)
func (s *TransactionService) SetTransactionPlanned(c core.Context, uid int64, transactionId int64, planned bool) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		// Get the transaction to check if it's a transfer
		transaction := &models.Transaction{}
		has, err := sess.ID(transactionId).Where("uid=? AND deleted=?", uid, false).Get(transaction)
		if err != nil {
			return err
		} else if !has {
			return errs.ErrTransactionNotFound
		}

		updateModel := &models.Transaction{Planned: planned}
		updatedRows, err := sess.ID(transactionId).Cols("planned").Where("uid=? AND deleted=?", uid, false).Update(updateModel)
		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrTransactionNotFound
		}

		// Also update the related transfer transaction
		if transaction.RelatedId > 0 && (transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN) {
			relatedUpdate := &models.Transaction{Planned: planned}
			_, err = sess.ID(transaction.RelatedId).Cols("planned").Where("uid=? AND deleted=?", uid, false).Update(relatedUpdate)
			if err != nil {
				return err
			}
		}

		return nil
	})
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
		// Use the same 'now' timestamp captured at the start for consistency
		switch transaction.Type {
		case models.TRANSACTION_DB_TYPE_MODIFY_BALANCE:
			if transaction.RelatedAccountAmount != 0 {
				sourceAccount := &models.Account{UpdatedUnixTime: now}
				updatedRows, err := sess.ID(transaction.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", transaction.RelatedAccountAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(sourceAccount)
				if err != nil {
					return err
				} else if updatedRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
		case models.TRANSACTION_DB_TYPE_INCOME:
			if transaction.Amount != 0 {
				sourceAccount := &models.Account{UpdatedUnixTime: now}
				updatedRows, err := sess.ID(transaction.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", transaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(sourceAccount)
				if err != nil {
					return err
				} else if updatedRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
		case models.TRANSACTION_DB_TYPE_EXPENSE:
			if transaction.Amount != 0 {
				sourceAccount := &models.Account{UpdatedUnixTime: now}
				updatedRows, err := sess.ID(transaction.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", transaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(sourceAccount)
				if err != nil {
					return err
				} else if updatedRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
		case models.TRANSACTION_DB_TYPE_TRANSFER_OUT:
			if transaction.Amount != 0 {
				sourceAccount := &models.Account{UpdatedUnixTime: now}
				updatedSourceRows, err := sess.ID(transaction.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", transaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(sourceAccount)
				if err != nil {
					return err
				} else if updatedSourceRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
			if transaction.RelatedAccountAmount != 0 {
				destinationAccount := &models.Account{UpdatedUnixTime: now}
				updatedDestRows, err := sess.ID(transaction.RelatedAccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", transaction.RelatedAccountAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(destinationAccount)
				if err != nil {
					return err
				} else if updatedDestRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
		case models.TRANSACTION_DB_TYPE_TRANSFER_IN:
			// TRANSFER_IN: AccountId is the destination (receiving money), apply balance
			if transaction.Amount != 0 {
				destinationAccount := &models.Account{UpdatedUnixTime: now}
				updatedDestRows, err := sess.ID(transaction.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", transaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(destinationAccount)
				if err != nil {
					return err
				} else if updatedDestRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
			if transaction.RelatedAccountAmount != 0 {
				sourceAccount := &models.Account{UpdatedUnixTime: now}
				updatedSourceRows, err := sess.ID(transaction.RelatedAccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", transaction.RelatedAccountAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(sourceAccount)
				if err != nil {
					return err
				} else if updatedSourceRows < 1 {
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

		affectedCount, err = sess.Where("uid=? AND deleted=? AND source_template_id=? AND transaction_time>=? AND (planned=? OR type=?)",
			uid, false, sourceTransaction.SourceTemplateId, sourceTransaction.TransactionTime, true, models.TRANSACTION_DB_TYPE_TRANSFER_IN).
			Cols(updateCols...).Update(updateTransaction)

		log.Infof(c, "[transactions.ModifyAllFuturePlannedTransactions] update result: affectedCount=%d, err=%v", affectedCount, err)

		// Also update related TRANSFER_IN records with swapped amounts/accounts
		if sourceTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || sourceTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			relatedUpdateCols := []string{"updated_unix_time"}
			relatedUpdate := &models.Transaction{UpdatedUnixTime: now}

			if modifyReq.DestinationAmount != 0 {
				relatedUpdate.Amount = modifyReq.DestinationAmount
				relatedUpdateCols = append(relatedUpdateCols, "amount")
			}
			if modifyReq.SourceAmount != 0 {
				relatedUpdate.RelatedAccountAmount = modifyReq.SourceAmount
				relatedUpdateCols = append(relatedUpdateCols, "related_account_amount")
			}
			if modifyReq.DestinationAccountId != 0 {
				relatedUpdate.AccountId = modifyReq.DestinationAccountId
				relatedUpdateCols = append(relatedUpdateCols, "account_id")
			}
			if modifyReq.SourceAccountId != 0 {
				relatedUpdate.RelatedAccountId = modifyReq.SourceAccountId
				relatedUpdateCols = append(relatedUpdateCols, "related_account_id")
			}

			relatedUpdate.HideAmount = modifyReq.HideAmount
			relatedUpdateCols = append(relatedUpdateCols, "hide_amount")
			relatedUpdate.Comment = modifyReq.Comment
			relatedUpdateCols = append(relatedUpdateCols, "comment")

			// Update TRANSFER_IN records: they have type=TRANSFER_IN and same source_template_id
			relatedAffected, relErr := sess.Where("uid=? AND deleted=? AND source_template_id=? AND transaction_time>=? AND type=?",
				uid, false, sourceTransaction.SourceTemplateId, sourceTransaction.TransactionTime, models.TRANSACTION_DB_TYPE_TRANSFER_IN).
				Cols(relatedUpdateCols...).Update(relatedUpdate)
			if relErr != nil {
				log.Warnf(c, "[transactions.ModifyAllFuturePlannedTransactions] failed to update related TRANSFER_IN records: %s", relErr.Error())
			} else {
				log.Infof(c, "[transactions.ModifyAllFuturePlannedTransactions] updated %d related TRANSFER_IN records", relatedAffected)
			}
		}

		return err
	})

	if err != nil {
		return 0, err
	}

	return affectedCount, nil
}

// UnconfirmTransaction converts an actual transaction back to planned by reversing balance changes
func (s *TransactionService) UnconfirmTransaction(c core.Context, uid int64, transactionId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		transaction := &models.Transaction{}
		has, err := sess.ID(transactionId).Where("uid=? AND deleted=?", uid, false).Get(transaction)

		if err != nil {
			return err
		} else if !has {
			return errs.ErrTransactionNotFound
		}

		if transaction.Planned {
			return errs.ErrNothingWillBeUpdated
		}

		// Set planned=true
		transaction.Planned = true
		transaction.UpdatedUnixTime = now

		_, err = sess.ID(transactionId).Where("uid=? AND deleted=?", uid, false).Cols("planned", "updated_unix_time").Update(transaction)
		if err != nil {
			return err
		}

		// If there is a related transfer transaction, update it too
		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			relatedTransaction := &models.Transaction{
				Planned:         true,
				UpdatedUnixTime: now,
			}
			_, err = sess.ID(transaction.RelatedId).Where("uid=? AND deleted=?", uid, false).Cols("planned", "updated_unix_time").Update(relatedTransaction)
			if err != nil {
				return err
			}
		}

		// Reverse balance changes (opposite of ConfirmPlannedTransaction)
		// Use the same 'now' timestamp captured at the start for consistency
		switch transaction.Type {
		case models.TRANSACTION_DB_TYPE_MODIFY_BALANCE:
			if transaction.RelatedAccountAmount != 0 {
				sourceAccount := &models.Account{UpdatedUnixTime: now}
				updatedRows, err := sess.ID(transaction.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", transaction.RelatedAccountAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(sourceAccount)
				if err != nil {
					return err
				} else if updatedRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
		case models.TRANSACTION_DB_TYPE_INCOME:
			if transaction.Amount != 0 {
				sourceAccount := &models.Account{UpdatedUnixTime: now}
				updatedRows, err := sess.ID(transaction.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", transaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(sourceAccount)
				if err != nil {
					return err
				} else if updatedRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
		case models.TRANSACTION_DB_TYPE_EXPENSE:
			if transaction.Amount != 0 {
				sourceAccount := &models.Account{UpdatedUnixTime: now}
				updatedRows, err := sess.ID(transaction.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", transaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(sourceAccount)
				if err != nil {
					return err
				} else if updatedRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
		case models.TRANSACTION_DB_TYPE_TRANSFER_OUT:
			if transaction.Amount != 0 {
				sourceAccount := &models.Account{UpdatedUnixTime: now}
				updatedSourceRows, err := sess.ID(transaction.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", transaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(sourceAccount)
				if err != nil {
					return err
				} else if updatedSourceRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
			if transaction.RelatedAccountAmount != 0 {
				destinationAccount := &models.Account{UpdatedUnixTime: now}
				updatedDestRows, err := sess.ID(transaction.RelatedAccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", transaction.RelatedAccountAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(destinationAccount)
				if err != nil {
					return err
				} else if updatedDestRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
		case models.TRANSACTION_DB_TYPE_TRANSFER_IN:
			// TRANSFER_IN: AccountId is the destination (received money), reverse it
			if transaction.Amount != 0 {
				destinationAccount := &models.Account{UpdatedUnixTime: now}
				updatedDestRows, err := sess.ID(transaction.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", transaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(destinationAccount)
				if err != nil {
					return err
				} else if updatedDestRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
			if transaction.RelatedAccountAmount != 0 {
				sourceAccount := &models.Account{UpdatedUnixTime: now}
				updatedSourceRows, err := sess.ID(transaction.RelatedAccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", transaction.RelatedAccountAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(sourceAccount)
				if err != nil {
					return err
				} else if updatedSourceRows < 1 {
					return errs.ErrDatabaseOperationFailed
				}
			}
		}

		return nil
	})
}
