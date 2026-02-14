package api

import (
	"sort"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// AssetsApi represents asset api
type AssetsApi struct {
	assets *services.AssetService
}

// Initialize an asset api singleton instance
var (
	Assets = &AssetsApi{
		assets: services.Assets,
	}
)

// AssetListHandler returns asset list of current user
func (a *AssetsApi) AssetListHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	assets, err := a.assets.GetAllAssetsByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[assets.AssetListHandler] failed to get assets for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	assetResps := make(models.AssetInfoResponseSlice, len(assets))

	for i := 0; i < len(assets); i++ {
		assetResps[i] = assets[i].ToAssetInfoResponse()
	}

	sort.Sort(assetResps)

	return assetResps, nil
}

// AssetGetHandler returns one specific asset of current user
func (a *AssetsApi) AssetGetHandler(c *core.WebContext) (any, *errs.Error) {
	var assetGetReq models.AssetGetRequest
	err := c.ShouldBindQuery(&assetGetReq)

	if err != nil {
		log.Warnf(c, "[assets.AssetGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	asset, err := a.assets.GetAssetByAssetId(c, uid, assetGetReq.Id)

	if err != nil {
		log.Errorf(c, "[assets.AssetGetHandler] failed to get asset \"id:%d\" for user \"uid:%d\", because %s", assetGetReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return asset.ToAssetInfoResponse(), nil
}

// AssetCreateHandler saves a new asset by request parameters for current user
func (a *AssetsApi) AssetCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var assetCreateReq models.AssetCreateRequest
	err := c.ShouldBindJSON(&assetCreateReq)

	if err != nil {
		log.Warnf(c, "[assets.AssetCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	maxOrderId, err := a.assets.GetMaxDisplayOrder(c, uid)

	if err != nil {
		log.Errorf(c, "[assets.AssetCreateHandler] failed to get max display order for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	asset := &models.Asset{
		Uid:                    uid,
		Name:                   assetCreateReq.Name,
		CfoId:                  assetCreateReq.CfoId,
		LocationId:             assetCreateReq.LocationId,
		AssetType:              assetCreateReq.AssetType,
		PurchaseDate:           assetCreateReq.PurchaseDate,
		PurchaseCost:           assetCreateReq.PurchaseCost,
		UsefulLifeMonths:       assetCreateReq.UsefulLifeMonths,
		SalvageValue:           assetCreateReq.SalvageValue,
		Status:                 assetCreateReq.Status,
		CommissionDate:         assetCreateReq.CommissionDate,
		DecommissionDate:       assetCreateReq.DecommissionDate,
		InstalledCapacityWatts: assetCreateReq.InstalledCapacityWatts,
		Comment:                assetCreateReq.Comment,
		DisplayOrder:           maxOrderId + 1,
	}

	err = a.assets.CreateAsset(c, asset)

	if err != nil {
		log.Errorf(c, "[assets.AssetCreateHandler] failed to create asset \"id:%d\" for user \"uid:%d\", because %s", asset.AssetId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[assets.AssetCreateHandler] user \"uid:%d\" has created a new asset \"id:%d\" successfully", uid, asset.AssetId)

	return asset.ToAssetInfoResponse(), nil
}

// AssetModifyHandler saves an existed asset by request parameters for current user
func (a *AssetsApi) AssetModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var assetModifyReq models.AssetModifyRequest
	err := c.ShouldBindJSON(&assetModifyReq)

	if err != nil {
		log.Warnf(c, "[assets.AssetModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	asset, err := a.assets.GetAssetByAssetId(c, uid, assetModifyReq.Id)

	if err != nil {
		log.Errorf(c, "[assets.AssetModifyHandler] failed to get asset \"id:%d\" for user \"uid:%d\", because %s", assetModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	newAsset := &models.Asset{
		AssetId:                asset.AssetId,
		Uid:                    uid,
		Name:                   assetModifyReq.Name,
		CfoId:                  assetModifyReq.CfoId,
		LocationId:             assetModifyReq.LocationId,
		AssetType:              assetModifyReq.AssetType,
		PurchaseDate:           assetModifyReq.PurchaseDate,
		PurchaseCost:           assetModifyReq.PurchaseCost,
		UsefulLifeMonths:       assetModifyReq.UsefulLifeMonths,
		SalvageValue:           assetModifyReq.SalvageValue,
		Status:                 assetModifyReq.Status,
		CommissionDate:         assetModifyReq.CommissionDate,
		DecommissionDate:       assetModifyReq.DecommissionDate,
		InstalledCapacityWatts: assetModifyReq.InstalledCapacityWatts,
		Comment:                assetModifyReq.Comment,
		Hidden:                 assetModifyReq.Hidden,
		DisplayOrder:           asset.DisplayOrder,
	}

	nameChanged := newAsset.Name != asset.Name

	if !nameChanged &&
		newAsset.CfoId == asset.CfoId &&
		newAsset.LocationId == asset.LocationId &&
		newAsset.AssetType == asset.AssetType &&
		newAsset.PurchaseDate == asset.PurchaseDate &&
		newAsset.PurchaseCost == asset.PurchaseCost &&
		newAsset.UsefulLifeMonths == asset.UsefulLifeMonths &&
		newAsset.SalvageValue == asset.SalvageValue &&
		newAsset.Status == asset.Status &&
		newAsset.CommissionDate == asset.CommissionDate &&
		newAsset.DecommissionDate == asset.DecommissionDate &&
		newAsset.InstalledCapacityWatts == asset.InstalledCapacityWatts &&
		newAsset.Comment == asset.Comment &&
		newAsset.Hidden == asset.Hidden {
		return nil, errs.ErrNothingWillBeUpdated
	}

	err = a.assets.ModifyAsset(c, newAsset, nameChanged)

	if err != nil {
		log.Errorf(c, "[assets.AssetModifyHandler] failed to update asset \"id:%d\" for user \"uid:%d\", because %s", assetModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[assets.AssetModifyHandler] user \"uid:%d\" has updated asset \"id:%d\" successfully", uid, assetModifyReq.Id)

	return newAsset.ToAssetInfoResponse(), nil
}

// AssetHideHandler hides an asset by request parameters for current user
func (a *AssetsApi) AssetHideHandler(c *core.WebContext) (any, *errs.Error) {
	var assetHideReq models.AssetHideRequest
	err := c.ShouldBindJSON(&assetHideReq)

	if err != nil {
		log.Warnf(c, "[assets.AssetHideHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.assets.HideAsset(c, uid, []int64{assetHideReq.Id}, assetHideReq.Hidden)

	if err != nil {
		log.Errorf(c, "[assets.AssetHideHandler] failed to hide asset \"id:%d\" for user \"uid:%d\", because %s", assetHideReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[assets.AssetHideHandler] user \"uid:%d\" has hidden asset \"id:%d\"", uid, assetHideReq.Id)
	return true, nil
}

// AssetMoveHandler moves display order of existed assets by request parameters for current user
func (a *AssetsApi) AssetMoveHandler(c *core.WebContext) (any, *errs.Error) {
	var assetMoveReq models.AssetMoveRequest
	err := c.ShouldBindJSON(&assetMoveReq)

	if err != nil {
		log.Warnf(c, "[assets.AssetMoveHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	assets := make([]*models.Asset, len(assetMoveReq.NewDisplayOrders))

	for i := 0; i < len(assetMoveReq.NewDisplayOrders); i++ {
		newDisplayOrder := assetMoveReq.NewDisplayOrders[i]
		asset := &models.Asset{
			Uid:          uid,
			AssetId:      newDisplayOrder.Id,
			DisplayOrder: newDisplayOrder.DisplayOrder,
		}

		assets[i] = asset
	}

	err = a.assets.ModifyAssetDisplayOrders(c, uid, assets)

	if err != nil {
		log.Errorf(c, "[assets.AssetMoveHandler] failed to move assets for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[assets.AssetMoveHandler] user \"uid:%d\" has moved assets", uid)
	return true, nil
}

// AssetDeleteHandler deletes an existed asset by request parameters for current user
func (a *AssetsApi) AssetDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var assetDeleteReq models.AssetDeleteRequest
	err := c.ShouldBindJSON(&assetDeleteReq)

	if err != nil {
		log.Warnf(c, "[assets.AssetDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.assets.DeleteAsset(c, uid, assetDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[assets.AssetDeleteHandler] failed to delete asset \"id:%d\" for user \"uid:%d\", because %s", assetDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[assets.AssetDeleteHandler] user \"uid:%d\" has deleted asset \"id:%d\"", uid, assetDeleteReq.Id)
	return true, nil
}
