package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func newTestInvestorPaymentService(t *testing.T) (*InvestorPaymentService, *InvestorDealService, *testDB) {
	t.Helper()
	tdb := newTestDB(t)
	uuidContainer := initUuidContainer(t)
	paySvc := &InvestorPaymentService{
		ServiceUsingDB:   ServiceUsingDB{container: tdb.container},
		ServiceUsingUuid: ServiceUsingUuid{container: uuidContainer},
	}
	dealSvc := &InvestorDealService{
		ServiceUsingDB:   ServiceUsingDB{container: tdb.container},
		ServiceUsingUuid: ServiceUsingUuid{container: uuidContainer},
	}
	return paySvc, dealSvc, tdb
}

func createTestDeal(t *testing.T, dealSvc *InvestorDealService, uid int64) *models.InvestorDeal {
	t.Helper()
	deal := &models.InvestorDeal{
		Uid:              uid,
		InvestorName:     "TestInvestor",
		InvestmentAmount: 500000,
		Currency:         "RUB",
		DealType:         models.INVESTOR_DEAL_TYPE_LOAN,
	}
	err := dealSvc.CreateDeal(nil, deal)
	assert.Nil(t, err)
	return deal
}

func TestInvestorPaymentServiceCreateAndGet(t *testing.T) {
	svc, dealSvc, tdb := newTestInvestorPaymentService(t)
	defer tdb.close()

	deal := createTestDeal(t, dealSvc, 1)

	payment := &models.InvestorPayment{
		Uid:         1,
		DealId:      deal.DealId,
		Amount:      25000,
		PaymentType: models.INVESTOR_PAYMENT_TYPE_INTEREST,
		Comment:     "Monthly interest",
	}

	err := svc.CreatePayment(nil, payment)
	assert.Nil(t, err)
	assert.True(t, payment.PaymentId > 0)

	got, err := svc.GetPaymentByPaymentId(nil, 1, payment.PaymentId)
	assert.Nil(t, err)
	assert.Equal(t, deal.DealId, got.DealId)
	assert.Equal(t, int64(25000), got.Amount)
	assert.Equal(t, models.INVESTOR_PAYMENT_TYPE_INTEREST, got.PaymentType)
	assert.Equal(t, "Monthly interest", got.Comment)
}

func TestInvestorPaymentServiceGetAllByDealId(t *testing.T) {
	svc, dealSvc, tdb := newTestInvestorPaymentService(t)
	defer tdb.close()

	deal := createTestDeal(t, dealSvc, 1)

	for i := 0; i < 3; i++ {
		payment := &models.InvestorPayment{
			Uid:         1,
			DealId:      deal.DealId,
			Amount:      int64(i+1) * 10000,
			PaymentType: models.INVESTOR_PAYMENT_TYPE_MIXED,
		}
		assert.Nil(t, svc.CreatePayment(nil, payment))
	}

	all, err := svc.GetAllPaymentsByDealId(nil, 1, deal.DealId)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(all))
}

func TestInvestorPaymentServiceModify(t *testing.T) {
	svc, dealSvc, tdb := newTestInvestorPaymentService(t)
	defer tdb.close()

	deal := createTestDeal(t, dealSvc, 1)

	payment := &models.InvestorPayment{
		Uid:         1,
		DealId:      deal.DealId,
		Amount:      15000,
		PaymentType: models.INVESTOR_PAYMENT_TYPE_PRINCIPAL,
	}
	assert.Nil(t, svc.CreatePayment(nil, payment))

	payment.Amount = 20000
	payment.Comment = "Updated payment"
	err := svc.ModifyPayment(nil, payment)
	assert.Nil(t, err)

	got, err := svc.GetPaymentByPaymentId(nil, 1, payment.PaymentId)
	assert.Nil(t, err)
	assert.Equal(t, int64(20000), got.Amount)
	assert.Equal(t, "Updated payment", got.Comment)
}

func TestInvestorPaymentServiceDelete(t *testing.T) {
	svc, dealSvc, tdb := newTestInvestorPaymentService(t)
	defer tdb.close()

	deal := createTestDeal(t, dealSvc, 1)

	payment := &models.InvestorPayment{
		Uid:         1,
		DealId:      deal.DealId,
		Amount:      10000,
		PaymentType: models.INVESTOR_PAYMENT_TYPE_INTEREST,
	}
	assert.Nil(t, svc.CreatePayment(nil, payment))

	err := svc.DeletePayment(nil, 1, payment.PaymentId)
	assert.Nil(t, err)

	_, err = svc.GetPaymentByPaymentId(nil, 1, payment.PaymentId)
	assert.Equal(t, errs.ErrInvestorPaymentNotFound, err)
}

func TestInvestorPaymentServiceInvalidUid(t *testing.T) {
	svc, _, tdb := newTestInvestorPaymentService(t)
	defer tdb.close()

	_, err := svc.GetAllPaymentsByDealId(nil, 0, 1)
	assert.Equal(t, errs.ErrUserIdInvalid, err)

	_, err = svc.GetPaymentByPaymentId(nil, 0, 1)
	assert.Equal(t, errs.ErrUserIdInvalid, err)

	err = svc.CreatePayment(nil, &models.InvestorPayment{Uid: 0, DealId: 1})
	assert.Equal(t, errs.ErrUserIdInvalid, err)
}

func TestInvestorPaymentServiceUserIsolation(t *testing.T) {
	svc, dealSvc, tdb := newTestInvestorPaymentService(t)
	defer tdb.close()

	deal1 := createTestDeal(t, dealSvc, 1)
	deal2 := &models.InvestorDeal{
		Uid:              2,
		InvestorName:     "User2Investor",
		InvestmentAmount: 300000,
		Currency:         "RUB",
		DealType:         models.INVESTOR_DEAL_TYPE_EQUITY,
	}
	assert.Nil(t, dealSvc.CreateDeal(nil, deal2))

	p1 := &models.InvestorPayment{Uid: 1, DealId: deal1.DealId, Amount: 10000, PaymentType: models.INVESTOR_PAYMENT_TYPE_PRINCIPAL, Comment: "user1pay"}
	p2 := &models.InvestorPayment{Uid: 2, DealId: deal2.DealId, Amount: 20000, PaymentType: models.INVESTOR_PAYMENT_TYPE_INTEREST, Comment: "user2pay"}

	assert.Nil(t, svc.CreatePayment(nil, p1))
	assert.Nil(t, svc.CreatePayment(nil, p2))

	list1, err := svc.GetAllPaymentsByUid(nil, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list1))
	assert.Equal(t, "user1pay", list1[0].Comment)

	list2, err := svc.GetAllPaymentsByUid(nil, 2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list2))
	assert.Equal(t, "user2pay", list2[0].Comment)
}

func TestInvestorPaymentServiceInvalidDealId(t *testing.T) {
	svc, _, tdb := newTestInvestorPaymentService(t)
	defer tdb.close()

	payment := &models.InvestorPayment{
		Uid:         1,
		DealId:      0,
		Amount:      10000,
		PaymentType: models.INVESTOR_PAYMENT_TYPE_PRINCIPAL,
	}
	err := svc.CreatePayment(nil, payment)
	assert.Equal(t, errs.ErrInvestorDealIdInvalid, err)
}
