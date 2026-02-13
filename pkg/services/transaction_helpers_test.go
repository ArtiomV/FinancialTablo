package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func TestGetRelatedTransferTransaction_TransferOut(t *testing.T) {
	originalTransaction := &models.Transaction{
		TransactionId:        1001,
		Uid:                  100,
		Deleted:              false,
		Type:                 models.TRANSACTION_DB_TYPE_TRANSFER_OUT,
		CategoryId:           5001,
		TransactionTime:      1000000,
		TimezoneUtcOffset:    480,
		AccountId:            2001,
		Amount:               10000,
		RelatedId:            1002,
		RelatedAccountId:     2002,
		RelatedAccountAmount: 15000,
		Comment:              "Transfer to savings",
		GeoLongitude:         116.397128,
		GeoLatitude:          39.916527,
		CreatedIp:            "127.0.0.1",
		CreatedUnixTime:      1700000000,
		UpdatedUnixTime:      1700000001,
		DeletedUnixTime:      0,
	}

	relatedTransaction := Transactions.GetRelatedTransferTransaction(originalTransaction)

	assert.NotNil(t, relatedTransaction)
	assert.Equal(t, int64(1002), relatedTransaction.TransactionId)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_TRANSFER_IN, relatedTransaction.Type)
	assert.Equal(t, int64(2002), relatedTransaction.AccountId)
	assert.Equal(t, int64(15000), relatedTransaction.Amount)
	assert.Equal(t, int64(1001), relatedTransaction.RelatedId)
	assert.Equal(t, int64(2001), relatedTransaction.RelatedAccountId)
	assert.Equal(t, int64(10000), relatedTransaction.RelatedAccountAmount)
	assert.Equal(t, int64(1000001), relatedTransaction.TransactionTime)
	assert.Equal(t, int64(100), relatedTransaction.Uid)
	assert.Equal(t, false, relatedTransaction.Deleted)
	assert.Equal(t, int64(5001), relatedTransaction.CategoryId)
	assert.Equal(t, int16(480), relatedTransaction.TimezoneUtcOffset)
	assert.Equal(t, "Transfer to savings", relatedTransaction.Comment)
	assert.Equal(t, 116.397128, relatedTransaction.GeoLongitude)
	assert.Equal(t, 39.916527, relatedTransaction.GeoLatitude)
	assert.Equal(t, "127.0.0.1", relatedTransaction.CreatedIp)
	assert.Equal(t, int64(1700000000), relatedTransaction.CreatedUnixTime)
	assert.Equal(t, int64(1700000001), relatedTransaction.UpdatedUnixTime)
	assert.Equal(t, int64(0), relatedTransaction.DeletedUnixTime)
}

func TestGetRelatedTransferTransaction_TransferIn(t *testing.T) {
	originalTransaction := &models.Transaction{
		TransactionId:        1002,
		Uid:                  100,
		Deleted:              false,
		Type:                 models.TRANSACTION_DB_TYPE_TRANSFER_IN,
		CategoryId:           5001,
		TransactionTime:      1000001,
		TimezoneUtcOffset:    480,
		AccountId:            2002,
		Amount:               15000,
		RelatedId:            1001,
		RelatedAccountId:     2001,
		RelatedAccountAmount: 10000,
		Comment:              "Transfer from checking",
		GeoLongitude:         0,
		GeoLatitude:          0,
		CreatedIp:            "192.168.1.1",
		CreatedUnixTime:      1700000000,
		UpdatedUnixTime:      1700000002,
		DeletedUnixTime:      0,
	}

	relatedTransaction := Transactions.GetRelatedTransferTransaction(originalTransaction)

	assert.NotNil(t, relatedTransaction)
	assert.Equal(t, int64(1001), relatedTransaction.TransactionId)
	assert.Equal(t, models.TRANSACTION_DB_TYPE_TRANSFER_OUT, relatedTransaction.Type)
	assert.Equal(t, int64(2001), relatedTransaction.AccountId)
	assert.Equal(t, int64(10000), relatedTransaction.Amount)
	assert.Equal(t, int64(1002), relatedTransaction.RelatedId)
	assert.Equal(t, int64(2002), relatedTransaction.RelatedAccountId)
	assert.Equal(t, int64(15000), relatedTransaction.RelatedAccountAmount)
	assert.Equal(t, int64(1000000), relatedTransaction.TransactionTime)
	assert.Equal(t, int64(100), relatedTransaction.Uid)
	assert.Equal(t, false, relatedTransaction.Deleted)
	assert.Equal(t, int64(5001), relatedTransaction.CategoryId)
	assert.Equal(t, int16(480), relatedTransaction.TimezoneUtcOffset)
	assert.Equal(t, "Transfer from checking", relatedTransaction.Comment)
	assert.Equal(t, "192.168.1.1", relatedTransaction.CreatedIp)
	assert.Equal(t, int64(1700000000), relatedTransaction.CreatedUnixTime)
	assert.Equal(t, int64(1700000002), relatedTransaction.UpdatedUnixTime)
}

