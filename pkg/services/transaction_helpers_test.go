package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"xorm.io/builder"

	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// testBuildCondToSQL is a helper to convert builder.Cond to SQL string and args for testing
func testBuildCondToSQL(cond builder.Cond) (string, []interface{}, error) {
	return builder.ToSQL(cond)
}

// ==================== GetRelatedTransferTransaction ====================

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

// ==================== GetTransactionMapByList ====================

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

// ==================== GetTransactionIds ====================

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

// ==================== isAccountIdValid ====================

func TestIsAccountIdValid_ModifyBalance_NoRelatedAccount(t *testing.T) {
	transaction := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_MODIFY_BALANCE,
		AccountId:        2001,
		RelatedAccountId: 0,
	}

	err := Transactions.isAccountIdValid(transaction)
	assert.Nil(t, err)
}

func TestIsAccountIdValid_ModifyBalance_SameRelatedAccount(t *testing.T) {
	transaction := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_MODIFY_BALANCE,
		AccountId:        2001,
		RelatedAccountId: 2001,
	}

	err := Transactions.isAccountIdValid(transaction)
	assert.Nil(t, err)
}

func TestIsAccountIdValid_ModifyBalance_DifferentRelatedAccount(t *testing.T) {
	transaction := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_MODIFY_BALANCE,
		AccountId:        2001,
		RelatedAccountId: 2002,
	}

	err := Transactions.isAccountIdValid(transaction)
	assert.Equal(t, errs.ErrTransactionDestinationAccountCannotBeSet, err)
}

func TestIsAccountIdValid_Income_NoRelatedAccount(t *testing.T) {
	transaction := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_INCOME,
		AccountId:        2001,
		RelatedAccountId: 0,
	}

	err := Transactions.isAccountIdValid(transaction)
	assert.Nil(t, err)
}

func TestIsAccountIdValid_Income_WithRelatedAccount(t *testing.T) {
	transaction := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_INCOME,
		AccountId:        2001,
		RelatedAccountId: 2002,
	}

	err := Transactions.isAccountIdValid(transaction)
	assert.Equal(t, errs.ErrTransactionDestinationAccountCannotBeSet, err)
}

func TestIsAccountIdValid_Income_WithRelatedAmount(t *testing.T) {
	transaction := &models.Transaction{
		Type:                 models.TRANSACTION_DB_TYPE_INCOME,
		AccountId:            2001,
		RelatedAccountId:     0,
		RelatedAccountAmount: 5000,
	}

	err := Transactions.isAccountIdValid(transaction)
	assert.Equal(t, errs.ErrTransactionDestinationAmountCannotBeSet, err)
}

func TestIsAccountIdValid_Expense_NoRelatedAccount(t *testing.T) {
	transaction := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_EXPENSE,
		AccountId:        2001,
		RelatedAccountId: 0,
	}

	err := Transactions.isAccountIdValid(transaction)
	assert.Nil(t, err)
}

func TestIsAccountIdValid_Expense_WithRelatedAccount(t *testing.T) {
	transaction := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_EXPENSE,
		AccountId:        2001,
		RelatedAccountId: 2002,
	}

	err := Transactions.isAccountIdValid(transaction)
	assert.Equal(t, errs.ErrTransactionDestinationAccountCannotBeSet, err)
}

func TestIsAccountIdValid_Expense_WithRelatedAmount(t *testing.T) {
	transaction := &models.Transaction{
		Type:                 models.TRANSACTION_DB_TYPE_EXPENSE,
		AccountId:            2001,
		RelatedAccountId:     0,
		RelatedAccountAmount: 5000,
	}

	err := Transactions.isAccountIdValid(transaction)
	assert.Equal(t, errs.ErrTransactionDestinationAmountCannotBeSet, err)
}

