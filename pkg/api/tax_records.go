package api

import (
	"sort"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// TaxRecordsApi represents tax records api
type TaxRecordsApi struct {
	taxRecords *services.TaxRecordService
}

// Initialize a tax records api singleton instance
var (
	TaxRecordsAPI = &TaxRecordsApi{
		taxRecords: services.TaxRecords,
	}
)

// TaxRecordListHandler returns tax record list of current user
func (a *TaxRecordsApi) TaxRecordListHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	records, err := a.taxRecords.GetAllTaxRecordsByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[tax_records.TaxRecordListHandler] failed to get tax records for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	recordResps := make([]*models.TaxRecordInfoResponse, len(records))

	for i := 0; i < len(records); i++ {
		recordResps[i] = records[i].ToTaxRecordInfoResponse()
	}

	sort.Slice(recordResps, func(i, j int) bool {
		if recordResps[i].PeriodYear != recordResps[j].PeriodYear {
			return recordResps[i].PeriodYear > recordResps[j].PeriodYear
		}
		return recordResps[i].PeriodQuarter > recordResps[j].PeriodQuarter
	})

	return recordResps, nil
}

// TaxRecordCreateHandler saves a new tax record by request parameters for current user
func (a *TaxRecordsApi) TaxRecordCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var taxRecordCreateReq models.TaxRecordCreateRequest
	err := c.ShouldBindJSON(&taxRecordCreateReq)

	if err != nil {
		log.Warnf(c, "[tax_records.TaxRecordCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	record := &models.TaxRecord{
		Uid:           uid,
		CfoId:         taxRecordCreateReq.CfoId,
		TaxType:       taxRecordCreateReq.TaxType,
		PeriodYear:    taxRecordCreateReq.PeriodYear,
		PeriodQuarter: taxRecordCreateReq.PeriodQuarter,
		TaxableIncome: taxRecordCreateReq.TaxableIncome,
		TaxAmount:     taxRecordCreateReq.TaxAmount,
		PaidAmount:    taxRecordCreateReq.PaidAmount,
		DueDate:       taxRecordCreateReq.DueDate,
		Status:        taxRecordCreateReq.Status,
		Comment:       taxRecordCreateReq.Comment,
	}

	err = a.taxRecords.CreateTaxRecord(c, record)

	if err != nil {
		log.Errorf(c, "[tax_records.TaxRecordCreateHandler] failed to create tax record \"id:%d\" for user \"uid:%d\", because %s", record.TaxId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[tax_records.TaxRecordCreateHandler] user \"uid:%d\" has created a new tax record \"id:%d\" successfully", uid, record.TaxId)

	return record.ToTaxRecordInfoResponse(), nil
}

// TaxRecordModifyHandler saves an existed tax record by request parameters for current user
func (a *TaxRecordsApi) TaxRecordModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var taxRecordModifyReq models.TaxRecordModifyRequest
	err := c.ShouldBindJSON(&taxRecordModifyReq)

	if err != nil {
		log.Warnf(c, "[tax_records.TaxRecordModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	existingRecord, err := a.taxRecords.GetTaxRecordByTaxId(c, uid, taxRecordModifyReq.Id)

	if err != nil {
		log.Errorf(c, "[tax_records.TaxRecordModifyHandler] failed to get tax record \"id:%d\" for user \"uid:%d\", because %s", taxRecordModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	newRecord := &models.TaxRecord{
		TaxId:         existingRecord.TaxId,
		Uid:           uid,
		CfoId:         taxRecordModifyReq.CfoId,
		TaxType:       taxRecordModifyReq.TaxType,
		PeriodYear:    taxRecordModifyReq.PeriodYear,
		PeriodQuarter: taxRecordModifyReq.PeriodQuarter,
		TaxableIncome: taxRecordModifyReq.TaxableIncome,
		TaxAmount:     taxRecordModifyReq.TaxAmount,
		PaidAmount:    taxRecordModifyReq.PaidAmount,
		DueDate:       taxRecordModifyReq.DueDate,
		Status:        taxRecordModifyReq.Status,
		Comment:       taxRecordModifyReq.Comment,
	}

	if newRecord.CfoId == existingRecord.CfoId &&
		newRecord.TaxType == existingRecord.TaxType &&
		newRecord.PeriodYear == existingRecord.PeriodYear &&
		newRecord.PeriodQuarter == existingRecord.PeriodQuarter &&
		newRecord.TaxableIncome == existingRecord.TaxableIncome &&
		newRecord.TaxAmount == existingRecord.TaxAmount &&
		newRecord.PaidAmount == existingRecord.PaidAmount &&
		newRecord.DueDate == existingRecord.DueDate &&
		newRecord.Status == existingRecord.Status &&
		newRecord.Comment == existingRecord.Comment {
		return nil, errs.ErrNothingWillBeUpdated
	}

	err = a.taxRecords.ModifyTaxRecord(c, newRecord)

	if err != nil {
		log.Errorf(c, "[tax_records.TaxRecordModifyHandler] failed to update tax record \"id:%d\" for user \"uid:%d\", because %s", taxRecordModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[tax_records.TaxRecordModifyHandler] user \"uid:%d\" has updated tax record \"id:%d\" successfully", uid, taxRecordModifyReq.Id)

	return newRecord.ToTaxRecordInfoResponse(), nil
}

// TaxRecordDeleteHandler deletes an existed tax record by request parameters for current user
func (a *TaxRecordsApi) TaxRecordDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var taxRecordDeleteReq models.TaxRecordDeleteRequest
	err := c.ShouldBindJSON(&taxRecordDeleteReq)

	if err != nil {
		log.Warnf(c, "[tax_records.TaxRecordDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.taxRecords.DeleteTaxRecord(c, uid, taxRecordDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[tax_records.TaxRecordDeleteHandler] failed to delete tax record \"id:%d\" for user \"uid:%d\", because %s", taxRecordDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[tax_records.TaxRecordDeleteHandler] user \"uid:%d\" has deleted tax record \"id:%d\"", uid, taxRecordDeleteReq.Id)
	return true, nil
}
