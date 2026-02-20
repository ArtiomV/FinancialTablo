package api

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/duplicatechecker"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// TransactionCreateHandler saves a new transaction by request parameters for current user
func (a *TransactionsApi) TransactionCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionCreateReq models.TransactionCreateRequest
	err := c.ShouldBindJSON(&transactionCreateReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionCreateHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	tagIds, err := utils.StringArrayToInt64Array(transactionCreateReq.TagIds)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionCreateHandler] parse tag ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionTagIdInvalid
	}

	if len(tagIds) > models.MaximumTagsCountOfTransaction {
		return nil, errs.ErrTransactionHasTooManyTags
	}

	pictureIds, err := utils.StringArrayToInt64Array(transactionCreateReq.PictureIds)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionCreateHandler] parse picture ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionPictureIdInvalid
	}

	if len(pictureIds) > models.MaximumPicturesCountOfTransaction {
		return nil, errs.ErrTransactionHasTooManyPictures
	}

	if transactionCreateReq.Type < models.TRANSACTION_TYPE_MODIFY_BALANCE || transactionCreateReq.Type > models.TRANSACTION_TYPE_TRANSFER {
		log.Warnf(c, "[transactions.TransactionCreateHandler] transaction type is invalid")
		return nil, errs.ErrTransactionTypeInvalid
	}

	if transactionCreateReq.Type == models.TRANSACTION_TYPE_MODIFY_BALANCE && transactionCreateReq.CategoryId != 0 {
		log.Warnf(c, "[transactions.TransactionCreateHandler] balance modification transaction cannot set category id")
		return nil, errs.ErrBalanceModificationTransactionCannotSetCategory
	}

	if transactionCreateReq.Type != models.TRANSACTION_TYPE_TRANSFER && transactionCreateReq.DestinationAccountId != 0 {
		log.Warnf(c, "[transactions.TransactionCreateHandler] non-transfer transaction destination account cannot be set")
		return nil, errs.ErrTransactionDestinationAccountCannotBeSet
	} else if transactionCreateReq.Type == models.TRANSACTION_TYPE_TRANSFER && transactionCreateReq.SourceAccountId == transactionCreateReq.DestinationAccountId {
		log.Warnf(c, "[transactions.TransactionCreateHandler] transfer transaction source account must not be destination account")
		return nil, errs.ErrTransactionSourceAndDestinationIdCannotBeEqual
	}

	if transactionCreateReq.Type != models.TRANSACTION_TYPE_TRANSFER && transactionCreateReq.DestinationAmount != 0 {
		log.Warnf(c, "[transactions.TransactionCreateHandler] non-transfer transaction destination amount cannot be set")
		return nil, errs.ErrTransactionDestinationAmountCannotBeSet
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionCreateHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	transaction := a.createNewTransactionModel(uid, &transactionCreateReq, c.ClientIP())
	transactionEditable := user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone)

	if !transactionEditable {
		return nil, errs.ErrCannotCreateTransactionWithThisTransactionTime
	}

	var pictureInfos []*models.TransactionPictureInfo

	if len(pictureIds) > 0 {
		pictureInfos, err = a.transactionPictures.GetNewPictureInfosByPictureIds(c, uid, pictureIds)

		if err != nil {
			log.Errorf(c, "[transactions.TransactionCreateHandler] failed to get transactions pictures for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		notExistsPictureIds := utils.Int64SliceMinus(pictureIds, a.transactionPictures.GetTransactionPictureIds(pictureInfos))

		if len(notExistsPictureIds) > 0 {
			log.Errorf(c, "[transactions.TransactionCreateHandler] some pictures \"ids:%s\" does not exists for user \"uid:%d\"", strings.Join(utils.Int64ArrayToStringArray(notExistsPictureIds), ","), uid)
			return nil, errs.ErrTransactionPictureNotFound
		}
	}

	if a.CurrentConfig().EnableDuplicateSubmissionsCheck && transactionCreateReq.ClientSessionId != "" {
		found, remark := a.GetSubmissionRemark(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_TRANSACTION, uid, transactionCreateReq.ClientSessionId)

		if found {
			log.Infof(c, "[transactions.TransactionCreateHandler] another transaction \"id:%s\" has been created for user \"uid:%d\"", remark, uid)
			transactionId, err := utils.StringToInt64(remark)

			if err == nil {
				transaction, err = a.transactions.GetTransactionByTransactionId(c, uid, transactionId)

				if err != nil {
					log.Errorf(c, "[transactions.TransactionCreateHandler] failed to get existed transaction \"id:%d\" for user \"uid:%d\", because %s", transactionId, uid, err.Error())
					return nil, errs.Or(err, errs.ErrOperationFailed)
				}

				transactionResp := transaction.ToTransactionInfoResponse(tagIds, transactionEditable)
				transactionResp.Pictures = a.GetTransactionPictureInfoResponseList(pictureInfos)

				return transactionResp, nil
			}
		}
	}

	// If splits are provided, override amount and category from splits
	if len(transactionCreateReq.Splits) > 0 {
		var totalAmount int64
		for _, split := range transactionCreateReq.Splits {
			totalAmount += split.Amount
		}
		transaction.Amount = totalAmount
		transaction.CategoryId = transactionCreateReq.Splits[0].CategoryId
	}

	err = a.transactions.CreateTransaction(c, transaction, tagIds, pictureIds)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionCreateHandler] failed to create transaction \"id:%d\" for user \"uid:%d\", because %s", transaction.TransactionId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionCreateHandler] user \"uid:%d\" has created a new transaction \"id:%d\" successfully", uid, transaction.TransactionId)

	// Save splits if provided
	var splitResponses []models.TransactionSplitResponse
	if len(transactionCreateReq.Splits) > 0 {
		splitErr := a.transactionSplits.CreateSplits(c, uid, transaction.TransactionId, transactionCreateReq.Splits)
		if splitErr != nil {
			log.Errorf(c, "[transactions.TransactionCreateHandler] failed to create splits for transaction \"id:%d\" for user \"uid:%d\", because %s", transaction.TransactionId, uid, splitErr.Error())
		} else {
			for _, s := range transactionCreateReq.Splits {
				splitResponses = append(splitResponses, models.TransactionSplitResponse{
					CategoryId: s.CategoryId,
					Amount:     s.Amount,
					
					TagIds:     s.TagIds,
				})
			}
		}
	}

	// Handle repeatable transaction: create a template and generate planned future transactions
	if transactionCreateReq.Repeatable && transactionCreateReq.RepeatFrequencyType > models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DISABLED {
		// Determine if the base transaction should be planned (date > today)
		transactionUnixTime := utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime)
		now := time.Now()
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		if transactionUnixTime > todayStart.Unix() {
			transaction.Planned = true
			// Update the base transaction to set Planned=true
			updateErr := a.transactions.SetTransactionPlanned(c, uid, transaction.TransactionId, true)
			if updateErr != nil {
				log.Warnf(c, "[transactions.TransactionCreateHandler] failed to set transaction \"id:%d\" as planned for user \"uid:%d\", because %s", transaction.TransactionId, uid, updateErr.Error())
			}
		}

		// Create a TransactionTemplate for the repeatable transaction
		tagIdStrs := utils.Int64ArrayToStringArray(tagIds)
		template := &models.TransactionTemplate{
			Uid:                        uid,
			TemplateType:               models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE,
			Name:                       fmt.Sprintf("Repeat: %s", transactionCreateReq.Comment),
			Type:                       transactionCreateReq.Type,
			CategoryId:                 transactionCreateReq.CategoryId,
			AccountId:                  transactionCreateReq.SourceAccountId,
			ScheduledFrequencyType:     transactionCreateReq.RepeatFrequencyType,
			ScheduledFrequency:         transactionCreateReq.RepeatFrequency,
			TagIds:                     strings.Join(tagIdStrs, ","),
			Amount:                     transactionCreateReq.SourceAmount,
			RelatedAccountId:           transactionCreateReq.DestinationAccountId,
			RelatedAccountAmount:       transactionCreateReq.DestinationAmount,
			HideAmount:                 transactionCreateReq.HideAmount,
			Comment:                    transactionCreateReq.Comment,
			ScheduledTimezoneUtcOffset: transactionCreateReq.UtcOffset,
		}

		templateErr := a.transactionTemplates.CreateTemplate(c, template)
		if templateErr != nil {
			log.Errorf(c, "[transactions.TransactionCreateHandler] failed to create template for repeatable transaction for user \"uid:%d\", because %s", uid, templateErr.Error())
		} else {
			log.Infof(c, "[transactions.TransactionCreateHandler] user \"uid:%d\" has created template \"id:%d\" for repeatable transaction", uid, template.TemplateId)

			// Set SourceTemplateId on the base transaction
			setTemplateErr := a.transactions.SetTransactionSourceTemplateId(c, uid, transaction.TransactionId, template.TemplateId)
			if setTemplateErr != nil {
				log.Warnf(c, "[transactions.TransactionCreateHandler] failed to set source template id on transaction \"id:%d\" for user \"uid:%d\", because %s", transaction.TransactionId, uid, setTemplateErr.Error())
			} else {
				transaction.SourceTemplateId = template.TemplateId
			}

			// Generate planned future transactions
			plannedCount, genErr := a.transactions.GeneratePlannedTransactions(c, transaction, tagIds, transactionCreateReq.RepeatFrequencyType, transactionCreateReq.RepeatFrequency, template.TemplateId)
			if genErr != nil {
				log.Errorf(c, "[transactions.TransactionCreateHandler] failed to generate all planned transactions for user \"uid:%d\", generated %d, because %s", uid, plannedCount, genErr.Error())
			} else {
				log.Infof(c, "[transactions.TransactionCreateHandler] user \"uid:%d\" has generated %d planned transactions for template \"id:%d\"", uid, plannedCount, template.TemplateId)
			}
		}
	}

	a.SetSubmissionRemarkIfEnable(duplicatechecker.DUPLICATE_CHECKER_TYPE_NEW_TRANSACTION, uid, transactionCreateReq.ClientSessionId, utils.Int64ToString(transaction.TransactionId))
	transactionResp := transaction.ToTransactionInfoResponse(tagIds, transactionEditable)
	transactionResp.Pictures = a.GetTransactionPictureInfoResponseList(pictureInfos)
	transactionResp.Splits = splitResponses

	return transactionResp, nil
}