func TestIsAccountIdValid_TransferOut_DifferentAccounts(t *testing.T) {
	transaction := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_TRANSFER_OUT,
		AccountId:        2001,
		RelatedAccountId: 2002,
	}

	err := Transactions.isAccountIdValid(transaction)
	assert.Nil(t, err)
}

func TestIsAccountIdValid_TransferOut_SameAccounts(t *testing.T) {
	transaction := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_TRANSFER_OUT,
		AccountId:        2001,
		RelatedAccountId: 2001,
	}

	err := Transactions.isAccountIdValid(transaction)
	assert.Equal(t, errs.ErrTransactionSourceAndDestinationIdCannotBeEqual, err)
}

func TestIsAccountIdValid_TransferIn_AlwaysInvalid(t *testing.T) {
	transaction := &models.Transaction{
		Type:             models.TRANSACTION_DB_TYPE_TRANSFER_IN,
		AccountId:        2001,
		RelatedAccountId: 2002,
	}

	err := Transactions.isAccountIdValid(transaction)
	assert.Equal(t, errs.ErrTransactionTypeInvalid, err)
}

func TestIsAccountIdValid_InvalidType(t *testing.T) {
	transaction := &models.Transaction{
		Type:      models.TransactionDbType(99),
		AccountId: 2001,
	}

	err := Transactions.isAccountIdValid(transaction)
	assert.Equal(t, errs.ErrTransactionTypeInvalid, err)
}

// ==================== getRelatedUpdateColumns ====================

func TestGetRelatedUpdateColumns_Empty(t *testing.T) {
	result := Transactions.getRelatedUpdateColumns([]string{})
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}

func TestGetRelatedUpdateColumns_AccountIdSwap(t *testing.T) {
	result := Transactions.getRelatedUpdateColumns([]string{"account_id"})
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "related_account_id", result[0])
}

func TestGetRelatedUpdateColumns_RelatedAccountIdSwap(t *testing.T) {
	result := Transactions.getRelatedUpdateColumns([]string{"related_account_id"})
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "account_id", result[0])
}

func TestGetRelatedUpdateColumns_AmountSwap(t *testing.T) {
	result := Transactions.getRelatedUpdateColumns([]string{"amount"})
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "related_account_amount", result[0])
}

func TestGetRelatedUpdateColumns_RelatedAccountAmountSwap(t *testing.T) {
	result := Transactions.getRelatedUpdateColumns([]string{"related_account_amount"})
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "amount", result[0])
}

func TestGetRelatedUpdateColumns_OtherColumnUnchanged(t *testing.T) {
	result := Transactions.getRelatedUpdateColumns([]string{"comment", "category_id", "transaction_time"})
	assert.Equal(t, 3, len(result))
	assert.Equal(t, "comment", result[0])
	assert.Equal(t, "category_id", result[1])
	assert.Equal(t, "transaction_time", result[2])
}

func TestGetRelatedUpdateColumns_MixedColumns(t *testing.T) {
	result := Transactions.getRelatedUpdateColumns([]string{
		"updated_unix_time",
		"account_id",
		"amount",
		"related_account_id",
		"related_account_amount",
		"comment",
	})
	assert.Equal(t, 6, len(result))
	assert.Equal(t, "updated_unix_time", result[0])
	assert.Equal(t, "related_account_id", result[1])
	assert.Equal(t, "related_account_amount", result[2])
	assert.Equal(t, "account_id", result[3])
	assert.Equal(t, "amount", result[4])
	assert.Equal(t, "comment", result[5])
}

// ==================== buildTransactionQueryCondition ====================

func TestBuildTransactionQueryCondition_BasicUidAndDeleted(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid: 100,
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "uid")
	assert.Contains(t, sql, "deleted")
	assert.Contains(t, args, int64(100))
	assert.Contains(t, args, false)
}

func TestBuildTransactionQueryCondition_WithMaxTransactionTime(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:                100,
		MaxTransactionTime: 1700000000,
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "transaction_time")
	assert.Contains(t, args, int64(1700000000))
}

