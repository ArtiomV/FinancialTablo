package api

import (
	"sort"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// CounterpartiesApi represents counterparty api
type CounterpartiesApi struct {
	counterparties *services.CounterpartyService
}

// Initialize a counterparty api singleton instance
var (
	Counterparties = &CounterpartiesApi{
		counterparties: services.Counterparties,
	}
)

// CounterpartyListHandler returns counterparty list of current user
func (a *CounterpartiesApi) CounterpartyListHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	counterparties, err := a.counterparties.GetAllCounterpartiesByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[counterparties.CounterpartyListHandler] failed to get counterparties for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	counterpartyResps := make(models.CounterpartyInfoResponseSlice, len(counterparties))

	for i := 0; i < len(counterparties); i++ {
		counterpartyResps[i] = counterparties[i].ToCounterpartyInfoResponse()
	}

	sort.Sort(counterpartyResps)

	return counterpartyResps, nil
}

// CounterpartyGetHandler returns one specific counterparty of current user
func (a *CounterpartiesApi) CounterpartyGetHandler(c *core.WebContext) (any, *errs.Error) {
	var counterpartyGetReq models.CounterpartyGetRequest
	err := c.ShouldBindQuery(&counterpartyGetReq)

	if err != nil {
		log.Warnf(c, "[counterparties.CounterpartyGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	counterparty, err := a.counterparties.GetCounterpartyByCounterpartyId(c, uid, counterpartyGetReq.Id)

	if err != nil {
		log.Errorf(c, "[counterparties.CounterpartyGetHandler] failed to get counterparty \"id:%d\" for user \"uid:%d\", because %s", counterpartyGetReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	counterpartyResp := counterparty.ToCounterpartyInfoResponse()

	return counterpartyResp, nil
}

// CounterpartyCreateHandler saves a new counterparty by request parameters for current user
func (a *CounterpartiesApi) CounterpartyCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var counterpartyCreateReq models.CounterpartyCreateRequest
	err := c.ShouldBindJSON(&counterpartyCreateReq)

	if err != nil {
		log.Warnf(c, "[counterparties.CounterpartyCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	if counterpartyCreateReq.Type < models.COUNTERPARTY_TYPE_PERSON || counterpartyCreateReq.Type > models.COUNTERPARTY_TYPE_COMPANY {
		log.Warnf(c, "[counterparties.CounterpartyCreateHandler] counterparty type invalid, type is %d", counterpartyCreateReq.Type)
		return nil, errs.ErrCounterpartyTypeInvalid
	}

	uid := c.GetCurrentUid()

	maxOrderId, err := a.counterparties.GetMaxDisplayOrder(c, uid)

	if err != nil {
		log.Errorf(c, "[counterparties.CounterpartyCreateHandler] failed to get max display order for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	counterparty := a.createNewCounterpartyModel(uid, &counterpartyCreateReq, maxOrderId+1)

	err = a.counterparties.CreateCounterparty(c, counterparty)

	if err != nil {
		log.Errorf(c, "[counterparties.CounterpartyCreateHandler] failed to create counterparty \"id:%d\" for user \"uid:%d\", because %s", counterparty.CounterpartyId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[counterparties.CounterpartyCreateHandler] user \"uid:%d\" has created a new counterparty \"id:%d\" successfully", uid, counterparty.CounterpartyId)

	counterpartyResp := counterparty.ToCounterpartyInfoResponse()

	return counterpartyResp, nil
}

// CounterpartyModifyHandler saves an existed counterparty by request parameters for current user
func (a *CounterpartiesApi) CounterpartyModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var counterpartyModifyReq models.CounterpartyModifyRequest
	err := c.ShouldBindJSON(&counterpartyModifyReq)

	if err != nil {
		log.Warnf(c, "[counterparties.CounterpartyModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	if counterpartyModifyReq.Type < models.COUNTERPARTY_TYPE_PERSON || counterpartyModifyReq.Type > models.COUNTERPARTY_TYPE_COMPANY {
		log.Warnf(c, "[counterparties.CounterpartyModifyHandler] counterparty type invalid, type is %d", counterpartyModifyReq.Type)
		return nil, errs.ErrCounterpartyTypeInvalid
	}

	uid := c.GetCurrentUid()
	counterparty, err := a.counterparties.GetCounterpartyByCounterpartyId(c, uid, counterpartyModifyReq.Id)

	if err != nil {
		log.Errorf(c, "[counterparties.CounterpartyModifyHandler] failed to get counterparty \"id:%d\" for user \"uid:%d\", because %s", counterpartyModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	newCounterparty := &models.Counterparty{
		CounterpartyId: counterparty.CounterpartyId,
		Uid:            uid,
		Name:           counterpartyModifyReq.Name,
		Type:           counterpartyModifyReq.Type,
		Icon:           counterpartyModifyReq.Icon,
		Color:          counterpartyModifyReq.Color,
		Comment:        counterpartyModifyReq.Comment,
		Hidden:         counterpartyModifyReq.Hidden,
		DisplayOrder:   counterparty.DisplayOrder,
	}

	counterpartyNameChanged := newCounterparty.Name != counterparty.Name

	if !counterpartyNameChanged &&
		newCounterparty.Type == counterparty.Type &&
		newCounterparty.Icon == counterparty.Icon &&
		newCounterparty.Color == counterparty.Color &&
		newCounterparty.Comment == counterparty.Comment &&
		newCounterparty.Hidden == counterparty.Hidden {
		return nil, errs.ErrNothingWillBeUpdated
	}

	err = a.counterparties.ModifyCounterparty(c, newCounterparty, counterpartyNameChanged)

	if err != nil {
		log.Errorf(c, "[counterparties.CounterpartyModifyHandler] failed to update counterparty \"id:%d\" for user \"uid:%d\", because %s", counterpartyModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[counterparties.CounterpartyModifyHandler] user \"uid:%d\" has updated counterparty \"id:%d\" successfully", uid, counterpartyModifyReq.Id)

	counterpartyResp := newCounterparty.ToCounterpartyInfoResponse()

	return counterpartyResp, nil
}

// CounterpartyHideHandler hides a counterparty by request parameters for current user
func (a *CounterpartiesApi) CounterpartyHideHandler(c *core.WebContext) (any, *errs.Error) {
	var counterpartyHideReq models.CounterpartyHideRequest
	err := c.ShouldBindJSON(&counterpartyHideReq)

	if err != nil {
		log.Warnf(c, "[counterparties.CounterpartyHideHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.counterparties.HideCounterparty(c, uid, []int64{counterpartyHideReq.Id}, counterpartyHideReq.Hidden)

	if err != nil {
		log.Errorf(c, "[counterparties.CounterpartyHideHandler] failed to hide counterparty \"id:%d\" for user \"uid:%d\", because %s", counterpartyHideReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[counterparties.CounterpartyHideHandler] user \"uid:%d\" has hidden counterparty \"id:%d\"", uid, counterpartyHideReq.Id)
	return true, nil
}

// CounterpartyMoveHandler moves display order of existed counterparties by request parameters for current user
func (a *CounterpartiesApi) CounterpartyMoveHandler(c *core.WebContext) (any, *errs.Error) {
	var counterpartyMoveReq models.CounterpartyMoveRequest
	err := c.ShouldBindJSON(&counterpartyMoveReq)

	if err != nil {
		log.Warnf(c, "[counterparties.CounterpartyMoveHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	counterparties := make([]*models.Counterparty, len(counterpartyMoveReq.NewDisplayOrders))

	for i := 0; i < len(counterpartyMoveReq.NewDisplayOrders); i++ {
		newDisplayOrder := counterpartyMoveReq.NewDisplayOrders[i]
		counterparty := &models.Counterparty{
			Uid:            uid,
			CounterpartyId: newDisplayOrder.Id,
			DisplayOrder:   newDisplayOrder.DisplayOrder,
		}

		counterparties[i] = counterparty
	}

	err = a.counterparties.ModifyCounterpartyDisplayOrders(c, uid, counterparties)

	if err != nil {
		log.Errorf(c, "[counterparties.CounterpartyMoveHandler] failed to move counterparties for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[counterparties.CounterpartyMoveHandler] user \"uid:%d\" has moved counterparties", uid)
	return true, nil
}

// CounterpartyDeleteHandler deletes an existed counterparty by request parameters for current user
func (a *CounterpartiesApi) CounterpartyDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var counterpartyDeleteReq models.CounterpartyDeleteRequest
	err := c.ShouldBindJSON(&counterpartyDeleteReq)

	if err != nil {
		log.Warnf(c, "[counterparties.CounterpartyDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.counterparties.DeleteCounterparty(c, uid, counterpartyDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[counterparties.CounterpartyDeleteHandler] failed to delete counterparty \"id:%d\" for user \"uid:%d\", because %s", counterpartyDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[counterparties.CounterpartyDeleteHandler] user \"uid:%d\" has deleted counterparty \"id:%d\"", uid, counterpartyDeleteReq.Id)
	return true, nil
}

func (a *CounterpartiesApi) createNewCounterpartyModel(uid int64, counterpartyCreateReq *models.CounterpartyCreateRequest, order int32) *models.Counterparty {
	return &models.Counterparty{
		Uid:          uid,
		Name:         counterpartyCreateReq.Name,
		Type:         counterpartyCreateReq.Type,
		Icon:         counterpartyCreateReq.Icon,
		Color:        counterpartyCreateReq.Color,
		Comment:      counterpartyCreateReq.Comment,
		DisplayOrder: order,
	}
}