func TestGetRelatedTransferTransaction_Income(t *testing.T) {
	originalTransaction := &models.Transaction{
		TransactionId:   1003,
		Uid:             100,
		Type:            models.TRANSACTION_DB_TYPE_INCOME,
		TransactionTime: 1000000,
		AccountId:       2001,
		Amount:          50000,
	}

	relatedTransaction := Transactions.GetRelatedTransferTransaction(originalTransaction)

	assert.Nil(t, relatedTransaction)
}

func TestGetRelatedTransferTransaction_Expense(t *testing.T) {
	originalTransaction := &models.Transaction{
		TransactionId:   1004,
		Uid:             100,
		Type:            models.TRANSACTION_DB_TYPE_EXPENSE,
		TransactionTime: 1000000,
		AccountId:       2001,
		Amount:          30000,
	}

	relatedTransaction := Transactions.GetRelatedTransferTransaction(originalTransaction)

	assert.Nil(t, relatedTransaction)
}

func TestGetRelatedTransferTransaction_ModifyBalance(t *testing.T) {
	originalTransaction := &models.Transaction{
		TransactionId:   1005,
		Uid:             100,
		Type:            models.TRANSACTION_DB_TYPE_MODIFY_BALANCE,
		TransactionTime: 1000000,
		AccountId:       2001,
		Amount:          20000,
	}

	relatedTransaction := Transactions.GetRelatedTransferTransaction(originalTransaction)

	assert.Nil(t, relatedTransaction)
}

func TestGetTransactionMapByList_EmptyList(t *testing.T) {
	transactions := make([]*models.Transaction, 0)
	actualTransactionMap := Transactions.GetTransactionMapByList(transactions)

	assert.NotNil(t, actualTransactionMap)
	assert.Equal(t, 0, len(actualTransactionMap))
}

func TestGetTransactionMapByList_MultipleTransactions(t *testing.T) {
	transactions := []*models.Transaction{
		{
			TransactionId: 1001,
			Uid:           100,
			Type:          models.TRANSACTION_DB_TYPE_INCOME,
			Amount:        10000,
			Comment:       "Salary",
		},
		{
			TransactionId: 1002,
			Uid:           100,
			Type:          models.TRANSACTION_DB_TYPE_EXPENSE,
			Amount:        5000,
			Comment:       "Groceries",
		},
		{
			TransactionId: 1003,
			Uid:           100,
			Type:          models.TRANSACTION_DB_TYPE_TRANSFER_OUT,
			Amount:        20000,
			Comment:       "Transfer",
		},
	}
	actualTransactionMap := Transactions.GetTransactionMapByList(transactions)

	assert.Equal(t, 3, len(actualTransactionMap))
	assert.Contains(t, actualTransactionMap, int64(1001))
	assert.Contains(t, actualTransactionMap, int64(1002))
	assert.Contains(t, actualTransactionMap, int64(1003))
	assert.Equal(t, "Salary", actualTransactionMap[1001].Comment)
	assert.Equal(t, "Groceries", actualTransactionMap[1002].Comment)
	assert.Equal(t, "Transfer", actualTransactionMap[1003].Comment)
	assert.Equal(t, int64(10000), actualTransactionMap[1001].Amount)
	assert.Equal(t, int64(5000), actualTransactionMap[1002].Amount)
	assert.Equal(t, int64(20000), actualTransactionMap[1003].Amount)
}

func TestGetTransactionIds_EmptyList(t *testing.T) {
	transactions := make([]*models.Transaction, 0)
	actualIds := Transactions.GetTransactionIds(transactions)

	assert.NotNil(t, actualIds)
	assert.Equal(t, 0, len(actualIds))
}

func TestGetTransactionIds_MultipleTransactions(t *testing.T) {
	transactions := []*models.Transaction{
		{
			TransactionId: 1001,
		},
		{
			TransactionId: 1002,
		},
		{
			TransactionId: 1003,
		},
	}
	actualIds := Transactions.GetTransactionIds(transactions)

	assert.Equal(t, 3, len(actualIds))
	assert.Equal(t, int64(1001), actualIds[0])
	assert.Equal(t, int64(1002), actualIds[1])
	assert.Equal(t, int64(1003), actualIds[2])
}
