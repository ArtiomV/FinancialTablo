package models

// CFO represents CFO (Cost/Financial center) data stored in database
type CFO struct {
	CfoId           int64  `xorm:"PK"`
	Uid             int64  `xorm:"INDEX(IDX_cfo_uid_deleted_order) NOT NULL"`
	Deleted         bool   `xorm:"INDEX(IDX_cfo_uid_deleted_order) NOT NULL"`
	Name            string `xorm:"VARCHAR(64) NOT NULL"`
	Color           string `xorm:"VARCHAR(6) NOT NULL"`
	Comment         string `xorm:"VARCHAR(255) NOT NULL"`
	DisplayOrder    int32  `xorm:"INDEX(IDX_cfo_uid_deleted_order) NOT NULL"`
	Hidden          bool   `xorm:"NOT NULL"`
	CreatedUnixTime int64
	UpdatedUnixTime int64
	DeletedUnixTime int64
}

// CFOGetRequest represents all parameters of CFO getting request
type CFOGetRequest struct {
	Id int64 `form:"id,string" binding:"required,min=1"`
}

// CFOCreateRequest represents all parameters of CFO creation request
type CFOCreateRequest struct {
	Name            string `json:"name" binding:"required,notBlank,max=64"`
	Color           string `json:"color" binding:"required,len=6,validHexRGBColor"`
	Comment         string `json:"comment" binding:"max=255"`
	ClientSessionId string `json:"clientSessionId"`
}

// CFOModifyRequest represents all parameters of CFO modification request
type CFOModifyRequest struct {
	Id      int64  `json:"id,string" binding:"required,min=1"`
	Name    string `json:"name" binding:"required,notBlank,max=64"`
	Color   string `json:"color" binding:"required,len=6,validHexRGBColor"`
	Comment string `json:"comment" binding:"max=255"`
	Hidden  bool   `json:"hidden"`
}

// CFOHideRequest represents all parameters of CFO hiding request
type CFOHideRequest struct {
	Id     int64 `json:"id,string" binding:"required,min=1"`
	Hidden bool  `json:"hidden"`
}

// CFOMoveRequest represents all parameters of CFO moving request
type CFOMoveRequest struct {
	NewDisplayOrders []*CFONewDisplayOrderRequest `json:"newDisplayOrders" binding:"required,min=1"`
}

// CFONewDisplayOrderRequest represents a data pair of id and display order
type CFONewDisplayOrderRequest struct {
	Id           int64 `json:"id,string" binding:"required,min=1"`
	DisplayOrder int32 `json:"displayOrder"`
}

// CFODeleteRequest represents all parameters of CFO deleting request
type CFODeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

// CFOInfoResponse represents a view-object of CFO
type CFOInfoResponse struct {
	Id           int64  `json:"id,string"`
	Name         string `json:"name"`
	Color        string `json:"color"`
	Comment      string `json:"comment"`
	DisplayOrder int32  `json:"displayOrder"`
	Hidden       bool   `json:"hidden"`
}

// ToCFOInfoResponse returns a view-object according to database model
func (c *CFO) ToCFOInfoResponse() *CFOInfoResponse {
	return &CFOInfoResponse{
		Id:           c.CfoId,
		Name:         c.Name,
		Color:        c.Color,
		Comment:      c.Comment,
		DisplayOrder: c.DisplayOrder,
		Hidden:       c.Hidden,
	}
}

// CFOInfoResponseSlice represents the slice data structure of CFOInfoResponse
type CFOInfoResponseSlice []*CFOInfoResponse

// Len returns the count of items
func (s CFOInfoResponseSlice) Len() int {
	return len(s)
}

// Swap swaps two items
func (s CFOInfoResponseSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first item is less than the second one
func (s CFOInfoResponseSlice) Less(i, j int) bool {
	return s[i].DisplayOrder < s[j].DisplayOrder
}
