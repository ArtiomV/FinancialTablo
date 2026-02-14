package errs

import "net/http"

// Error codes related to locations
var (
	ErrLocationIdInvalid            = NewNormalError(NormalSubcategoryLocation, 0, http.StatusBadRequest, "location id is invalid")
	ErrLocationNotFound             = NewNormalError(NormalSubcategoryLocation, 1, http.StatusBadRequest, "location not found")
	ErrLocationNameIsEmpty          = NewNormalError(NormalSubcategoryLocation, 2, http.StatusBadRequest, "location name is empty")
	ErrLocationNameAlreadyExists    = NewNormalError(NormalSubcategoryLocation, 3, http.StatusBadRequest, "location name already exists")
	ErrLocationInUseCannotBeDeleted = NewNormalError(NormalSubcategoryLocation, 4, http.StatusBadRequest, "location is in use and cannot be deleted")
)
