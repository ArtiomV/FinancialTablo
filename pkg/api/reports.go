package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// ReportsApi represents reports api
type ReportsApi struct {
	reports *services.ReportService
}

// Initialize a reports api singleton instance
var (
	ReportsAPI = &ReportsApi{
		reports: services.Reports,
	}
)

// CashFlowHandler returns cash flow report
func (a *ReportsApi) CashFlowHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.ReportRequest
	err := c.ShouldBindQuery(&req)

	if err != nil {
		log.Warnf(c, "[reports.CashFlowHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	result, err := a.reports.GetCashFlow(c, uid, req.CfoId, req.StartTime, req.EndTime)

	if err != nil {
		log.Errorf(c, "[reports.CashFlowHandler] failed to get cash flow for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return result, nil
}

// PnLHandler returns profit and loss report
func (a *ReportsApi) PnLHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.ReportRequest
	err := c.ShouldBindQuery(&req)

	if err != nil {
		log.Warnf(c, "[reports.PnLHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	result, err := a.reports.GetPnL(c, uid, req.CfoId, req.StartTime, req.EndTime)

	if err != nil {
		log.Errorf(c, "[reports.PnLHandler] failed to get P&L for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return result, nil
}

// BalanceHandler returns balance sheet report
func (a *ReportsApi) BalanceHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.ReportRequest
	err := c.ShouldBindQuery(&req)

	if err != nil {
		log.Warnf(c, "[reports.BalanceHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	result, err := a.reports.GetBalance(c, uid, req.CfoId)

	if err != nil {
		log.Errorf(c, "[reports.BalanceHandler] failed to get balance for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return result, nil
}

// PaymentCalendarHandler returns payment calendar
func (a *ReportsApi) PaymentCalendarHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.ReportRequest
	err := c.ShouldBindQuery(&req)

	if err != nil {
		log.Warnf(c, "[reports.PaymentCalendarHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	result, err := a.reports.GetPaymentCalendar(c, uid, req.StartTime, req.EndTime)

	if err != nil {
		log.Errorf(c, "[reports.PaymentCalendarHandler] failed to get payment calendar for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return result, nil
}
