package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func newTestTaxRecordService(t *testing.T) (*TaxRecordService, *testDB) {
	t.Helper()
	tdb := newTestDB(t)
	uuidContainer := initUuidContainer(t)
	svc := &TaxRecordService{
		ServiceUsingDB:   ServiceUsingDB{container: tdb.container},
		ServiceUsingUuid: ServiceUsingUuid{container: uuidContainer},
	}
	return svc, tdb
}

func TestTaxRecordServiceCreateAndGet(t *testing.T) {
	svc, tdb := newTestTaxRecordService(t)
	defer tdb.close()

	rec := &models.TaxRecord{
		Uid:           1,
		TaxType:       models.TAX_TYPE_INCOME,
		PeriodYear:    2025,
		PeriodQuarter: 1,
		TaxAmount:     50000,
		Currency:      "RUB",
		Status:        models.TAX_STATUS_PENDING,
	}

	err := svc.CreateTaxRecord(nil, rec)
	assert.Nil(t, err)
	assert.True(t, rec.TaxId > 0)

	got, err := svc.GetTaxRecordByTaxId(nil, 1, rec.TaxId)
	assert.Nil(t, err)
	assert.Equal(t, models.TAX_TYPE_INCOME, got.TaxType)
	assert.Equal(t, int32(2025), got.PeriodYear)
	assert.Equal(t, int32(1), got.PeriodQuarter)
	assert.Equal(t, int64(50000), got.TaxAmount)
	assert.Equal(t, "RUB", got.Currency)
}

func TestTaxRecordServiceGetAll(t *testing.T) {
	svc, tdb := newTestTaxRecordService(t)
	defer tdb.close()

	for i := 0; i < 3; i++ {
		rec := &models.TaxRecord{
			Uid:           1,
			TaxType:       models.TAX_TYPE_VAT,
			PeriodYear:    2025,
			PeriodQuarter: int32(i + 1),
			TaxAmount:     int64(i+1) * 10000,
			Currency:      "RUB",
			Status:        models.TAX_STATUS_PENDING,
		}
		assert.Nil(t, svc.CreateTaxRecord(nil, rec))
	}

	all, err := svc.GetAllTaxRecordsByUid(nil, 1)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(all))
}

func TestTaxRecordServiceModify(t *testing.T) {
	svc, tdb := newTestTaxRecordService(t)
	defer tdb.close()

	rec := &models.TaxRecord{
		Uid:       1,
		TaxType:   models.TAX_TYPE_PROPERTY,
		TaxAmount: 30000,
		Currency:  "RUB",
		Status:    models.TAX_STATUS_PENDING,
	}
	assert.Nil(t, svc.CreateTaxRecord(nil, rec))

	rec.TaxAmount = 45000
	rec.Status = models.TAX_STATUS_PAID
	rec.PaidAmount = 45000
	err := svc.ModifyTaxRecord(nil, rec)
	assert.Nil(t, err)

	got, err := svc.GetTaxRecordByTaxId(nil, 1, rec.TaxId)
	assert.Nil(t, err)
	assert.Equal(t, int64(45000), got.TaxAmount)
	assert.Equal(t, models.TAX_STATUS_PAID, got.Status)
	assert.Equal(t, int64(45000), got.PaidAmount)
}

func TestTaxRecordServiceDelete(t *testing.T) {
	svc, tdb := newTestTaxRecordService(t)
	defer tdb.close()

	rec := &models.TaxRecord{
		Uid:       1,
		TaxType:   models.TAX_TYPE_OTHER,
		TaxAmount: 5000,
		Currency:  "RUB",
		Status:    models.TAX_STATUS_PENDING,
	}
	assert.Nil(t, svc.CreateTaxRecord(nil, rec))

	err := svc.DeleteTaxRecord(nil, 1, rec.TaxId)
	assert.Nil(t, err)

	_, err = svc.GetTaxRecordByTaxId(nil, 1, rec.TaxId)
	assert.Equal(t, errs.ErrTaxRecordNotFound, err)
}

func TestTaxRecordServiceInvalidUid(t *testing.T) {
	svc, tdb := newTestTaxRecordService(t)
	defer tdb.close()

	_, err := svc.GetAllTaxRecordsByUid(nil, 0)
	assert.Equal(t, errs.ErrUserIdInvalid, err)

	_, err = svc.GetTaxRecordByTaxId(nil, 0, 1)
	assert.Equal(t, errs.ErrUserIdInvalid, err)

	err = svc.CreateTaxRecord(nil, &models.TaxRecord{Uid: 0, Currency: "RUB"})
	assert.Equal(t, errs.ErrUserIdInvalid, err)
}

func TestTaxRecordServiceUserIsolation(t *testing.T) {
	svc, tdb := newTestTaxRecordService(t)
	defer tdb.close()

	r1 := &models.TaxRecord{Uid: 1, TaxType: models.TAX_TYPE_INCOME, TaxAmount: 10000, Currency: "RUB", Status: models.TAX_STATUS_PENDING, Comment: "user1"}
	r2 := &models.TaxRecord{Uid: 2, TaxType: models.TAX_TYPE_VAT, TaxAmount: 20000, Currency: "USD", Status: models.TAX_STATUS_PENDING, Comment: "user2"}

	assert.Nil(t, svc.CreateTaxRecord(nil, r1))
	assert.Nil(t, svc.CreateTaxRecord(nil, r2))

	list1, err := svc.GetAllTaxRecordsByUid(nil, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list1))
	assert.Equal(t, "user1", list1[0].Comment)

	list2, err := svc.GetAllTaxRecordsByUid(nil, 2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list2))
	assert.Equal(t, "user2", list2[0].Comment)
}
