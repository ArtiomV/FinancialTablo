package api

import (
	"sort"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// ObligationsApi represents obligations api
type ObligationsApi struct {
	obligations *services.ObligationService
}

// Initialize an obligations api singleton instance
var (
	ObligationsAPI = &ObligationsApi{
		obligations: services.Obligations,
	}
)

// ObligationListHandler returns obligation list of current user
func (a *ObligationsApi) ObligationListHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	obligations, err := a.obligations.GetAllObligationsByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[obligations.ObligationListHandler] failed to get obligations for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	obligationResps := make([]*models.ObligationInfoResponse, len(obligations))

	for i := 0; i < len(obligations); i++ {
		obligationResps[i] = obligations[i].ToObligationInfoResponse()
	}

	sort.Slice(obligationResps, func(i, j int) bool {
		return obligationResps[i].DueDate > obligationResps[j].DueDate
	})

	return obligationResps, nil
}

// ObligationCreateHandler saves a new obligation by request parameters for current user
func (a *ObligationsApi) ObligationCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var obligationCreateReq models.ObligationCreateRequest
	err := c.ShouldBindJSON(&obligationCreateReq)

	if err != nil {
		log.Warnf(c, "[obligations.ObligationCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	obligation := &models.Obligation{
		Uid:            uid,
		ObligationType: obligationCreateReq.ObligationType,
		CounterpartyId: obligationCreateReq.CounterpartyId,
		CfoId:          obligationCreateReq.CfoId,
		Amount:         obligationCreateReq.Amount,
		Currency:       obligationCreateReq.Currency,
		DueDate:        obligationCreateReq.DueDate,
		Status:         obligationCreateReq.Status,
		PaidAmount:     obligationCreateReq.PaidAmount,
		Comment:        obligationCreateReq.Comment,
	}

	err = a.obligations.CreateObligation(c, obligation)

	if err != nil {
		log.Errorf(c, "[obligations.ObligationCreateHandler] failed to create obligation \"id:%d\" for user \"uid:%d\", because %s", obligation.ObligationId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[obligations.ObligationCreateHandler] user \"uid:%d\" has created a new obligation \"id:%d\" successfully", uid, obligation.ObligationId)

	return obligation.ToObligationInfoResponse(), nil
}

// ObligationModifyHandler saves an existed obligation by request parameters for current user
func (a *ObligationsApi) ObligationModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var obligationModifyReq models.ObligationModifyRequest
	err := c.ShouldBindJSON(&obligationModifyReq)

	if err != nil {
		log.Warnf(c, "[obligations.ObligationModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	existingObligation, err := a.obligations.GetObligationByObligationId(c, uid, obligationModifyReq.Id)

	if err != nil {
		log.Errorf(c, "[obligations.ObligationModifyHandler] failed to get obligation \"id:%d\" for user \"uid:%d\", because %s", obligationModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	newObligation := &models.Obligation{
		ObligationId:   existingObligation.ObligationId,
		Uid:            uid,
		ObligationType: obligationModifyReq.ObligationType,
		CounterpartyId: obligationModifyReq.CounterpartyId,
		CfoId:          obligationModifyReq.CfoId,
		Amount:         obligationModifyReq.Amount,
		Currency:       obligationModifyReq.Currency,
		DueDate:        obligationModifyReq.DueDate,
		Status:         obligationModifyReq.Status,
		PaidAmount:     obligationModifyReq.PaidAmount,
		Comment:        obligationModifyReq.Comment,
	}

	if newObligation.ObligationType == existingObligation.ObligationType &&
		newObligation.CounterpartyId == existingObligation.CounterpartyId &&
		newObligation.CfoId == existingObligation.CfoId &&
		newObligation.Amount == existingObligation.Amount &&
		newObligation.Currency == existingObligation.Currency &&
		newObligation.DueDate == existingObligation.DueDate &&
		newObligation.Status == existingObligation.Status &&
		newObligation.PaidAmount == existingObligation.PaidAmount &&
		newObligation.Comment == existingObligation.Comment {
		return nil, errs.ErrNothingWillBeUpdated
	}

	err = a.obligations.ModifyObligation(c, newObligation)

	if err != nil {
		log.Errorf(c, "[obligations.ObligationModifyHandler] failed to update obligation \"id:%d\" for user \"uid:%d\", because %s", obligationModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[obligations.ObligationModifyHandler] user \"uid:%d\" has updated obligation \"id:%d\" successfully", uid, obligationModifyReq.Id)

	return newObligation.ToObligationInfoResponse(), nil
}

// ObligationDeleteHandler deletes an existed obligation by request parameters for current user
func (a *ObligationsApi) ObligationDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var obligationDeleteReq models.ObligationDeleteRequest
	err := c.ShouldBindJSON(&obligationDeleteReq)

	if err != nil {
		log.Warnf(c, "[obligations.ObligationDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.obligations.DeleteObligation(c, uid, obligationDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[obligations.ObligationDeleteHandler] failed to delete obligation \"id:%d\" for user \"uid:%d\", because %s", obligationDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[obligations.ObligationDeleteHandler] user \"uid:%d\" has deleted obligation \"id:%d\"", uid, obligationDeleteReq.Id)
	return true, nil
}
