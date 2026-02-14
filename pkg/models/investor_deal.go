package models

// InvestorDealType represents deal type
type InvestorDealType byte

// Investor deal types
const (
	INVESTOR_DEAL_TYPE_LOAN          InvestorDealType = 1
	INVESTOR_DEAL_TYPE_EQUITY        InvestorDealType = 2
	INVESTOR_DEAL_TYPE_REVENUE_SHARE InvestorDealType = 3
	INVESTOR_DEAL_TYPE_OTHER         InvestorDealType = 4
)

// InvestorDeal represents investor deal data stored in database
type InvestorDeal struct {
	DealId             int64            `xorm:"PK"`
	Uid                int64            `xorm:"INDEX(IDX_investor_deal_uid_deleted) NOT NULL"`
	Deleted            bool             `xorm:"INDEX(IDX_investor_deal_uid_deleted) NOT NULL"`
	InvestorName       string           `xorm:"VARCHAR(64) NOT NULL"`
	CfoId              int64            `xorm:"NOT NULL DEFAULT 0"`
	InvestmentDate     int64            `xorm:"NOT NULL DEFAULT 0"`
	InvestmentAmount   int64            `xorm:"NOT NULL DEFAULT 0"`
	Currency           string           `xorm:"VARCHAR(3) NOT NULL DEFAULT 'RUB'"`
	DealType           InvestorDealType `xorm:"NOT NULL DEFAULT 1"`
	AnnualRate         int32            `xorm:"NOT NULL DEFAULT 0"`
	ProfitSharePct     int32            `xorm:"NOT NULL DEFAULT 0"`
	FixedPayment       int64            `xorm:"NOT NULL DEFAULT 0"`
	RepaymentStartDate int64            `xorm:"NOT NULL DEFAULT 0"`
	RepaymentEndDate   int64            `xorm:"NOT NULL DEFAULT 0"`
	TotalToRepay       int64            `xorm:"NOT NULL DEFAULT 0"`
	Comment            string           `xorm:"VARCHAR(255) NOT NULL DEFAULT ''"`
	CreatedUnixTime    int64
	UpdatedUnixTime    int64
	DeletedUnixTime    int64
}

// InvestorDealGetRequest represents all parameters of deal getting request
type InvestorDealGetRequest struct {
	Id int64 `form:"id,string" binding:"required,min=1"`
}

// InvestorDealCreateRequest represents all parameters of deal creation request
type InvestorDealCreateRequest struct {
	InvestorName       string           `json:"investorName" binding:"required,notBlank,max=64"`
	CfoId              int64            `json:"cfoId,string"`
	InvestmentDate     int64            `json:"investmentDate"`
	InvestmentAmount   int64            `json:"investmentAmount"`
	Currency           string           `json:"currency" binding:"required,max=3"`
	DealType           InvestorDealType `json:"dealType"`
	AnnualRate         int32            `json:"annualRate"`
	ProfitSharePct     int32            `json:"profitSharePct"`
	FixedPayment       int64            `json:"fixedPayment"`
	RepaymentStartDate int64            `json:"repaymentStartDate"`
	RepaymentEndDate   int64            `json:"repaymentEndDate"`
	TotalToRepay       int64            `json:"totalToRepay"`
	Comment            string           `json:"comment" binding:"max=255"`
}

// InvestorDealModifyRequest represents all parameters of deal modification request
type InvestorDealModifyRequest struct {
	Id                 int64            `json:"id,string" binding:"required,min=1"`
	InvestorName       string           `json:"investorName" binding:"required,notBlank,max=64"`
	CfoId              int64            `json:"cfoId,string"`
	InvestmentDate     int64            `json:"investmentDate"`
	InvestmentAmount   int64            `json:"investmentAmount"`
	Currency           string           `json:"currency" binding:"required,max=3"`
	DealType           InvestorDealType `json:"dealType"`
	AnnualRate         int32            `json:"annualRate"`
	ProfitSharePct     int32            `json:"profitSharePct"`
	FixedPayment       int64            `json:"fixedPayment"`
	RepaymentStartDate int64            `json:"repaymentStartDate"`
	RepaymentEndDate   int64            `json:"repaymentEndDate"`
	TotalToRepay       int64            `json:"totalToRepay"`
	Comment            string           `json:"comment" binding:"max=255"`
}

// InvestorDealDeleteRequest represents all parameters of deal deleting request
type InvestorDealDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

// InvestorDealInfoResponse represents a view-object of investor deal
type InvestorDealInfoResponse struct {
	Id                 int64            `json:"id,string"`
	InvestorName       string           `json:"investorName"`
	CfoId              int64            `json:"cfoId,string"`
	InvestmentDate     int64            `json:"investmentDate"`
	InvestmentAmount   int64            `json:"investmentAmount"`
	Currency           string           `json:"currency"`
	DealType           InvestorDealType `json:"dealType"`
	AnnualRate         int32            `json:"annualRate"`
	ProfitSharePct     int32            `json:"profitSharePct"`
	FixedPayment       int64            `json:"fixedPayment"`
	RepaymentStartDate int64            `json:"repaymentStartDate"`
	RepaymentEndDate   int64            `json:"repaymentEndDate"`
	TotalToRepay       int64            `json:"totalToRepay"`
	Comment            string           `json:"comment"`
}

// ToInvestorDealInfoResponse returns a view-object according to database model
func (d *InvestorDeal) ToInvestorDealInfoResponse() *InvestorDealInfoResponse {
	return &InvestorDealInfoResponse{
		Id:                 d.DealId,
		InvestorName:       d.InvestorName,
		CfoId:              d.CfoId,
		InvestmentDate:     d.InvestmentDate,
		InvestmentAmount:   d.InvestmentAmount,
		Currency:           d.Currency,
		DealType:           d.DealType,
		AnnualRate:         d.AnnualRate,
		ProfitSharePct:     d.ProfitSharePct,
		FixedPayment:       d.FixedPayment,
		RepaymentStartDate: d.RepaymentStartDate,
		RepaymentEndDate:   d.RepaymentEndDate,
		TotalToRepay:       d.TotalToRepay,
		Comment:            d.Comment,
	}
}

// InvestorDealInfoResponseSlice represents the slice data structure of InvestorDealInfoResponse
type InvestorDealInfoResponseSlice []*InvestorDealInfoResponse

// Len returns the count of items
func (s InvestorDealInfoResponseSlice) Len() int {
	return len(s)
}

// Swap swaps two items
func (s InvestorDealInfoResponseSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first item is less than the second one (by investment date desc)
func (s InvestorDealInfoResponseSlice) Less(i, j int) bool {
	return s[i].InvestmentDate > s[j].InvestmentDate
}
