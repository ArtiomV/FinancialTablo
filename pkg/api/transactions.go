package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/duplicatechecker"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

const pageCountForAccountStatement = 1000

// TransactionsApi represents transaction api
type TransactionsApi struct {
	ApiUsingConfig
	ApiUsingDuplicateChecker
	transactions          *services.TransactionService
	transactionCategories *services.TransactionCategoryService
	transactionTags       *services.TransactionTagService
	transactionTagGroups  *services.TransactionTagGroupService
	transactionPictures   *services.TransactionPictureService
	transactionTemplates  *services.TransactionTemplateService
	transactionSplits     *services.TransactionSplitService
	accounts              *services.AccountService
	counterparties        *services.CounterpartyService
	users                 *services.UserService
}

// Initialize a transaction api singleton instance
var (
	Transactions = &TransactionsApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		ApiUsingDuplicateChecker: ApiUsingDuplicateChecker{
			ApiUsingConfig: ApiUsingConfig{
				container: settings.Container,
			},
			container: duplicatechecker.Container,
		},
		transactions:          services.Transactions,
		transactionCategories: services.TransactionCategories,
		transactionTags:       services.TransactionTags,
		transactionTagGroups:  services.TransactionTagGroups,
		transactionPictures:   services.TransactionPictures,
		transactionTemplates:  services.TransactionTemplates,
		transactionSplits:     services.TransactionSplits,
		accounts:              services.Accounts,
		counterparties:        services.Counterparties,
		users:                 services.Users,
	}
)
