package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// ===== Mock implementations for ReportService dependencies =====

type mockAssetProvider struct {
	assets []*models.Asset
	err    error
}

func (m *mockAssetProvider) GetAllAssetsByUid(_ core.Context, _ int64) ([]*models.Asset, error) {
	return m.assets, m.err
}
func (m *mockAssetProvider) GetAssetByAssetId(_ core.Context, _ int64, _ int64) (*models.Asset, error) {
	return nil, nil
}
func (m *mockAssetProvider) GetMaxDisplayOrder(_ core.Context, _ int64) (int32, error) {
	return 0, nil
}
func (m *mockAssetProvider) CreateAsset(_ core.Context, _ *models.Asset) error  { return nil }
func (m *mockAssetProvider) ModifyAsset(_ core.Context, _ *models.Asset, _ bool) error {
	return nil
}
func (m *mockAssetProvider) HideAsset(_ core.Context, _ int64, _ []int64, _ bool) error {
	return nil
}
func (m *mockAssetProvider) ModifyAssetDisplayOrders(_ core.Context, _ int64, _ []*models.Asset) error {
	return nil
}
func (m *mockAssetProvider) DeleteAsset(_ core.Context, _ int64, _ int64) error { return nil }
func (m *mockAssetProvider) ExistsAssetName(_ core.Context, _ int64, _ string) (bool, error) {
	return false, nil
}

type mockTaxRecordProvider struct {
	records []*models.TaxRecord
	err     error
}

func (m *mockTaxRecordProvider) GetAllTaxRecordsByUid(_ core.Context, _ int64) ([]*models.TaxRecord, error) {
	return m.records, m.err
}
func (m *mockTaxRecordProvider) GetTaxRecordByTaxId(_ core.Context, _ int64, _ int64) (*models.TaxRecord, error) {
	return nil, nil
}
func (m *mockTaxRecordProvider) CreateTaxRecord(_ core.Context, _ *models.TaxRecord) error {
	return nil
}
func (m *mockTaxRecordProvider) ModifyTaxRecord(_ core.Context, _ *models.TaxRecord) error {
	return nil
}
func (m *mockTaxRecordProvider) DeleteTaxRecord(_ core.Context, _ int64, _ int64) error { return nil }

type mockInvestorDealProvider struct {
	deals []*models.InvestorDeal
	err   error
}

func (m *mockInvestorDealProvider) GetAllDealsByUid(_ core.Context, _ int64) ([]*models.InvestorDeal, error) {
	return m.deals, m.err
}
func (m *mockInvestorDealProvider) GetDealByDealId(_ core.Context, _ int64, _ int64) (*models.InvestorDeal, error) {
	return nil, nil
}
func (m *mockInvestorDealProvider) CreateDeal(_ core.Context, _ *models.InvestorDeal) error {
	return nil
}
func (m *mockInvestorDealProvider) ModifyDeal(_ core.Context, _ *models.InvestorDeal) error {
	return nil
}
func (m *mockInvestorDealProvider) DeleteDeal(_ core.Context, _ int64, _ int64) error { return nil }

type mockInvestorPaymentProvider struct {
	paymentsByDeal map[int64][]*models.InvestorPayment
	err            error
}

func (m *mockInvestorPaymentProvider) GetAllPaymentsByDealId(_ core.Context, _ int64, dealId int64) ([]*models.InvestorPayment, error) {
	return m.paymentsByDeal[dealId], m.err
}
func (m *mockInvestorPaymentProvider) GetAllPaymentsByDealIds(_ core.Context, _ int64, dealIds []int64) (map[int64][]*models.InvestorPayment, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[int64][]*models.InvestorPayment)
	for _, id := range dealIds {
		if payments, ok := m.paymentsByDeal[id]; ok {
			result[id] = payments
		}
	}
	return result, nil
}
func (m *mockInvestorPaymentProvider) GetAllPaymentsByUid(_ core.Context, _ int64) ([]*models.InvestorPayment, error) {
	return nil, nil
}
func (m *mockInvestorPaymentProvider) GetPaymentByPaymentId(_ core.Context, _ int64, _ int64) (*models.InvestorPayment, error) {
	return nil, nil
}
func (m *mockInvestorPaymentProvider) CreatePayment(_ core.Context, _ *models.InvestorPayment) error {
	return nil
}
func (m *mockInvestorPaymentProvider) ModifyPayment(_ core.Context, _ *models.InvestorPayment) error {
	return nil
}
func (m *mockInvestorPaymentProvider) DeletePayment(_ core.Context, _ int64, _ int64) error {
	return nil
}

// newTestReportService creates a ReportService with mock dependencies (no DB needed)
func newTestReportService(
	assets AssetProvider,
	taxes TaxRecordProvider,
	deals InvestorDealProvider,
	payments InvestorPaymentProvider,
) *ReportService {
	return &ReportService{
		assets:   assets,
		taxes:    taxes,
		deals:    deals,
		payments: payments,
	}
}

// ===== Integration tests using mocks =====
// These tests validate the report methods' business logic end-to-end
// using mock providers, without requiring a database.
//
// NOTE: GetCashFlow and GetPnL also run SQL queries via ServiceUsingDB,
// so full integration tests for those require a real database.
// The tests below cover the non-SQL logic paths.

func TestReportService_ValidateTimeRange_Integration(t *testing.T) {
	svc := newTestReportService(
		&mockAssetProvider{},
		&mockTaxRecordProvider{},
		&mockInvestorDealProvider{},
		&mockInvestorPaymentProvider{},
	)

	// Invalid: start >= end
	_, err := svc.GetCashFlow(nil, 1, 0, 1000, 1000)
	assert.NotNil(t, err)

	_, err = svc.GetPnL(nil, 1, 0, 2000, 1000)
	assert.NotNil(t, err)

	_, err = svc.GetPaymentCalendar(nil, 1, 5000, 3000)
	assert.NotNil(t, err)
}

func TestReportService_ValidateUid_Integration(t *testing.T) {
	svc := newTestReportService(
		&mockAssetProvider{},
		&mockTaxRecordProvider{},
		&mockInvestorDealProvider{},
		&mockInvestorPaymentProvider{},
	)

	_, err := svc.GetCashFlow(nil, 0, 0, 1000, 2000)
	assert.NotNil(t, err)

	_, err = svc.GetPnL(nil, -1, 0, 1000, 2000)
	assert.NotNil(t, err)

	_, err = svc.GetBalance(nil, 0, 0)
	assert.NotNil(t, err)

	_, err = svc.GetPaymentCalendar(nil, -5, 1000, 2000)
	assert.NotNil(t, err)
}

// TODO: Full integration tests with SQLite in-memory
// To test GetCashFlow, GetPnL, GetBalance, and GetPaymentCalendar end-to-end
// with real SQL queries, set up an in-memory SQLite database:
//
//   engine, _ := xorm.NewEngine("sqlite3", ":memory:")
//   engine.Sync2(new(models.Transaction), new(models.TransactionCategory), ...)
//   db := &datastore.Database{...}
//   svc := &ReportService{ServiceUsingDB: ServiceUsingDB{container: ...}, ...}
//
// Then insert test data and verify the full report output.
