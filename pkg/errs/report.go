package errs

import "net/http"

// Error codes related to reports
var (
	ErrReportStartTimeAfterEndTime = NewNormalError(NormalSubcategoryReport, 0, http.StatusBadRequest, "start time must be before end time")
	ErrReportTimeRangeTooLong      = NewNormalError(NormalSubcategoryReport, 1, http.StatusBadRequest, "time range exceeds maximum allowed period")
)
