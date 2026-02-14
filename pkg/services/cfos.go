// cfos.go provides CRUD for Centers of Financial Responsibility (CFO).
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

// CFOService represents CFO service
type CFOService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a CFO service singleton instance
var (
	CFOs = &CFOService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetAllCFOsByUid returns all CFO models of user
func (s *CFOService) GetAllCFOsByUid(c core.Context, uid int64) ([]*models.CFO, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var cfos []*models.CFO
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).OrderBy("display_order asc").Find(&cfos)

	return cfos, err
}

// GetCFOByCFOId returns a CFO model according to CFO id
func (s *CFOService) GetCFOByCFOId(c core.Context, uid int64, cfoId int64) (*models.CFO, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if cfoId <= 0 {
		return nil, errs.ErrCFOIdInvalid
	}

	cfo := &models.CFO{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(cfoId).Where("uid=? AND deleted=?", uid, false).Get(cfo)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrCFONotFound
	}

	return cfo, nil
}

// GetMaxDisplayOrder returns the max display order
func (s *CFOService) GetMaxDisplayOrder(c core.Context, uid int64) (int32, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	cfo := &models.CFO{}
	has, err := s.UserDataDB(uid).NewSession(c).Cols("uid", "deleted", "display_order").Where("uid=? AND deleted=?", uid, false).OrderBy("display_order desc").Limit(1).Get(cfo)

	if err != nil {
		return 0, err
	}

	if has {
		return cfo.DisplayOrder, nil
	} else {
		return 0, nil
	}
}

// CreateCFO saves a new CFO model to database
func (s *CFOService) CreateCFO(c core.Context, cfo *models.CFO) error {
	if cfo.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	exists, err := s.ExistsCFOName(c, cfo.Uid, cfo.Name)

	if err != nil {
		return err
	} else if exists {
		return errs.ErrCFONameAlreadyExists
	}

	cfo.CfoId = s.GenerateUuid(uuid.UUID_TYPE_CFO)

	if cfo.CfoId < 1 {
		return errs.ErrSystemIsBusy
	}

	cfo.Deleted = false
	cfo.CreatedUnixTime = time.Now().Unix()
	cfo.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(cfo.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(cfo)
		return err
	})
}

// ModifyCFO saves an existed CFO model to database
func (s *CFOService) ModifyCFO(c core.Context, cfo *models.CFO, nameChanged bool) error {
	if cfo.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if nameChanged {
		exists, err := s.ExistsCFOName(c, cfo.Uid, cfo.Name)

		if err != nil {
			return err
		} else if exists {
			return errs.ErrCFONameAlreadyExists
		}
	}

	cfo.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(cfo.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.ID(cfo.CfoId).Cols("name", "color", "comment", "hidden", "updated_unix_time").Where("uid=? AND deleted=?", cfo.Uid, false).Update(cfo)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrCFONotFound
		}

		return err
	})
}

// HideCFO updates hidden field of given CFOs
func (s *CFOService) HideCFO(c core.Context, uid int64, ids []int64, hidden bool) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.CFO{
		Hidden:          hidden,
		UpdatedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.Cols("hidden", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).In("cfo_id", ids).Update(updateModel)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrCFONotFound
		}

		return err
	})
}

// ModifyCFODisplayOrders updates display order of given CFOs
func (s *CFOService) ModifyCFODisplayOrders(c core.Context, uid int64, cfos []*models.CFO) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	for i := 0; i < len(cfos); i++ {
		cfos[i].UpdatedUnixTime = time.Now().Unix()
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		for i := 0; i < len(cfos); i++ {
			cfo := cfos[i]
			updatedRows, err := sess.ID(cfo.CfoId).Cols("display_order", "updated_unix_time").Where("uid=? AND deleted=?", uid, false).Update(cfo)

			if err != nil {
				return err
			} else if updatedRows < 1 {
				return errs.ErrCFONotFound
			}
		}

		return nil
	})
}

// DeleteCFO deletes an existed CFO from database
func (s *CFOService) DeleteCFO(c core.Context, uid int64, cfoId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.CFO{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		deletedRows, err := sess.ID(cfoId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrCFONotFound
		}

		return err
	})
}

// ExistsCFOName returns whether the given CFO name exists
func (s *CFOService) ExistsCFOName(c core.Context, uid int64, name string) (bool, error) {
	if name == "" {
		return false, errs.ErrCFONameIsEmpty
	}

	return s.UserDataDB(uid).NewSession(c).Cols("name").Where("uid=? AND deleted=? AND name=?", uid, false, name).Exist(&models.CFO{})
}
