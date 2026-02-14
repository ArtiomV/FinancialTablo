package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func newTestCFOService(t *testing.T) (*CFOService, *testDB) {
	t.Helper()
	tdb := newTestDB(t)
	uuidContainer := initUuidContainer(t)
	svc := &CFOService{
		ServiceUsingDB:   ServiceUsingDB{container: tdb.container},
		ServiceUsingUuid: ServiceUsingUuid{container: uuidContainer},
	}
	return svc, tdb
}

func TestCFOServiceCreateAndGet(t *testing.T) {
	svc, tdb := newTestCFOService(t)
	defer tdb.close()

	cfo := &models.CFO{
		Uid:   1,
		Name:  "Head Office",
		Color: "FF0000",
	}

	err := svc.CreateCFO(nil, cfo)
	assert.Nil(t, err)
	assert.True(t, cfo.CfoId > 0)

	got, err := svc.GetCFOByCFOId(nil, 1, cfo.CfoId)
	assert.Nil(t, err)
	assert.Equal(t, "Head Office", got.Name)
	assert.Equal(t, "FF0000", got.Color)
}

func TestCFOServiceGetAll(t *testing.T) {
	svc, tdb := newTestCFOService(t)
	defer tdb.close()

	for i := 0; i < 3; i++ {
		cfo := &models.CFO{
			Uid:   1,
			Name:  "CFO" + string(rune('A'+i)),
			Color: "00FF00",
		}
		assert.Nil(t, svc.CreateCFO(nil, cfo))
	}

	all, err := svc.GetAllCFOsByUid(nil, 1)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(all))
}

func TestCFOServiceModify(t *testing.T) {
	svc, tdb := newTestCFOService(t)
	defer tdb.close()

	cfo := &models.CFO{
		Uid:   1,
		Name:  "OldCFO",
		Color: "AABBCC",
	}
	assert.Nil(t, svc.CreateCFO(nil, cfo))

	cfo.Name = "NewCFO"
	cfo.Color = "DDEEFF"
	err := svc.ModifyCFO(nil, cfo, true)
	assert.Nil(t, err)

	got, err := svc.GetCFOByCFOId(nil, 1, cfo.CfoId)
	assert.Nil(t, err)
	assert.Equal(t, "NewCFO", got.Name)
	assert.Equal(t, "DDEEFF", got.Color)
}

func TestCFOServiceDelete(t *testing.T) {
	svc, tdb := newTestCFOService(t)
	defer tdb.close()

	cfo := &models.CFO{
		Uid:   1,
		Name:  "ToDelete",
		Color: "112233",
	}
	assert.Nil(t, svc.CreateCFO(nil, cfo))

	err := svc.DeleteCFO(nil, 1, cfo.CfoId)
	assert.Nil(t, err)

	_, err = svc.GetCFOByCFOId(nil, 1, cfo.CfoId)
	assert.Equal(t, errs.ErrCFONotFound, err)
}

func TestCFOServiceInvalidUid(t *testing.T) {
	svc, tdb := newTestCFOService(t)
	defer tdb.close()

	_, err := svc.GetAllCFOsByUid(nil, 0)
	assert.Equal(t, errs.ErrUserIdInvalid, err)

	_, err = svc.GetCFOByCFOId(nil, 0, 1)
	assert.Equal(t, errs.ErrUserIdInvalid, err)

	err = svc.CreateCFO(nil, &models.CFO{Uid: 0, Name: "x", Color: "AABBCC"})
	assert.Equal(t, errs.ErrUserIdInvalid, err)
}

func TestCFOServiceUserIsolation(t *testing.T) {
	svc, tdb := newTestCFOService(t)
	defer tdb.close()

	c1 := &models.CFO{Uid: 1, Name: "User1CFO", Color: "111111"}
	c2 := &models.CFO{Uid: 2, Name: "User2CFO", Color: "222222"}

	assert.Nil(t, svc.CreateCFO(nil, c1))
	assert.Nil(t, svc.CreateCFO(nil, c2))

	list1, err := svc.GetAllCFOsByUid(nil, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list1))
	assert.Equal(t, "User1CFO", list1[0].Name)

	list2, err := svc.GetAllCFOsByUid(nil, 2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list2))
	assert.Equal(t, "User2CFO", list2[0].Name)
}

func TestCFOServiceDuplicateName(t *testing.T) {
	svc, tdb := newTestCFOService(t)
	defer tdb.close()

	c1 := &models.CFO{Uid: 1, Name: "DupName", Color: "AAAAAA"}
	assert.Nil(t, svc.CreateCFO(nil, c1))

	c2 := &models.CFO{Uid: 1, Name: "DupName", Color: "BBBBBB"}
	err := svc.CreateCFO(nil, c2)
	assert.Equal(t, errs.ErrCFONameAlreadyExists, err)
}
