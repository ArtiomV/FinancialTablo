package api

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/converters"
	"github.com/mayswind/ezbookkeeping/pkg/converters/converter"
	"github.com/mayswind/ezbookkeeping/pkg/converters/datatable"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/duplicatechecker"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// TransactionParseImportDsvFileDataHandler returns the parsed file data by request parameters for current user
func (a *TransactionsApi) TransactionParseImportDsvFileDataHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	form, err := c.MultipartForm()

	if err != nil {
		log.Errorf(c, "[transactions.TransactionParseImportDsvFileDataHandler] failed to get multi-part form data for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrParameterInvalid
	}

	fileTypes := form.Value["fileType"]

	if len(fileTypes) < 1 || fileTypes[0] == "" {
		return nil, errs.ErrImportFileTypeIsEmpty
	}

	fileType := fileTypes[0]

	if !converters.IsCustomDelimiterSeparatedValuesFileType(fileType) {
		return nil, errs.Or(err, errs.ErrImportFileTypeNotSupported)
	}

	fileEncodings := form.Value["fileEncoding"]

	if len(fileEncodings) < 1 || fileEncodings[0] == "" {
		return nil, errs.ErrImportFileEncodingIsEmpty
	}

	fileEncoding := fileEncodings[0]
	dataParser, err := converters.CreateNewDelimiterSeparatedValuesDataParser(fileType, fileEncoding)

	if err != nil {
		return nil, errs.Or(err, errs.ErrImportFileTypeNotSupported)
	}

	importFiles := form.File["file"]

	if len(importFiles) < 1 {
		log.Warnf(c, "[transactions.TransactionParseImportDsvFileDataHandler] there is no import file in request for user \"uid:%d\"", uid)
		return nil, errs.ErrNoFilesUpload
	}

	if importFiles[0].Size < 1 {
		log.Warnf(c, "[transactions.TransactionParseImportDsvFileDataHandler] the size of import file in request is zero for user \"uid:%d\"", uid)
		return nil, errs.ErrUploadedFileEmpty
	}

	if importFiles[0].Size > int64(a.CurrentConfig().MaxImportFileSize) {
		log.Warnf(c, "[transactions.TransactionParseImportDsvFileDataHandler] the upload file size \"%d\" exceeds the maximum size \"%d\" of import file for user \"uid:%d\"", importFiles[0].Size, a.CurrentConfig().MaxImportFileSize, uid)
		return nil, errs.ErrExceedMaxUploadFileSize
	}

	importFile, err := importFiles[0].Open()

	if err != nil {
		log.Errorf(c, "[transactions.TransactionParseImportDsvFileDataHandler] failed to get import file from request for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	defer importFile.Close()
	fileData, err := io.ReadAll(importFile)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionParseImportDsvFileDataHandler] failed to read import file data for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	allLines, err := dataParser.ParseDsvFileLines(c, fileData)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionParseImportDsvFileDataHandler] failed to parse import file data for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return allLines, nil
}

