package errs

import "net/http"

var (
	ErrObligationIdInvalid  = NewNormalError(NormalSubcategoryObligation, 0, http.StatusBadRequest, "obligation id is invalid")
	ErrObligationNotFound   = NewNormalError(NormalSubcategoryObligation, 1, http.StatusNotFound, "obligation not found")
	ErrTaxRecordIdInvalid   = NewNormalError(NormalSubcategoryTaxRecord, 0, http.StatusBadRequest, "tax record id is invalid")
	ErrTaxRecordNotFound    = NewNormalError(NormalSubcategoryTaxRecord, 1, http.StatusNotFound, "tax record not found")
)