func TestBuildTransactionQueryCondition_WithMinTransactionTime(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:                100,
		MinTransactionTime: 1600000000,
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "transaction_time")
	assert.Contains(t, args, int64(1600000000))
}

func TestBuildTransactionQueryCondition_WithTransactionTimeRange(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:                100,
		MinTransactionTime: 1600000000,
		MaxTransactionTime: 1700000000,
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "transaction_time")
	assert.Contains(t, args, int64(1600000000))
	assert.Contains(t, args, int64(1700000000))
}

func TestBuildTransactionQueryCondition_WithIncomeType(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:             100,
		TransactionType: models.TRANSACTION_TYPE_INCOME,
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "type")
	assert.Contains(t, args, models.TRANSACTION_DB_TYPE_INCOME)
}

func TestBuildTransactionQueryCondition_WithExpenseType(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:             100,
		TransactionType: models.TRANSACTION_TYPE_EXPENSE,
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "type")
	assert.Contains(t, args, models.TRANSACTION_DB_TYPE_EXPENSE)
}

func TestBuildTransactionQueryCondition_WithTransferType_NoAccounts(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:             100,
		TransactionType: models.TRANSACTION_TYPE_TRANSFER,
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "type")
	assert.Contains(t, args, models.TRANSACTION_DB_TYPE_TRANSFER_OUT)
}

func TestBuildTransactionQueryCondition_WithTransferType_OneAccount(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:             100,
		TransactionType: models.TRANSACTION_TYPE_TRANSFER,
		AccountIds:      []int64{2001},
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "type")
	// Should include both TRANSFER_OUT and TRANSFER_IN
	assert.Contains(t, args, models.TRANSACTION_DB_TYPE_TRANSFER_OUT)
	assert.Contains(t, args, models.TRANSACTION_DB_TYPE_TRANSFER_IN)
}

func TestBuildTransactionQueryCondition_WithTransferType_MultipleAccounts(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:             100,
		TransactionType: models.TRANSACTION_TYPE_TRANSFER,
		AccountIds:      []int64{2001, 2002},
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "type")
	assert.Contains(t, sql, "related_account_id")
	assert.Contains(t, args, models.TRANSACTION_DB_TYPE_TRANSFER_OUT)
	assert.Contains(t, args, models.TRANSACTION_DB_TYPE_TRANSFER_IN)
}

func TestBuildTransactionQueryCondition_WithCategoryIds(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:         100,
		CategoryIds: []int64{3001, 3002, 3003},
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "category_id")
	assert.Contains(t, args, int64(3001))
	assert.Contains(t, args, int64(3002))
	assert.Contains(t, args, int64(3003))
}

func TestBuildTransactionQueryCondition_WithAccountIds(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:        100,
		AccountIds: []int64{2001, 2002},
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "account_id")
	assert.Contains(t, args, int64(2001))
	assert.Contains(t, args, int64(2002))
}

func TestBuildTransactionQueryCondition_WithAmountFilterGt(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:          100,
		AmountFilter: "gt:5000",
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "amount")
	assert.Contains(t, args, int64(5000))
}

func TestBuildTransactionQueryCondition_WithAmountFilterLt(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:          100,
		AmountFilter: "lt:10000",
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "amount")
	assert.Contains(t, args, int64(10000))
}

func TestBuildTransactionQueryCondition_WithAmountFilterEq(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:          100,
		AmountFilter: "eq:7500",
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "amount")
	assert.Contains(t, args, int64(7500))
}

func TestBuildTransactionQueryCondition_WithAmountFilterNe(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:          100,
		AmountFilter: "ne:0",
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "amount")
	assert.Contains(t, args, int64(0))
}

func TestBuildTransactionQueryCondition_WithAmountFilterBt(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:          100,
		AmountFilter: "bt:1000:5000",
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "amount")
	assert.Contains(t, args, int64(1000))
	assert.Contains(t, args, int64(5000))
}

