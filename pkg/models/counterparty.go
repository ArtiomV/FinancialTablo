package models

// CounterpartyType represents counterparty type
type CounterpartyType byte

// Counterparty types
const (
	COUNTERPARTY_TYPE_PERSON  CounterpartyType = 1
	COUNTERPARTY_TYPE_COMPANY CounterpartyType = 2
)

// Counterparty represents counterparty data stored in database
type Counterparty struct {
	CounterpartyId  int64            `xorm:"PK"`
	Uid             int64            `xorm:"INDEX(IDX_counterparty_uid_deleted_order) NOT NULL"`
	Deleted         bool             `xorm:"INDEX(IDX_counterparty_uid_deleted_order) NOT NULL"`
	Type            CounterpartyType `xorm:"NOT NULL"`
	Name            string           `xorm:"VARCHAR(64) NOT NULL"`
	Comment         string           `xorm:"VARCHAR(255) NOT NULL"`
	Icon            int64            `xorm:"NOT NULL"`
	Color           string           `xorm:"VARCHAR(6) NOT NULL"`
	DisplayOrder    int32            `xorm:"INDEX(IDX_counterparty_uid_deleted_order) NOT NULL"`
	Hidden          bool             `xorm:"NOT NULL"`
	CreatedUnixTime int64
	UpdatedUnixTime int64
	DeletedUnixTime int64
}

// CounterpartyListRequest represents all parameters of counterparty listing request
type CounterpartyListRequest struct {
	Type CounterpartyType `form:"type" binding:"min=0"`
}

// CounterpartyGetRequest represents all parameters of counterparty getting request
type CounterpartyGetRequest struct {
	Id int64 `form:"id,string" binding:"required,min=1"`
}

// CounterpartyCreateRequest represents all parameters of counterparty creation request
type CounterpartyCreateRequest struct {
	Name            string           `json:"name" binding:"required,notBlank,max=64"`
	Type            CounterpartyType `json:"type" binding:"required"`
	Icon            int64            `json:"icon,string" binding:"min=0"`
	Color           string           `json:"color" binding:"required,len=6,validHexRGBColor"`
	Comment         string           `json:"comment" binding:"max=255"`
	ClientSessionId string           `json:"clientSessionId"`
}

// CounterpartyModifyRequest represents all parameters of counterparty modification request
type CounterpartyModifyRequest struct {
	Id      int64            `json:"id,string" binding:"required,min=1"`
	Name    string           `json:"name" binding:"required,notBlank,max=64"`
	Type    CounterpartyType `json:"type" binding:"required"`
	Icon    int64            `json:"icon,string" binding:"min=0"`
	Color   string           `json:"color" binding:"required,len=6,validHexRGBColor"`
	Comment string           `json:"comment" binding:"max=255"`
	Hidden  bool             `json:"hidden"`
}

// CounterpartyHideRequest represents all parameters of counterparty hiding request
type CounterpartyHideRequest struct {
	Id     int64 `json:"id,string" binding:"required,min=1"`
	Hidden bool  `json:"hidden"`
}

// CounterpartyMoveRequest represents all parameters of counterparty moving request
type CounterpartyMoveRequest struct {
	NewDisplayOrders []*CounterpartyNewDisplayOrderRequest `json:"newDisplayOrders" binding:"required,min=1"`
}

// CounterpartyNewDisplayOrderRequest represents a data pair of id and display order
type CounterpartyNewDisplayOrderRequest struct {
	Id           int64 `json:"id,string" binding:"required,min=1"`
	DisplayOrder int32 `json:"displayOrder"`
}

// CounterpartyDeleteRequest represents all parameters of counterparty deleting request
type CounterpartyDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

// CounterpartyInfoResponse represents a view-object of counterparty
type CounterpartyInfoResponse struct {
	Id           int64            `json:"id,string"`
	Name         string           `json:"name"`
	Type         CounterpartyType `json:"type"`
	Icon         int64            `json:"icon,string"`
	Color        string           `json:"color"`
	Comment      string           `json:"comment"`
	DisplayOrder int32            `json:"displayOrder"`
	Hidden       bool             `json:"hidden"`
}

// ToCounterpartyInfoResponse returns a view-object according to database model
func (c *Counterparty) ToCounterpartyInfoResponse() *CounterpartyInfoResponse {
	return &CounterpartyInfoResponse{
		Id:           c.CounterpartyId,
		Name:         c.Name,
		Type:         c.Type,
		Icon:         c.Icon,
		Color:        c.Color,
		Comment:      c.Comment,
		DisplayOrder: c.DisplayOrder,
		Hidden:       c.Hidden,
	}
}

// CounterpartyInfoResponseSlice represents the slice data structure of CounterpartyInfoResponse
type CounterpartyInfoResponseSlice []*CounterpartyInfoResponse

// Len returns the count of items
func (s CounterpartyInfoResponseSlice) Len() int {
	return len(s)
}

// Swap swaps two items
func (s CounterpartyInfoResponseSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first item is less than the second one
func (s CounterpartyInfoResponseSlice) Less(i, j int) bool {
	return s[i].DisplayOrder < s[j].DisplayOrder
}
