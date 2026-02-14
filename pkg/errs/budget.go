package errs

import "net/http"

var (
	ErrBudgetIdInvalid  = NewNormalError(NormalSubcategoryBudget, 0, http.StatusBadRequest, "budget id is invalid")
	ErrBudgetNotFound   = NewNormalError(NormalSubcategoryBudget, 1, http.StatusNotFound, "budget not found")
)
