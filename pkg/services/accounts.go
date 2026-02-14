// accounts.go defines the AccountService singleton and core account operations.
package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// AccountService represents account service
type AccountService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a account service singleton instance
var (
	Accounts = &AccountService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)
