package migrations

import (
	"embed"
	"strings"
	"time"

	"xorm.io/xorm"
)

//go:embed *.sql
var sqlFiles embed.FS

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	SQL         string
}

// AllMigrations returns all registered migrations
func AllMigrations() []*Migration {
	return []*Migration{
		{
			Version:     1,
			Description: "Create all new tables (CFO, Location, Asset, Obligation, TaxRecord, Budget, InvestorDeal, InvestorPayment, Counterparty, InsightsExplorer)",
			SQL:         mustReadSQL("round2_new_tables.sql"),
		},
	}
}

func mustReadSQL(name string) string {
	data, err := sqlFiles.ReadFile(name)
	if err != nil {
		panic("migration file not found: " + name)
	}
	return string(data)
}

// MigrationRecord tracks applied migrations in the database
type MigrationRecord struct {
	Version         int    `xorm:"PK"`
	Description     string `xorm:"VARCHAR(255)"`
	AppliedUnixTime int64
}

// TableName returns the table name for migration records
func (MigrationRecord) TableName() string {
	return "schema_migration"
}

// RunMigrations applies all pending migrations to the database
func RunMigrations(engine *xorm.Engine) error {
	// Ensure migration tracking table exists
	if err := engine.Sync2(new(MigrationRecord)); err != nil {
		return err
	}

	// Get already applied versions
	var applied []MigrationRecord
	if err := engine.Find(&applied); err != nil {
		return err
	}
	appliedMap := make(map[int]bool)
	for _, a := range applied {
		appliedMap[a.Version] = true
	}

	// Apply pending migrations
	for _, m := range AllMigrations() {
		if appliedMap[m.Version] {
			continue
		}

		// Execute SQL statements one by one
		statements := splitStatements(m.SQL)
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := engine.Exec(stmt); err != nil {
				return err
			}
		}

		// Record migration
		record := &MigrationRecord{
			Version:         m.Version,
			Description:     m.Description,
			AppliedUnixTime: time.Now().Unix(),
		}
		if _, err := engine.Insert(record); err != nil {
			return err
		}
	}

	return nil
}

// splitStatements splits SQL text by semicolons, skipping comment-only lines
func splitStatements(sql string) []string {
	var result []string
	var current strings.Builder

	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(line)

		// Skip comment-only lines
		if strings.HasPrefix(trimmed, "--") {
			continue
		}

		current.WriteString(line)
		current.WriteString("\n")

		if strings.HasSuffix(trimmed, ";") {
			result = append(result, current.String())
			current.Reset()
		}
	}

	// Last statement without trailing semicolon
	if remaining := strings.TrimSpace(current.String()); remaining != "" {
		result = append(result, remaining)
	}

	return result
}
