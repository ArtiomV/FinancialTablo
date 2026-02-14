package models

// TaxType represents tax type
type TaxType byte

const (
	TAX_TYPE_INCOME    TaxType = 1
	TAX_TYPE_VAT       TaxType = 2
	TAX_TYPE_PROPERTY  TaxType = 3
	TAX_TYPE_OTHER     TaxType = 4
)

// TaxStatus represents tax status
type TaxStatus byte

const (
	TAX_STATUS_PENDING TaxStatus = 1
	TAX_STATUS_PAID    TaxStatus = 2
	TAX_STATUS_OVERDUE TaxStatus = 3
)

// TaxRecord represents tax record data stored in database
type TaxRecord struct {
	TaxId           int64     `xorm:"PK"`
	Uid             int64     `xorm:"INDEX(IDX_tax_record_uid_deleted) NOT NULL"`
	Deleted         bool      `xorm:"INDEX(IDX_tax_record_uid_deleted) NOT NULL"`
	CfoId           int64     `xorm:"NOT NULL DEFAULT 0"`
	TaxType         TaxType   `xorm:"NOT NULL DEFAULT 1"`
	PeriodYear      int32     `xorm:"NOT NULL DEFAULT 0"`
	PeriodQuarter   int32     `xorm:"NOT NULL DEFAULT 0"`
	TaxableIncome   int64     `xorm:"NOT NULL DEFAULT 0"`
	TaxAmount       int64     `xorm:"NOT NULL DEFAULT 0"`
	PaidAmount      int64     `xorm:"NOT NULL DEFAULT 0"`
	DueDate         int64     `xorm:"NOT NULL DEFAULT 0"`
	Status          TaxStatus `xorm:"NOT NULL DEFAULT 1"`
	Comment         string    `xorm:"VARCHAR(255) NOT NULL DEFAULT ''"`
	CreatedUnixTime int64
	UpdatedUnixTime int64
	DeletedUnixTime int64
}

type TaxRecordGetRequest struct {
	Id int64 `form:"id,string" binding:"required,min=1"`
}

type TaxRecordCreateRequest struct {
	CfoId         int64     `json:"cfoId,string"`
	TaxType       TaxType   `json:"taxType"`
	PeriodYear    int32     `json:"periodYear"`
	PeriodQuarter int32     `json:"periodQuarter"`
	TaxableIncome int64     `json:"taxableIncome"`
	TaxAmount     int64     `json:"taxAmount"`
	PaidAmount    int64     `json:"paidAmount"`
	DueDate       int64     `json:"dueDate"`
	Status        TaxStatus `json:"status"`
	Comment       string    `json:"comment" binding:"max=255"`
}

type TaxRecordModifyRequest struct {
	Id            int64     `json:"id,string" binding:"required,min=1"`
	CfoId         int64     `json:"cfoId,string"`
	TaxType       TaxType   `json:"taxType"`
	PeriodYear    int32     `json:"periodYear"`
	PeriodQuarter int32     `json:"periodQuarter"`
	TaxableIncome int64     `json:"taxableIncome"`
	TaxAmount     int64     `json:"taxAmount"`
	PaidAmount    int64     `json:"paidAmount"`
	DueDate       int64     `json:"dueDate"`
	Status        TaxStatus `json:"status"`
	Comment       string    `json:"comment" binding:"max=255"`
}

type TaxRecordDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

type TaxRecordInfoResponse struct {
	Id            int64     `json:"id,string"`
	CfoId         int64     `json:"cfoId,string"`
	TaxType       TaxType   `json:"taxType"`
	PeriodYear    int32     `json:"periodYear"`
	PeriodQuarter int32     `json:"periodQuarter"`
	TaxableIncome int64     `json:"taxableIncome"`
	TaxAmount     int64     `json:"taxAmount"`
	PaidAmount    int64     `json:"paidAmount"`
	DueDate       int64     `json:"dueDate"`
	Status        TaxStatus `json:"status"`
	Comment       string    `json:"comment"`
}

func (t *TaxRecord) ToTaxRecordInfoResponse() *TaxRecordInfoResponse {
	return &TaxRecordInfoResponse{
		Id:            t.TaxId,
		CfoId:         t.CfoId,
		TaxType:       t.TaxType,
		PeriodYear:    t.PeriodYear,
		PeriodQuarter: t.PeriodQuarter,
		TaxableIncome: t.TaxableIncome,
		TaxAmount:     t.TaxAmount,
		PaidAmount:    t.PaidAmount,
		DueDate:       t.DueDate,
		Status:        t.Status,
		Comment:       t.Comment,
	}
}
