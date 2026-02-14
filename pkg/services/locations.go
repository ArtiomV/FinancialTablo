// locations.go provides CRUD for physical locations linked to assets.
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

// LocationService represents location service
type LocationService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a location service singleton instance
var (
	Locations = &LocationService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetAllLocationsByUid returns all location models of user
func (s *LocationService) GetAllLocationsByUid(c core.Context, uid int64) ([]*models.Location, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var locations []*models.Location
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).OrderBy("display_order asc").Find(&locations)

	return locations, err
}

// GetLocationByLocationId returns a location model according to location id
func (s *LocationService) GetLocationByLocationId(c core.Context, uid int64, locationId int64) (*models.Location, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if locationId <= 0 {
		return nil, errs.ErrLocationIdInvalid
	}

	location := &models.Location{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(locationId).Where("uid=? AND deleted=?", uid, false).Get(location)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrLocationNotFound
	}

	return location, nil
}

// GetMaxDisplayOrder returns the max display order
func (s *LocationService) GetMaxDisplayOrder(c core.Context, uid int64) (int32, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	location := &models.Location{}
	has, err := s.UserDataDB(uid).NewSession(c).Cols("uid", "deleted", "display_order").Where("uid=? AND deleted=?", uid, false).OrderBy("display_order desc").Limit(1).Get(location)

	if err != nil {
		return 0, err
	}

	if has {
		return location.DisplayOrder, nil
	} else {
		return 0, nil
	}
}

// CreateLocation saves a new location model to database
func (s *LocationService) CreateLocation(c core.Context, location *models.Location) error {
	if location.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	exists, err := s.ExistsLocationName(c, location.Uid, location.Name)

	if err != nil {
		return err
	} else if exists {
		return errs.ErrLocationNameAlreadyExists
	}

	location.LocationId = s.GenerateUuid(uuid.UUID_TYPE_LOCATION)

	if location.LocationId < 1 {
		return errs.ErrSystemIsBusy
	}

	location.Deleted = false
	location.CreatedUnixTime = time.Now().Unix()
	location.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(location.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(location)
		return err
	})
}

// ModifyLocation saves an existed location model to database
func (s *LocationService) ModifyLocation(c core.Context, location *models.Location, nameChanged bool) error {
	if location.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if nameChanged {
		exists, err := s.ExistsLocationName(c, location.Uid, location.Name)

		if err != nil {
			return err
		} else if exists {
			return errs.ErrLocationNameAlreadyExists
		}
	}

	location.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(location.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.ID(location.LocationId).Cols("name", "cfo_id", "address", "location_type", "monthly_rent", "monthly_electricity", "monthly_internet", "comment", "hidden", "updated_unix_time").Where("uid=? AND deleted=?", location.Uid, false).Update(location)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrLocationNotFound
		}

		return err
	})
}

// HideLocation updates hidden field of given locations
func (s *LocationService) HideLocation(c core.Context, uid int64, ids []int64, hidden bool) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.Location{
		Hidden:          hidden,
		UpdatedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.Cols("hidden", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).In("location_id", ids).Update(updateModel)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrLocationNotFound
		}

		return err
	})
}

// ModifyLocationDisplayOrders updates display order of given locations
func (s *LocationService) ModifyLocationDisplayOrders(c core.Context, uid int64, locations []*models.Location) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	for i := 0; i < len(locations); i++ {
		locations[i].UpdatedUnixTime = time.Now().Unix()
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(locations); i++ {
			location := locations[i]
			updatedRows, err := sess.ID(location.LocationId).Cols("display_order", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(location)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				return errs.ErrLocationNotFound
			}
		}

		return nil
	})
}

// DeleteLocation deletes an existed location from database
func (s *LocationService) DeleteLocation(c core.Context, uid int64, locationId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.Location{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		deletedRows, err := sess.ID(locationId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrLocationNotFound
		}

		return err
	})
}

// ExistsLocationName returns whether the given location name exists
func (s *LocationService) ExistsLocationName(c core.Context, uid int64, name string) (bool, error) {
	if name == "" {
		return false, errs.ErrLocationNameIsEmpty
	}

	return s.UserDataDB(uid).NewSession(c).Cols("name").Where("uid=? AND deleted=? AND name=?", uid, false, name).Exist(&models.Location{})
}
