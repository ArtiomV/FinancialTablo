// counterparties.go provides CRUD for transaction counterparties.
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

// CounterpartyService represents counterparty service
type CounterpartyService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a counterparty service singleton instance
var (
	Counterparties = &CounterpartyService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetTotalCounterpartyCountByUid returns total counterparty count of user
func (s *CounterpartyService) GetTotalCounterpartyCountByUid(c core.Context, uid int64) (int64, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	count, err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).Count(&models.Counterparty{})

	return count, err
}

// GetAllCounterpartiesByUid returns all counterparty models of user
func (s *CounterpartyService) GetAllCounterpartiesByUid(c core.Context, uid int64) ([]*models.Counterparty, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var counterparties []*models.Counterparty
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).OrderBy("display_order asc").Find(&counterparties)

	return counterparties, err
}

// GetCounterpartyByCounterpartyId returns a counterparty model according to counterparty id
func (s *CounterpartyService) GetCounterpartyByCounterpartyId(c core.Context, uid int64, counterpartyId int64) (*models.Counterparty, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if counterpartyId <= 0 {
		return nil, errs.ErrCounterpartyIdInvalid
	}

	counterparty := &models.Counterparty{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(counterpartyId).Where("uid=? AND deleted=?", uid, false).Get(counterparty)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrCounterpartyNotFound
	}

	return counterparty, nil
}

// GetCounterpartiesByCounterpartyIds returns counterparty models according to counterparty ids
func (s *CounterpartyService) GetCounterpartiesByCounterpartyIds(c core.Context, uid int64, counterpartyIds []int64) (map[int64]*models.Counterparty, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if counterpartyIds == nil {
		return nil, errs.ErrCounterpartyIdInvalid
	}

	var counterparties []*models.Counterparty
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).In("counterparty_id", counterpartyIds).Find(&counterparties)

	if err != nil {
		return nil, err
	}

	counterpartyMap := make(map[int64]*models.Counterparty)

	for i := 0; i < len(counterparties); i++ {
		counterparty := counterparties[i]
		counterpartyMap[counterparty.CounterpartyId] = counterparty
	}

	return counterpartyMap, err
}

// GetMaxDisplayOrder returns the max display order
func (s *CounterpartyService) GetMaxDisplayOrder(c core.Context, uid int64) (int32, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	counterparty := &models.Counterparty{}
	has, err := s.UserDataDB(uid).NewSession(c).Cols("uid", "deleted", "display_order").Where("uid=? AND deleted=?", uid, false).OrderBy("display_order desc").Limit(1).Get(counterparty)

	if err != nil {
		return 0, err
	}

	if has {
		return counterparty.DisplayOrder, nil
	} else {
		return 0, nil
	}
}

// CreateCounterparty saves a new counterparty model to database
func (s *CounterpartyService) CreateCounterparty(c core.Context, counterparty *models.Counterparty) error {
	if counterparty.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	exists, err := s.ExistsCounterpartyName(c, counterparty.Uid, counterparty.Name)

	if err != nil {
		return err
	} else if exists {
		return errs.ErrCounterpartyNameAlreadyExists
	}

	counterparty.CounterpartyId = s.GenerateUuid(uuid.UUID_TYPE_COUNTERPARTY)

	if counterparty.CounterpartyId < 1 {
		return errs.ErrSystemIsBusy
	}

	counterparty.Deleted = false
	counterparty.CreatedUnixTime = time.Now().Unix()
	counterparty.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(counterparty.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(counterparty)
		return err
	})
}

// ModifyCounterparty saves an existed counterparty model to database
func (s *CounterpartyService) ModifyCounterparty(c core.Context, counterparty *models.Counterparty, counterpartyNameChanged bool) error {
	if counterparty.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if counterpartyNameChanged {
		exists, err := s.ExistsCounterpartyName(c, counterparty.Uid, counterparty.Name)

		if err != nil {
			return err
		} else if exists {
			return errs.ErrCounterpartyNameAlreadyExists
		}
	}

	counterparty.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(counterparty.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.ID(counterparty.CounterpartyId).Cols("name", "type", "icon", "color", "comment", "hidden", "updated_unix_time").Where("uid=? AND deleted=?", counterparty.Uid, false).Update(counterparty)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrCounterpartyNotFound
		}

		return err
	})
}

// HideCounterparty updates hidden field of given counterparties
func (s *CounterpartyService) HideCounterparty(c core.Context, uid int64, ids []int64, hidden bool) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.Counterparty{
		Hidden:          hidden,
		UpdatedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.Cols("hidden", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).In("counterparty_id", ids).Update(updateModel)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrCounterpartyNotFound
		}

		return err
	})
}

// ModifyCounterpartyDisplayOrders updates display order of given counterparties
func (s *CounterpartyService) ModifyCounterpartyDisplayOrders(c core.Context, uid int64, counterparties []*models.Counterparty) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	for i := 0; i < len(counterparties); i++ {
		counterparties[i].UpdatedUnixTime = time.Now().Unix()
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(counterparties); i++ {
			counterparty := counterparties[i]
			updatedRows, err := sess.ID(counterparty.CounterpartyId).Cols("display_order", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(counterparty)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				return errs.ErrCounterpartyNotFound
			}
		}

		return nil
	})
}

// DeleteCounterparty deletes an existed counterparty from database
func (s *CounterpartyService) DeleteCounterparty(c core.Context, uid int64, counterpartyId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.Counterparty{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		deletedRows, err := sess.ID(counterpartyId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrCounterpartyNotFound
		}

		return err
	})
}

// DeleteAllCounterparties deletes all existed counterparties from database
func (s *CounterpartyService) DeleteAllCounterparties(c core.Context, uid int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.Counterparty{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		}

		return nil
	})
}

// ExistsCounterpartyName returns whether the given counterparty name exists
func (s *CounterpartyService) ExistsCounterpartyName(c core.Context, uid int64, name string) (bool, error) {
	if name == "" {
		return false, errs.ErrCounterpartyNameIsEmpty
	}

	return s.UserDataDB(uid).NewSession(c).Cols("name").Where("uid=? AND deleted=? AND name=?", uid, false, name).Exist(&models.Counterparty{})
}
