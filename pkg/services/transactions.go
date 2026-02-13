package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

const pageCountForLoadTransactionAmounts = 1000

type dayAccountKey struct {
	YearMonthDay int32
	AccountId    int64
}

type monthCategoryAccountKey struct {
	YearMonth        int32
	CategoryId       int64
	AccountId        int64
	RelatedAccountId int64
	Type             models.TransactionDbType
}

const maxBatchImportUuidCount = 65535
const batchImportMinProgressUpdateStep = 100
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
