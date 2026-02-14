// investor_deals.go provides CRUD for investor deals with repayment tracking.
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

// InvestorDealService represents investor deal service
type InvestorDealService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize an investor deal service singleton instance
var (
	InvestorDeals = &InvestorDealService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetAllDealsByUid returns all investor deal models of user
func (s *InvestorDealService) GetAllDealsByUid(c core.Context, uid int64) ([]*models.InvestorDeal, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var deals []*models.InvestorDeal
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).OrderBy("investment_date desc").Find(&deals)

	return deals, err
}

// GetDealByDealId returns an investor deal model according to deal id
func (s *InvestorDealService) GetDealByDealId(c core.Context, uid int64, dealId int64) (*models.InvestorDeal, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if dealId <= 0 {
		return nil, errs.ErrInvestorDealIdInvalid
	}

	deal := &models.InvestorDeal{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(dealId).Where("uid=? AND deleted=?", uid, false).Get(deal)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrInvestorDealNotFound
	}

	return deal, nil
}

// CreateDeal saves a new investor deal model to database
func (s *InvestorDealService) CreateDeal(c core.Context, deal *models.InvestorDeal) error {
	if deal.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if deal.InvestorName == "" {
		return errs.ErrInvestorDealInvestorNameIsEmpty
	}

	deal.DealId = s.GenerateUuid(uuid.UUID_TYPE_DEFAULT)

	if deal.DealId < 1 {
		return errs.ErrSystemIsBusy
	}

	deal.Deleted = false
	deal.CreatedUnixTime = time.Now().Unix()
	deal.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(deal.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(deal)
		return err
	})
}

// ModifyDeal saves an existed investor deal model to database
func (s *InvestorDealService) ModifyDeal(c core.Context, deal *models.InvestorDeal) error {
	if deal.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if deal.InvestorName == "" {
		return errs.ErrInvestorDealInvestorNameIsEmpty
	}

	deal.UpdatedUnixTime = time.Now().Unix()

	return s.UserDataDB(deal.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.ID(deal.DealId).Cols("investor_name", "cfo_id", "investment_date", "investment_amount", "currency", "deal_type", "annual_rate", "profit_share_pct", "fixed_payment", "repayment_start_date", "repayment_end_date", "total_to_repay", "comment", "updated_unix_time").Where("uid=? AND deleted=?", deal.Uid, false).Update(deal)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrInvestorDealNotFound
		}

		return err
	})
}

// DeleteDeal deletes an existed investor deal from database
func (s *InvestorDealService) DeleteDeal(c core.Context, uid int64, dealId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()

	updateModel := &models.InvestorDeal{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		deletedRows, err := sess.ID(dealId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrInvestorDealNotFound
		}

		return err
	})
}