// TransactionModifyHandler saves an existed transaction by request parameters for current user
func (a *TransactionsApi) TransactionModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionModifyReq models.TransactionModifyRequest
	err := c.ShouldBindJSON(&transactionModifyReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionModifyHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	tagIds, err := utils.StringArrayToInt64Array(transactionModifyReq.TagIds)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionModifyHandler] parse tag ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionTagIdInvalid
	}

	if len(tagIds) > models.MaximumTagsCountOfTransaction {
		return nil, errs.ErrTransactionHasTooManyTags
	}

	pictureIds, err := utils.StringArrayToInt64Array(transactionModifyReq.PictureIds)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionModifyHandler] parse picture ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionPictureIdInvalid
	}

	if len(pictureIds) > models.MaximumPicturesCountOfTransaction {
		return nil, errs.ErrTransactionHasTooManyPictures
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionModifyHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	transaction, err := a.transactions.GetTransactionByTransactionId(c, uid, transactionModifyReq.Id)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionModifyHandler] failed to get transaction \"id:%d\" for user \"uid:%d\", because %s", transactionModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		log.Warnf(c, "[transactions.TransactionModifyHandler] cannot modify transaction \"id:%d\" for user \"uid:%d\", because transaction type is transfer in", transactionModifyReq.Id, uid)
		return nil, errs.ErrTransactionTypeInvalid
	}

	if transaction.Type == models.TRANSACTION_DB_TYPE_MODIFY_BALANCE && transactionModifyReq.CategoryId != 0 {
		log.Warnf(c, "[transactions.TransactionModifyHandler] balance modification transaction cannot set category id")
		return nil, errs.ErrBalanceModificationTransactionCannotSetCategory
	} else if transaction.Type != models.TRANSACTION_DB_TYPE_MODIFY_BALANCE && transactionModifyReq.CategoryId == 0 {
		log.Warnf(c, "[transactions.TransactionModifyHandler] non-balance modification transaction must set category id")
		return nil, errs.ErrIncompleteOrIncorrectSubmission
	}

	allTransactionTagIds, err := a.transactionTags.GetAllTagIdsOfTransactions(c, uid, []int64{transaction.TransactionId})

	if err != nil {
		log.Errorf(c, "[transactions.TransactionModifyHandler] failed to get transactions tag ids for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactionTagIds := allTransactionTagIds[transaction.TransactionId]

	if transactionTagIds == nil {
		transactionTagIds = make([]int64, 0, 0)
	}

	transactionPictureInfos, err := a.transactionPictures.GetPictureInfosByTransactionId(c, uid, transaction.TransactionId)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionModifyHandler] failed to get transaction picture infos for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	transactionPictureIds := a.transactionPictures.GetTransactionPictureIds(transactionPictureInfos)

	// If splits are provided, override amount and category from splits
	modifySourceAmount := transactionModifyReq.SourceAmount
	modifyCategoryId := transactionModifyReq.CategoryId
	if len(transactionModifyReq.Splits) > 0 {
		var totalAmount int64
		for _, split := range transactionModifyReq.Splits {
			totalAmount += split.Amount
		}
		modifySourceAmount = totalAmount
		modifyCategoryId = transactionModifyReq.Splits[0].CategoryId
	}

	newTransaction := &models.Transaction{
		TransactionId:     transaction.TransactionId,
		Uid:               uid,
		CategoryId:        modifyCategoryId,
		TransactionTime:   utils.GetMinTransactionTimeFromUnixTime(transactionModifyReq.Time),
		TimezoneUtcOffset: transactionModifyReq.UtcOffset,
		AccountId:         transactionModifyReq.SourceAccountId,
		Amount:            modifySourceAmount,
		HideAmount:        transactionModifyReq.HideAmount,
		CounterpartyId:    transactionModifyReq.CounterpartyId,
		Comment:           transactionModifyReq.Comment,
	}

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
		newTransaction.RelatedAccountId = transactionModifyReq.DestinationAccountId
		newTransaction.RelatedAccountAmount = transactionModifyReq.DestinationAmount
	}

	if transactionModifyReq.GeoLocation != nil {
		newTransaction.GeoLongitude = transactionModifyReq.GeoLocation.Longitude
		newTransaction.GeoLatitude = transactionModifyReq.GeoLocation.Latitude
	}

	// If splits are provided, always allow the update (splits may have changed)
	hasSplitsChange := len(transactionModifyReq.Splits) > 0

	if !hasSplitsChange &&
		newTransaction.CategoryId == transaction.CategoryId &&
		utils.GetUnixTimeFromTransactionTime(newTransaction.TransactionTime) == utils.GetUnixTimeFromTransactionTime(transaction.TransactionTime) &&
		newTransaction.TimezoneUtcOffset == transaction.TimezoneUtcOffset &&
		newTransaction.AccountId == transaction.AccountId &&
		newTransaction.Amount == transaction.Amount &&
		(transaction.Type != models.TRANSACTION_DB_TYPE_TRANSFER_OUT || newTransaction.RelatedAccountId == transaction.RelatedAccountId) &&
		(transaction.Type != models.TRANSACTION_DB_TYPE_TRANSFER_OUT || newTransaction.RelatedAccountAmount == transaction.RelatedAccountAmount) &&
		newTransaction.HideAmount == transaction.HideAmount &&
		newTransaction.CounterpartyId == transaction.CounterpartyId &&
		newTransaction.Comment == transaction.Comment &&
		newTransaction.GeoLongitude == transaction.GeoLongitude &&
		newTransaction.GeoLatitude == transaction.GeoLatitude &&
		utils.Int64SliceEquals(tagIds, transactionTagIds) &&
		utils.Int64SliceEquals(pictureIds, transactionPictureIds) {
		return nil, errs.ErrNothingWillBeUpdated
	}

	transactionEditable := user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone)
	newTransactionEditable := user.CanEditTransactionByTransactionTime(newTransaction.TransactionTime, clientTimezone)

	if !transactionEditable || !newTransactionEditable {
		return nil, errs.ErrCannotModifyTransactionWithThisTransactionTime
	}

	var addTransactionTagIds []int64
	var removeTransactionTagIds []int64

	if !utils.Int64SliceEquals(tagIds, transactionTagIds) {
		removeTransactionTagIds = transactionTagIds
		addTransactionTagIds = tagIds
	}

	addTransactionPictureIds := utils.Int64SliceMinus(pictureIds, transactionPictureIds)
	removeTransactionPictureIds := utils.Int64SliceMinus(transactionPictureIds, pictureIds)
	var newPictureInfos []*models.TransactionPictureInfo

	if !utils.Int64SliceEquals(pictureIds, transactionPictureIds) {
		oldAndNewPictureIds := transactionPictureIds
		oldAndNewPictureInfoMap := a.transactionPictures.GetPictureInfoMapByList(transactionPictureInfos)

		if len(addTransactionPictureIds) > 0 {
			addPictureInfos, err := a.transactionPictures.GetNewPictureInfosByPictureIds(c, uid, addTransactionPictureIds)

			if err != nil {
				log.Errorf(c, "[transactions.TransactionModifyHandler] failed to get transactions pictures for user \"uid:%d\", because %s", uid, err.Error())
				return nil, errs.Or(err, errs.ErrOperationFailed)
			}

			oldAndNewPictureIds = append(oldAndNewPictureIds, a.transactionPictures.GetTransactionPictureIds(addPictureInfos)...)
			notExistsPictureIds := utils.Int64SliceMinus(pictureIds, oldAndNewPictureIds)

			if len(notExistsPictureIds) > 0 {
				log.Errorf(c, "[transactions.TransactionModifyHandler] some pictures \"ids:%s\" does not exists for user \"uid:%d\"", strings.Join(utils.Int64ArrayToStringArray(notExistsPictureIds), ","), uid)
				return nil, errs.ErrTransactionPictureNotFound
			}

			for i := 0; i < len(addPictureInfos); i++ {
				oldAndNewPictureInfoMap[addPictureInfos[i].PictureId] = addPictureInfos[i]
			}
		}

		for i := 0; i < len(pictureIds); i++ {
			pictureId := pictureIds[i]
			pictureInfo, exists := oldAndNewPictureInfoMap[pictureId]

			if exists {
				newPictureInfos = append(newPictureInfos, pictureInfo)
			}
		}
	}

	err = a.transactions.ModifyTransaction(c, newTransaction, len(transactionTagIds), addTransactionTagIds, removeTransactionTagIds, addTransactionPictureIds, removeTransactionPictureIds)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionModifyHandler] failed to update transaction \"id:%d\" for user \"uid:%d\", because %s", transactionModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionModifyHandler] user \"uid:%d\" has updated transaction \"id:%d\" successfully", uid, transactionModifyReq.Id)

	// Handle splits: replace old splits with new ones (or delete if no splits provided)
	var splitResponses []models.TransactionSplitResponse
	splitErr := a.transactionSplits.ReplaceSplits(c, uid, transaction.TransactionId, transactionModifyReq.Splits)
	if splitErr != nil {
		log.Errorf(c, "[transactions.TransactionModifyHandler] failed to replace splits for transaction \"id:%d\" for user \"uid:%d\", because %s", transaction.TransactionId, uid, splitErr.Error())
	} else if len(transactionModifyReq.Splits) > 0 {
		for _, s := range transactionModifyReq.Splits {
			splitResponses = append(splitResponses, models.TransactionSplitResponse{
				CategoryId: s.CategoryId,
				Amount:     s.Amount,
				
				TagIds:     s.TagIds,
			})
		}
	}

	newTransaction.Type = transaction.Type
	newTransaction.Planned = transaction.Planned
	newTransaction.SourceTemplateId = transaction.SourceTemplateId
	newTransactionResp := newTransaction.ToTransactionInfoResponse(tagIds, transactionEditable)
	newTransactionResp.Pictures = a.GetTransactionPictureInfoResponseList(newPictureInfos)
	newTransactionResp.Splits = splitResponses

	return newTransactionResp, nil
}

