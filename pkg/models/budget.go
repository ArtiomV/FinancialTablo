package models

// Budget represents budget data stored in database
type Budget struct {
	BudgetId        int64  `xorm:"PK"`
	Uid             int64  `xorm:"INDEX(IDX_budget_uid_deleted_year_month) NOT NULL"`
	Deleted         bool   `xorm:"INDEX(IDX_budget_uid_deleted_year_month) NOT NULL"`
	CfoId           int64  `xorm:"NOT NULL DEFAULT 0"`
	CategoryId      int64  `xorm:"NOT NULL DEFAULT 0"`
	Year            int32  `xorm:"INDEX(IDX_budget_uid_deleted_year_month) NOT NULL"`
	Month           int32  `xorm:"INDEX(IDX_budget_uid_deleted_year_month) NOT NULL"`
	PlannedAmount   int64  `xorm:"NOT NULL DEFAULT 0"`
	Comment         string `xorm:"VARCHAR(255) NOT NULL DEFAULT ''"`
	CreatedUnixTime int64
	UpdatedUnixTime int64
	DeletedUnixTime int64
}

// BudgetListRequest represents parameters to list budgets
type BudgetListRequest struct {
	Year  int32 `form:"year" binding:"required,min=2000,max=2100"`
	Month int32 `form:"month" binding:"required,min=1,max=12"`
	CfoId int64 `form:"cfoId,string"`
}

// BudgetSaveRequest represents parameters to save budgets (bulk)
type BudgetSaveRequest struct {
	Year    int32               `json:"year" binding:"required,min=2000,max=2100"`
	Month   int32               `json:"month" binding:"required,min=1,max=12"`
	CfoId   int64               `json:"cfoId,string"`
	Budgets []*BudgetItemRequest `json:"budgets" binding:"required"`
}

// BudgetItemRequest represents a single budget line
type BudgetItemRequest struct {
	CategoryId    int64  `json:"categoryId,string" binding:"required,min=1"`
	PlannedAmount int64  `json:"plannedAmount"`
	Comment       string `json:"comment" binding:"max=255"`
}

// PlanFactRequest represents parameters for plan-fact analysis
type PlanFactRequest struct {
	Year  int32 `form:"year" binding:"required,min=2000,max=2100"`
	Month int32 `form:"month" binding:"required,min=1,max=12"`
	CfoId int64 `form:"cfoId,string"`
}

// BudgetInfoResponse represents a view-object of budget
type BudgetInfoResponse struct {
	Id            int64  `json:"id,string"`
	CfoId         int64  `json:"cfoId,string"`
	CategoryId    int64  `json:"categoryId,string"`
	Year          int32  `json:"year"`
	Month         int32  `json:"month"`
	PlannedAmount int64  `json:"plannedAmount"`
	Comment       string `json:"comment"`
}

// ToBudgetInfoResponse returns a view-object according to database model
func (b *Budget) ToBudgetInfoResponse() *BudgetInfoResponse {
	return &BudgetInfoResponse{
		Id:            b.BudgetId,
		CfoId:         b.CfoId,
		CategoryId:    b.CategoryId,
		Year:          b.Year,
		Month:         b.Month,
		PlannedAmount: b.PlannedAmount,
		Comment:       b.Comment,
	}
}

// PlanFactLineResponse represents a plan-fact line for one category
type PlanFactLineResponse struct {
	CategoryId    int64  `json:"categoryId,string"`
	CategoryName  string `json:"categoryName"`
	CategoryType  int32  `json:"categoryType"`
	PlannedAmount int64  `json:"plannedAmount"`
	FactAmount    int64  `json:"factAmount"`
	Deviation     int64  `json:"deviation"`
	DeviationPct  *int32 `json:"deviationPct"`
}

// PlanFactResponse represents plan-fact analysis result
type PlanFactResponse struct {
	Year  int32                   `json:"year"`
	Month int32                   `json:"month"`
	CfoId int64                   `json:"cfoId,string"`
	Lines []*PlanFactLineResponse `json:"lines"`
}
