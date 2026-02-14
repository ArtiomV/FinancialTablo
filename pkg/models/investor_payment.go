package models

// InvestorPaymentType represents payment type
type InvestorPaymentType byte

// Investor payment types
const (
	INVESTOR_PAYMENT_TYPE_PRINCIPAL InvestorPaymentType = 1
	INVESTOR_PAYMENT_TYPE_INTEREST InvestorPaymentType = 2
	INVESTOR_PAYMENT_TYPE_MIXED    InvestorPaymentType = 3
)

// InvestorPayment represents investor payment data stored in database
type InvestorPayment struct {
	PaymentId       int64               `xorm:"PK"`
	Uid             int64               `xorm:"INDEX(IDX_investor_payment_uid_deleted) NOT NULL"`
	Deleted         bool                `xorm:"INDEX(IDX_investor_payment_uid_deleted) NOT NULL"`
	DealId          int64               `xorm:"INDEX NOT NULL"`
	PaymentDate     int64               `xorm:"NOT NULL DEFAULT 0"`
	Amount          int64               `xorm:"NOT NULL DEFAULT 0"`
	PaymentType     InvestorPaymentType `xorm:"NOT NULL DEFAULT 1"`
	TransactionId   int64               `xorm:"NOT NULL DEFAULT 0"`
	Comment         string              `xorm:"VARCHAR(255) NOT NULL DEFAULT ''"`
	CreatedUnixTime int64
	UpdatedUnixTime int64
	DeletedUnixTime int64
}

// InvestorPaymentGetRequest represents all parameters of payment getting request
type InvestorPaymentGetRequest struct {
	Id int64 `form:"id,string" binding:"required,min=1"`
}

// InvestorPaymentListByDealRequest represents all parameters to list payments by deal
type InvestorPaymentListByDealRequest struct {
	DealId int64 `form:"dealId,string" binding:"required,min=1"`
}

// InvestorPaymentCreateRequest represents all parameters of payment creation request
type InvestorPaymentCreateRequest struct {
	DealId      int64               `json:"dealId,string" binding:"required,min=1"`
	PaymentDate int64               `json:"paymentDate"`
	Amount      int64               `json:"amount"`
	PaymentType InvestorPaymentType `json:"paymentType"`
	TransactionId int64             `json:"transactionId,string"`
	Comment     string              `json:"comment" binding:"max=255"`
}

// InvestorPaymentModifyRequest represents all parameters of payment modification request
type InvestorPaymentModifyRequest struct {
	Id          int64               `json:"id,string" binding:"required,min=1"`
	DealId      int64               `json:"dealId,string" binding:"required,min=1"`
	PaymentDate int64               `json:"paymentDate"`
	Amount      int64               `json:"amount"`
	PaymentType InvestorPaymentType `json:"paymentType"`
	TransactionId int64             `json:"transactionId,string"`
	Comment     string              `json:"comment" binding:"max=255"`
}

// InvestorPaymentDeleteRequest represents all parameters of payment deleting request
type InvestorPaymentDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

// InvestorPaymentInfoResponse represents a view-object of investor payment
type InvestorPaymentInfoResponse struct {
	Id            int64               `json:"id,string"`
	DealId        int64               `json:"dealId,string"`
	PaymentDate   int64               `json:"paymentDate"`
	Amount        int64               `json:"amount"`
	PaymentType   InvestorPaymentType `json:"paymentType"`
	TransactionId int64               `json:"transactionId,string"`
	Comment       string              `json:"comment"`
}

// ToInvestorPaymentInfoResponse returns a view-object according to database model
func (p *InvestorPayment) ToInvestorPaymentInfoResponse() *InvestorPaymentInfoResponse {
	return &InvestorPaymentInfoResponse{
		Id:            p.PaymentId,
		DealId:        p.DealId,
		PaymentDate:   p.PaymentDate,
		Amount:        p.Amount,
		PaymentType:   p.PaymentType,
		TransactionId: p.TransactionId,
		Comment:       p.Comment,
	}
}

// InvestorPaymentInfoResponseSlice represents the slice data structure of InvestorPaymentInfoResponse
type InvestorPaymentInfoResponseSlice []*InvestorPaymentInfoResponse

// Len returns the count of items
func (s InvestorPaymentInfoResponseSlice) Len() int {
	return len(s)
}

// Swap swaps two items
func (s InvestorPaymentInfoResponseSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first item is less than the second one (by payment date desc)
func (s InvestorPaymentInfoResponseSlice) Less(i, j int) bool {
	return s[i].PaymentDate > s[j].PaymentDate
}