// TransactionConfirmHandler confirms a planned transaction for current user
func (a *TransactionsApi) TransactionConfirmHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionConfirmReq models.TransactionConfirmRequest
	err := c.ShouldBindJSON(&transactionConfirmReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionConfirmHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionConfirmHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()

	transaction, err := a.transactions.ConfirmPlannedTransaction(c, uid, transactionConfirmReq.Id, clientTimezone)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionConfirmHandler] failed to confirm planned transaction \"id:%d\" for user \"uid:%d\", because %s", transactionConfirmReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionConfirmHandler] user \"uid:%d\" has confirmed planned transaction \"id:%d\" successfully", uid, transaction.TransactionId)

	transactionTagIds, err := a.transactionTags.GetAllTagIdsOfTransactions(c, uid, []int64{transaction.TransactionId})

	if err != nil {
		log.Errorf(c, "[transactions.TransactionConfirmHandler] failed to get transaction tags for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	tagIds := transactionTagIds[transaction.TransactionId]

	if tagIds == nil {
		tagIds = make([]int64, 0)
	}

	transactionResp := transaction.ToTransactionInfoResponse(tagIds, true)

	return transactionResp, nil
}

// TransactionModifyAllFutureHandler modifies all future planned transactions for current user
func (a *TransactionsApi) TransactionModifyAllFutureHandler(c *core.WebContext) (any, *errs.Error) {
	var modifyReq models.TransactionModifyAllFutureRequest

	// Pre-process: read body, replace empty strings with "0" for json:",string" int64 fields
	bodyBytes, readErr := io.ReadAll(c.Request.Body)
	if readErr != nil {
		log.Warnf(c, "[transactions.TransactionModifyAllFutureHandler] failed to read body, because %s", readErr.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(readErr)
	}

	var rawMap map[string]interface{}
	if jsonErr := json.Unmarshal(bodyBytes, &rawMap); jsonErr == nil {
		for _, key := range []string{"id", "categoryId", "sourceAccountId", "destinationAccountId", "counterpartyId"} {
			if val, ok := rawMap[key]; ok {
				if strVal, isStr := val.(string); isStr && strVal == "" {
					rawMap[key] = "0"
				}
			}
		}
		bodyBytes, _ = json.Marshal(rawMap)
	}

	err := json.Unmarshal(bodyBytes, &modifyReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionModifyAllFutureHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	log.Infof(c, "[transactions.TransactionModifyAllFutureHandler] request: id=%d, sourceAmount=%d, categoryId=%d, sourceAccountId=%d, destAccountId=%d, destAmount=%d, counterpartyId=%d, comment=%s",
		modifyReq.Id, modifyReq.SourceAmount, modifyReq.CategoryId, modifyReq.SourceAccountId, modifyReq.DestinationAccountId, modifyReq.DestinationAmount, modifyReq.CounterpartyId, modifyReq.Comment)

	affectedCount, err := a.transactions.ModifyAllFuturePlannedTransactions(c, uid, modifyReq.Id, &modifyReq)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionModifyAllFutureHandler] failed to modify future planned transactions for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionModifyAllFutureHandler] user \"uid:%d\" has modified %d future planned transactions successfully", uid, affectedCount)

	return map[string]int64{"affectedCount": affectedCount}, nil
}

