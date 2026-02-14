package api

import (
	"sort"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// CFOsApi represents CFO api
type CFOsApi struct {
	cfos services.CFOProvider
}

// NewCFOsApi creates a new CFOsApi instance
func NewCFOsApi(c services.CFOProvider) *CFOsApi {
	return &CFOsApi{cfos: c}
}

// Initialize a CFO api singleton instance
var (
	CFOs = NewCFOsApi(services.CFOs)
)

// CFOListHandler returns CFO list of current user
func (a *CFOsApi) CFOListHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	cfos, err := a.cfos.GetAllCFOsByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[cfos.CFOListHandler] failed to get cfos for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	cfoResps := make(models.CFOInfoResponseSlice, len(cfos))

	for i := 0; i < len(cfos); i++ {
		cfoResps[i] = cfos[i].ToCFOInfoResponse()
	}

	sort.Sort(cfoResps)

	return cfoResps, nil
}

// CFOGetHandler returns one specific CFO of current user
func (a *CFOsApi) CFOGetHandler(c *core.WebContext) (any, *errs.Error) {
	var cfoGetReq models.CFOGetRequest
	err := c.ShouldBindQuery(&cfoGetReq)

	if err != nil {
		log.Warnf(c, "[cfos.CFOGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	cfo, err := a.cfos.GetCFOByCFOId(c, uid, cfoGetReq.Id)

	if err != nil {
		log.Errorf(c, "[cfos.CFOGetHandler] failed to get cfo \"id:%d\" for user \"uid:%d\", because %s", cfoGetReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	cfoResp := cfo.ToCFOInfoResponse()

	return cfoResp, nil
}

// CFOCreateHandler saves a new CFO by request parameters for current user
func (a *CFOsApi) CFOCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var cfoCreateReq models.CFOCreateRequest
	err := c.ShouldBindJSON(&cfoCreateReq)

	if err != nil {
		log.Warnf(c, "[cfos.CFOCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	maxOrderId, err := a.cfos.GetMaxDisplayOrder(c, uid)

	if err != nil {
		log.Errorf(c, "[cfos.CFOCreateHandler] failed to get max display order for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	cfo := &models.CFO{
		Uid:          uid,
		Name:         cfoCreateReq.Name,
		Color:        cfoCreateReq.Color,
		Comment:      cfoCreateReq.Comment,
		DisplayOrder: maxOrderId + 1,
	}

	err = a.cfos.CreateCFO(c, cfo)

	if err != nil {
		log.Errorf(c, "[cfos.CFOCreateHandler] failed to create cfo \"id:%d\" for user \"uid:%d\", because %s", cfo.CfoId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[cfos.CFOCreateHandler] user \"uid:%d\" has created a new cfo \"id:%d\" successfully", uid, cfo.CfoId)

	cfoResp := cfo.ToCFOInfoResponse()

	return cfoResp, nil
}

// CFOModifyHandler saves an existed CFO by request parameters for current user
func (a *CFOsApi) CFOModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var cfoModifyReq models.CFOModifyRequest
	err := c.ShouldBindJSON(&cfoModifyReq)

	if err != nil {
		log.Warnf(c, "[cfos.CFOModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	cfo, err := a.cfos.GetCFOByCFOId(c, uid, cfoModifyReq.Id)

	if err != nil {
		log.Errorf(c, "[cfos.CFOModifyHandler] failed to get cfo \"id:%d\" for user \"uid:%d\", because %s", cfoModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	newCFO := &models.CFO{
		CfoId:        cfo.CfoId,
		Uid:          uid,
		Name:         cfoModifyReq.Name,
		Color:        cfoModifyReq.Color,
		Comment:      cfoModifyReq.Comment,
		Hidden:       cfoModifyReq.Hidden,
		DisplayOrder: cfo.DisplayOrder,
	}

	nameChanged := newCFO.Name != cfo.Name

	if !nameChanged &&
		newCFO.Color == cfo.Color &&
		newCFO.Comment == cfo.Comment &&
		newCFO.Hidden == cfo.Hidden {
		return nil, errs.ErrNothingWillBeUpdated
	}

	err = a.cfos.ModifyCFO(c, newCFO, nameChanged)

	if err != nil {
		log.Errorf(c, "[cfos.CFOModifyHandler] failed to update cfo \"id:%d\" for user \"uid:%d\", because %s", cfoModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[cfos.CFOModifyHandler] user \"uid:%d\" has updated cfo \"id:%d\" successfully", uid, cfoModifyReq.Id)

	cfoResp := newCFO.ToCFOInfoResponse()

	return cfoResp, nil
}

// CFOHideHandler hides a CFO by request parameters for current user
func (a *CFOsApi) CFOHideHandler(c *core.WebContext) (any, *errs.Error) {
	var cfoHideReq models.CFOHideRequest
	err := c.ShouldBindJSON(&cfoHideReq)

	if err != nil {
		log.Warnf(c, "[cfos.CFOHideHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.cfos.HideCFO(c, uid, []int64{cfoHideReq.Id}, cfoHideReq.Hidden)

	if err != nil {
		log.Errorf(c, "[cfos.CFOHideHandler] failed to hide cfo \"id:%d\" for user \"uid:%d\", because %s", cfoHideReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[cfos.CFOHideHandler] user \"uid:%d\" has hidden cfo \"id:%d\"", uid, cfoHideReq.Id)
	return true, nil
}

// CFOMoveHandler moves display order of existed CFOs by request parameters for current user
func (a *CFOsApi) CFOMoveHandler(c *core.WebContext) (any, *errs.Error) {
	var cfoMoveReq models.CFOMoveRequest
	err := c.ShouldBindJSON(&cfoMoveReq)

	if err != nil {
		log.Warnf(c, "[cfos.CFOMoveHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	cfos := make([]*models.CFO, len(cfoMoveReq.NewDisplayOrders))

	for i := 0; i < len(cfoMoveReq.NewDisplayOrders); i++ {
		newDisplayOrder := cfoMoveReq.NewDisplayOrders[i]
		cfo := &models.CFO{
			Uid:          uid,
			CfoId:        newDisplayOrder.Id,
			DisplayOrder: newDisplayOrder.DisplayOrder,
		}

		cfos[i] = cfo
	}

	err = a.cfos.ModifyCFODisplayOrders(c, uid, cfos)

	if err != nil {
		log.Errorf(c, "[cfos.CFOMoveHandler] failed to move cfos for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[cfos.CFOMoveHandler] user \"uid:%d\" has moved cfos", uid)
	return true, nil
}

// CFODeleteHandler deletes an existed CFO by request parameters for current user
func (a *CFOsApi) CFODeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var cfoDeleteReq models.CFODeleteRequest
	err := c.ShouldBindJSON(&cfoDeleteReq)

	if err != nil {
		log.Warnf(c, "[cfos.CFODeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.cfos.DeleteCFO(c, uid, cfoDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[cfos.CFODeleteHandler] failed to delete cfo \"id:%d\" for user \"uid:%d\", because %s", cfoDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[cfos.CFODeleteHandler] user \"uid:%d\" has deleted cfo \"id:%d\"", uid, cfoDeleteReq.Id)
	return true, nil
}
