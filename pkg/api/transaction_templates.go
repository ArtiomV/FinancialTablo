package api

import (
	"sort"
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/duplicatechecker"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

const maximumTagsCountOfTemplate = 10

// TransactionTemplatesApi represents transaction template api
type TransactionTemplatesApi struct {
	ApiUsingConfig
	ApiUsingDuplicateChecker
	templates         *services.TransactionTemplateService
	transactions      *services.TransactionService
	transactionSplits *services.TransactionSplitService
	transactionTags   *services.TransactionTagService
}

// Initialize a transaction template api singleton instance
var (
	TransactionTemplates = &TransactionTemplatesApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		ApiUsingDuplicateChecker: ApiUsingDuplicateChecker{
			ApiUsingConfig: ApiUsingConfig{
				container: settings.Container,
			},
			container: duplicatechecker.Container,
		},
		templates:         services.TransactionTemplates,
		transactions:      services.Transactions,
		transactionSplits: services.TransactionSplits,
		transactionTags:   services.TransactionTags,
	}
)

// TemplateListHandler returns transaction template list of current user
func (a *TransactionTemplatesApi) TemplateListHandler(c *core.WebContext) (any, *errs.Error) {
	var templateListReq models.TransactionTemplateListRequest
	err := c.ShouldBindQuery(&templateListReq)

	if err != nil {
		log.Warnf(c, "[transaction_templates.TemplateListHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	if templateListReq.TemplateType < models.TRANSACTION_TEMPLATE_TYPE_NORMAL || templateListReq.TemplateType > models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE {
		log.Warnf(c, "[transaction_templates.TemplateListHandler] template type invalid, type is %d", templateListReq.TemplateType)
		return nil, errs.ErrTransactionTemplateTypeInvalid
	}

	if templateListReq.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE && !a.CurrentConfig().EnableScheduledTransaction {
		return nil, errs.ErrScheduledTransactionNotEnabled
	}

	uid := c.GetCurrentUid()
	templates, err := a.templates.GetAllTemplatesByUid(c, uid, templateListReq.TemplateType)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateListHandler] failed to get templates for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	templateResps := make(models.TransactionTemplateInfoResponseSlice, len(templates))
	serverUtcOffset := utils.GetServerTimezoneOffsetMinutes()

	for i := 0; i < len(templates); i++ {
		templateResps[i] = templates[i].ToTransactionTemplateInfoResponse(serverUtcOffset)
	}

	sort.Sort(templateResps)

	return templateResps, nil
}

