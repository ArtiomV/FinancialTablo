package errs

import "net/http"

// Error codes related to CFOs
var (
	ErrCFOIdInvalid            = NewNormalError(NormalSubcategoryCFO, 0, http.StatusBadRequest, "cfo id is invalid")
	ErrCFONotFound             = NewNormalError(NormalSubcategoryCFO, 1, http.StatusBadRequest, "cfo not found")
	ErrCFONameIsEmpty          = NewNormalError(NormalSubcategoryCFO, 2, http.StatusBadRequest, "cfo name is empty")
	ErrCFONameAlreadyExists    = NewNormalError(NormalSubcategoryCFO, 3, http.StatusBadRequest, "cfo name already exists")
	ErrCFOInUseCannotBeDeleted = NewNormalError(NormalSubcategoryCFO, 4, http.StatusBadRequest, "cfo is in use and cannot be deleted")
)
