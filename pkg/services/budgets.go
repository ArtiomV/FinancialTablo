// budgets.go provides CRUD operations for budget entries and plan-fact analysis.
package services

import (
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// BudgetService represents budget service
type BudgetService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a budget service singleton instance
var (
	Budgets = &BudgetService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetBudgetsByYearMonth returns budget models for given year+month+cfo
func (s *BudgetService) GetBudgetsByYearMonth(c core.Context, uid int64, year int32, month int32, cfoId int64) ([]*models.Budget, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var budgets []*models.Budget
	sess := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=? AND year=? AND month=?", uid, false, year, month)

	if cfoId > 0 {
		sess = sess.And("cfo_id=?", cfoId)
	}

	err := sess.Find(&budgets)

	return budgets, err
}

// SaveBudgets saves budgets in bulk (upsert for year+month+cfoId+categoryId)
func (s *BudgetService) SaveBudgets(c core.Context, uid int64, year int32, month int32, cfoId int64, items []*models.BudgetItemRequest) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		// Get existing budgets for this period
		var existing []*models.Budget
		existQuery := sess.Where("uid=? AND deleted=? AND year=? AND month=? AND cfo_id=?", uid, false, year, month, cfoId)
		err := existQuery.Find(&existing)

		if err != nil {
			return err
		}

		// Map existing by categoryId
		existingMap := make(map[int64]*models.Budget)
		for _, b := range existing {
			existingMap[b.CategoryId] = b
		}

		now := time.Now().Unix()

		for _, item := range items {
			if existBudget, ok := existingMap[item.CategoryId]; ok {
				// Update existing
				existBudget.PlannedAmount = item.PlannedAmount
				existBudget.Comment = item.Comment
				existBudget.UpdatedUnixTime = now

				_, err := sess.ID(existBudget.BudgetId).Cols("planned_amount", "comment", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(existBudget)
				if err != nil {
					return err
				}
			} else {
				// Create new
				newBudget := &models.Budget{
					BudgetId:        s.GenerateUuid(uuid.UUID_TYPE_DEFAULT),
					Uid:             uid,
					Deleted:         false,
					CfoId:           cfoId,
					CategoryId:      item.CategoryId,
					Year:            year,
					Month:           month,
					PlannedAmount:   item.PlannedAmount,
					Comment:         item.Comment,
					CreatedUnixTime: now,
					UpdatedUnixTime: now,
				}

				if newBudget.BudgetId < 1 {
					return errs.ErrSystemIsBusy
				}

				_, err := sess.Insert(newBudget)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// GetFactAmountsByYearMonth returns fact (actual) amounts grouped by categoryId for given period
func (s *BudgetService) GetFactAmountsByYearMonth(c core.Context, uid int64, startTime int64, endTime int64, cfoId int64) (map[int64]int64, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	type FactResult struct {
		CategoryId int64 `xorm:"category_id"`
		TotalAmount int64 `xorm:"total_amount"`
	}

	var results []FactResult
	sess := s.UserDataDB(uid).NewSession(c).
		Table("\"transaction\"").
		Select("category_id, SUM(amount) as total_amount").
		Where("uid=? AND deleted=? AND transaction_time>=? AND transaction_time<?", uid, false, startTime, endTime)

	if cfoId > 0 {
		sess = sess.And("cfo_id=?", cfoId)
	}

	err := sess.GroupBy("category_id").Find(&results)

	if err != nil {
		return nil, err
	}

	factMap := make(map[int64]int64)
	for _, r := range results {
		factMap[r.CategoryId] = r.TotalAmount
	}

	return factMap, nil
}