// TemplateGetHandler returns one specific transaction template of current user
func (a *TransactionTemplatesApi) TemplateGetHandler(c *core.WebContext) (any, *errs.Error) {
	var templateGetReq models.TransactionTemplateGetRequest
	err := c.ShouldBindQuery(&templateGetReq)

	if err != nil {
		log.Warnf(c, "[transaction_templates.TemplateGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	template, err := a.templates.GetTemplateByTemplateId(c, uid, templateGetReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateGetHandler] failed to get template \"id:%d\" for user \"uid:%d\", because %s", templateGetReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE && !a.CurrentConfig().EnableScheduledTransaction {
		return nil, errs.ErrScheduledTransactionNotEnabled
	}

	serverUtcOffset := utils.GetServerTimezoneOffsetMinutes()
	templateResp := template.ToTransactionTemplateInfoResponse(serverUtcOffset)

	return templateResp, nil
}

// TemplateCreateHandler saves a new transaction template by request parameters for current user
func (a *TransactionTemplatesApi) TemplateCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var templateCreateReq models.TransactionTemplateCreateRequest
	err := c.ShouldBindJSON(&templateCreateReq)

	if err != nil {
		log.Warnf(c, "[transaction_templates.TemplateCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	if templateCreateReq.TemplateType < models.TRANSACTION_TEMPLATE_TYPE_NORMAL || templateCreateReq.TemplateType > models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE {
		log.Warnf(c, "[transaction_templates.TemplateCreateHandler] template type invalid, type is %d", templateCreateReq.TemplateType)
		return nil, errs.ErrTransactionTemplateTypeInvalid
	}

	if templateCreateReq.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE && !a.CurrentConfig().EnableScheduledTransaction {
		return nil, errs.ErrScheduledTransactionNotEnabled
	}

	if templateCreateReq.Type <= models.TRANSACTION_TYPE_MODIFY_BALANCE || templateCreateReq.Type > models.TRANSACTION_TYPE_TRANSFER {
		log.Warnf(c, "[transaction_templates.TemplateCreateHandler] transaction type invalid, type is %d", templateCreateReq.Type)
		return nil, errs.ErrTransactionTypeInvalid
	}

	if templateCreateReq.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE {
		if templateCreateReq.ScheduledFrequencyType == nil ||
			templateCreateReq.ScheduledFrequency == nil ||
			templateCreateReq.ScheduledTimezoneUtcOffset == nil {
			return nil, errs.ErrScheduledTransactionFrequencyInvalid
		}

		if *templateCreateReq.ScheduledFrequencyType == models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DISABLED && *templateCreateReq.ScheduledFrequency != "" {
			return nil, errs.ErrScheduledTransactionFrequencyInvalid
		} else if *templateCreateReq.ScheduledFrequencyType != models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DISABLED && *templateCreateReq.ScheduledFrequency == "" {
			return nil, errs.ErrScheduledTransactionFrequencyInvalid
		}
	}

	if len(templateCreateReq.TagIds) > maximumTagsCountOfTemplate {
		return nil, errs.ErrTransactionTemplateHasTooManyTags
	}

	uid := c.GetCurrentUid()

	maxOrderId, err := a.templates.GetMaxDisplayOrder(c, uid, templateCreateReq.TemplateType)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateCreateHandler] failed to get max display order for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	serverUtcOffset := utils.GetServerTimezoneOffsetMinutes()
	template, err := a.createNewTemplateModel(uid, &templateCreateReq, maxOrderId+1)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateCreateHandler] failed to create new template for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if a.CurrentConfig().EnableDuplicateSubmissionsCheck && templateCreateReq.ClientSessionId != "" {
		found, remark := a.GetSubmissionRemark(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_TEMPLATE, uid, templateCreateReq.ClientSessionId)

		if found {
			log.Infof(c, "[transaction_templates.TemplateCreateHandler] another template \"id:%s\" has been created for user \"uid:%d\"", remark, uid)
			templateId, err := utils.StringToInt64(remark)

			if err == nil {
				template, err = a.templates.GetTemplateByTemplateId(c, uid, templateId)

				if err != nil {
					log.Errorf(c, "[transaction_templates.TemplateCreateHandler] failed to get existed template \"id:%d\" for user \"uid:%d\", because %s", templateId, uid, err.Error())
					return nil, errs.Or(err, errs.ErrOperationFailed)
				}

				templateResp := template.ToTransactionTemplateInfoResponse(serverUtcOffset)

				return templateResp, nil
			}
		}
	}

	err = a.templates.CreateTemplate(c, template)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateCreateHandler] failed to create template \"id:%d\" for user \"uid:%d\", because %s", template.TemplateId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_templates.TemplateCreateHandler] user \"uid:%d\" has created a new template \"id:%d\" successfully", uid, template.TemplateId)

	a.SetSubmissionRemarkIfEnable(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_TEMPLATE, uid, templateCreateReq.ClientSessionId, utils.Int64ToString(template.TemplateId))
	templateResp := template.ToTransactionTemplateInfoResponse(serverUtcOffset)

	return templateResp, nil
}

