package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func newTestObligationService(t *testing.T) (*ObligationService, *testDB) {
	t.Helper()
	tdb := newTestDB(t)
	uuidContainer := initUuidContainer(t)
	svc := &ObligationService{
		ServiceUsingDB:   ServiceUsingDB{container: tdb.container},
		ServiceUsingUuid: ServiceUsingUuid{container: uuidContainer},
	}
	return svc, tdb
}

func TestObligationServiceCreateAndGet(t *testing.T) {
	svc, tdb := newTestObligationService(t)
	defer tdb.close()

	ob := &models.Obligation{
		Uid:            1,
		ObligationType: models.OBLIGATION_TYPE_RECEIVABLE,
		Amount:         100000,
		Currency:       "RUB",
		Status:         models.OBLIGATION_STATUS_ACTIVE,
		Comment:        "Test obligation",
	}

	err := svc.CreateObligation(nil, ob)
	assert.Nil(t, err)
	assert.True(t, ob.ObligationId > 0)

	got, err := svc.GetObligationByObligationId(nil, 1, ob.ObligationId)
	assert.Nil(t, err)
	assert.Equal(t, models.OBLIGATION_TYPE_RECEIVABLE, got.ObligationType)
	assert.Equal(t, int64(100000), got.Amount)
	assert.Equal(t, "RUB", got.Currency)
	assert.Equal(t, "Test obligation", got.Comment)
}

func TestObligationServiceGetAll(t *testing.T) {
	svc, tdb := newTestObligationService(t)
	defer tdb.close()

	for i := 0; i < 3; i++ {
		ob := &models.Obligation{
			Uid:            1,
			ObligationType: models.OBLIGATION_TYPE_PAYABLE,
			Amount:         int64(i+1) * 10000,
			Currency:       "RUB",
			Status:         models.OBLIGATION_STATUS_ACTIVE,
		}
		assert.Nil(t, svc.CreateObligation(nil, ob))
	}

	all, err := svc.GetAllObligationsByUid(nil, 1)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(all))
}

func TestObligationServiceModify(t *testing.T) {
	svc, tdb := newTestObligationService(t)
	defer tdb.close()

	ob := &models.Obligation{
		Uid:            1,
		ObligationType: models.OBLIGATION_TYPE_RECEIVABLE,
		Amount:         50000,
		Currency:       "RUB",
		Status:         models.OBLIGATION_STATUS_ACTIVE,
	}
	assert.Nil(t, svc.CreateObligation(nil, ob))

	ob.Amount = 75000
	ob.Status = models.OBLIGATION_STATUS_PARTIAL
	ob.PaidAmount = 25000
	err := svc.ModifyObligation(nil, ob)
	assert.Nil(t, err)

	got, err := svc.GetObligationByObligationId(nil, 1, ob.ObligationId)
	assert.Nil(t, err)
	assert.Equal(t, int64(75000), got.Amount)
	assert.Equal(t, models.OBLIGATION_STATUS_PARTIAL, got.Status)
	assert.Equal(t, int64(25000), got.PaidAmount)
}

func TestObligationServiceDelete(t *testing.T) {
	svc, tdb := newTestObligationService(t)
	defer tdb.close()

	ob := &models.Obligation{
		Uid:            1,
		ObligationType: models.OBLIGATION_TYPE_PAYABLE,
		Amount:         30000,
		Currency:       "RUB",
		Status:         models.OBLIGATION_STATUS_ACTIVE,
	}
	assert.Nil(t, svc.CreateObligation(nil, ob))

	err := svc.DeleteObligation(nil, 1, ob.ObligationId)
	assert.Nil(t, err)

	_, err = svc.GetObligationByObligationId(nil, 1, ob.ObligationId)
	assert.Equal(t, errs.ErrObligationNotFound, err)
}

func TestObligationServiceInvalidUid(t *testing.T) {
	svc, tdb := newTestObligationService(t)
	defer tdb.close()

	_, err := svc.GetAllObligationsByUid(nil, 0)
	assert.Equal(t, errs.ErrUserIdInvalid, err)

	_, err = svc.GetObligationByObligationId(nil, 0, 1)
	assert.Equal(t, errs.ErrUserIdInvalid, err)

	err = svc.CreateObligation(nil, &models.Obligation{Uid: 0, Currency: "RUB"})
	assert.Equal(t, errs.ErrUserIdInvalid, err)
}

func TestObligationServiceUserIsolation(t *testing.T) {
	svc, tdb := newTestObligationService(t)
	defer tdb.close()

	o1 := &models.Obligation{Uid: 1, ObligationType: models.OBLIGATION_TYPE_RECEIVABLE, Amount: 10000, Currency: "RUB", Status: models.OBLIGATION_STATUS_ACTIVE, Comment: "user1"}
	o2 := &models.Obligation{Uid: 2, ObligationType: models.OBLIGATION_TYPE_PAYABLE, Amount: 20000, Currency: "USD", Status: models.OBLIGATION_STATUS_ACTIVE, Comment: "user2"}

	assert.Nil(t, svc.CreateObligation(nil, o1))
	assert.Nil(t, svc.CreateObligation(nil, o2))

	list1, err := svc.GetAllObligationsByUid(nil, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list1))
	assert.Equal(t, "user1", list1[0].Comment)

	list2, err := svc.GetAllObligationsByUid(nil, 2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list2))
	assert.Equal(t, "user2", list2[0].Comment)
}
