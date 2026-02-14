// tax_records.go provides CRUD for tax obligations with due date tracking.
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

// TaxRecordService represents tax record service
type TaxRecordService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a tax record service singleton instance
var (
	TaxRecords = &TaxRecordService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetAllTaxRecordsByUid returns all tax record models of user
func (s *TaxRecordService) GetAllTaxRecordsByUid(c core.Context, uid int64) ([]*models.TaxRecord, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var records []*models.TaxRecord
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).OrderBy("period_year desc, period_quarter desc").Find(&records)

	return records, err
}

// GetTaxRecordByTaxId returns a tax record model according to tax id
func (s *TaxRecordService) GetTaxRecordByTaxId(c core.Context, uid int64, taxId int64) (*models.TaxRecord, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if taxId <= 0 {
		return nil, errs.ErrTaxRecordIdInvalid
	}

	record := &models.TaxRecord{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(taxId).Where("uid=? AND deleted=?", uid, false).Get(record)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrTaxRecordNotFound
	}

	return record, nil
}

// CreateTaxRecord saves a new tax record model to database
func (s *TaxRecordService) CreateTaxRecord(c core.Context, record *models.TaxRecord) error {
	if record.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	record.TaxId = s.GenerateUuid(uuid.UUID_TYPE_DEFAULT)

	if record.TaxId < 1 {
		return errs.ErrSystemIsBusy
	}

	record.Deleted = false
	record.CreatedUnixTime = time.Now().Unix()
	record.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(record.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(record)
		return err
	})
}

// ModifyTaxRecord saves an existed tax record model to database
func (s *TaxRecordService) ModifyTaxRecord(c core.Context, record *models.TaxRecord) error {
	if record.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	record.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(record.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.ID(record.TaxId).Cols("cfo_id", "tax_type", "period_year", "period_quarter", "taxable_income", "tax_amount", "paid_amount", "due_date", "status", "comment", "updated_unix_time").Where("uid=? AND deleted=?", record.Uid, false).Update(record)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrTaxRecordNotFound
		}

		return err
	})
}

// DeleteTaxRecord deletes an existed tax record from database
func (s *TaxRecordService) DeleteTaxRecord(c core.Context, uid int64, taxId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.TaxRecord{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		deletedRows, err := sess.ID(taxId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrTaxRecordNotFound
		}

		return err
	})
}
