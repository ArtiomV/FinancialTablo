package errs

import "net/http"

// Error codes related to transaction categories
var (
	ErrTransactionCategoryIdInvalid            = NewNormalError(NormalSubcategoryCategory, 0, http.StatusBadRequest, "transaction category id is invalid")
	ErrTransactionCategoryNotFound             = NewNormalError(NormalSubcategoryCategory, 1, http.StatusBadRequest, "transaction category not found")
	ErrTransactionCategoryTypeInvalid          = NewNormalError(NormalSubcategoryCategory, 2, http.StatusBadRequest, "transaction category type is invalid")
	ErrTransactionCategoryInUseCannotBeDeleted = NewNormalError(NormalSubcategoryCategory, 6, http.StatusBadRequest, "transaction category is in use and cannot be deleted")
)
