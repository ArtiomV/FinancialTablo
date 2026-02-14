package models

// AssetType represents asset type
type AssetType byte

// Asset types
const (
	ASSET_TYPE_EQUIPMENT  AssetType = 1
	ASSET_TYPE_FURNITURE  AssetType = 2
	ASSET_TYPE_VEHICLE    AssetType = 3
	ASSET_TYPE_ELECTRONICS AssetType = 4
	ASSET_TYPE_REAL_ESTATE AssetType = 5
	ASSET_TYPE_OTHER      AssetType = 6
)

// AssetStatus represents asset status
type AssetStatus byte

// Asset statuses
const (
	ASSET_STATUS_ACTIVE       AssetStatus = 1
	ASSET_STATUS_DECOMMISSIONED AssetStatus = 2
	ASSET_STATUS_SOLD         AssetStatus = 3
)

// Asset represents asset data stored in database
type Asset struct {
	AssetId                int64       `xorm:"PK"`
	Uid                    int64       `xorm:"INDEX(IDX_asset_uid_deleted_order) NOT NULL"`
	Deleted                bool        `xorm:"INDEX(IDX_asset_uid_deleted_order) NOT NULL"`
	CfoId                  int64       `xorm:"NOT NULL DEFAULT 0"`
	LocationId             int64       `xorm:"NOT NULL DEFAULT 0"`
	Name                   string      `xorm:"VARCHAR(64) NOT NULL"`
	AssetType              AssetType   `xorm:"NOT NULL DEFAULT 1"`
	PurchaseDate           int64       `xorm:"NOT NULL DEFAULT 0"`
	PurchaseCost           int64       `xorm:"NOT NULL DEFAULT 0"`
	UsefulLifeMonths       int32       `xorm:"NOT NULL DEFAULT 0"`
	SalvageValue           int64       `xorm:"NOT NULL DEFAULT 0"`
	Status                 AssetStatus `xorm:"NOT NULL DEFAULT 1"`
	CommissionDate         int64       `xorm:"NOT NULL DEFAULT 0"`
	DecommissionDate       int64       `xorm:"NOT NULL DEFAULT 0"`
	InstalledCapacityWatts int64       `xorm:"NOT NULL DEFAULT 0"`
	Comment                string      `xorm:"VARCHAR(255) NOT NULL"`
	DisplayOrder           int32       `xorm:"INDEX(IDX_asset_uid_deleted_order) NOT NULL"`
	Hidden                 bool        `xorm:"NOT NULL"`
	CreatedUnixTime        int64
	UpdatedUnixTime        int64
	DeletedUnixTime        int64
}

// AssetGetRequest represents all parameters of asset getting request
type AssetGetRequest struct {
	Id int64 `form:"id,string" binding:"required,min=1"`
}

// AssetCreateRequest represents all parameters of asset creation request
type AssetCreateRequest struct {
	Name                   string      `json:"name" binding:"required,notBlank,max=64"`
	CfoId                  int64       `json:"cfoId,string"`
	LocationId             int64       `json:"locationId,string"`
	AssetType              AssetType   `json:"assetType" binding:"min=0"`
	PurchaseDate           int64       `json:"purchaseDate"`
	PurchaseCost           int64       `json:"purchaseCost" binding:"min=0"`
	UsefulLifeMonths       int32       `json:"usefulLifeMonths" binding:"min=0"`
	SalvageValue           int64       `json:"salvageValue" binding:"min=0"`
	Status                 AssetStatus `json:"status"`
	CommissionDate         int64       `json:"commissionDate"`
	DecommissionDate       int64       `json:"decommissionDate"`
	InstalledCapacityWatts int64       `json:"installedCapacityWatts" binding:"min=0"`
	Comment                string      `json:"comment" binding:"max=255"`
	ClientSessionId        string      `json:"clientSessionId"`
}