// TransactionDeleteAllFutureHandler deletes all future planned transactions for current user
func (a *TransactionsApi) TransactionDeleteAllFutureHandler(c *core.WebContext) (any, *errs.Error) {
	var deleteReq models.TransactionDeleteRequest
	err := c.ShouldBindJSON(&deleteReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionDeleteAllFutureHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	affectedCount, err := a.transactions.DeleteAllFuturePlannedTransactions(c, uid, deleteReq.Id)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionDeleteAllFutureHandler] failed to delete future planned transactions for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionDeleteAllFutureHandler] user \"uid:%d\" has deleted %d future planned transactions successfully", uid, affectedCount)

	return map[string]int64{"affectedCount": affectedCount}, nil
}

// TransactionMoveAllBetweenAccountsHandler moves all transactions from one account to another account for current user
func (a *TransactionsApi) TransactionMoveAllBetweenAccountsHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionMoveReq models.TransactionMoveBetweenAccountsRequest
	err := c.ShouldBindJSON(&transactionMoveReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionMoveAllBetweenAccountsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	if transactionMoveReq.FromAccountId == transactionMoveReq.ToAccountId {
		return nil, errs.ErrCannotMoveTransactionToSameAccount
	}

	uid := c.GetCurrentUid()
	accountMap, err := a.accounts.GetAccountsByAccountIds(c, uid, []int64{transactionMoveReq.FromAccountId, transactionMoveReq.ToAccountId})

	if err != nil {
		log.Errorf(c, "[transactions.TransactionMoveAllBetweenAccountsHandler] failed to get accounts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	fromAccount, exists := accountMap[transactionMoveReq.FromAccountId]

	if !exists {
		return nil, errs.ErrSourceAccountNotFound
	}

	toAccount, exists := accountMap[transactionMoveReq.ToAccountId]

	if !exists {
		return nil, errs.ErrDestinationAccountNotFound
	}

	if fromAccount.Hidden || toAccount.Hidden {
		return nil, errs.ErrCannotMoveTransactionFromOrToHiddenAccount
	}

	if fromAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS || toAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
		return nil, errs.ErrCannotMoveTransactionFromOrToParentAccount
	}

	if fromAccount.Currency != toAccount.Currency {
		return nil, errs.ErrCannotMoveTransactionBetweenAccountsWithDifferentCurrencies
	}

	err = a.transactions.MoveAllTransactionsBetweenAccounts(c, uid, transactionMoveReq.FromAccountId, transactionMoveReq.ToAccountId)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionMoveAllBetweenAccountsHandler] failed to move all transactions from account \"id:%d\" to account \"id:%d\" for user \"uid:%d\", because %s", transactionMoveReq.FromAccountId, transactionMoveReq.ToAccountId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionMoveAllBetweenAccountsHandler] user \"uid:%d\" has moved all transactions from account \"id:%d\" to account \"id:%d\" successfully", uid, transactionMoveReq.FromAccountId, transactionMoveReq.ToAccountId)
	return true, nil
}

