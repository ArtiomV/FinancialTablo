package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func newTestInvestorDealService(t *testing.T) (*InvestorDealService, *testDB) {
	t.Helper()
	tdb := newTestDB(t)
	uuidContainer := initUuidContainer(t)
	svc := &InvestorDealService{
		ServiceUsingDB:   ServiceUsingDB{container: tdb.container},
		ServiceUsingUuid: ServiceUsingUuid{container: uuidContainer},
	}
	return svc, tdb
}

func TestInvestorDealServiceCreateAndGet(t *testing.T) {
	svc, tdb := newTestInvestorDealService(t)
	defer tdb.close()

	deal := &models.InvestorDeal{
		Uid:              1,
		InvestorName:     "John Doe",
		InvestmentAmount: 1000000,
		Currency:         "RUB",
		DealType:         models.INVESTOR_DEAL_TYPE_LOAN,
		AnnualRate:       1500,
	}

	err := svc.CreateDeal(nil, deal)
	assert.Nil(t, err)
	assert.True(t, deal.DealId > 0)

	got, err := svc.GetDealByDealId(nil, 1, deal.DealId)
	assert.Nil(t, err)
	assert.Equal(t, "John Doe", got.InvestorName)
	assert.Equal(t, int64(1000000), got.InvestmentAmount)
	assert.Equal(t, "RUB", got.Currency)
	assert.Equal(t, models.INVESTOR_DEAL_TYPE_LOAN, got.DealType)
	assert.Equal(t, int32(1500), got.AnnualRate)
}

func TestInvestorDealServiceGetAll(t *testing.T) {
	svc, tdb := newTestInvestorDealService(t)
	defer tdb.close()

	for i := 0; i < 3; i++ {
		deal := &models.InvestorDeal{
			Uid:              1,
			InvestorName:     "Investor" + string(rune('A'+i)),
			InvestmentAmount: int64(i+1) * 500000,
			Currency:         "RUB",
			DealType:         models.INVESTOR_DEAL_TYPE_EQUITY,
		}
		assert.Nil(t, svc.CreateDeal(nil, deal))
	}

	all, err := svc.GetAllDealsByUid(nil, 1)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(all))
}

func TestInvestorDealServiceModify(t *testing.T) {
	svc, tdb := newTestInvestorDealService(t)
	defer tdb.close()

	deal := &models.InvestorDeal{
		Uid:              1,
		InvestorName:     "OldInvestor",
		InvestmentAmount: 200000,
		Currency:         "RUB",
		DealType:         models.INVESTOR_DEAL_TYPE_LOAN,
	}
	assert.Nil(t, svc.CreateDeal(nil, deal))

	deal.InvestorName = "NewInvestor"
	deal.InvestmentAmount = 300000
	err := svc.ModifyDeal(nil, deal)
	assert.Nil(t, err)

	got, err := svc.GetDealByDealId(nil, 1, deal.DealId)
	assert.Nil(t, err)
	assert.Equal(t, "NewInvestor", got.InvestorName)
	assert.Equal(t, int64(300000), got.InvestmentAmount)
}

func TestInvestorDealServiceDelete(t *testing.T) {
	svc, tdb := newTestInvestorDealService(t)
	defer tdb.close()

	deal := &models.InvestorDeal{
		Uid:              1,
		InvestorName:     "ToDelete",
		InvestmentAmount: 100000,
		Currency:         "RUB",
		DealType:         models.INVESTOR_DEAL_TYPE_OTHER,
	}
	assert.Nil(t, svc.CreateDeal(nil, deal))

	err := svc.DeleteDeal(nil, 1, deal.DealId)
	assert.Nil(t, err)

	_, err = svc.GetDealByDealId(nil, 1, deal.DealId)
	assert.Equal(t, errs.ErrInvestorDealNotFound, err)
}

func TestInvestorDealServiceInvalidUid(t *testing.T) {
	svc, tdb := newTestInvestorDealService(t)
	defer tdb.close()

	_, err := svc.GetAllDealsByUid(nil, 0)
	assert.Equal(t, errs.ErrUserIdInvalid, err)

	_, err = svc.GetDealByDealId(nil, 0, 1)
	assert.Equal(t, errs.ErrUserIdInvalid, err)

	err = svc.CreateDeal(nil, &models.InvestorDeal{Uid: 0, InvestorName: "x", Currency: "RUB"})
	assert.Equal(t, errs.ErrUserIdInvalid, err)
}

func TestInvestorDealServiceUserIsolation(t *testing.T) {
	svc, tdb := newTestInvestorDealService(t)
	defer tdb.close()

	d1 := &models.InvestorDeal{Uid: 1, InvestorName: "User1Investor", InvestmentAmount: 10000, Currency: "RUB", DealType: models.INVESTOR_DEAL_TYPE_LOAN}
	d2 := &models.InvestorDeal{Uid: 2, InvestorName: "User2Investor", InvestmentAmount: 20000, Currency: "USD", DealType: models.INVESTOR_DEAL_TYPE_EQUITY}

	assert.Nil(t, svc.CreateDeal(nil, d1))
	assert.Nil(t, svc.CreateDeal(nil, d2))

	list1, err := svc.GetAllDealsByUid(nil, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list1))
	assert.Equal(t, "User1Investor", list1[0].InvestorName)

	list2, err := svc.GetAllDealsByUid(nil, 2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list2))
	assert.Equal(t, "User2Investor", list2[0].InvestorName)
}

func TestInvestorDealServiceEmptyName(t *testing.T) {
	svc, tdb := newTestInvestorDealService(t)
	defer tdb.close()

	deal := &models.InvestorDeal{
		Uid:          1,
		InvestorName: "",
		Currency:     "RUB",
		DealType:     models.INVESTOR_DEAL_TYPE_LOAN,
	}
	err := svc.CreateDeal(nil, deal)
	assert.Equal(t, errs.ErrInvestorDealInvestorNameIsEmpty, err)
}
