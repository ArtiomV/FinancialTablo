// transactions.go defines the TransactionService and its singleton instances.
// Central service for all transaction-related operations.
package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// pageCountForLoadTransactionAmounts is the page size used when loading all transactions for balance/statistics calculations
const pageCountForLoadTransactionAmounts = 1000

// dayAccountKey is a composite key for grouping daily account balances by date and account
type dayAccountKey struct {
	YearMonthDay int32
	AccountId    int64
}

// monthCategoryAccountKey is a composite key for grouping monthly inflow/outflow by month, category, account, and type
type monthCategoryAccountKey struct {
	YearMonth        int32
	CategoryId       int64
	AccountId        int64
	RelatedAccountId int64
	Type             models.TransactionDbType
}

// maxBatchImportUuidCount is the maximum number of UUIDs that can be pre-generated for batch import
const maxBatchImportUuidCount = 65535

// batchImportMinProgressUpdateStep is the minimum number of transactions between progress updates during batch import
const batchImportMinProgressUpdateStep = 100

// batchImportProgressStepDivisor divides the total transaction count to determine the progress update interval
const batchImportProgressStepDivisor = 100

// TransactionService represents transaction service
type TransactionService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a transaction service singleton instance
var (
	Transactions = &TransactionService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)
