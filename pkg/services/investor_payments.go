// investor_payments.go provides CRUD for individual investor payment records.
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

// InvestorPaymentService represents investor payment service
type InvestorPaymentService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize an investor payment service singleton instance
var (
	InvestorPayments = &InvestorPaymentService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetAllPaymentsByDealId returns all investor payment models for a deal
func (s *InvestorPaymentService) GetAllPaymentsByDealId(c core.Context, uid int64, dealId int64) ([]*models.InvestorPayment, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if dealId <= 0 {
		return nil, errs.ErrInvestorDealIdInvalid
	}

	var payments []*models.InvestorPayment
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=? AND deal_id=?", uid, false, dealId).OrderBy("payment_date desc").Find(&payments)

	return payments, err
}

// GetAllPaymentsByDealIds returns all investor payments for the given deal IDs
func (s *InvestorPaymentService) GetAllPaymentsByDealIds(c core.Context, uid int64, dealIds []int64) (map[int64][]*models.InvestorPayment, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if len(dealIds) == 0 {
		return make(map[int64][]*models.InvestorPayment), nil
	}

	var payments []*models.InvestorPayment
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).In("deal_id", dealIds).Find(&payments)

	if err != nil {
		return nil, err
	}

	result := make(map[int64][]*models.InvestorPayment)
	for _, p := range payments {
		result[p.DealId] = append(result[p.DealId], p)
	}

	return result, nil
}

// GetAllPaymentsByUid returns all investor payment models for a user
func (s *InvestorPaymentService) GetAllPaymentsByUid(c core.Context, uid int64) ([]*models.InvestorPayment, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var payments []*models.InvestorPayment
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).OrderBy("payment_date desc").Find(&payments)

	return payments, err
}

// GetPaymentByPaymentId returns an investor payment model according to payment id
func (s *InvestorPaymentService) GetPaymentByPaymentId(c core.Context, uid int64, paymentId int64) (*models.InvestorPayment, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if paymentId <= 0 {
		return nil, errs.ErrInvestorPaymentIdInvalid
	}

	payment := &models.InvestorPayment{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(paymentId).Where("uid=? AND deleted=?", uid, false).Get(payment)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrInvestorPaymentNotFound
	}

	return payment, nil
}

// CreatePayment saves a new investor payment model to database
func (s *InvestorPaymentService) CreatePayment(c core.Context, payment *models.InvestorPayment) error {
	if payment.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if payment.DealId <= 0 {
		return errs.ErrInvestorDealIdInvalid
	}

	payment.PaymentId = s.GenerateUuid(uuid.UUID_TYPE_DEFAULT)

	if payment.PaymentId < 1 {
		return errs.ErrSystemIsBusy
	}

	payment.Deleted = false
	payment.CreatedUnixTime = time.Now().Unix()
	payment.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(payment.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(payment)
		return err
	})
}

// ModifyPayment saves an existed investor payment model to database
func (s *InvestorPaymentService) ModifyPayment(c core.Context, payment *models.InvestorPayment) error {
	if payment.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	payment.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(payment.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.ID(payment.PaymentId).Cols("deal_id", "payment_date", "amount", "payment_type", "transaction_id", "comment", "updated_unix_time").Where("uid=? AND deleted=?", payment.Uid, false).Update(payment)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrInvestorPaymentNotFound
		}

		return err
	})
}

// DeletePayment deletes an existed investor payment from database
func (s *InvestorPaymentService) DeletePayment(c core.Context, uid int64, paymentId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.InvestorPayment{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		deletedRows, err := sess.ID(paymentId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrInvestorPaymentNotFound
		}

		return err
	})
}
