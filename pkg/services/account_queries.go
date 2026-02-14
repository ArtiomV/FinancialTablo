// account_queries.go implements account listing and lookup queries.
package services

import (
	"strings"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// GetTotalAccountCountByUid returns total account count of user
func (s *AccountService) GetTotalAccountCountByUid(c core.Context, uid int64) (int64, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	count, err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).Count(&models.Account{})

	return count, err
}

// GetAllAccountsByUid returns all account models of user
func (s *AccountService) GetAllAccountsByUid(c core.Context, uid int64) ([]*models.Account, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var accounts []*models.Account
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).OrderBy("parent_account_id asc, display_order asc").Find(&accounts)

	return accounts, err
}

// GetAccountByAccountId returns account model according to account id
func (s *AccountService) GetAccountByAccountId(c core.Context, uid int64, accountId int64) (*models.Account, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if accountId <= 0 {
		return nil, errs.ErrAccountIdInvalid
	}

	account := &models.Account{}
	has, err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=? AND account_id=?", uid, false, accountId).Get(account)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrAccountNotFound
	}

	return account, err
}

// GetAccountAndSubAccountsByAccountId returns account model and sub-account models according to account id
func (s *AccountService) GetAccountAndSubAccountsByAccountId(c core.Context, uid int64, accountId int64) ([]*models.Account, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if accountId <= 0 {
		return nil, errs.ErrAccountIdInvalid
	}

	var accounts []*models.Account
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=? AND (account_id=? OR parent_account_id=?)", uid, false, accountId, accountId).OrderBy("parent_account_id asc, display_order asc").Find(&accounts)

	return accounts, err
}

// GetSubAccountsByAccountId returns sub-account models according to account id
func (s *AccountService) GetSubAccountsByAccountId(c core.Context, uid int64, accountId int64) ([]*models.Account, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if accountId <= 0 {
		return nil, errs.ErrAccountIdInvalid
	}

	var accounts []*models.Account
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=? AND parent_account_id=?", uid, false, accountId).OrderBy("display_order asc").Find(&accounts)

	return accounts, err
}

// GetSubAccountsByAccountIds returns sub-account models according to account ids
func (s *AccountService) GetSubAccountsByAccountIds(c core.Context, uid int64, accountIds []int64) ([]*models.Account, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if len(accountIds) <= 0 {
		return nil, errs.ErrAccountIdInvalid
	}

	condition := "uid=? AND deleted=?"
	conditionParams := make([]any, 0, len(accountIds)+2)
	conditionParams = append(conditionParams, uid)
	conditionParams = append(conditionParams, false)

	var accountIdConditions strings.Builder

	for i := 0; i < len(accountIds); i++ {
		if accountIds[i] <= 0 {
			return nil, errs.ErrAccountIdInvalid
		}

		if accountIdConditions.Len() > 0 {
			accountIdConditions.WriteString(",")
		}

		accountIdConditions.WriteString("?")
		conditionParams = append(conditionParams, accountIds[i])
	}

	if accountIdConditions.Len() > 1 {
		condition = condition + " AND parent_account_id IN (" + accountIdConditions.String() + ")"
	} else {
		condition = condition + " AND parent_account_id = " + accountIdConditions.String()
	}

	var accounts []*models.Account
	err := s.UserDataDB(uid).NewSession(c).Where(condition, conditionParams...).OrderBy("display_order asc").Find(&accounts)

	return accounts, err
}

// GetAccountsByAccountIds returns account models according to account ids
func (s *AccountService) GetAccountsByAccountIds(c core.Context, uid int64, accountIds []int64) (map[int64]*models.Account, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if accountIds == nil {
		return nil, errs.ErrAccountIdInvalid
	}

	var accounts []*models.Account
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).In("account_id", accountIds).Find(&accounts)

	if err != nil {
		return nil, err
	}

	accountMap := s.GetAccountMapByList(accounts)
	return accountMap, err
}

// GetMaxDisplayOrder returns the max display order according to account category
func (s *AccountService) GetMaxDisplayOrder(c core.Context, uid int64, category models.AccountCategory) (int32, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	account := &models.Account{}
	has, err := s.UserDataDB(uid).NewSession(c).Cols("uid", "deleted", "parent_account_id", "display_order").Where("uid=? AND deleted=? AND parent_account_id=? AND category=?", uid, false, models.LevelOneAccountParentId, category).OrderBy("display_order desc").Limit(1).Get(account)

	if err != nil {
		return 0, err
	}

	if has {
		return account.DisplayOrder, nil
	} else {
		return 0, nil
	}
}

