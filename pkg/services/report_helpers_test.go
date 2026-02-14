package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// === calculateResidualValue edge cases ===

func TestCalculateResidualValue_SalvageExceedsCost(t *testing.T) {
	// Edge case: salvage > cost (misconfigured asset)
	asset := &models.Asset{
		PurchaseCost:     50000,
		SalvageValue:     100000,
		UsefulLifeMonths: 12,
		CommissionDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	asOf := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	result := calculateResidualValue(asset, asOf)
	// monthly depr = (50000 - 100000) / 12 = negative → accumulated negative → residual > cost
	// Function should handle this gracefully
	assert.True(t, result >= asset.SalvageValue)
}

func TestCalculateResidualValue_OneMonthUsefulLife_EdgeCase(t *testing.T) {
	asset := &models.Asset{
		PurchaseCost:     120000,
		SalvageValue:     0,
		UsefulLifeMonths: 1,
		CommissionDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	asOf := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, int64(0), calculateResidualValue(asset, asOf))
}

func TestCalculateResidualValue_ExactEndOfLife(t *testing.T) {
	asset := &models.Asset{
		PurchaseCost:     120000,
		SalvageValue:     20000,
		UsefulLifeMonths: 10,
		CommissionDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
	}
	// Exactly 10 months after commission
	asOf := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, int64(20000), calculateResidualValue(asset, asOf))
}

// === validateTimeRange ===

func TestValidateTimeRange_Valid_Helper(t *testing.T) {
	assert.Nil(t, validateTimeRange(1000, 2000))
}

func TestValidateTimeRange_Equal_Helper(t *testing.T) {
	assert.NotNil(t, validateTimeRange(1000, 1000))
}

func TestValidateTimeRange_Reversed_Helper(t *testing.T) {
	assert.NotNil(t, validateTimeRange(2000, 1000))
}

func TestValidateTimeRange_ExceedsMax_Helper(t *testing.T) {
	// 10 years + 1 second
	tenYears := int64(10 * 365 * 24 * 60 * 60)
	assert.NotNil(t, validateTimeRange(1000, 1000+tenYears+1))
}

func TestValidateTimeRange_ExactlyMax_Helper(t *testing.T) {
	tenYears := int64(10 * 365 * 24 * 60 * 60)
	assert.Nil(t, validateTimeRange(1000, 1000+tenYears))
}

// === matchesCfo ===

func TestMatchesCfo_NegativeFilter_Helper(t *testing.T) {
	assert.True(t, matchesCfo(-1, 5))
}

func TestMatchesCfo_ZeroEntityCfo(t *testing.T) {
	// Filter is set, entity has no CFO assigned
	assert.False(t, matchesCfo(5, 0))
}

// === monthsBetween additional edge cases ===

func TestMonthsBetween_LeapYearBoundary(t *testing.T) {
	from := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 3, 29, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, int64(1), monthsBetween(from, to))
}

func TestMonthsBetween_EndOfYear(t *testing.T) {
	from := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, int64(1), monthsBetween(from, to)) // Dec → Jan = 1 month
}
