package errs

import "net/http"

// Error codes related to counterparties
var (
	ErrCounterpartyIdInvalid            = NewNormalError(NormalSubcategoryCounterparty, 0, http.StatusBadRequest, "counterparty id is invalid")
	ErrCounterpartyNotFound             = NewNormalError(NormalSubcategoryCounterparty, 1, http.StatusBadRequest, "counterparty not found")
	ErrCounterpartyTypeInvalid          = NewNormalError(NormalSubcategoryCounterparty, 2, http.StatusBadRequest, "counterparty type is invalid")
	ErrCounterpartyNameIsEmpty          = NewNormalError(NormalSubcategoryCounterparty, 3, http.StatusBadRequest, "counterparty name is empty")
	ErrCounterpartyNameAlreadyExists    = NewNormalError(NormalSubcategoryCounterparty, 4, http.StatusBadRequest, "counterparty name already exists")
	ErrCounterpartyInUseCannotBeDeleted = NewNormalError(NormalSubcategoryCounterparty, 5, http.StatusBadRequest, "counterparty is in use and cannot be deleted")
)