// GetMaxSubAccountDisplayOrder returns the max display order of sub-account according to account category and parent account id
func (s *AccountService) GetMaxSubAccountDisplayOrder(c core.Context, uid int64, category models.AccountCategory, parentAccountId int64) (int32, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	if parentAccountId <= 0 {
		return 0, errs.ErrAccountIdInvalid
	}

	account := &models.Account{}
	has, err := s.UserDataDB(uid).NewSession(c).Cols("uid", "deleted", "parent_account_id", "display_order").Where("uid=? AND deleted=? AND parent_account_id=? AND category=?", uid, false, parentAccountId, category).OrderBy("display_order desc").Limit(1).Get(account)

	if err != nil {
		return 0, err
	}

	if has {
		return account.DisplayOrder, nil
	} else {
		return 0, nil
	}
}

// GetAccountMapByList returns an account map by a list
func (s *AccountService) GetAccountMapByList(accounts []*models.Account) map[int64]*models.Account {
	accountMap := make(map[int64]*models.Account)

	for i := 0; i < len(accounts); i++ {
		account := accounts[i]
		accountMap[account.AccountId] = account
	}
	return accountMap
}

// GetVisibleAccountNameMapByList returns visible account map by a list
func (s *AccountService) GetVisibleAccountNameMapByList(accounts []*models.Account) map[string]*models.Account {
	accountMap := make(map[string]*models.Account)

	for i := 0; i < len(accounts); i++ {
		account := accounts[i]

		if account.Hidden {
			continue
		}

		if account.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
			continue
		}

		accountMap[account.Name] = account
	}
	return accountMap
}

// GetAccountNames returns a list with account names from account models list
func (s *AccountService) GetAccountNames(accounts []*models.Account) []string {
	accountNames := make([]string, len(accounts))

	for i := 0; i < len(accounts); i++ {
		accountNames[i] = accounts[i].Name
	}

	return accountNames
}

// GetAccountOrSubAccountIds returns a list of account ids or sub-account ids according to given account ids
func (s *AccountService) GetAccountOrSubAccountIds(c core.Context, accountIds string, uid int64) ([]int64, error) {
	if accountIds == "" || accountIds == "0" {
		return nil, nil
	}

	requestAccountIds, err := utils.StringArrayToInt64Array(strings.Split(accountIds, ","))

	if err != nil {
		return nil, errs.Or(err, errs.ErrAccountIdInvalid)
	}

	var allAccountIds []int64

	if len(requestAccountIds) > 0 {
		allSubAccounts, err := s.GetSubAccountsByAccountIds(c, uid, requestAccountIds)

		if err != nil {
			return nil, err
		}

		accountIdsMap := make(map[int64]int32, len(requestAccountIds))

		for i := 0; i < len(requestAccountIds); i++ {
			accountIdsMap[requestAccountIds[i]] = 0
		}

		for i := 0; i < len(allSubAccounts); i++ {
			subAccount := allSubAccounts[i]

			if refCount, exists := accountIdsMap[subAccount.ParentAccountId]; exists {
				accountIdsMap[subAccount.ParentAccountId] = refCount + 1
			} else {
				accountIdsMap[subAccount.ParentAccountId] = 1
			}

			if _, exists := accountIdsMap[subAccount.AccountId]; exists {
				delete(accountIdsMap, subAccount.AccountId)
			}

			allAccountIds = append(allAccountIds, subAccount.AccountId)
		}

		for accountId, refCount := range accountIdsMap {
			if refCount < 1 {
				allAccountIds = append(allAccountIds, accountId)
			}
		}
	}

	return allAccountIds, nil
}

// GetAccountOrSubAccountIdsByAccountName returns a list of account ids or sub-account ids according to given account name
func (s *AccountService) GetAccountOrSubAccountIdsByAccountName(accounts []*models.Account, accountName string) []int64 {
	accountIds := make([]int64, 0)
	parentAccountIds := make([]int64, 0)
	childAccountByParentAccountId := make(map[int64][]*models.Account)

	for i := 0; i < len(accounts); i++ {
		account := accounts[i]

		if account.Name == accountName {
			switch account.Type {
			case models.ACCOUNT_TYPE_SINGLE_ACCOUNT:
				accountIds = append(accountIds, account.AccountId)
			case models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS:
				parentAccountIds = append(parentAccountIds, account.AccountId)
			}
		} else if account.ParentAccountId > 0 {
			childAccounts, exists := childAccountByParentAccountId[account.ParentAccountId]

			if !exists {
				childAccounts = make([]*models.Account, 0)
			}

			childAccounts = append(childAccounts, account)
			childAccountByParentAccountId[account.ParentAccountId] = childAccounts
		}
	}

	for i := 0; i < len(parentAccountIds); i++ {
		parentAccountId := parentAccountIds[i]

		if childAccounts, exists := childAccountByParentAccountId[parentAccountId]; exists {
			for j := 0; j < len(childAccounts); j++ {
				childAccount := childAccounts[j]
				accountIds = append(accountIds, childAccount.AccountId)
			}
		}
	}

	return accountIds
}