// TransactionParseImportFileHandler returns the parsed transaction data by request parameters for current user
func (a *TransactionsApi) TransactionParseImportFileHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	form, err := c.MultipartForm()

	if err != nil {
		log.Errorf(c, "[transactions.TransactionParseImportFileHandler] failed to get multi-part form data for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrParameterInvalid
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionParseImportFileHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	fileTypes := form.Value["fileType"]

	if len(fileTypes) < 1 || fileTypes[0] == "" {
		return nil, errs.ErrImportFileTypeIsEmpty
	}

	fileType := fileTypes[0]

	textualOptions := form.Value["options"]
	textualOption := ""

	if len(textualOptions) > 0 {
		textualOption = textualOptions[0]
	}

	additionalOptions := converter.ParseImporterOptions(textualOption)

	var dataImporter converter.TransactionDataImporter

	if converters.IsCustomDelimiterSeparatedValuesFileType(fileType) {
		fileEncodings := form.Value["fileEncoding"]

		if len(fileEncodings) < 1 || fileEncodings[0] == "" {
			return nil, errs.ErrImportFileEncodingIsEmpty
		}

		fileEncoding := fileEncodings[0]

		columnMappings := form.Value["columnMapping"]

		if len(columnMappings) < 1 || columnMappings[0] == "" {
			return nil, errs.ErrImportFileColumnMappingInvalid
		}

		var columnIndexMapping = map[datatable.TransactionDataTableColumn]int{}
		err = json.Unmarshal([]byte(columnMappings[0]), &columnIndexMapping)

		if err != nil {
			log.Errorf(c, "[transactions.TransactionParseImportFileHandler] failed to parse column mapping for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.ErrImportFileColumnMappingInvalid
		}

		transactionTypeMappings := form.Value["transactionTypeMapping"]

		if len(transactionTypeMappings) < 1 || transactionTypeMappings[0] == "" {
			return nil, errs.ErrImportFileTransactionTypeMappingInvalid
		}

		var transactionTypeNameMapping = map[string]models.TransactionType{}
		err = json.Unmarshal([]byte(transactionTypeMappings[0]), &transactionTypeNameMapping)

		if err != nil {
			log.Errorf(c, "[transactions.TransactionParseImportFileHandler] failed to parse transaction type mapping for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.ErrImportFileTransactionTypeMappingInvalid
		}

		hasHeaderLines := form.Value["hasHeaderLine"]
		hasHeaderLine := false

		if len(hasHeaderLines) > 0 {
			hasHeaderLine = hasHeaderLines[0] == "true"
		}

		timeFormats := form.Value["timeFormat"]

		if len(timeFormats) < 1 || timeFormats[0] == "" {
			return nil, errs.ErrImportFileTransactionTimeFormatInvalid
		}

		timezoneFormats := form.Value["timezoneFormat"]
		timezoneFormat := ""

		if len(timezoneFormats) > 0 {
			timezoneFormat = timezoneFormats[0]
		}

		amountDecimalSeparators := form.Value["amountDecimalSeparator"]
		amountDecimalSeparator := ""

		if len(amountDecimalSeparators) > 0 {
			amountDecimalSeparator = amountDecimalSeparators[0]
		}

		amountDigitGroupingSymbols := form.Value["amountDigitGroupingSymbol"]
		amountDigitGroupingSymbol := ""

		if len(amountDigitGroupingSymbols) > 0 {
			amountDigitGroupingSymbol = amountDigitGroupingSymbols[0]
		}

		geoLocationSeparators := form.Value["geoSeparator"]
		geoLocationSeparator := ""

		if len(geoLocationSeparators) > 0 {
			geoLocationSeparator = geoLocationSeparators[0]
		}

		geoLocationOrders := form.Value["geoOrder"]
		geoLocationOrder := ""

		if len(geoLocationOrders) > 0 {
			geoLocationOrder = geoLocationOrders[0]
		}

		transactionTagSeparators := form.Value["tagSeparator"]
		transactionTagSeparator := ""

		if len(transactionTagSeparators) > 0 {
			transactionTagSeparator = transactionTagSeparators[0]
		}

		dataImporter, err = converters.CreateNewDelimiterSeparatedValuesDataImporter(fileType, fileEncoding, columnIndexMapping, transactionTypeNameMapping, hasHeaderLine, timeFormats[0], timezoneFormat, amountDecimalSeparator, amountDigitGroupingSymbol, geoLocationSeparator, geoLocationOrder, transactionTagSeparator)
	} else {
		dataImporter, err = converters.GetTransactionDataImporter(fileType)
	}

	if err != nil {
		return nil, errs.Or(err, errs.ErrImportFileTypeNotSupported)
	}

	importFiles := form.File["file"]

	if len(importFiles) < 1 {
		log.Warnf(c, "[transactions.TransactionParseImportFileHandler] there is no import file in request for user \"uid:%d\"", uid)
		return nil, errs.ErrNoFilesUpload
	}

	if importFiles[0].Size < 1 {
		log.Warnf(c, "[transactions.TransactionParseImportFileHandler] the size of import file in request is zero for user \"uid:%d\"", uid)
		return nil, errs.ErrUploadedFileEmpty
	}

	if importFiles[0].Size > int64(a.CurrentConfig().MaxImportFileSize) {
		log.Warnf(c, "[transactions.TransactionParseImportFileHandler] the upload file size \"%d\" exceeds the maximum size \"%d\" of import file for user \"uid:%d\"", importFiles[0].Size, a.CurrentConfig().MaxImportFileSize, uid)
		return nil, errs.ErrExceedMaxUploadFileSize
	}

	importFile, err := importFiles[0].Open()

	if err != nil {
		log.Errorf(c, "[transactions.TransactionParseImportFileHandler] failed to get import file from request for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrOperationFailed
	}

	defer importFile.Close()
	fileData, err := io.ReadAll(importFile)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionParseImportFileHandler] failed to read import file data for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionParseImportFileHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	if user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_IMPORT_TRANSACTION) {
		return nil, errs.ErrNotPermittedToPerformThisAction
	}

	accounts, err := a.accounts.GetAllAccountsByUid(c, user.Uid)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionParseImportFileHandler] failed to get accounts for user \"uid:%d\", because %s", user.Uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	accountMap := a.accounts.GetVisibleAccountNameMapByList(accounts)

	categories, err := a.transactionCategories.GetAllCategoriesByUid(c, user.Uid, 0)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionParseImportFileHandler] failed to get categories for user \"uid:%d\", because %s", user.Uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	expenseCategoryMap, incomeCategoryMap, transferCategoryMap := a.transactionCategories.GetVisibleCategoryNameMapByList(categories)

	tags, err := a.transactionTags.GetAllTagsByUid(c, user.Uid)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionParseImportFileHandler] failed to get tags for user \"uid:%d\", because %s", user.Uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	tagMap := a.transactionTags.GetVisibleTagNameMapByList(tags)

	parsedTransactions, allNewAccounts, allNewSubExpenseCategories, allNewSubIncomeCategories, allNewSubTransferCategories, allNewTags, err := dataImporter.ParseImportedData(c, user, fileData, clientTimezone, additionalOptions, accountMap, expenseCategoryMap, incomeCategoryMap, transferCategoryMap, tagMap)

	if err != nil {
		log.Errorf(c, "[transactions.TransactionParseImportFileHandler] failed to parse imported data for user \"uid:%d\", because %s", user.Uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	// Auto-create missing accounts
	if len(allNewAccounts) > 0 {
		for _, newAccount := range allNewAccounts {
			newAccount.Uid = user.Uid
			newAccount.Category = models.ACCOUNT_CATEGORY_CASH
			newAccount.Type = models.ACCOUNT_TYPE_SINGLE_ACCOUNT
			newAccount.Icon = 1
			newAccount.Color = "588a6a"

			maxOrder, orderErr := a.accounts.GetMaxDisplayOrder(c, user.Uid, newAccount.Category)
			if orderErr != nil {
				log.Warnf(c, "[transactions.TransactionParseImportFileHandler] failed to get max display order for account, because %s", orderErr.Error())
				maxOrder = 0
			}
			newAccount.DisplayOrder = maxOrder + 1

			createErr := a.accounts.CreateAccounts(c, newAccount, 0, nil, nil, clientTimezone)
			if createErr != nil {
				log.Errorf(c, "[transactions.TransactionParseImportFileHandler] failed to auto-create account \"%s\" for user \"uid:%d\", because %s", newAccount.Name, user.Uid, createErr.Error())
				return nil, errs.Or(createErr, errs.ErrOperationFailed)
			}

			log.Infof(c, "[transactions.TransactionParseImportFileHandler] auto-created account \"%s\" (id:%d) for user \"uid:%d\"", newAccount.Name, newAccount.AccountId, user.Uid)

			// Update account IDs in all parsed transactions
			for _, t := range parsedTransactions {
				if t.OriginalSourceAccountName == newAccount.Name {
					t.AccountId = newAccount.AccountId
				}
				if t.OriginalDestinationAccountName == newAccount.Name {
					t.RelatedAccountId = newAccount.AccountId
				}
			}
		}
	}

	// Auto-create missing categories (with hierarchy support)
	allNewCategories := make([]*models.TransactionCategory, 0)
	allNewCategories = append(allNewCategories, allNewSubExpenseCategories...)
	allNewCategories = append(allNewCategories, allNewSubIncomeCategories...)
	allNewCategories = append(allNewCategories, allNewSubTransferCategories...)

	// Build a set of child category names (those that have OriginalParentCategoryName)
	childCategoryNames := make(map[string]bool)
	for _, t := range parsedTransactions {
		if t.OriginalParentCategoryName != "" && t.OriginalCategoryName != "" {
			childCategoryNames[t.OriginalCategoryName] = true
		}
	}

	// Separate parent categories and child categories
	parentCategories := make([]*models.TransactionCategory, 0)
	childCategories := make([]*models.TransactionCategory, 0)

	for _, cat := range allNewCategories {
		if childCategoryNames[cat.Name] {
			childCategories = append(childCategories, cat)
		} else {
			parentCategories = append(parentCategories, cat)
		}
	}

	// Reorder: parents first, then children
	orderedCategories := make([]*models.TransactionCategory, 0, len(allNewCategories))
	orderedCategories = append(orderedCategories, parentCategories...)
	orderedCategories = append(orderedCategories, childCategories...)

	categoryTypeMaxOrderMap := make(map[models.TransactionCategoryType]int32)
	// Map of parent category name+type to created ID for resolving child references
	parentNameToIdMap := make(map[string]int64)

	for _, newCategory := range orderedCategories {
		if strings.TrimSpace(newCategory.Name) == "" {
			continue // skip categories with empty names
		}

		newCategory.Uid = user.Uid
		newCategory.Icon = 1
		newCategory.Color = "588a6a"

		isChild := childCategoryNames[newCategory.Name]

		// If this is a child category, resolve parent reference by OriginalParentCategoryName
		if isChild {
			parentName := ""
			for _, t := range parsedTransactions {
				if t.OriginalCategoryName == newCategory.Name && t.OriginalParentCategoryName != "" {
					parentName = t.OriginalParentCategoryName
					break
				}
			}

			if parentName != "" {
				parentKey := fmt.Sprintf("%s:%d", parentName, newCategory.Type)
				if parentId, ok := parentNameToIdMap[parentKey]; ok {
					newCategory.ParentCategoryId = parentId
				}
			}
		}

		maxOrder, exists := categoryTypeMaxOrderMap[newCategory.Type]
		if !exists {
			var orderErr error
			maxOrder, orderErr = a.transactionCategories.GetMaxDisplayOrder(c, user.Uid, newCategory.Type)
			if orderErr != nil {
				log.Warnf(c, "[transactions.TransactionParseImportFileHandler] failed to get max display order for category, because %s", orderErr.Error())
				maxOrder = 0
			}
		}
		newCategory.DisplayOrder = maxOrder + 1
		categoryTypeMaxOrderMap[newCategory.Type] = maxOrder + 1

		createErr := a.transactionCategories.CreateCategory(c, newCategory)
		if createErr != nil {
			log.Errorf(c, "[transactions.TransactionParseImportFileHandler] failed to auto-create category \"%s\" for user \"uid:%d\", because %s", newCategory.Name, user.Uid, createErr.Error())
			return nil, errs.Or(createErr, errs.ErrOperationFailed)
		}

		log.Infof(c, "[transactions.TransactionParseImportFileHandler] auto-created category \"%s\" (id:%d, type:%d, parentId:%d) for user \"uid:%d\"", newCategory.Name, newCategory.CategoryId, newCategory.Type, newCategory.ParentCategoryId, user.Uid)

		// If this is a parent category, store its ID for child resolution
		if !isChild {
			parentKey := fmt.Sprintf("%s:%d", newCategory.Name, newCategory.Type)
			parentNameToIdMap[parentKey] = newCategory.CategoryId
		}

		// Update category IDs in all parsed transactions that reference this category
		for _, t := range parsedTransactions {
			if isChild {
				// For child categories, match by OriginalCategoryName (subcategory name)
				if t.OriginalCategoryName == newCategory.Name && t.CategoryId == 0 {
					t.CategoryId = newCategory.CategoryId
				}
			} else {
				// For parent categories, don't assign to transactions (children will be assigned instead)
				// But if a transaction has no subcategory and its category name matches, assign it
				if t.OriginalCategoryName == newCategory.Name && t.OriginalParentCategoryName == "" && t.CategoryId == 0 {
					t.CategoryId = newCategory.CategoryId
				}
			}
		}
	}

	// Auto-create missing counterparties from OriginalCounterpartyName field
	existingCounterparties, cpErr := a.counterparties.GetAllCounterpartiesByUid(c, user.Uid)
	if cpErr != nil {
		log.Warnf(c, "[transactions.TransactionParseImportFileHandler] failed to get counterparties for user \"uid:%d\", because %s", user.Uid, cpErr.Error())
	} else {
		counterpartyNameMap := make(map[string]int64)
		for _, cp := range existingCounterparties {
			if !cp.Deleted && !cp.Hidden {
				counterpartyNameMap[cp.Name] = cp.CounterpartyId
			}
		}

		cpMaxOrder, cpOrderErr := a.counterparties.GetMaxDisplayOrder(c, user.Uid)
		if cpOrderErr != nil {
			cpMaxOrder = 0
		}

		for _, t := range parsedTransactions {
			counterpartyName := strings.TrimSpace(t.OriginalCounterpartyName)
			if counterpartyName == "" {
				continue
			}

			// Check if counterparty already exists
			if cpId, exists := counterpartyNameMap[counterpartyName]; exists {
				t.CounterpartyId = cpId
				continue
			}

			// Create new counterparty
			cpMaxOrder++
			newCounterparty := &models.Counterparty{
				Uid:          user.Uid,
				Name:         counterpartyName,
				Type:         models.COUNTERPARTY_TYPE_COMPANY,
				Icon:         0,
				Color:        "588a6a",
				DisplayOrder: cpMaxOrder,
			}

			createErr := a.counterparties.CreateCounterparty(c, newCounterparty)
			if createErr != nil {
				log.Warnf(c, "[transactions.TransactionParseImportFileHandler] failed to auto-create counterparty \"%s\" for user \"uid:%d\", because %s", counterpartyName, user.Uid, createErr.Error())
				continue
			}

			log.Infof(c, "[transactions.TransactionParseImportFileHandler] auto-created counterparty \"%s\" (id:%d) for user \"uid:%d\"", counterpartyName, newCounterparty.CounterpartyId, user.Uid)

			counterpartyNameMap[counterpartyName] = newCounterparty.CounterpartyId
			t.CounterpartyId = newCounterparty.CounterpartyId
		}
	}

	// Auto-create missing tags
	if len(allNewTags) > 0 {
		tagMaxOrder, tagOrderErr := a.transactionTags.GetMaxDisplayOrder(c, user.Uid, 0)
		if tagOrderErr != nil {
			tagMaxOrder = 0
		}

		for _, newTag := range allNewTags {
			newTag.Uid = user.Uid
			tagMaxOrder++
			newTag.DisplayOrder = tagMaxOrder

			createErr := a.transactionTags.CreateTag(c, newTag)
			if createErr != nil {
				log.Warnf(c, "[transactions.TransactionParseImportFileHandler] failed to auto-create tag \"%s\" for user \"uid:%d\", because %s", newTag.Name, user.Uid, createErr.Error())
				continue
			}

			log.Infof(c, "[transactions.TransactionParseImportFileHandler] auto-created tag \"%s\" (id:%d) for user \"uid:%d\"", newTag.Name, newTag.TagId, user.Uid)

			// Update tag IDs in all parsed transactions
			oldTagIdStr := utils.Int64ToString(newTag.TagId)
			for _, t := range parsedTransactions {
				for idx, tid := range t.TagIds {
					if tid == oldTagIdStr || tid == "0" {
						// Find by name match
						for _, origName := range t.OriginalTagNames {
							if origName == newTag.Name {
								t.TagIds[idx] = utils.Int64ToString(newTag.TagId)
								break
							}
						}
					}
				}
			}
		}
	}

	parsedTransactionRespsList := parsedTransactions.ToImportTransactionResponseList()

	if len(parsedTransactionRespsList) < 1 {
		return nil, errs.ErrNoDataToImport
	}

	parsedTransactionResps := &models.ImportTransactionResponsePageWrapper{
		Items:      parsedTransactionRespsList,
		TotalCount: int64(len(parsedTransactionRespsList)),
	}

	return parsedTransactionResps, nil
}

// TransactionImportHandler imports transactions by request parameters for current user
func (a *TransactionsApi) TransactionImportHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionImportReq models.TransactionImportRequest
	err := c.ShouldBindJSON(&transactionImportReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionImportHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[transactions.TransactionImportHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()

	if a.CurrentConfig().EnableDuplicateSubmissionsCheck && transactionImportReq.ClientSessionId != "" {
		found, remark := a.GetSubmissionRemark(duplicatechecker.DUPLICATE_CHECKER_TYPE_IMPORT_TRANSACTIONS, uid, transactionImportReq.ClientSessionId)

		if found {
			items := strings.Split(remark, ":")

			if len(items) >= 2 {
				if items[0] == "finished" {
					log.Infof(c, "[transactions.TransactionImportHandler] another \"%s\" transactions has been imported for user \"uid:%d\"", items[1], uid)
					count, err := utils.StringToInt(items[1])

					if err == nil {
						return count, nil
					}
				} else if items[0] == "processing" {
					return nil, errs.ErrRepeatedRequest
				}
			} else {
				log.Warnf(c, "[transactions.TransactionImportHandler] another transaction import task may be executing, but remark \"%s\" is invalid", remark)
			}
		}
	}

	newTransactionTagIdsMap := make(map[int][]int64, len(transactionImportReq.Transactions))

	for i := 0; i < len(transactionImportReq.Transactions); i++ {
		transactionCreateReq := transactionImportReq.Transactions[i]
		tagIds, err := utils.StringArrayToInt64Array(transactionCreateReq.TagIds)

		if err != nil {
			log.Warnf(c, "[transactions.TransactionImportHandler] parse tag ids failed of transaction \"index:%d\", because %s", i, err.Error())
			return nil, errs.ErrTransactionTagIdInvalid
		}

		if len(tagIds) > models.MaximumTagsCountOfTransaction {
			return nil, errs.ErrTransactionHasTooManyTags
		}

		if transactionCreateReq.Type < models.TRANSACTION_TYPE_MODIFY_BALANCE || transactionCreateReq.Type > models.TRANSACTION_TYPE_TRANSFER {
			log.Warnf(c, "[transactions.TransactionImportHandler] transaction type of transaction \"index:%d\" is invalid", i)
			return nil, errs.ErrTransactionTypeInvalid
		}

		if transactionCreateReq.Type == models.TRANSACTION_TYPE_MODIFY_BALANCE && transactionCreateReq.CategoryId != 0 {
			log.Warnf(c, "[transactions.TransactionImportHandler] balance modification transaction \"index:%d\" cannot set category id", i)
			return nil, errs.ErrBalanceModificationTransactionCannotSetCategory
		}

		if transactionCreateReq.Type != models.TRANSACTION_TYPE_TRANSFER && transactionCreateReq.DestinationAccountId != 0 {
			log.Warnf(c, "[transactions.TransactionImportHandler] non-transfer transaction \"index:%d\" destination account cannot be set", i)
			return nil, errs.ErrTransactionDestinationAccountCannotBeSet
		} else if transactionCreateReq.Type == models.TRANSACTION_TYPE_TRANSFER && transactionCreateReq.SourceAccountId == transactionCreateReq.DestinationAccountId {
			log.Warnf(c, "[transactions.TransactionImportHandler] transfer transaction \"index:%d\" source account must not be destination account", i)
			return nil, errs.ErrTransactionSourceAndDestinationIdCannotBeEqual
		}

		if transactionCreateReq.Type != models.TRANSACTION_TYPE_TRANSFER && transactionCreateReq.DestinationAmount != 0 {
			log.Warnf(c, "[transactions.TransactionImportHandler] non-transfer transaction \"index:%d\" destination amount cannot be set", i)
			return nil, errs.ErrTransactionDestinationAmountCannotBeSet
		}

		newTransactionTagIdsMap[i] = tagIds
	}

	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transactions.TransactionImportHandler] failed to get user, because %s", err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	if user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_IMPORT_TRANSACTION) {
		return nil, errs.ErrNotPermittedToPerformThisAction
	}

	newTransactions := make([]*models.Transaction, len(transactionImportReq.Transactions))
	now := utils.GetMinTransactionTimeFromUnixTime(time.Now().Unix())

	for i := 0; i < len(transactionImportReq.Transactions); i++ {
		transactionCreateReq := transactionImportReq.Transactions[i]
		transaction := a.createNewTransactionModel(uid, transactionCreateReq, c.ClientIP())
		transactionEditable := user.CanEditTransactionByTransactionTime(transaction.TransactionTime, clientTimezone)

		if !transactionEditable {
			return nil, errs.ErrCannotCreateTransactionWithThisTransactionTime
		}

		// Mark future-dated transactions as planned
		if transaction.TransactionTime > now {
			transaction.Planned = true
		}

		newTransactions[i] = transaction
	}

	err = a.transactions.BatchCreateTransactions(c, user.Uid, newTransactions, newTransactionTagIdsMap, func(currentProcess float64) {
		a.SetSubmissionRemarkIfEnable(duplicatechecker.DUPLICATE_CHECKER_TYPE_IMPORT_TRANSACTIONS, uid, transactionImportReq.ClientSessionId, fmt.Sprintf("processing:%.2f", currentProcess))
	})
	count := len(newTransactions)

	if err != nil {
		a.RemoveSubmissionRemarkIfEnable(duplicatechecker.DUPLICATE_CHECKER_TYPE_IMPORT_TRANSACTIONS, uid, transactionImportReq.ClientSessionId)
		log.Errorf(c, "[transactions.TransactionImportHandler] failed to import %d transactions for user \"uid:%d\", because %s", count, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transactions.TransactionImportHandler] user \"uid:%d\" has imported %d transactions successfully", uid, count)

	a.SetSubmissionRemarkIfEnable(duplicatechecker.DUPLICATE_CHECKER_TYPE_IMPORT_TRANSACTIONS, uid, transactionImportReq.ClientSessionId, fmt.Sprintf("finished:%d", count))

	return count, nil
}

// TransactionImportProcessHandler returns the process of specified transaction import task by request parameters for current user
func (a *TransactionsApi) TransactionImportProcessHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionImportProcessReq models.TransactionImportProcessRequest
	err := c.ShouldBindQuery(&transactionImportProcessReq)

	if err != nil {
		log.Warnf(c, "[transactions.TransactionImportProcessHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	if !a.CurrentConfig().EnableDuplicateSubmissionsCheck {
		return nil, nil
	}

	found, remark := a.GetSubmissionRemark(duplicatechecker.DUPLICATE_CHECKER_TYPE_IMPORT_TRANSACTIONS, uid, transactionImportProcessReq.ClientSessionId)

	if !found {
		return nil, nil
	}

	items := strings.Split(remark, ":")

	if len(items) < 2 {
		return nil, nil
	}

	if items[0] == "finished" {
		return 100, nil
	} else if items[0] != "processing" {
		return nil, nil
	}

	process, err := utils.StringToFloat64(items[1])

	if err != nil {
		log.Warnf(c, "[transactions.TransactionImportProcessHandler] parse process failed, because %s", err.Error())
		return nil, nil
	}

	if process < 0 {
		return nil, nil
	} else if process >= 100 {
		process = 100
	}

	return process, nil
}
