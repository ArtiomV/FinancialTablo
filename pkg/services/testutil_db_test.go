package services

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// testDB holds a test database instance
type testDB struct {
	engine    *xorm.Engine
	container *datastore.DataStoreContainer
}

// newTestDB creates an in-memory SQLite database with all required tables synced
func newTestDB(t *testing.T) *testDB {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create test engine: %v", err)
	}

	// Sync all models that services use
	err = engine.Sync2(
		new(models.Transaction),
		new(models.TransactionCategory),
		new(models.Account),
		new(models.Asset),
		new(models.Obligation),
		new(models.TaxRecord),
		new(models.CFO),
		new(models.Location),
		new(models.Budget),
		new(models.InvestorDeal),
		new(models.InvestorPayment),
	)
	if err != nil {
		t.Fatalf("failed to sync tables: %v", err)
	}

	// Create EngineGroup from single engine
	group, err := xorm.NewEngineGroup(engine, []*xorm.Engine{})
	if err != nil {
		t.Fatalf("failed to create engine group: %v", err)
	}

	db := datastore.NewDatabaseForTest(group, "sqlite3")
	container := datastore.NewContainerForTest(db)

	return &testDB{engine: engine, container: container}
}

// close releases the test database
func (tdb *testDB) close() {
	tdb.engine.Close()
}