// TemplateModifyHandler saves an existed transaction template by request parameters for current user
func (a *TransactionTemplatesApi) TemplateModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var templateModifyReq models.TransactionTemplateModifyRequest
	err := c.ShouldBindJSON(&templateModifyReq)

	if err != nil {
		log.Warnf(c, "[transaction_templates.TemplateModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	if templateModifyReq.Type <= models.TRANSACTION_TYPE_MODIFY_BALANCE || templateModifyReq.Type > models.TRANSACTION_TYPE_TRANSFER {
		log.Warnf(c, "[transaction_templates.TemplateModifyHandler] transaction type invalid, type is %d", templateModifyReq.Type)
		return nil, errs.ErrTransactionTypeInvalid
	}

	uid := c.GetCurrentUid()
	template, err := a.templates.GetTemplateByTemplateId(c, uid, templateModifyReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateModifyHandler] failed to get template \"id:%d\" for user \"uid:%d\", because %s", templateModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE && !a.CurrentConfig().EnableScheduledTransaction {
		return nil, errs.ErrScheduledTransactionNotEnabled
	}

	if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE {
		if templateModifyReq.ScheduledFrequencyType == nil ||
			templateModifyReq.ScheduledFrequency == nil ||
			templateModifyReq.ScheduledTimezoneUtcOffset == nil {
			return nil, errs.ErrScheduledTransactionFrequencyInvalid
		}

		if *templateModifyReq.ScheduledFrequencyType == models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DISABLED && *templateModifyReq.ScheduledFrequency != "" {
			return nil, errs.ErrScheduledTransactionFrequencyInvalid
		} else if *templateModifyReq.ScheduledFrequencyType != models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DISABLED && *templateModifyReq.ScheduledFrequency == "" {
			return nil, errs.ErrScheduledTransactionFrequencyInvalid
		}
	}

	if len(templateModifyReq.TagIds) > maximumTagsCountOfTemplate {
		return nil, errs.ErrTransactionTemplateHasTooManyTags
	}

	newTemplate := &models.TransactionTemplate{
		TemplateId:           template.TemplateId,
		Uid:                  uid,
		Name:                 templateModifyReq.Name,
		Type:                 templateModifyReq.Type,
		CategoryId:           templateModifyReq.CategoryId,
		AccountId:            templateModifyReq.SourceAccountId,
		TagIds:               strings.Join(templateModifyReq.TagIds, ","),
		Amount:               templateModifyReq.SourceAmount,
		RelatedAccountId:     templateModifyReq.DestinationAccountId,
		RelatedAccountAmount: templateModifyReq.DestinationAmount,
		HideAmount:           templateModifyReq.HideAmount,
		Comment:              templateModifyReq.Comment,
	}

	if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE {
		newTemplate.ScheduledFrequencyType = *templateModifyReq.ScheduledFrequencyType
		newTemplate.ScheduledFrequency = a.getOrderedFrequencyValues(*templateModifyReq.ScheduledFrequency)
		newTemplate.ScheduledAt = a.getUTCScheduledAt(*templateModifyReq.ScheduledTimezoneUtcOffset)
		newTemplate.ScheduledTimezoneUtcOffset = *templateModifyReq.ScheduledTimezoneUtcOffset

		if templateModifyReq.ScheduledStartDate != nil {
			startTime, err := utils.ParseFromLongDateFirstTime(*templateModifyReq.ScheduledStartDate, *templateModifyReq.ScheduledTimezoneUtcOffset)

			if err != nil {
				log.Errorf(c, "[transaction_templates.TemplateModifyHandler] failed to parse scheduled start date for user \"uid:%d\", because %s", uid, err.Error())
				return nil, errs.Or(err, errs.ErrOperationFailed)
			}

			startUnixTime := startTime.Unix()
			newTemplate.ScheduledStartTime = &startUnixTime
		}

		if templateModifyReq.ScheduledEndDate != nil {
			endTime, err := utils.ParseFromLongDateLastTime(*templateModifyReq.ScheduledEndDate, *templateModifyReq.ScheduledTimezoneUtcOffset)

			if err != nil {
				log.Errorf(c, "[transaction_templates.TemplateModifyHandler] failed to parse scheduled end date for user \"uid:%d\", because %s", uid, err.Error())
				return nil, errs.Or(err, errs.ErrOperationFailed)
			}

			endUnixTime := endTime.Unix()
			newTemplate.ScheduledEndTime = &endUnixTime
		}

		if newTemplate.ScheduledStartTime != nil && newTemplate.ScheduledEndTime != nil && *newTemplate.ScheduledStartTime > *newTemplate.ScheduledEndTime {
			return nil, errs.ErrScheduledTransactionTemplateStartDataLaterThanEndDate
		}
	}

	if newTemplate.Name == template.Name &&
		newTemplate.Type == template.Type &&
		newTemplate.CategoryId == template.CategoryId &&
		newTemplate.AccountId == template.AccountId &&
		newTemplate.TagIds == template.TagIds &&
		newTemplate.Amount == template.Amount &&
		newTemplate.RelatedAccountId == template.RelatedAccountId &&
		newTemplate.RelatedAccountAmount == template.RelatedAccountAmount &&
		newTemplate.HideAmount == template.HideAmount &&
		newTemplate.Comment == template.Comment {
		if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_NORMAL {
			return nil, errs.ErrNothingWillBeUpdated
		} else if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE {
			if newTemplate.ScheduledFrequencyType == template.ScheduledFrequencyType &&
				newTemplate.ScheduledFrequency == template.ScheduledFrequency &&
				int64PtrEqual(newTemplate.ScheduledStartTime, template.ScheduledStartTime) &&
				int64PtrEqual(newTemplate.ScheduledEndTime, template.ScheduledEndTime) &&
				newTemplate.ScheduledAt == template.ScheduledAt &&
				newTemplate.ScheduledTimezoneUtcOffset == template.ScheduledTimezoneUtcOffset {
				return nil, errs.ErrNothingWillBeUpdated
			}
		}
	}

	// Detect if frequency changed (for planned transaction regeneration)
	frequencyChanged := false
	if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE {
		frequencyChanged = newTemplate.ScheduledFrequencyType != template.ScheduledFrequencyType ||
			newTemplate.ScheduledFrequency != template.ScheduledFrequency
	}

	err = a.templates.ModifyTemplate(c, newTemplate)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateModifyHandler] failed to update template \"id:%d\" for user \"uid:%d\", because %s", templateModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_templates.TemplateModifyHandler] user \"uid:%d\" has updated template \"id:%d\" successfully", uid, templateModifyReq.Id)

	// If frequency changed, regenerate planned transactions
	if frequencyChanged {
		a.regeneratePlannedTransactions(c, uid, template.TemplateId, newTemplate)
	}

	serverUtcOffset := utils.GetServerTimezoneOffsetMinutes()
	newTemplate.TemplateType = template.TemplateType
	newTemplate.DisplayOrder = template.DisplayOrder
	newTemplate.Hidden = template.Hidden
	templateResp := newTemplate.ToTransactionTemplateInfoResponse(serverUtcOffset)

	return templateResp, nil
}

