package models

// ObligationType represents obligation type
type ObligationType byte

const (
	OBLIGATION_TYPE_RECEIVABLE ObligationType = 1
	OBLIGATION_TYPE_PAYABLE    ObligationType = 2
)

// ObligationStatus represents obligation status
type ObligationStatus byte

const (
	OBLIGATION_STATUS_ACTIVE   ObligationStatus = 1
	OBLIGATION_STATUS_PARTIAL  ObligationStatus = 2
	OBLIGATION_STATUS_PAID     ObligationStatus = 3
)

// Obligation represents obligation data stored in database
type Obligation struct {
	ObligationId    int64            `xorm:"PK"`
	Uid             int64            `xorm:"INDEX(IDX_obligation_uid_deleted) NOT NULL"`
	Deleted         bool             `xorm:"INDEX(IDX_obligation_uid_deleted) NOT NULL"`
	ObligationType  ObligationType   `xorm:"NOT NULL DEFAULT 1"`
	CounterpartyId  int64            `xorm:"NOT NULL DEFAULT 0"`
	CfoId           int64            `xorm:"NOT NULL DEFAULT 0"`
	Amount          int64            `xorm:"NOT NULL DEFAULT 0"`
	Currency        string           `xorm:"VARCHAR(3) NOT NULL DEFAULT 'RUB'"`
	DueDate         int64            `xorm:"NOT NULL DEFAULT 0"`
	Status          ObligationStatus `xorm:"NOT NULL DEFAULT 1"`
	PaidAmount      int64            `xorm:"NOT NULL DEFAULT 0"`
	Comment         string           `xorm:"VARCHAR(255) NOT NULL DEFAULT ''"`
	CreatedUnixTime int64
	UpdatedUnixTime int64
	DeletedUnixTime int64
}

type ObligationGetRequest struct {
	Id int64 `form:"id,string" binding:"required,min=1"`
}

type ObligationCreateRequest struct {
	ObligationType ObligationType   `json:"obligationType"`
	CounterpartyId int64            `json:"counterpartyId,string"`
	CfoId          int64            `json:"cfoId,string"`
	Amount         int64            `json:"amount"`
	Currency       string           `json:"currency" binding:"required,max=3"`
	DueDate        int64            `json:"dueDate"`
	Status         ObligationStatus `json:"status"`
	PaidAmount     int64            `json:"paidAmount"`
	Comment        string           `json:"comment" binding:"max=255"`
}

type ObligationModifyRequest struct {
	Id             int64            `json:"id,string" binding:"required,min=1"`
	ObligationType ObligationType   `json:"obligationType"`
	CounterpartyId int64            `json:"counterpartyId,string"`
	CfoId          int64            `json:"cfoId,string"`
	Amount         int64            `json:"amount"`
	Currency       string           `json:"currency" binding:"required,max=3"`
	DueDate        int64            `json:"dueDate"`
	Status         ObligationStatus `json:"status"`
	PaidAmount     int64            `json:"paidAmount"`
	Comment        string           `json:"comment" binding:"max=255"`
}

type ObligationDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

type ObligationInfoResponse struct {
	Id             int64            `json:"id,string"`
	ObligationType ObligationType   `json:"obligationType"`
	CounterpartyId int64            `json:"counterpartyId,string"`
	CfoId          int64            `json:"cfoId,string"`
	Amount         int64            `json:"amount"`
	Currency       string           `json:"currency"`
	DueDate        int64            `json:"dueDate"`
	Status         ObligationStatus `json:"status"`
	PaidAmount     int64            `json:"paidAmount"`
	Comment        string           `json:"comment"`
}

func (o *Obligation) ToObligationInfoResponse() *ObligationInfoResponse {
	return &ObligationInfoResponse{
		Id:             o.ObligationId,
		ObligationType: o.ObligationType,
		CounterpartyId: o.CounterpartyId,
		CfoId:          o.CfoId,
		Amount:         o.Amount,
		Currency:       o.Currency,
		DueDate:        o.DueDate,
		Status:         o.Status,
		PaidAmount:     o.PaidAmount,
		Comment:        o.Comment,
	}
}
