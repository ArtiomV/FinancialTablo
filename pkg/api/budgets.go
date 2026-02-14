package api

import (
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// BudgetsApi represents budgets api
type BudgetsApi struct {
	budgets    *services.BudgetService
	categories *services.TransactionCategoryService
}

// Initialize a budgets api singleton instance
var (
	BudgetsAPI = &BudgetsApi{
		budgets:    services.Budgets,
		categories: services.TransactionCategories,
	}
)

// BudgetListHandler returns budget list for given period
func (a *BudgetsApi) BudgetListHandler(c *core.WebContext) (any, *errs.Error) {
	var budgetListReq models.BudgetListRequest
	err := c.ShouldBindQuery(&budgetListReq)

	if err != nil {
		log.Warnf(c, "[budgets.BudgetListHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	budgets, err := a.budgets.GetBudgetsByYearMonth(c, uid, budgetListReq.Year, budgetListReq.Month, budgetListReq.CfoId)

	if err != nil {
		log.Errorf(c, "[budgets.BudgetListHandler] failed to get budgets for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	budgetResps := make([]*models.BudgetInfoResponse, len(budgets))

	for i := 0; i < len(budgets); i++ {
		budgetResps[i] = budgets[i].ToBudgetInfoResponse()
	}

	return budgetResps, nil
}

// BudgetSaveHandler saves budgets in bulk
func (a *BudgetsApi) BudgetSaveHandler(c *core.WebContext) (any, *errs.Error) {
	var budgetSaveReq models.BudgetSaveRequest
	err := c.ShouldBindJSON(&budgetSaveReq)

	if err != nil {
		log.Warnf(c, "[budgets.BudgetSaveHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	err = a.budgets.SaveBudgets(c, uid, budgetSaveReq.Year, budgetSaveReq.Month, budgetSaveReq.CfoId, budgetSaveReq.Budgets)

	if err != nil {
		log.Errorf(c, "[budgets.BudgetSaveHandler] failed to save budgets for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[budgets.BudgetSaveHandler] user \"uid:%d\" has saved budgets for %d-%02d", uid, budgetSaveReq.Year, budgetSaveReq.Month)

	return true, nil
}

// PlanFactHandler returns plan-fact analysis
func (a *BudgetsApi) PlanFactHandler(c *core.WebContext) (any, *errs.Error) {
	var planFactReq models.PlanFactRequest
	err := c.ShouldBindQuery(&planFactReq)

	if err != nil {
		log.Warnf(c, "[budgets.PlanFactHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	// Calculate time range for the month (UTC)
	startTime := time.Date(int(planFactReq.Year), time.Month(planFactReq.Month), 1, 0, 0, 0, 0, time.UTC).Unix()
	endTime := time.Date(int(planFactReq.Year), time.Month(planFactReq.Month)+1, 1, 0, 0, 0, 0, time.UTC).Unix()

	// Get budgets
	budgets, err := a.budgets.GetBudgetsByYearMonth(c, uid, planFactReq.Year, planFactReq.Month, planFactReq.CfoId)

	if err != nil {
		log.Errorf(c, "[budgets.PlanFactHandler] failed to get budgets for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	// Get fact amounts
	factMap, err := a.budgets.GetFactAmountsByYearMonth(c, uid, startTime, endTime, planFactReq.CfoId)

	if err != nil {
		log.Errorf(c, "[budgets.PlanFactHandler] failed to get fact amounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	// Get all categories to resolve names
	allCategories, err := a.categories.GetAllCategoriesByUid(c, uid, 0)

	if err != nil {
		log.Errorf(c, "[budgets.PlanFactHandler] failed to get categories for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	categoryMap := make(map[int64]*models.TransactionCategory)
	for _, cat := range allCategories {
		categoryMap[cat.CategoryId] = cat
	}

	// Collect all unique category IDs
	categoryIds := make(map[int64]bool)
	for _, b := range budgets {
		categoryIds[b.CategoryId] = true
	}
	for catId := range factMap {
		categoryIds[catId] = true
	}

	// Build budget map
	budgetMap := make(map[int64]int64)
	for _, b := range budgets {
		budgetMap[b.CategoryId] = b.PlannedAmount
	}

	// Build plan-fact lines
	lines := make([]*models.PlanFactLineResponse, 0)

	for catId := range categoryIds {
		planned := budgetMap[catId]
		fact := factMap[catId]
		deviation := fact - planned

		var deviationPct *int32
		if planned != 0 {
			pct := int32(float64(deviation) / float64(planned) * 100)
			deviationPct = &pct
		}

		catName := ""
		var catType int32
		if cat, ok := categoryMap[catId]; ok {
			catName = cat.Name
			catType = int32(cat.Type)
		}

		lines = append(lines, &models.PlanFactLineResponse{
			CategoryId:    catId,
			CategoryName:  catName,
			CategoryType:  catType,
			PlannedAmount: planned,
			FactAmount:    fact,
			Deviation:     deviation,
			DeviationPct:  deviationPct,
		})
	}

	return &models.PlanFactResponse{
		Year:  planFactReq.Year,
		Month: planFactReq.Month,
		CfoId: planFactReq.CfoId,
		Lines: lines,
	}, nil
}
