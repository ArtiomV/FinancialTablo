package errs

import "net/http"

var (
	ErrAssetIdInvalid          = NewNormalError(NormalSubcategoryAsset, 0, http.StatusBadRequest, "asset id is invalid")
	ErrAssetNotFound           = NewNormalError(NormalSubcategoryAsset, 1, http.StatusNotFound, "asset not found")
	ErrAssetNameIsEmpty        = NewNormalError(NormalSubcategoryAsset, 2, http.StatusBadRequest, "asset name is empty")
	ErrAssetNameAlreadyExists  = NewNormalError(NormalSubcategoryAsset, 3, http.StatusConflict, "asset name already exists")
	ErrAssetInUseCannotBeDeleted = NewNormalError(NormalSubcategoryAsset, 4, http.StatusConflict, "asset is in use and cannot be deleted")
)
