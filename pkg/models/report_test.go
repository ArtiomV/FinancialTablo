package models

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

// newBindingValidator creates a validator that reads "binding" tags (like gin does)
func newBindingValidator() *validator.Validate {
	v := validator.New()
	v.SetTagName("binding")
	return v
}

func TestReportRequest_EndTimeMustBeGreaterThanStartTime(t *testing.T) {
	validate := newBindingValidator()

	// EndTime > StartTime → valid
	req := ReportRequest{StartTime: 100, EndTime: 200}
	err := validate.Struct(req)
	assert.NoError(t, err)

	// EndTime == StartTime → invalid (gtfield means strictly greater)
	req2 := ReportRequest{StartTime: 100, EndTime: 100}
	err2 := validate.Struct(req2)
	assert.Error(t, err2)

	// EndTime < StartTime → invalid
	req3 := ReportRequest{StartTime: 200, EndTime: 100}
	err3 := validate.Struct(req3)
	assert.Error(t, err3)
}

func TestReportRequest_RequiredFields(t *testing.T) {
	validate := newBindingValidator()

	// Missing StartTime → invalid
	req := ReportRequest{StartTime: 0, EndTime: 200}
	err := validate.Struct(req)
	assert.Error(t, err)

	// Missing EndTime → invalid
	req2 := ReportRequest{StartTime: 100, EndTime: 0}
	err2 := validate.Struct(req2)
	assert.Error(t, err2)
}
