package api

import (
	"sort"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// LocationsApi represents location api
type LocationsApi struct {
	locations *services.LocationService
}

// Initialize a location api singleton instance
var (
	Locations = &LocationsApi{
		locations: services.Locations,
	}
)

// LocationListHandler returns location list of current user
func (a *LocationsApi) LocationListHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	locations, err := a.locations.GetAllLocationsByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[locations.LocationListHandler] failed to get locations for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	locationResps := make(models.LocationInfoResponseSlice, len(locations))

	for i := 0; i < len(locations); i++ {
		locationResps[i] = locations[i].ToLocationInfoResponse()
	}

	sort.Sort(locationResps)

	return locationResps, nil
}

// LocationGetHandler returns one specific location of current user
func (a *LocationsApi) LocationGetHandler(c *core.WebContext) (any, *errs.Error) {
	var locationGetReq models.LocationGetRequest
	err := c.ShouldBindQuery(&locationGetReq)

	if err != nil {
		log.Warnf(c, "[locations.LocationGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	location, err := a.locations.GetLocationByLocationId(c, uid, locationGetReq.Id)

	if err != nil {
		log.Errorf(c, "[locations.LocationGetHandler] failed to get location \"id:%d\" for user \"uid:%d\", because %s", locationGetReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return location.ToLocationInfoResponse(), nil
}

// LocationCreateHandler saves a new location by request parameters for current user
func (a *LocationsApi) LocationCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var locationCreateReq models.LocationCreateRequest
	err := c.ShouldBindJSON(&locationCreateReq)

	if err != nil {
		log.Warnf(c, "[locations.LocationCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	maxOrderId, err := a.locations.GetMaxDisplayOrder(c, uid)

	if err != nil {
		log.Errorf(c, "[locations.LocationCreateHandler] failed to get max display order for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	location := &models.Location{
		Uid:                uid,
		Name:               locationCreateReq.Name,
		CfoId:              locationCreateReq.CfoId,
		Address:            locationCreateReq.Address,
		LocationType:       locationCreateReq.LocationType,
		MonthlyRent:        locationCreateReq.MonthlyRent,
		MonthlyElectricity: locationCreateReq.MonthlyElectricity,
		MonthlyInternet:    locationCreateReq.MonthlyInternet,
		Comment:            locationCreateReq.Comment,
		DisplayOrder:       maxOrderId + 1,
	}

	err = a.locations.CreateLocation(c, location)

	if err != nil {
		log.Errorf(c, "[locations.LocationCreateHandler] failed to create location \"id:%d\" for user \"uid:%d\", because %s", location.LocationId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[locations.LocationCreateHandler] user \"uid:%d\" has created a new location \"id:%d\" successfully", uid, location.LocationId)

	return location.ToLocationInfoResponse(), nil
}

// LocationModifyHandler saves an existed location by request parameters for current user
func (a *LocationsApi) LocationModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var locationModifyReq models.LocationModifyRequest
	err := c.ShouldBindJSON(&locationModifyReq)

	if err != nil {
		log.Warnf(c, "[locations.LocationModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	location, err := a.locations.GetLocationByLocationId(c, uid, locationModifyReq.Id)

	if err != nil {
		log.Errorf(c, "[locations.LocationModifyHandler] failed to get location \"id:%d\" for user \"uid:%d\", because %s", locationModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	newLocation := &models.Location{
		LocationId:         location.LocationId,
		Uid:                uid,
		Name:               locationModifyReq.Name,
		CfoId:              locationModifyReq.CfoId,
		Address:            locationModifyReq.Address,
		LocationType:       locationModifyReq.LocationType,
		MonthlyRent:        locationModifyReq.MonthlyRent,
		MonthlyElectricity: locationModifyReq.MonthlyElectricity,
		MonthlyInternet:    locationModifyReq.MonthlyInternet,
		Comment:            locationModifyReq.Comment,
		Hidden:             locationModifyReq.Hidden,
		DisplayOrder:       location.DisplayOrder,
	}

	nameChanged := newLocation.Name != location.Name

	if !nameChanged &&
		newLocation.CfoId == location.CfoId &&
		newLocation.Address == location.Address &&
		newLocation.LocationType == location.LocationType &&
		newLocation.MonthlyRent == location.MonthlyRent &&
		newLocation.MonthlyElectricity == location.MonthlyElectricity &&
		newLocation.MonthlyInternet == location.MonthlyInternet &&
		newLocation.Comment == location.Comment &&
		newLocation.Hidden == location.Hidden {
		return nil, errs.ErrNothingWillBeUpdated
	}

	err = a.locations.ModifyLocation(c, newLocation, nameChanged)

	if err != nil {
		log.Errorf(c, "[locations.LocationModifyHandler] failed to update location \"id:%d\" for user \"uid:%d\", because %s", locationModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[locations.LocationModifyHandler] user \"uid:%d\" has updated location \"id:%d\" successfully", uid, locationModifyReq.Id)

	return newLocation.ToLocationInfoResponse(), nil
}

// LocationHideHandler hides a location by request parameters for current user
func (a *LocationsApi) LocationHideHandler(c *core.WebContext) (any, *errs.Error) {
	var locationHideReq models.LocationHideRequest
	err := c.ShouldBindJSON(&locationHideReq)

	if err != nil {
		log.Warnf(c, "[locations.LocationHideHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.locations.HideLocation(c, uid, []int64{locationHideReq.Id}, locationHideReq.Hidden)

	if err != nil {
		log.Errorf(c, "[locations.LocationHideHandler] failed to hide location \"id:%d\" for user \"uid:%d\", because %s", locationHideReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[locations.LocationHideHandler] user \"uid:%d\" has hidden location \"id:%d\"", uid, locationHideReq.Id)
	return true, nil
}

// LocationMoveHandler moves display order of existed locations by request parameters for current user
func (a *LocationsApi) LocationMoveHandler(c *core.WebContext) (any, *errs.Error) {
	var locationMoveReq models.LocationMoveRequest
	err := c.ShouldBindJSON(&locationMoveReq)

	if err != nil {
		log.Warnf(c, "[locations.LocationMoveHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	locations := make([]*models.Location, len(locationMoveReq.NewDisplayOrders))

	for i := 0; i < len(locationMoveReq.NewDisplayOrders); i++ {
		newDisplayOrder := locationMoveReq.NewDisplayOrders[i]
		location := &models.Location{
			Uid:          uid,
			LocationId:   newDisplayOrder.Id,
			DisplayOrder: newDisplayOrder.DisplayOrder,
		}

		locations[i] = location
	}

	err = a.locations.ModifyLocationDisplayOrders(c, uid, locations)

	if err != nil {
		log.Errorf(c, "[locations.LocationMoveHandler] failed to move locations for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[locations.LocationMoveHandler] user \"uid:%d\" has moved locations", uid)
	return true, nil
}

// LocationDeleteHandler deletes an existed location by request parameters for current user
func (a *LocationsApi) LocationDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var locationDeleteReq models.LocationDeleteRequest
	err := c.ShouldBindJSON(&locationDeleteReq)

	if err != nil {
		log.Warnf(c, "[locations.LocationDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.locations.DeleteLocation(c, uid, locationDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[locations.LocationDeleteHandler] failed to delete location \"id:%d\" for user \"uid:%d\", because %s", locationDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[locations.LocationDeleteHandler] user \"uid:%d\" has deleted location \"id:%d\"", uid, locationDeleteReq.Id)
	return true, nil
}