// TransactionDeleteHandler deletes an existed transaction by request parameters for current user
func (a *TransactionsApi) TransactionDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionDeleteReq models.TransactionDeleteRequest
	err := c.ShouldBindJSON(&transactionDeleteReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionDeleteHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionDeleteHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	transaction, err := a.transactions.GetTransactionByTransactionId(c, uid, transactionDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionDeleteHandler] failed to get transaction \"id:%d\" for user \"uid:%d\", because %s", transactionDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		log.Warnf(c, "[transactions.TransactionDeleteHandler] cannot delete transaction \"id:%d\" for user \"uid:%d\", because transaction type is transfer in", transactionDeleteReq.Id, uid)
		return nil, errs.ErrTransactionTypeInvalid
	}

	transactionEditable := user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone)

	if !transactionEditable {
		return nil, errs.ErrCannotDeleteTransactionWithThisTransactionTime
	}

	err = a.transactions.DeleteTransaction(c, uid, transactionDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionDeleteHandler] failed to delete transaction \"id:%d\" for user \"uid:%d\", because %s", transactionDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionDeleteHandler] user \"uid:%d\" has deleted transaction \"id:%d\"", uid, transactionDeleteReq.Id)
	return true, nil
}

func (a *TransactionsApi) createNewTransactionModel(uid int64, transactionCreateReq *models.TransactionCreateRequest, clientIp string) *models.Transaction {
	var transactionDbType models.TransactionDbType

	switch transactionCreateReq.Type {
	case models.TRANSACTION_TYPE_MODIFY_BALANCE:
		transactionDbType = models.TRANSACTION_DB_TYPE_MODIFY_BALANCE
	case models.TRANSACTION_TYPE_EXPENSE:
		transactionDbType = models.TRANSACTION_DB_TYPE_EXPENSE
	case models.TRANSACTION_TYPE_INCOME:
		transactionDbType = models.TRANSACTION_DB_TYPE_INCOME
	case models.TRANSACTION_TYPE_TRANSFER:
		transactionDbType = models.TRANSACTION_DB_TYPE_TRANSFER_OUT
	}

	transaction := &models.Transaction{
		Uid:               uid,
		Type:              transactionDbType,
		CategoryId:        transactionCreateReq.CategoryId,
		TransactionTime:   utils.GetMinTransactionTimeFromUnixTime(transactionCreateReq.Time),
		TimezoneUtcOffset: transactionCreateReq.UtcOffset,
		AccountId:         transactionCreateReq.SourceAccountId,
		Amount:            transactionCreateReq.SourceAmount,
		HideAmount:        transactionCreateReq.HideAmount,
		CounterpartyId:    transactionCreateReq.CounterpartyId,
		Comment:           transactionCreateReq.Comment,
		CreatedIp:         clientIp,
	}

	if transactionCreateReq.Type == models.TRANSACTION_TYPE_TRANSFER {
		transaction.RelatedAccountId = transactionCreateReq.DestinationAccountId
		transaction.RelatedAccountAmount = transactionCreateReq.DestinationAmount
	}

	if transactionCreateReq.GeoLocation != nil {
		transaction.GeoLongitude = transactionCreateReq.GeoLocation.Longitude
		transaction.GeoLatitude = transactionCreateReq.GeoLocation.Latitude
	}

	return transaction
}

