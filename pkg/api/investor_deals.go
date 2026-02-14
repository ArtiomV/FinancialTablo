package api

import (
	"sort"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// InvestorDealsApi represents investor deals api
type InvestorDealsApi struct {
	deals    *services.InvestorDealService
	payments *services.InvestorPaymentService
}

// Initialize an investor deals api singleton instance
var (
	InvestorDeals = &InvestorDealsApi{
		deals:    services.InvestorDeals,
		payments: services.InvestorPayments,
	}
)

// DealListHandler returns investor deal list of current user
func (a *InvestorDealsApi) DealListHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	deals, err := a.deals.GetAllDealsByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[investor_deals.DealListHandler] failed to get deals for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	dealResps := make(models.InvestorDealInfoResponseSlice, len(deals))

	for i := 0; i < len(deals); i++ {
		dealResps[i] = deals[i].ToInvestorDealInfoResponse()
	}

	sort.Sort(dealResps)

	return dealResps, nil
}

// DealGetHandler returns one specific investor deal of current user
func (a *InvestorDealsApi) DealGetHandler(c *core.WebContext) (any, *errs.Error) {
	var dealGetReq models.InvestorDealGetRequest
	err := c.ShouldBindQuery(&dealGetReq)

	if err != nil {
		log.Warnf(c, "[investor_deals.DealGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	deal, err := a.deals.GetDealByDealId(c, uid, dealGetReq.Id)

	if err != nil {
		log.Errorf(c, "[investor_deals.DealGetHandler] failed to get deal \"id:%d\" for user \"uid:%d\", because %s", dealGetReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return deal.ToInvestorDealInfoResponse(), nil
}

// DealCreateHandler saves a new investor deal by request parameters for current user
func (a *InvestorDealsApi) DealCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var dealCreateReq models.InvestorDealCreateRequest
	err := c.ShouldBindJSON(&dealCreateReq)

	if err != nil {
		log.Warnf(c, "[investor_deals.DealCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	deal := &models.InvestorDeal{
		Uid:                uid,
		InvestorName:       dealCreateReq.InvestorName,
		CfoId:              dealCreateReq.CfoId,
		InvestmentDate:     dealCreateReq.InvestmentDate,
		InvestmentAmount:   dealCreateReq.InvestmentAmount,
		Currency:           dealCreateReq.Currency,
		DealType:           dealCreateReq.DealType,
		AnnualRate:         dealCreateReq.AnnualRate,
		ProfitSharePct:     dealCreateReq.ProfitSharePct,
		FixedPayment:       dealCreateReq.FixedPayment,
		RepaymentStartDate: dealCreateReq.RepaymentStartDate,
		RepaymentEndDate:   dealCreateReq.RepaymentEndDate,
		TotalToRepay:       dealCreateReq.TotalToRepay,
		Comment:            dealCreateReq.Comment,
	}

	err = a.deals.CreateDeal(c, deal)

	if err != nil {
		log.Errorf(c, "[investor_deals.DealCreateHandler] failed to create deal \"id:%d\" for user \"uid:%d\", because %s", deal.DealId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[investor_deals.DealCreateHandler] user \"uid:%d\" has created a new deal \"id:%d\" successfully", uid, deal.DealId)

	return deal.ToInvestorDealInfoResponse(), nil
}

// DealModifyHandler saves an existed investor deal by request parameters for current user
func (a *InvestorDealsApi) DealModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var dealModifyReq models.InvestorDealModifyRequest
	err := c.ShouldBindJSON(&dealModifyReq)

	if err != nil {
		log.Warnf(c, "[investor_deals.DealModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	existingDeal, err := a.deals.GetDealByDealId(c, uid, dealModifyReq.Id)

	if err != nil {
		log.Errorf(c, "[investor_deals.DealModifyHandler] failed to get deal \"id:%d\" for user \"uid:%d\", because %s", dealModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	newDeal := &models.InvestorDeal{
		DealId:             existingDeal.DealId,
		Uid:                uid,
		InvestorName:       dealModifyReq.InvestorName,
		CfoId:              dealModifyReq.CfoId,
		InvestmentDate:     dealModifyReq.InvestmentDate,
		InvestmentAmount:   dealModifyReq.InvestmentAmount,
		Currency:           dealModifyReq.Currency,
		DealType:           dealModifyReq.DealType,
		AnnualRate:         dealModifyReq.AnnualRate,
		ProfitSharePct:     dealModifyReq.ProfitSharePct,
		FixedPayment:       dealModifyReq.FixedPayment,
		RepaymentStartDate: dealModifyReq.RepaymentStartDate,
		RepaymentEndDate:   dealModifyReq.RepaymentEndDate,
		TotalToRepay:       dealModifyReq.TotalToRepay,
		Comment:            dealModifyReq.Comment,
	}

	if newDeal.InvestorName == existingDeal.InvestorName &&
		newDeal.CfoId == existingDeal.CfoId &&
		newDeal.InvestmentDate == existingDeal.InvestmentDate &&
		newDeal.InvestmentAmount == existingDeal.InvestmentAmount &&
		newDeal.Currency == existingDeal.Currency &&
		newDeal.DealType == existingDeal.DealType &&
		newDeal.AnnualRate == existingDeal.AnnualRate &&
		newDeal.ProfitSharePct == existingDeal.ProfitSharePct &&
		newDeal.FixedPayment == existingDeal.FixedPayment &&
		newDeal.RepaymentStartDate == existingDeal.RepaymentStartDate &&
		newDeal.RepaymentEndDate == existingDeal.RepaymentEndDate &&
		newDeal.TotalToRepay == existingDeal.TotalToRepay &&
		newDeal.Comment == existingDeal.Comment {
		return nil, errs.ErrNothingWillBeUpdated
	}

	err = a.deals.ModifyDeal(c, newDeal)

	if err != nil {
		log.Errorf(c, "[investor_deals.DealModifyHandler] failed to update deal \"id:%d\" for user \"uid:%d\", because %s", dealModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[investor_deals.DealModifyHandler] user \"uid:%d\" has updated deal \"id:%d\" successfully", uid, dealModifyReq.Id)

	return newDeal.ToInvestorDealInfoResponse(), nil
}

// DealDeleteHandler deletes an existed investor deal by request parameters for current user
func (a *InvestorDealsApi) DealDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var dealDeleteReq models.InvestorDealDeleteRequest
	err := c.ShouldBindJSON(&dealDeleteReq)

	if err != nil {
		log.Warnf(c, "[investor_deals.DealDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.deals.DeleteDeal(c, uid, dealDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[investor_deals.DealDeleteHandler] failed to delete deal \"id:%d\" for user \"uid:%d\", because %s", dealDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[investor_deals.DealDeleteHandler] user \"uid:%d\" has deleted deal \"id:%d\"", uid, dealDeleteReq.Id)
	return true, nil
}

// PaymentListHandler returns investor payment list for a deal
func (a *InvestorDealsApi) PaymentListHandler(c *core.WebContext) (any, *errs.Error) {
	var paymentListReq models.InvestorPaymentListByDealRequest
	err := c.ShouldBindQuery(&paymentListReq)

	if err != nil {
		log.Warnf(c, "[investor_deals.PaymentListHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	payments, err := a.payments.GetAllPaymentsByDealId(c, uid, paymentListReq.DealId)

	if err != nil {
		log.Errorf(c, "[investor_deals.PaymentListHandler] failed to get payments for deal \"id:%d\" user \"uid:%d\", because %s", paymentListReq.DealId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	paymentResps := make(models.InvestorPaymentInfoResponseSlice, len(payments))

	for i := 0; i < len(payments); i++ {
		paymentResps[i] = payments[i].ToInvestorPaymentInfoResponse()
	}

	sort.Sort(paymentResps)

	return paymentResps, nil
}

// PaymentCreateHandler saves a new investor payment by request parameters for current user
func (a *InvestorDealsApi) PaymentCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var paymentCreateReq models.InvestorPaymentCreateRequest
	err := c.ShouldBindJSON(&paymentCreateReq)

	if err != nil {
		log.Warnf(c, "[investor_deals.PaymentCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	// Verify the deal exists
	_, err = a.deals.GetDealByDealId(c, uid, paymentCreateReq.DealId)

	if err != nil {
		log.Errorf(c, "[investor_deals.PaymentCreateHandler] failed to get deal \"id:%d\" for user \"uid:%d\", because %s", paymentCreateReq.DealId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	payment := &models.InvestorPayment{
		Uid:           uid,
		DealId:        paymentCreateReq.DealId,
		PaymentDate:   paymentCreateReq.PaymentDate,
		Amount:        paymentCreateReq.Amount,
		PaymentType:   paymentCreateReq.PaymentType,
		TransactionId: paymentCreateReq.TransactionId,
		Comment:       paymentCreateReq.Comment,
	}

	err = a.payments.CreatePayment(c, payment)

	if err != nil {
		log.Errorf(c, "[investor_deals.PaymentCreateHandler] failed to create payment \"id:%d\" for user \"uid:%d\", because %s", payment.PaymentId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[investor_deals.PaymentCreateHandler] user \"uid:%d\" has created a new payment \"id:%d\" successfully", uid, payment.PaymentId)

	return payment.ToInvestorPaymentInfoResponse(), nil
}

// PaymentModifyHandler saves an existed investor payment by request parameters for current user
func (a *InvestorDealsApi) PaymentModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var paymentModifyReq models.InvestorPaymentModifyRequest
	err := c.ShouldBindJSON(&paymentModifyReq)

	if err != nil {
		log.Warnf(c, "[investor_deals.PaymentModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	existingPayment, err := a.payments.GetPaymentByPaymentId(c, uid, paymentModifyReq.Id)

	if err != nil {
		log.Errorf(c, "[investor_deals.PaymentModifyHandler] failed to get payment \"id:%d\" for user \"uid:%d\", because %s", paymentModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	newPayment := &models.InvestorPayment{
		PaymentId:     existingPayment.PaymentId,
		Uid:           uid,
		DealId:        paymentModifyReq.DealId,
		PaymentDate:   paymentModifyReq.PaymentDate,
		Amount:        paymentModifyReq.Amount,
		PaymentType:   paymentModifyReq.PaymentType,
		TransactionId: paymentModifyReq.TransactionId,
		Comment:       paymentModifyReq.Comment,
	}

	if newPayment.DealId == existingPayment.DealId &&
		newPayment.PaymentDate == existingPayment.PaymentDate &&
		newPayment.Amount == existingPayment.Amount &&
		newPayment.PaymentType == existingPayment.PaymentType &&
		newPayment.TransactionId == existingPayment.TransactionId &&
		newPayment.Comment == existingPayment.Comment {
		return nil, errs.ErrNothingWillBeUpdated
	}

	err = a.payments.ModifyPayment(c, newPayment)

	if err != nil {
		log.Errorf(c, "[investor_deals.PaymentModifyHandler] failed to update payment \"id:%d\" for user \"uid:%d\", because %s", paymentModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[investor_deals.PaymentModifyHandler] user \"uid:%d\" has updated payment \"id:%d\" successfully", uid, paymentModifyReq.Id)

	return newPayment.ToInvestorPaymentInfoResponse(), nil
}

// PaymentDeleteHandler deletes an existed investor payment by request parameters for current user
func (a *InvestorDealsApi) PaymentDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var paymentDeleteReq models.InvestorPaymentDeleteRequest
	err := c.ShouldBindJSON(&paymentDeleteReq)

	if err != nil {
		log.Warnf(c, "[investor_deals.PaymentDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.payments.DeletePayment(c, uid, paymentDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[investor_deals.PaymentDeleteHandler] failed to delete payment \"id:%d\" for user \"uid:%d\", because %s", paymentDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[investor_deals.PaymentDeleteHandler] user \"uid:%d\" has deleted payment \"id:%d\"", uid, paymentDeleteReq.Id)
	return true, nil
}