// TemplateRegeneratePlannedHandler regenerates planned transactions for a scheduled template
func (a *TransactionTemplatesApi) TemplateRegeneratePlannedHandler(c *core.WebContext) (any, *errs.Error) {
	type regenerateReq struct {
		Id int64 `json:"id,string" binding:"required,min=1"`
	}

	var req regenerateReq
	err := c.ShouldBindJSON(&req)

	if err != nil {
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	template, err := a.templates.GetTemplateByTemplateId(c, uid, req.Id)

	if err != nil {
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if template.TemplateType != models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE {
		return nil, errs.ErrTransactionTemplateNotFound
	}

	newTemplate := &models.TransactionTemplate{
		TemplateId:                 template.TemplateId,
		Uid:                        uid,
		Type:                       template.Type,
		CategoryId:                 template.CategoryId,
		AccountId:                  template.AccountId,
		Amount:                     template.Amount,
		RelatedAccountId:           template.RelatedAccountId,
		RelatedAccountAmount:       template.RelatedAccountAmount,
		HideAmount:                 template.HideAmount,
		Comment:                    template.Comment,
		TagIds:                     template.TagIds,
		ScheduledFrequencyType:     template.ScheduledFrequencyType,
		ScheduledFrequency:         template.ScheduledFrequency,
		ScheduledTimezoneUtcOffset: template.ScheduledTimezoneUtcOffset,
	}

	a.regeneratePlannedTransactions(c, uid, template.TemplateId, newTemplate)

	return true, nil
}

// TemplateHideHandler hides a transaction template by request parameters for current user
func (a *TransactionTemplatesApi) TemplateHideHandler(c *core.WebContext) (any, *errs.Error) {
	var templateHideReq models.TransactionTemplateHideRequest
	err := c.ShouldBindJSON(&templateHideReq)

	if err != nil {
		log.Warnf(c, "[transaction_templates.TemplateHideHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	template, err := a.templates.GetTemplateByTemplateId(c, uid, templateHideReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateHideHandler] failed to get template \"id:%d\" for user \"uid:%d\", because %s", templateHideReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE && !a.CurrentConfig().EnableScheduledTransaction {
		return nil, errs.ErrScheduledTransactionNotEnabled
	}

	err = a.templates.HideTemplate(c, uid, []int64{templateHideReq.Id}, templateHideReq.Hidden)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateHideHandler] failed to hide template \"id:%d\" for user \"uid:%d\", because %s", templateHideReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_templates.TemplateHideHandler] user \"uid:%d\" has hidden template \"id:%d\"", uid, templateHideReq.Id)
	return true, nil
}

// TemplateMoveHandler moves display order of existed transaction templates by request parameters for current user
func (a *TransactionTemplatesApi) TemplateMoveHandler(c *core.WebContext) (any, *errs.Error) {
	var templateMoveReq models.TransactionTemplateMoveRequest
	err := c.ShouldBindJSON(&templateMoveReq)

	if err != nil {
		log.Warnf(c, "[transaction_templates.TemplateMoveHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	if len(templateMoveReq.NewDisplayOrders) > 0 {
		template, err := a.templates.GetTemplateByTemplateId(c, uid, templateMoveReq.NewDisplayOrders[0].Id)

		if err != nil {
			log.Errorf(c, "[transaction_templates.TemplateMoveHandler] failed to get template \"id:%d\" for user \"uid:%d\", because %s", templateMoveReq.NewDisplayOrders[0].Id, uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE && !a.CurrentConfig().EnableScheduledTransaction {
			return nil, errs.ErrScheduledTransactionNotEnabled
		}
	}

	templates := make([]*models.TransactionTemplate, len(templateMoveReq.NewDisplayOrders))

	for i := 0; i < len(templateMoveReq.NewDisplayOrders); i++ {
		newDisplayOrder := templateMoveReq.NewDisplayOrders[i]
		template := &models.TransactionTemplate{
			Uid:          uid,
			TemplateId:   newDisplayOrder.Id,
			DisplayOrder: newDisplayOrder.DisplayOrder,
		}

		templates[i] = template
	}

	err = a.templates.ModifyTemplateDisplayOrders(c, uid, templates)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateMoveHandler] failed to move templates for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_templates.TemplateMoveHandler] user \"uid:%d\" has moved templates", uid)
	return true, nil
}

// TemplateDeleteHandler deletes an existed transaction template by request parameters for current user
func (a *TransactionTemplatesApi) TemplateDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var templateDeleteReq models.TransactionTemplateDeleteRequest
	err := c.ShouldBindJSON(&templateDeleteReq)

	if err != nil {
		log.Warnf(c, "[transaction_templates.TemplateDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	template, err := a.templates.GetTemplateByTemplateId(c, uid, templateDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateDeleteHandler] failed to get template \"id:%d\" for user \"uid:%d\", because %s", templateDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if template.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE && !a.CurrentConfig().EnableScheduledTransaction {
		return nil, errs.ErrScheduledTransactionNotEnabled
	}

	err = a.templates.DeleteTemplate(c, uid, templateDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[transaction_templates.TemplateDeleteHandler] failed to delete template \"id:%d\" for user \"uid:%d\", because %s", templateDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_templates.TemplateDeleteHandler] user \"uid:%d\" has deleted template \"id:%d\"", uid, templateDeleteReq.Id)
	return true, nil
}

func (a *TransactionTemplatesApi) createNewTemplateModel(uid int64, templateCreateReq *models.TransactionTemplateCreateRequest, order int32) (*models.TransactionTemplate, error) {
	template := &models.TransactionTemplate{
		Uid:                  uid,
		TemplateType:         templateCreateReq.TemplateType,
		Name:                 templateCreateReq.Name,
		Type:                 templateCreateReq.Type,
		CategoryId:           templateCreateReq.CategoryId,
		AccountId:            templateCreateReq.SourceAccountId,
		TagIds:               strings.Join(templateCreateReq.TagIds, ","),
		Amount:               templateCreateReq.SourceAmount,
		RelatedAccountId:     templateCreateReq.DestinationAccountId,
		RelatedAccountAmount: templateCreateReq.DestinationAmount,
		HideAmount:           templateCreateReq.HideAmount,
		Comment:              templateCreateReq.Comment,
		DisplayOrder:         order,
	}

	if templateCreateReq.TemplateType == models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE {
		template.ScheduledFrequencyType = *templateCreateReq.ScheduledFrequencyType
		template.ScheduledFrequency = a.getOrderedFrequencyValues(*templateCreateReq.ScheduledFrequency)
		template.ScheduledAt = a.getUTCScheduledAt(*templateCreateReq.ScheduledTimezoneUtcOffset)
		template.ScheduledTimezoneUtcOffset = *templateCreateReq.ScheduledTimezoneUtcOffset

		if templateCreateReq.ScheduledStartDate != nil {
			startTime, err := utils.ParseFromLongDateFirstTime(*templateCreateReq.ScheduledStartDate, *templateCreateReq.ScheduledTimezoneUtcOffset)

			if err != nil {
				return nil, err
			}

			startUnixTime := startTime.Unix()
			template.ScheduledStartTime = &startUnixTime
		}

		if templateCreateReq.ScheduledEndDate != nil {
			endTime, err := utils.ParseFromLongDateLastTime(*templateCreateReq.ScheduledEndDate, *templateCreateReq.ScheduledTimezoneUtcOffset)

			if err != nil {
				return nil, err
			}

			endUnixTime := endTime.Unix()
			template.ScheduledEndTime = &endUnixTime
		}

		if template.ScheduledStartTime != nil && template.ScheduledEndTime != nil && *template.ScheduledStartTime > *template.ScheduledEndTime {
			return nil, errs.ErrScheduledTransactionTemplateStartDataLaterThanEndDate
		}
	}

	return template, nil
}

// int64PtrEqual compares two *int64 pointers by value (not by address)
func int64PtrEqual(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func (a *TransactionTemplatesApi) getUTCScheduledAt(scheduledTimezoneUtcOffset int16) int16 {
	templateTimeZone := time.FixedZone("Template Timezone", int(scheduledTimezoneUtcOffset)*60)
	transactionTime := time.Date(2020, 1, 1, 0, 0, 0, 0, templateTimeZone)
	transactionTimeInUTC := transactionTime.In(time.UTC)

	minutesElapsedOfDayInUtc := transactionTimeInUTC.Hour()*60 + transactionTimeInUTC.Minute()

	return int16(minutesElapsedOfDayInUtc)
}

func (a *TransactionTemplatesApi) getOrderedFrequencyValues(frequencyValue string) string {
	if frequencyValue == "" {
		return ""
	}

	items := strings.Split(frequencyValue, ",")
	values := make([]int, 0, len(items))
	valueExistMap := make(map[int]bool)

	for i := 0; i < len(items); i++ {
		value, err := utils.StringToInt(items[i])

		if err != nil {
			continue
		}

		if _, exists := valueExistMap[value]; !exists {
			values = append(values, value)
			valueExistMap[value] = true
		}
	}

	sort.Ints(values)

	var sortedFrequencyValueBuilder strings.Builder

	for i := 0; i < len(values); i++ {
		if sortedFrequencyValueBuilder.Len() > 0 {
			sortedFrequencyValueBuilder.WriteRune(',')
		}

		sortedFrequencyValueBuilder.WriteString(utils.IntToString(values[i]))
	}

	return sortedFrequencyValueBuilder.String()
}

// regeneratePlannedTransactions deletes all future planned transactions for a template and generates new ones
func (a *TransactionTemplatesApi) regeneratePlannedTransactions(c *core.WebContext, uid int64, templateId int64, newTemplate *models.TransactionTemplate) {
	now := time.Now().Unix()
	nowTransactionTime := utils.GetMinTransactionTimeFromUnixTime(now)

	// Step 1: Load splits from an existing transaction BEFORE deleting old planned ones
	// Search through all transactions of this template (incl. soft-deleted) to find one with splits
	var splitReqs []models.TransactionSplitCreateRequest
	recentTransactions, rtErr := a.transactions.GetTransactionsByTemplateId(c, uid, templateId, 50)
	if rtErr == nil && len(recentTransactions) > 0 {
		for _, rt := range recentTransactions {
			txSplits, tsErr := a.transactionSplits.GetSplitsByTransactionId(c, uid, rt.TransactionId)
			if tsErr == nil && len(txSplits) > 0 {
				splitReqs = make([]models.TransactionSplitCreateRequest, len(txSplits))
				for i, sp := range txSplits {
					splitReqs[i] = models.TransactionSplitCreateRequest{
						CategoryId: sp.CategoryId,
						Amount:     sp.Amount,
						TagIds:     sp.GetTagIdStringSlice(),
					}
				}
				log.Infof(c, "[transaction_templates.regeneratePlannedTransactions] loaded %d splits from transaction \"id:%d\" for template \"id:%d\"", len(splitReqs), rt.TransactionId, templateId)
				break
			}
		}
	}

	// Step 2: Delete all future planned transactions with this template
	deletedCount, err := a.transactions.DeleteAllPlannedTransactionsByTemplate(c, uid, templateId, nowTransactionTime)
	if err != nil {
		log.Warnf(c, "[transaction_templates.regeneratePlannedTransactions] failed to delete old planned transactions for template \"id:%d\", because %s", templateId, err.Error())
		return
	}
	log.Infof(c, "[transaction_templates.regeneratePlannedTransactions] deleted %d old planned transactions for template \"id:%d\"", deletedCount, templateId)

	// Step 3: Create a base transaction from the template to use for generation
	var transactionDbType models.TransactionDbType
	switch newTemplate.Type {
	case models.TRANSACTION_TYPE_EXPENSE:
		transactionDbType = models.TRANSACTION_DB_TYPE_EXPENSE
	case models.TRANSACTION_TYPE_INCOME:
		transactionDbType = models.TRANSACTION_DB_TYPE_INCOME
	case models.TRANSACTION_TYPE_TRANSFER:
		transactionDbType = models.TRANSACTION_DB_TYPE_TRANSFER_OUT
	default:
		log.Warnf(c, "[transaction_templates.regeneratePlannedTransactions] invalid transaction type %d for template \"id:%d\"", newTemplate.Type, templateId)
		return
	}

	// Build base transaction: prefer data from the most recent transaction of this template
	// (which includes user edits like amount, comment, counterparty), fall back to template
	// for fields not stored in transactions (timezone, frequency, etc.)
	var baseAmount int64
	var baseComment string
	var baseCounterpartyId int64
	var baseCategoryId int64
	var baseAccountId int64
	var baseRelatedAccountId int64
	var baseRelatedAccountAmount int64
	var baseHideAmount bool
	var tagIds []int64

	if rtErr == nil && len(recentTransactions) > 0 {
		// Find the most recently UPDATED transaction (the one the user just edited)
		// recentTransactions is sorted by transaction_time DESC, but we want the most recently modified
		recent := recentTransactions[0]
		for _, rt := range recentTransactions[1:] {
			if rt.UpdatedUnixTime > recent.UpdatedUnixTime {
				recent = rt
			}
		}
		baseAmount = recent.Amount
		baseComment = recent.Comment
		baseCounterpartyId = recent.CounterpartyId
		baseCategoryId = recent.CategoryId
		baseAccountId = recent.AccountId
		baseRelatedAccountId = recent.RelatedAccountId
		baseRelatedAccountAmount = recent.RelatedAccountAmount
		baseHideAmount = recent.HideAmount
		// Get tags from recent transaction
		recentTags, tagErr := a.transactionTags.GetAllTagIdsOfTransactions(c, uid, []int64{recent.TransactionId})
		if tagErr == nil && len(recentTags) > 0 {
			if ids, ok := recentTags[recent.TransactionId]; ok {
				tagIds = ids
			}
		}
		if tagIds == nil {
			tagIds = newTemplate.GetTagIds()
		}
		log.Infof(c, "[transaction_templates.regeneratePlannedTransactions] using data from recent transaction \"id:%d\" (amount=%d, comment=%s)", recent.TransactionId, recent.Amount, recent.Comment)
	} else {
		baseAmount = newTemplate.Amount
		baseComment = newTemplate.Comment
		baseCounterpartyId = 0
		baseCategoryId = newTemplate.CategoryId
		baseAccountId = newTemplate.AccountId
		baseRelatedAccountId = newTemplate.RelatedAccountId
		baseRelatedAccountAmount = newTemplate.RelatedAccountAmount
		baseHideAmount = newTemplate.HideAmount
		tagIds = newTemplate.GetTagIds()
	}

	baseTransaction := &models.Transaction{
		Uid:                  uid,
		Type:                 transactionDbType,
		CategoryId:           baseCategoryId,
		TransactionTime:      nowTransactionTime,
		TimezoneUtcOffset:    newTemplate.ScheduledTimezoneUtcOffset,
		AccountId:            baseAccountId,
		Amount:               baseAmount,
		RelatedAccountId:     baseRelatedAccountId,
		RelatedAccountAmount: baseRelatedAccountAmount,
		HideAmount:           baseHideAmount,
		Comment:              baseComment,
		CounterpartyId:       baseCounterpartyId,
		CreatedIp:            "127.0.0.1",
	}

	// Step 4: Generate new planned transactions with splits
	count, err := a.transactions.GeneratePlannedTransactions(c, baseTransaction, tagIds, newTemplate.ScheduledFrequencyType, newTemplate.ScheduledFrequency, templateId, splitReqs)
	if err != nil {
		log.Warnf(c, "[transaction_templates.regeneratePlannedTransactions] failed to generate new planned transactions for template \"id:%d\", because %s", templateId, err.Error())
		return
	}
	log.Infof(c, "[transaction_templates.regeneratePlannedTransactions] generated %d new planned transactions for template \"id:%d\"", count, templateId)
}

// TemplateUpdateFrequencyHandler updates only the scheduled frequency of a template
func (a *TransactionTemplatesApi) TemplateUpdateFrequencyHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.TransactionTemplateUpdateFrequencyRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[transaction_templates.TemplateUpdateFrequencyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	template, getErr := a.templates.GetTemplateByTemplateId(c, uid, req.Id)

	if getErr != nil {
		log.Errorf(c, "[transaction_templates.TemplateUpdateFrequencyHandler] failed to get template \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, getErr.Error())
		return nil, errs.Or(getErr, errs.ErrOperationFailed)
	}

	if template.TemplateType != models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE {
		return nil, errs.ErrTransactionTemplateNotFound
	}

	if template.ScheduledFrequencyType == req.ScheduledFrequencyType &&
		template.ScheduledFrequency == req.ScheduledFrequency {
		return nil, errs.ErrNothingWillBeUpdated
	}

	template.ScheduledFrequencyType = req.ScheduledFrequencyType
	template.ScheduledFrequency = req.ScheduledFrequency
	template.UpdatedUnixTime = time.Now().Unix()

	updateErr := a.templates.ModifyTemplate(c, template)

	if updateErr != nil {
		log.Errorf(c, "[transaction_templates.TemplateUpdateFrequencyHandler] failed to update template \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, updateErr.Error())
		return nil, errs.Or(updateErr, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_templates.TemplateUpdateFrequencyHandler] user \"uid:%d\" has updated template \"id:%d\" frequency to type=%d freq=%s", uid, req.Id, req.ScheduledFrequencyType, req.ScheduledFrequency)

	// Regenerate planned transactions with the new frequency
	newTemplate := &models.TransactionTemplate{
		TemplateId:                 template.TemplateId,
		Uid:                        uid,
		Type:                       template.Type,
		CategoryId:                 template.CategoryId,
		AccountId:                  template.AccountId,
		Amount:                     template.Amount,
		RelatedAccountId:           template.RelatedAccountId,
		RelatedAccountAmount:       template.RelatedAccountAmount,
		HideAmount:                 template.HideAmount,
		Comment:                    template.Comment,
		TagIds:                     template.TagIds,
		ScheduledFrequencyType:     template.ScheduledFrequencyType,
		ScheduledFrequency:         template.ScheduledFrequency,
		ScheduledTimezoneUtcOffset: template.ScheduledTimezoneUtcOffset,
	}

	a.regeneratePlannedTransactions(c, uid, template.TemplateId, newTemplate)

	return true, nil
}