// TransactionSetPlannedHandler sets or unsets the planned flag for a transaction for current user
func (a *TransactionsApi) TransactionSetPlannedHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionSetPlannedReq models.TransactionSetPlannedRequest
	err := c.ShouldBindJSON(&transactionSetPlannedReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionSetPlannedHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	if transactionSetPlannedReq.Planned {
		// Converting actual -> planned: need to reverse balance changes
		err = a.transactions.UnconfirmTransaction(c, uid, transactionSetPlannedReq.Id)
	} else {
		// Converting planned -> actual: just flip the flag (confirm should be used instead normally)
		err = a.transactions.SetTransactionPlanned(c, uid, transactionSetPlannedReq.Id, false)
	}

	if err != nil {
		log.Errorf(c, "[transactions.TransactionSetPlannedHandler] failed to set planned flag for transaction \"id:%d\" for user \"uid:%d\", because %s", transactionSetPlannedReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionSetPlannedHandler] user \"uid:%d\" has set planned flag for transaction \"id:%d\" to %v successfully", uid, transactionSetPlannedReq.Id, transactionSetPlannedReq.Planned)

	return true, nil
}


// TransactionMakeRepeatableHandler makes an existing transaction repeatable by creating a template and planned transactions
func (a *TransactionsApi) TransactionMakeRepeatableHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.TransactionMakeRepeatableRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionMakeRepeatableHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	transaction, err := a.transactions.GetTransactionByTransactionId(c, uid, req.Id)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionMakeRepeatableHandler] failed to get transaction \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		log.Warnf(c, "[transactions.TransactionMakeRepeatableHandler] cannot make transaction \"id:%d\" repeatable for user \"uid:%d\", because transaction type is transfer in", req.Id, uid)
		return nil, errs.ErrTransactionTypeInvalid
	}

	if transaction.SourceTemplateId != 0 {
		log.Warnf(c, "[transactions.TransactionMakeRepeatableHandler] transaction \"id:%d\" for user \"uid:%d\" is already repeatable", req.Id, uid)
		return nil, errs.ErrTransactionAlreadyRepeatable
	}

	transactionType, err := transaction.Type.ToTransactionType()

	if err != nil {
		log.Errorf(c, "[transactions.TransactionMakeRepeatableHandler] failed to convert transaction type for transaction \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, err.Error())
		return nil, errs.ErrTransactionTypeInvalid
	}

	// Get tag IDs for the transaction
	allTransactionTagIds, err := a.transactionTags.GetAllTagIdsOfTransactions(c, uid, []int64{transaction.TransactionId})

	if err != nil {
		log.Errorf(c, "[transactions.TransactionMakeRepeatableHandler] failed to get transaction tag ids for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	tagIds := allTransactionTagIds[transaction.TransactionId]
	tagIdStrs := utils.Int64ArrayToStringArray(tagIds)

	// Create a TransactionTemplate for the repeatable transaction
	template := &models.TransactionTemplate{
		Uid:                    uid,
		TemplateType:           models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE,
		Name:                   fmt.Sprintf("Repeat: %s", transaction.Comment),
		Type:                   transactionType,
		CategoryId:             transaction.CategoryId,
		AccountId:              transaction.AccountId,
		ScheduledFrequencyType: models.TransactionScheduleFrequencyType(req.RepeatFrequencyType),
		ScheduledFrequency:     req.RepeatFrequency,
		TagIds:                 strings.Join(tagIdStrs, ","),
		Amount:                 transaction.Amount,
		RelatedAccountId:       transaction.RelatedAccountId,
		RelatedAccountAmount:   transaction.RelatedAccountAmount,
		HideAmount:             transaction.HideAmount,
		Comment:                transaction.Comment,
	}

	templateErr := a.transactionTemplates.CreateTemplate(c, template)

	if templateErr != nil {
		log.Errorf(c, "[transactions.TransactionMakeRepeatableHandler] failed to create template for user \"uid:%d\", because %s", uid, templateErr.Error())
		return nil, errs.Or(templateErr, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionMakeRepeatableHandler] user \"uid:%d\" has created template \"id:%d\" for repeatable transaction", uid, template.TemplateId)

	// Set SourceTemplateId on the transaction
	setTemplateErr := a.transactions.SetTransactionSourceTemplateId(c, uid, transaction.TransactionId, template.TemplateId)

	if setTemplateErr != nil {
		log.Warnf(c, "[transactions.TransactionMakeRepeatableHandler] failed to set source template id on transaction \"id:%d\" for user \"uid:%d\", because %s", transaction.TransactionId, uid, setTemplateErr.Error())
	} else {
		transaction.SourceTemplateId = template.TemplateId
	}

	// Generate planned future transactions
	plannedCount, genErr := a.transactions.GeneratePlannedTransactions(c, transaction, tagIds, models.TransactionScheduleFrequencyType(req.RepeatFrequencyType), req.RepeatFrequency, template.TemplateId)

	if genErr != nil {
		log.Errorf(c, "[transactions.TransactionMakeRepeatableHandler] failed to generate planned transactions for user \"uid:%d\", generated %d, because %s", uid, plannedCount, genErr.Error())
	} else {
		log.Infof(c, "[transactions.TransactionMakeRepeatableHandler] user \"uid:%d\" has generated %d planned transactions for template \"id:%d\"", uid, plannedCount, template.TemplateId)
	}

	return map[string]any{
		"templateId":   template.TemplateId,
		"plannedCount": plannedCount,
	}, nil
}
