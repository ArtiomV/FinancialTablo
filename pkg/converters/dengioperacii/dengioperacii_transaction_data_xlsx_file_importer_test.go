package dengioperacii

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/converters/converter"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

func TestDengioperaciiTransactionDataXlsxImporterParseImportedData(t *testing.T) {
	importer := DengioperaciiTransactionDataXlsxFileImporter
	context := core.NewNullContext()

	user := &models.User{
		Uid:             1234567890,
		DefaultCurrency: "MDL",
	}

	testdata, err := os.ReadFile("../../../testdata/dengioperacii_test_file.xlsx")
	assert.Nil(t, err)

	allNewTransactions, allNewAccounts, allNewSubExpenseCategories, allNewSubIncomeCategories, allNewSubTransferCategories, _, err := importer.ParseImportedData(context, user, testdata, time.UTC, converter.DefaultImporterOptions, nil, nil, nil, nil, nil)
	assert.Nil(t, err)

	// Basic checks: data was parsed
	assert.True(t, len(allNewTransactions) > 0, "Should have parsed transactions")
	assert.True(t, len(allNewAccounts) > 0, "Should have parsed accounts")
	assert.True(t, len(allNewSubExpenseCategories) > 0, "Should have expense categories")
	assert.True(t, len(allNewSubIncomeCategories) > 0, "Should have income categories")

	t.Logf("Parsed %d transactions", len(allNewTransactions))
	t.Logf("Parsed %d accounts", len(allNewAccounts))
	t.Logf("Parsed %d expense categories", len(allNewSubExpenseCategories))
	t.Logf("Parsed %d income categories", len(allNewSubIncomeCategories))
	t.Logf("Parsed %d transfer categories", len(allNewSubTransferCategories))

	// Verify first few transactions have correct structure
	// The file is sorted by date ascending after import, so first transaction should be the oldest
	firstTxn := allNewTransactions[0]
	assert.Equal(t, int64(1234567890), firstTxn.Uid)
	assert.True(t, firstTxn.Type == models.TRANSACTION_DB_TYPE_INCOME ||
		firstTxn.Type == models.TRANSACTION_DB_TYPE_EXPENSE ||
		firstTxn.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT,
		"Transaction type should be valid")
	assert.True(t, firstTxn.Amount != 0, "Amount should not be zero")

	// Check that dates are properly formatted
	formattedTime := utils.FormatUnixTimeToLongDateTime(utils.GetUnixTimeFromTransactionTime(firstTxn.TransactionTime), time.UTC)
	t.Logf("First transaction time: %s", formattedTime)
	assert.Contains(t, formattedTime, "-", "Time should be in YYYY-MM-DD format")

	// Check that account names and currencies were parsed
	for _, acct := range allNewAccounts {
		t.Logf("Account: %s (%s)", acct.Name, acct.Currency)
		assert.NotEmpty(t, acct.Name, "Account name should not be empty")
		assert.NotEmpty(t, acct.Currency, "Account currency should not be empty")
	}

	// Show some sample transactions for verification
	count := 5
	if len(allNewTransactions) < count {
		count = len(allNewTransactions)
	}
	for i := 0; i < count; i++ {
		txn := allNewTransactions[i]
		formattedTime := utils.FormatUnixTimeToLongDateTime(utils.GetUnixTimeFromTransactionTime(txn.TransactionTime), time.UTC)
		t.Logf("Txn[%d]: time=%s type=%d amount=%d account=%s category=%s comment=%s",
			i, formattedTime, txn.Type, txn.Amount, txn.OriginalSourceAccountName, txn.OriginalCategoryName, txn.Comment)
	}
}
