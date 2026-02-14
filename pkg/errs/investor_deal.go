package errs

import "net/http"

var (
	ErrInvestorDealIdInvalid            = NewNormalError(NormalSubcategoryInvestorDeal, 0, http.StatusBadRequest, "investor deal id is invalid")
	ErrInvestorDealNotFound             = NewNormalError(NormalSubcategoryInvestorDeal, 1, http.StatusNotFound, "investor deal not found")
	ErrInvestorDealInvestorNameIsEmpty  = NewNormalError(NormalSubcategoryInvestorDeal, 2, http.StatusBadRequest, "investor name is empty")
	ErrInvestorPaymentIdInvalid         = NewNormalError(NormalSubcategoryInvestorDeal, 3, http.StatusBadRequest, "investor payment id is invalid")
	ErrInvestorPaymentNotFound          = NewNormalError(NormalSubcategoryInvestorDeal, 4, http.StatusNotFound, "investor payment not found")
)
