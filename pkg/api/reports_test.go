package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReportsApi_NotNil(t *testing.T) {
	api := NewReportsApi(nil)
	assert.NotNil(t, api)
	assert.Nil(t, api.reports)
}

func TestReportsAPI_Singleton(t *testing.T) {
	assert.NotNil(t, ReportsAPI)
	assert.NotNil(t, ReportsAPI.reports)
}
