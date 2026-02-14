package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

func initUuidContainer(t *testing.T) *uuid.UuidContainer {
	t.Helper()
	err := uuid.InitializeUuidGenerator(&settings.Config{
		UuidGeneratorType: settings.InternalUuidGeneratorType,
		UuidServerId:      1,
	})
	if err != nil {
		t.Fatalf("failed to init uuid generator: %v", err)
	}
	return uuid.Container
}

func newTestAssetService(t *testing.T) (*AssetService, *testDB) {
	t.Helper()
	tdb := newTestDB(t)
	uuidContainer := initUuidContainer(t)
	svc := &AssetService{
		ServiceUsingDB:   ServiceUsingDB{container: tdb.container},
		ServiceUsingUuid: ServiceUsingUuid{container: uuidContainer},
	}
	return svc, tdb
}

func TestAssetServiceCreateAndGet(t *testing.T) {
	svc, tdb := newTestAssetService(t)
	defer tdb.close()

	asset := &models.Asset{
		Uid:       1,
		Name:      "Test Laptop",
		AssetType: models.ASSET_TYPE_ELECTRONICS,
		Status:    models.ASSET_STATUS_ACTIVE,
	}

	err := svc.CreateAsset(nil, asset)
	assert.Nil(t, err)
	assert.True(t, asset.AssetId > 0)

	got, err := svc.GetAssetByAssetId(nil, 1, asset.AssetId)
	assert.Nil(t, err)
	assert.Equal(t, "Test Laptop", got.Name)
	assert.Equal(t, models.ASSET_TYPE_ELECTRONICS, got.AssetType)
	assert.Equal(t, models.ASSET_STATUS_ACTIVE, got.Status)
}

func TestAssetServiceGetAll(t *testing.T) {
	svc, tdb := newTestAssetService(t)
	defer tdb.close()

	for i := 0; i < 3; i++ {
		asset := &models.Asset{
			Uid:       1,
			Name:      "Asset" + string(rune('A'+i)),
			AssetType: models.ASSET_TYPE_EQUIPMENT,
			Status:    models.ASSET_STATUS_ACTIVE,
		}
		err := svc.CreateAsset(nil, asset)
		assert.Nil(t, err)
	}

	all, err := svc.GetAllAssetsByUid(nil, 1)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(all))
}

func TestAssetServiceModify(t *testing.T) {
	svc, tdb := newTestAssetService(t)
	defer tdb.close()

	asset := &models.Asset{
		Uid:       1,
		Name:      "Old Name",
		AssetType: models.ASSET_TYPE_VEHICLE,
		Status:    models.ASSET_STATUS_ACTIVE,
	}
	err := svc.CreateAsset(nil, asset)
	assert.Nil(t, err)

	asset.Name = "New Name"
	err = svc.ModifyAsset(nil, asset, true)
	assert.Nil(t, err)

	got, err := svc.GetAssetByAssetId(nil, 1, asset.AssetId)
	assert.Nil(t, err)
	assert.Equal(t, "New Name", got.Name)
}

func TestAssetServiceDelete(t *testing.T) {
	svc, tdb := newTestAssetService(t)
	defer tdb.close()

	asset := &models.Asset{
		Uid:       1,
		Name:      "ToDelete",
		AssetType: models.ASSET_TYPE_OTHER,
		Status:    models.ASSET_STATUS_ACTIVE,
	}
	err := svc.CreateAsset(nil, asset)
	assert.Nil(t, err)

	err = svc.DeleteAsset(nil, 1, asset.AssetId)
	assert.Nil(t, err)

	_, err = svc.GetAssetByAssetId(nil, 1, asset.AssetId)
	assert.Equal(t, errs.ErrAssetNotFound, err)
}

func TestAssetServiceInvalidUid(t *testing.T) {
	svc, tdb := newTestAssetService(t)
	defer tdb.close()

	_, err := svc.GetAllAssetsByUid(nil, 0)
	assert.Equal(t, errs.ErrUserIdInvalid, err)

	_, err = svc.GetAssetByAssetId(nil, 0, 1)
	assert.Equal(t, errs.ErrUserIdInvalid, err)

	err = svc.CreateAsset(nil, &models.Asset{Uid: 0, Name: "x"})
	assert.Equal(t, errs.ErrUserIdInvalid, err)
}

func TestAssetServiceUserIsolation(t *testing.T) {
	svc, tdb := newTestAssetService(t)
	defer tdb.close()

	a1 := &models.Asset{Uid: 1, Name: "User1Asset", AssetType: models.ASSET_TYPE_EQUIPMENT, Status: models.ASSET_STATUS_ACTIVE}
	a2 := &models.Asset{Uid: 2, Name: "User2Asset", AssetType: models.ASSET_TYPE_FURNITURE, Status: models.ASSET_STATUS_ACTIVE}

	assert.Nil(t, svc.CreateAsset(nil, a1))
	assert.Nil(t, svc.CreateAsset(nil, a2))

	list1, err := svc.GetAllAssetsByUid(nil, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list1))
	assert.Equal(t, "User1Asset", list1[0].Name)

	list2, err := svc.GetAllAssetsByUid(nil, 2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(list2))
	assert.Equal(t, "User2Asset", list2[0].Name)
}

func TestAssetServiceDuplicateName(t *testing.T) {
	svc, tdb := newTestAssetService(t)
	defer tdb.close()

	a1 := &models.Asset{Uid: 1, Name: "SameName", AssetType: models.ASSET_TYPE_EQUIPMENT, Status: models.ASSET_STATUS_ACTIVE}
	assert.Nil(t, svc.CreateAsset(nil, a1))

	a2 := &models.Asset{Uid: 1, Name: "SameName", AssetType: models.ASSET_TYPE_VEHICLE, Status: models.ASSET_STATUS_ACTIVE}
	err := svc.CreateAsset(nil, a2)
	assert.Equal(t, errs.ErrAssetNameAlreadyExists, err)
}
