package models

// LocationType represents location type
type LocationType byte

// Location types
const (
	LOCATION_TYPE_OFFICE     LocationType = 1
	LOCATION_TYPE_WAREHOUSE  LocationType = 2
	LOCATION_TYPE_STORE      LocationType = 3
	LOCATION_TYPE_PRODUCTION LocationType = 4
	LOCATION_TYPE_OTHER      LocationType = 5
)

// Location represents location data stored in database
type Location struct {
	LocationId         int64        `xorm:"PK"`
	Uid                int64        `xorm:"INDEX(IDX_location_uid_deleted_order) NOT NULL"`
	Deleted            bool         `xorm:"INDEX(IDX_location_uid_deleted_order) NOT NULL"`
	CfoId              int64        `xorm:"NOT NULL DEFAULT 0"`
	Name               string       `xorm:"VARCHAR(64) NOT NULL"`
	Address            string       `xorm:"VARCHAR(255) NOT NULL"`
	LocationType       LocationType `xorm:"NOT NULL DEFAULT 1"`
	MonthlyRent        int64        `xorm:"NOT NULL DEFAULT 0"`
	MonthlyElectricity int64        `xorm:"NOT NULL DEFAULT 0"`
	MonthlyInternet    int64        `xorm:"NOT NULL DEFAULT 0"`
	Comment            string       `xorm:"VARCHAR(255) NOT NULL"`
	DisplayOrder       int32        `xorm:"INDEX(IDX_location_uid_deleted_order) NOT NULL"`
	Hidden             bool         `xorm:"NOT NULL"`
	CreatedUnixTime    int64
	UpdatedUnixTime    int64
	DeletedUnixTime    int64
}

// LocationGetRequest represents all parameters of location getting request
type LocationGetRequest struct {
	Id int64 `form:"id,string" binding:"required,min=1"`
}

// LocationCreateRequest represents all parameters of location creation request
type LocationCreateRequest struct {
	Name               string       `json:"name" binding:"required,notBlank,max=64"`
	CfoId              int64        `json:"cfoId,string"`
	Address            string       `json:"address" binding:"max=255"`
	LocationType       LocationType `json:"locationType"`
	MonthlyRent        int64        `json:"monthlyRent"`
	MonthlyElectricity int64        `json:"monthlyElectricity"`
	MonthlyInternet    int64        `json:"monthlyInternet"`
	Comment            string       `json:"comment" binding:"max=255"`
	ClientSessionId    string       `json:"clientSessionId"`
}

// LocationModifyRequest represents all parameters of location modification request
type LocationModifyRequest struct {
	Id                 int64        `json:"id,string" binding:"required,min=1"`
	Name               string       `json:"name" binding:"required,notBlank,max=64"`
	CfoId              int64        `json:"cfoId,string"`
	Address            string       `json:"address" binding:"max=255"`
	LocationType       LocationType `json:"locationType"`
	MonthlyRent        int64        `json:"monthlyRent"`
	MonthlyElectricity int64        `json:"monthlyElectricity"`
	MonthlyInternet    int64        `json:"monthlyInternet"`
	Comment            string       `json:"comment" binding:"max=255"`
	Hidden             bool         `json:"hidden"`
}

// LocationHideRequest represents all parameters of location hiding request
type LocationHideRequest struct {
	Id     int64 `json:"id,string" binding:"required,min=1"`
	Hidden bool  `json:"hidden"`
}

// LocationMoveRequest represents all parameters of location moving request
type LocationMoveRequest struct {
	NewDisplayOrders []*LocationNewDisplayOrderRequest `json:"newDisplayOrders" binding:"required,min=1"`
}

// LocationNewDisplayOrderRequest represents a data pair of id and display order
type LocationNewDisplayOrderRequest struct {
	Id           int64 `json:"id,string" binding:"required,min=1"`
	DisplayOrder int32 `json:"displayOrder"`
}

// LocationDeleteRequest represents all parameters of location deleting request
type LocationDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

// LocationInfoResponse represents a view-object of location
type LocationInfoResponse struct {
	Id                 int64        `json:"id,string"`
	Name               string       `json:"name"`
	CfoId              int64        `json:"cfoId,string"`
	Address            string       `json:"address"`
	LocationType       LocationType `json:"locationType"`
	MonthlyRent        int64        `json:"monthlyRent"`
	MonthlyElectricity int64        `json:"monthlyElectricity"`
	MonthlyInternet    int64        `json:"monthlyInternet"`
	Comment            string       `json:"comment"`
	DisplayOrder       int32        `json:"displayOrder"`
	Hidden             bool         `json:"hidden"`
}

// ToLocationInfoResponse returns a view-object according to database model
func (l *Location) ToLocationInfoResponse() *LocationInfoResponse {
	return &LocationInfoResponse{
		Id:                 l.LocationId,
		Name:               l.Name,
		CfoId:              l.CfoId,
		Address:            l.Address,
		LocationType:       l.LocationType,
		MonthlyRent:        l.MonthlyRent,
		MonthlyElectricity: l.MonthlyElectricity,
		MonthlyInternet:    l.MonthlyInternet,
		Comment:            l.Comment,
		DisplayOrder:       l.DisplayOrder,
		Hidden:             l.Hidden,
	}
}

// LocationInfoResponseSlice represents the slice data structure of LocationInfoResponse
type LocationInfoResponseSlice []*LocationInfoResponse

// Len returns the count of items
func (s LocationInfoResponseSlice) Len() int {
	return len(s)
}

// Swap swaps two items
func (s LocationInfoResponseSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first item is less than the second one
func (s LocationInfoResponseSlice) Less(i, j int) bool {
	return s[i].DisplayOrder < s[j].DisplayOrder
}