// AssetModifyRequest represents all parameters of asset modification request
type AssetModifyRequest struct {
	Id                     int64       `json:"id,string" binding:"required,min=1"`
	Name                   string      `json:"name" binding:"required,notBlank,max=64"`
	CfoId                  int64       `json:"cfoId,string"`
	LocationId             int64       `json:"locationId,string"`
	AssetType              AssetType   `json:"assetType"`
	PurchaseDate           int64       `json:"purchaseDate"`
	PurchaseCost           int64       `json:"purchaseCost"`
	UsefulLifeMonths       int32       `json:"usefulLifeMonths"`
	SalvageValue           int64       `json:"salvageValue"`
	Status                 AssetStatus `json:"status"`
	CommissionDate         int64       `json:"commissionDate"`
	DecommissionDate       int64       `json:"decommissionDate"`
	InstalledCapacityWatts int64       `json:"installedCapacityWatts"`
	Comment                string      `json:"comment" binding:"max=255"`
	Hidden                 bool        `json:"hidden"`
}

// AssetHideRequest represents all parameters of asset hiding request
type AssetHideRequest struct {
	Id     int64 `json:"id,string" binding:"required,min=1"`
	Hidden bool  `json:"hidden"`
}

// AssetMoveRequest represents all parameters of asset moving request
type AssetMoveRequest struct {
	NewDisplayOrders []*AssetNewDisplayOrderRequest `json:"newDisplayOrders" binding:"required,min=1"`
}

// AssetNewDisplayOrderRequest represents a data pair of id and display order
type AssetNewDisplayOrderRequest struct {
	Id           int64 `json:"id,string" binding:"required,min=1"`
	DisplayOrder int32 `json:"displayOrder"`
}

// AssetDeleteRequest represents all parameters of asset deleting request
type AssetDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

// AssetInfoResponse represents a view-object of asset
type AssetInfoResponse struct {
	Id                     int64       `json:"id,string"`
	Name                   string      `json:"name"`
	CfoId                  int64       `json:"cfoId,string"`
	LocationId             int64       `json:"locationId,string"`
	AssetType              AssetType   `json:"assetType"`
	PurchaseDate           int64       `json:"purchaseDate"`
	PurchaseCost           int64       `json:"purchaseCost"`
	UsefulLifeMonths       int32       `json:"usefulLifeMonths"`
	SalvageValue           int64       `json:"salvageValue"`
	Status                 AssetStatus `json:"status"`
	CommissionDate         int64       `json:"commissionDate"`
	DecommissionDate       int64       `json:"decommissionDate"`
	InstalledCapacityWatts int64       `json:"installedCapacityWatts"`
	Comment                string      `json:"comment"`
	DisplayOrder           int32       `json:"displayOrder"`
	Hidden                 bool        `json:"hidden"`
}

// ToAssetInfoResponse returns a view-object according to database model
func (a *Asset) ToAssetInfoResponse() *AssetInfoResponse {
	return &AssetInfoResponse{
		Id:                     a.AssetId,
		Name:                   a.Name,
		CfoId:                  a.CfoId,
		LocationId:             a.LocationId,
		AssetType:              a.AssetType,
		PurchaseDate:           a.PurchaseDate,
		PurchaseCost:           a.PurchaseCost,
		UsefulLifeMonths:       a.UsefulLifeMonths,
		SalvageValue:           a.SalvageValue,
		Status:                 a.Status,
		CommissionDate:         a.CommissionDate,
		DecommissionDate:       a.DecommissionDate,
		InstalledCapacityWatts: a.InstalledCapacityWatts,
		Comment:                a.Comment,
		DisplayOrder:           a.DisplayOrder,
		Hidden:                 a.Hidden,
	}
}

// AssetInfoResponseSlice represents the slice data structure of AssetInfoResponse
type AssetInfoResponseSlice []*AssetInfoResponse

// Len returns the count of items
func (s AssetInfoResponseSlice) Len() int {
	return len(s)
}

// Swap swaps two items
func (s AssetInfoResponseSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first item is less than the second one
func (s AssetInfoResponseSlice) Less(i, j int) bool {
	return s[i].DisplayOrder < s[j].DisplayOrder
}
