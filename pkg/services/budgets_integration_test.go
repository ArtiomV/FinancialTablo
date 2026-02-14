package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func newTestBudgetService(t *testing.T) (*BudgetService, *testDB) {
	t.Helper()
	tdb := newTestDB(t)
	uuidContainer := initUuidContainer(t)
	svc := &BudgetService{
		ServiceUsingDB:   ServiceUsingDB{container: tdb.container},
		ServiceUsingUuid: ServiceUsingUuid{container: uuidContainer},
	}
	return svc, tdb
}

func TestBudgetServiceSaveAndGet(t *testing.T) {
	svc, tdb := newTestBudgetService(t)
	defer tdb.close()

	items := []*models.BudgetItemRequest{
		{CategoryId: 100, PlannedAmount: 50000, Comment: "Salary"},
		{CategoryId: 200, PlannedAmount: 30000, Comment: "Rent"},
	}

	err := svc.SaveBudgets(nil, 1, 2025, 6, 0, items)
	assert.Nil(t, err)

	budgets, err := svc.GetBudgetsByYearMonth(nil, 1, 2025, 6, 0)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(budgets))

	// Verify data
	budgetMap := make(map[int64]*models.Budget)
	for _, b := range budgets {
		budgetMap[b.CategoryId] = b
	}

	assert.Equal(t, int64(50000), budgetMap[100].PlannedAmount)
	assert.Equal(t, "Salary", budgetMap[100].Comment)
	assert.Equal(t, int64(30000), budgetMap[200].PlannedAmount)
	assert.Equal(t, "Rent", budgetMap[200].Comment)
}

func TestBudgetServiceSaveUpsert(t *testing.T) {
	svc, tdb := newTestBudgetService(t)
	defer tdb.close()

	// Create initial budget
	items := []*models.BudgetItemRequest{
		{CategoryId: 100, PlannedAmount: 50000, Comment: "Initial"},
	}
	assert.Nil(t, svc.SaveBudgets(nil, 1, 2025, 7, 0, items))

	// Update with upsert
	updatedItems := []*models.BudgetItemRequest{
		{CategoryId: 100, PlannedAmount: 75000, Comment: "Updated"},
		{CategoryId: 200, PlannedAmount: 20000, Comment: "New line"},
	}
	assert.Nil(t, svc.SaveBudgets(nil, 1, 2025, 7, 0, updatedItems))

	budgets, err := svc.GetBudgetsByYearMonth(nil, 1, 2025, 7, 0)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(budgets))

	budgetMap := make(map[int64]*models.Budget)
	for _, b := range budgets {
		budgetMap[b.CategoryId] = b
	}

	assert.Equal(t, int64(75000), budgetMap[100].PlannedAmount)
	assert.Equal(t, "Updated", budgetMap[100].Comment)
	assert.Equal(t, int64(20000), budgetMap[200].PlannedAmount)
	assert.Equal(t, "New line", budgetMap[200].Comment)
}

func TestBudgetServiceGetByYearMonthEmpty(t *testing.T) {
	svc, tdb := newTestBudgetService(t)
	defer tdb.close()

	budgets, err := svc.GetBudgetsByYearMonth(nil, 1, 2025, 12, 0)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(budgets))
}

func TestBudgetServiceInvalidUid(t *testing.T) {
	svc, tdb := newTestBudgetService(t)
	defer tdb.close()

	_, err := svc.GetBudgetsByYearMonth(nil, 0, 2025, 6, 0)
	assert.Equal(t, errs.ErrUserIdInvalid, err)

	err = svc.SaveBudgets(nil, 0, 2025, 6, 0, []*models.BudgetItemRequest{})
	assert.Equal(t, errs.ErrUserIdInvalid, err)
}

func TestBudgetServiceUserIsolation(t *testing.T) {
	svc, tdb := newTestBudgetService(t)
	defer tdb.close()

	items1 := []*models.BudgetItemRequest{
		{CategoryId: 100, PlannedAmount: 40000, Comment: "user1budget"},
	}
	items2 := []*models.BudgetItemRequest{
		{CategoryId: 100, PlannedAmount: 60000, Comment: "user2budget"},
	}

	assert.Nil(t, svc.SaveBudgets(nil, 1, 2025, 8, 0, items1))
	assert.Nil(t, svc.SaveBudgets(nil, 2, 2025, 8, 0, items2))

	list1, err := svc.GetBudgetsByYearMonth(nil, 1, 2025, 8, 0)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list1))
	assert.Equal(t, "user1budget", list1[0].Comment)

	list2, err := svc.GetBudgetsByYearMonth(nil, 2, 2025, 8, 0)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list2))
	assert.Equal(t, "user2budget", list2[0].Comment)
}

func TestBudgetServiceFilterByCfoId(t *testing.T) {
	svc, tdb := newTestBudgetService(t)
	defer tdb.close()

	items1 := []*models.BudgetItemRequest{
		{CategoryId: 100, PlannedAmount: 10000, Comment: "cfo1"},
	}
	items2 := []*models.BudgetItemRequest{
		{CategoryId: 100, PlannedAmount: 20000, Comment: "cfo2"},
	}

	assert.Nil(t, svc.SaveBudgets(nil, 1, 2025, 9, 10, items1))
	assert.Nil(t, svc.SaveBudgets(nil, 1, 2025, 9, 20, items2))

	// Filter by cfoId=10
	list, err := svc.GetBudgetsByYearMonth(nil, 1, 2025, 9, 10)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list))
	assert.Equal(t, "cfo1", list[0].Comment)

	// Get all (cfoId=0 means no filter)
	all, err := svc.GetBudgetsByYearMonth(nil, 1, 2025, 9, 0)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(all))
}