func TestBuildTransactionQueryCondition_WithAmountFilterNb(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:          100,
		AmountFilter: "nb:1000:5000",
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "amount")
	assert.Contains(t, args, int64(1000))
	assert.Contains(t, args, int64(5000))
}

func TestBuildTransactionQueryCondition_WithAmountFilterInvalid(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:          100,
		AmountFilter: "gt:notanumber",
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, _, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	// Should not contain amount filter since the value is invalid
	assert.NotContains(t, sql, "amount")
}

func TestBuildTransactionQueryCondition_WithAmountFilterBtMissingSecondValue(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:          100,
		AmountFilter: "bt:1000",
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, _, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	// bt requires 3 parts (bt:min:max), with only 2 parts it should be ignored
	assert.NotContains(t, sql, "amount")
}

func TestBuildTransactionQueryCondition_WithKeyword(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:     100,
		Keyword: "groceries",
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "comment")
	assert.Contains(t, args, "%groceries%")
}

func TestBuildTransactionQueryCondition_NoDuplicated_NoAccounts(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:          100,
		NoDuplicated: true,
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "type")
	// Should include MODIFY_BALANCE, INCOME, EXPENSE, TRANSFER_OUT but not TRANSFER_IN
	assert.Contains(t, args, models.TRANSACTION_DB_TYPE_MODIFY_BALANCE)
	assert.Contains(t, args, models.TRANSACTION_DB_TYPE_INCOME)
	assert.Contains(t, args, models.TRANSACTION_DB_TYPE_EXPENSE)
	assert.Contains(t, args, models.TRANSACTION_DB_TYPE_TRANSFER_OUT)
}

func TestBuildTransactionQueryCondition_NoDuplicated_OneAccount(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:          100,
		NoDuplicated: true,
		AccountIds:   []int64{2001},
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, _, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	// With one account and NoDuplicated, no type filter is added
	// (the default case with len(accountIds)==1 does nothing)
	assert.NotContains(t, sql, "TRANSFER_IN")
}

func TestBuildTransactionQueryCondition_NoDuplicated_MultipleAccounts(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:          100,
		NoDuplicated: true,
		AccountIds:   []int64{2001, 2002},
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "type")
	assert.Contains(t, sql, "related_account_id")
	assert.Contains(t, args, models.TRANSACTION_DB_TYPE_TRANSFER_IN)
}

func TestBuildTransactionQueryCondition_AllFilters(t *testing.T) {
	params := &models.TransactionQueryParams{
		Uid:                100,
		MaxTransactionTime: 1700000000,
		MinTransactionTime: 1600000000,
		TransactionType:    models.TRANSACTION_TYPE_EXPENSE,
		CategoryIds:        []int64{3001},
		AccountIds:         []int64{2001},
		AmountFilter:       "gt:1000",
		Keyword:            "test",
	}

	cond := Transactions.buildTransactionQueryCondition(params)
	sql, args, err := testBuildCondToSQL(cond)

	assert.Nil(t, err)
	assert.Contains(t, sql, "uid")
	assert.Contains(t, sql, "deleted")
	assert.Contains(t, sql, "transaction_time")
	assert.Contains(t, sql, "type")
	assert.Contains(t, sql, "category_id")
	assert.Contains(t, sql, "account_id")
	assert.Contains(t, sql, "amount")
	assert.Contains(t, sql, "comment")
	assert.Contains(t, args, int64(100))
	assert.Contains(t, args, int64(1700000000))
	assert.Contains(t, args, int64(1600000000))
	assert.Contains(t, args, models.TRANSACTION_DB_TYPE_EXPENSE)
	assert.Contains(t, args, int64(3001))
	assert.Contains(t, args, int64(2001))
	assert.Contains(t, args, int64(1000))
	assert.Contains(t, args, "%test%")
}
