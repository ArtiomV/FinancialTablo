// assets.go provides CRUD operations for fixed assets with depreciation tracking.
// Residual values use straight-line depreciation.
package services

import (
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// AssetService represents asset service
type AssetService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize an asset service singleton instance
var (
	Assets = &AssetService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetAllAssetsByUid returns all asset models of user
func (s *AssetService) GetAllAssetsByUid(c core.Context, uid int64) ([]*models.Asset, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var assets []*models.Asset
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).OrderBy("display_order asc").Find(&assets)

	return assets, err
}

// GetAssetByAssetId returns an asset model according to asset id
func (s *AssetService) GetAssetByAssetId(c core.Context, uid int64, assetId int64) (*models.Asset, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if assetId <= 0 {
		return nil, errs.ErrAssetIdInvalid
	}

	asset := &models.Asset{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(assetId).Where("uid=? AND deleted=?", uid, false).Get(asset)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrAssetNotFound
	}

	return asset, nil
}

// GetMaxDisplayOrder returns the max display order
func (s *AssetService) GetMaxDisplayOrder(c core.Context, uid int64) (int32, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	asset := &models.Asset{}
	has, err := s.UserDataDB(uid).NewSession(c).Cols("uid", "deleted", "display_order").Where("uid=? AND deleted=?", uid, false).OrderBy("display_order desc").Limit(1).Get(asset)

	if err != nil {
		return 0, err
	}

	if has {
		return asset.DisplayOrder, nil
	} else {
		return 0, nil
	}
}

// CreateAsset saves a new asset model to database
func (s *AssetService) CreateAsset(c core.Context, asset *models.Asset) error {
	if asset.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	exists, err := s.ExistsAssetName(c, asset.Uid, asset.Name)

	if err != nil {
		return err
	} else if exists {
		return errs.ErrAssetNameAlreadyExists
	}

	asset.AssetId = s.GenerateUuid(uuid.UUID_TYPE_ASSET)

	if asset.AssetId < 1 {
		return errs.ErrSystemIsBusy
	}

	asset.Deleted = false
	asset.CreatedUnixTime = time.Now().Unix()
	asset.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(asset.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(asset)
		return err
	})
}

// ModifyAsset saves an existed asset model to database
func (s *AssetService) ModifyAsset(c core.Context, asset *models.Asset, nameChanged bool) error {
	if asset.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if nameChanged {
		exists, err := s.ExistsAssetName(c, asset.Uid, asset.Name)

		if err != nil {
			return err
		} else if exists {
			return errs.ErrAssetNameAlreadyExists
		}
	}

	asset.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(asset.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.ID(asset.AssetId).Cols("name", "cfo_id", "location_id", "asset_type", "purchase_date", "purchase_cost", "useful_life_months", "salvage_value", "status", "commission_date", "decommission_date", "installed_capacity_watts", "comment", "hidden", "updated_unix_time").Where("uid=? AND deleted=?", asset.Uid, false).Update(asset)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrAssetNotFound
		}

		return err
	})
}

// HideAsset updates hidden field of given assets
func (s *AssetService) HideAsset(c core.Context, uid int64, ids []int64, hidden bool) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.Asset{
		Hidden:          hidden,
		UpdatedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.Cols("hidden", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).In("asset_id", ids).Update(updateModel)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrAssetNotFound
		}

		return err
	})
}

// ModifyAssetDisplayOrders updates display order of given assets
func (s *AssetService) ModifyAssetDisplayOrders(c core.Context, uid int64, assets []*models.Asset) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	for i := 0; i < len(assets); i++ {
		assets[i].UpdatedUnixTime = time.Now().Unix()
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(assets); i++ {
			asset := assets[i]
			updatedRows, err := sess.ID(asset.AssetId).Cols("display_order", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(asset)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				return errs.ErrAssetNotFound
			}
		}

		return nil
	})
}

// DeleteAsset deletes an existed asset from database
func (s *AssetService) DeleteAsset(c core.Context, uid int64, assetId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.Asset{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		deletedRows, err := sess.ID(assetId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrAssetNotFound
		}

		return err
	})
}

// ExistsAssetName returns whether the given asset name exists
func (s *AssetService) ExistsAssetName(c core.Context, uid int64, name string) (bool, error) {
	if name == "" {
		return false, errs.ErrAssetNameIsEmpty
	}

	return s.UserDataDB(uid).NewSession(c).Cols("name").Where("uid=? AND deleted=? AND name=?", uid, false, name).Exist(&models.Asset{})
}
