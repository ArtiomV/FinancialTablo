package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func newTestLocationService(t *testing.T) (*LocationService, *testDB) {
	t.Helper()
	tdb := newTestDB(t)
	uuidContainer := initUuidContainer(t)
	svc := &LocationService{
		ServiceUsingDB:   ServiceUsingDB{container: tdb.container},
		ServiceUsingUuid: ServiceUsingUuid{container: uuidContainer},
	}
	return svc, tdb
}

func TestLocationServiceCreateAndGet(t *testing.T) {
	svc, tdb := newTestLocationService(t)
	defer tdb.close()

	loc := &models.Location{
		Uid:          1,
		Name:         "Main Office",
		Address:      "123 Main St",
		LocationType: models.LOCATION_TYPE_OFFICE,
	}

	err := svc.CreateLocation(nil, loc)
	assert.Nil(t, err)
	assert.True(t, loc.LocationId > 0)

	got, err := svc.GetLocationByLocationId(nil, 1, loc.LocationId)
	assert.Nil(t, err)
	assert.Equal(t, "Main Office", got.Name)
	assert.Equal(t, "123 Main St", got.Address)
	assert.Equal(t, models.LOCATION_TYPE_OFFICE, got.LocationType)
}

func TestLocationServiceGetAll(t *testing.T) {
	svc, tdb := newTestLocationService(t)
	defer tdb.close()

	for i := 0; i < 3; i++ {
		loc := &models.Location{
			Uid:          1,
			Name:         "Loc" + string(rune('A'+i)),
			LocationType: models.LOCATION_TYPE_WAREHOUSE,
		}
		assert.Nil(t, svc.CreateLocation(nil, loc))
	}

	all, err := svc.GetAllLocationsByUid(nil, 1)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(all))
}

func TestLocationServiceModify(t *testing.T) {
	svc, tdb := newTestLocationService(t)
	defer tdb.close()

	loc := &models.Location{
		Uid:          1,
		Name:         "OldLoc",
		Address:      "Old Address",
		LocationType: models.LOCATION_TYPE_STORE,
	}
	assert.Nil(t, svc.CreateLocation(nil, loc))

	loc.Name = "NewLoc"
	loc.Address = "New Address"
	err := svc.ModifyLocation(nil, loc, true)
	assert.Nil(t, err)

	got, err := svc.GetLocationByLocationId(nil, 1, loc.LocationId)
	assert.Nil(t, err)
	assert.Equal(t, "NewLoc", got.Name)
	assert.Equal(t, "New Address", got.Address)
}

func TestLocationServiceDelete(t *testing.T) {
	svc, tdb := newTestLocationService(t)
	defer tdb.close()

	loc := &models.Location{
		Uid:          1,
		Name:         "ToDelete",
		LocationType: models.LOCATION_TYPE_OTHER,
	}
	assert.Nil(t, svc.CreateLocation(nil, loc))

	err := svc.DeleteLocation(nil, 1, loc.LocationId)
	assert.Nil(t, err)

	_, err = svc.GetLocationByLocationId(nil, 1, loc.LocationId)
	assert.Equal(t, errs.ErrLocationNotFound, err)
}

func TestLocationServiceInvalidUid(t *testing.T) {
	svc, tdb := newTestLocationService(t)
	defer tdb.close()

	_, err := svc.GetAllLocationsByUid(nil, 0)
	assert.Equal(t, errs.ErrUserIdInvalid, err)

	_, err = svc.GetLocationByLocationId(nil, 0, 1)
	assert.Equal(t, errs.ErrUserIdInvalid, err)

	err = svc.CreateLocation(nil, &models.Location{Uid: 0, Name: "x"})
	assert.Equal(t, errs.ErrUserIdInvalid, err)
}

func TestLocationServiceUserIsolation(t *testing.T) {
	svc, tdb := newTestLocationService(t)
	defer tdb.close()

	l1 := &models.Location{Uid: 1, Name: "User1Loc", LocationType: models.LOCATION_TYPE_OFFICE}
	l2 := &models.Location{Uid: 2, Name: "User2Loc", LocationType: models.LOCATION_TYPE_WAREHOUSE}

	assert.Nil(t, svc.CreateLocation(nil, l1))
	assert.Nil(t, svc.CreateLocation(nil, l2))

	list1, err := svc.GetAllLocationsByUid(nil, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list1))
	assert.Equal(t, "User1Loc", list1[0].Name)

	list2, err := svc.GetAllLocationsByUid(nil, 2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list2))
	assert.Equal(t, "User2Loc", list2[0].Name)
}

func TestLocationServiceDuplicateName(t *testing.T) {
	svc, tdb := newTestLocationService(t)
	defer tdb.close()

	l1 := &models.Location{Uid: 1, Name: "DupLoc", LocationType: models.LOCATION_TYPE_OFFICE}
	assert.Nil(t, svc.CreateLocation(nil, l1))

	l2 := &models.Location{Uid: 1, Name: "DupLoc", LocationType: models.LOCATION_TYPE_WAREHOUSE}
	err := svc.CreateLocation(nil, l2)
	assert.Equal(t, errs.ErrLocationNameAlreadyExists, err)
}
