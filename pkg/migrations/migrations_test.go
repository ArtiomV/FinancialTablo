package migrations

import (
	"testing"

	"github.com/stretchr/testify/assert"
	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
)

func TestRunMigrations_Success(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	assert.Nil(t, err)
	defer engine.Close()

	err = RunMigrations(engine)
	assert.Nil(t, err)

	// Check that migration was recorded
	var records []MigrationRecord
	err = engine.Find(&records)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(records))
	assert.Equal(t, 1, records[0].Version)
	assert.True(t, records[0].AppliedUnixTime > 0)
}

func TestRunMigrations_Idempotent(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	assert.Nil(t, err)
	defer engine.Close()

	// Run twice — should not fail
	err = RunMigrations(engine)
	assert.Nil(t, err)
	err = RunMigrations(engine)
	assert.Nil(t, err)

	var records []MigrationRecord
	engine.Find(&records)
	assert.Equal(t, 1, len(records)) // still only 1 record
}

func TestAllMigrations_NotEmpty(t *testing.T) {
	migrations := AllMigrations()
	assert.True(t, len(migrations) > 0)
	assert.True(t, len(migrations[0].SQL) > 100)
	assert.Equal(t, 1, migrations[0].Version)
}

func TestSplitStatements_Basic(t *testing.T) {
	sql := "CREATE TABLE foo (id INT);\nCREATE TABLE bar (id INT);"
	stmts := splitStatements(sql)
	assert.Equal(t, 2, len(stmts))
}

func TestSplitStatements_SkipsComments(t *testing.T) {
	sql := "-- this is a comment\nCREATE TABLE foo (id INT);\n-- another comment\nCREATE TABLE bar (id INT);"
	stmts := splitStatements(sql)
	assert.Equal(t, 2, len(stmts))
	for _, s := range stmts {
		assert.NotContains(t, s, "-- this is a comment")
	}
}

func TestRunMigrations_TablesCreated(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	assert.Nil(t, err)
	defer engine.Close()

	err = RunMigrations(engine)
	assert.Nil(t, err)

	// Verify some tables were created
	tables, err := engine.DBMetas()
	assert.Nil(t, err)

	tableNames := make(map[string]bool)
	for _, table := range tables {
		tableNames[table.Name] = true
	}

	assert.True(t, tableNames["c_f_o"], "c_f_o table should exist")
	assert.True(t, tableNames["asset"], "asset table should exist")
	assert.True(t, tableNames["obligation"], "obligation table should exist")
	assert.True(t, tableNames["schema_migration"], "schema_migration table should exist")
}
