// obligations.go provides CRUD for receivables (type=1) and payables (type=2).
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

// ObligationService represents obligation service
type ObligationService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize an obligation service singleton instance
var (
	Obligations = &ObligationService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetAllObligationsByUid returns all obligation models of user
func (s *ObligationService) GetAllObligationsByUid(c core.Context, uid int64) ([]*models.Obligation, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var obligations []*models.Obligation
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).OrderBy("due_date desc").Find(&obligations)

	return obligations, err
}

// GetObligationByObligationId returns an obligation model according to obligation id
func (s *ObligationService) GetObligationByObligationId(c core.Context, uid int64, obligationId int64) (*models.Obligation, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if obligationId <= 0 {
		return nil, errs.ErrObligationIdInvalid
	}

	obligation := &models.Obligation{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(obligationId).Where("uid=? AND deleted=?", uid, false).Get(obligation)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrObligationNotFound
	}

	return obligation, nil
}

// CreateObligation saves a new obligation model to database
func (s *ObligationService) CreateObligation(c core.Context, obligation *models.Obligation) error {
	if obligation.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	obligation.ObligationId = s.GenerateUuid(uuid.UUID_TYPE_DEFAULT)

	if obligation.ObligationId < 1 {
		return errs.ErrSystemIsBusy
	}

	obligation.Deleted = false
	obligation.CreatedUnixTime = time.Now().Unix()
	obligation.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(obligation.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(obligation)
		return err
	})
}

// ModifyObligation saves an existed obligation model to database
func (s *ObligationService) ModifyObligation(c core.Context, obligation *models.Obligation) error {
	if obligation.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	obligation.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(obligation.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.ID(obligation.ObligationId).Cols("obligation_type", "counterparty_id", "cfo_id", "amount", "currency", "due_date", "status", "paid_amount", "comment", "updated_unix_time").Where("uid=? AND deleted=?", obligation.Uid, false).Update(obligation)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrObligationNotFound
		}

		return err
	})
}

// DeleteObligation deletes an existed obligation from database
func (s *ObligationService) DeleteObligation(c core.Context, uid int64, obligationId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.Obligation{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		deletedRows, err := sess.ID(obligationId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrObligationNotFound
		}

		return err
	})
}
