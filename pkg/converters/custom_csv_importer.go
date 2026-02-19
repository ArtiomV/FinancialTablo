package converters

import (
	"encoding/csv"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/converters/converter"
	"github.com/mayswind/ezbookkeeping/pkg/converters/datatable"
	"github.com/mayswind/ezbookkeeping/pkg/converters/excel"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// Custom column names (Russian headers)
const (
	csvColDate            = "Дата"
	csvColAmount          = "Сумма"
	csvColAccount         = "Счет"
	csvColCurrency        = "Валюта"
	csvColCounterparty    = "Контрагент"
	csvColCounterpartyINN = "ИНН контрагент"
	csvColCategory        = "Статья"
	csvColParentCategory  = "Род. статья"
	csvColDescription     = "Описание"
)

// Fixed columns count (before dynamic tag group columns)
const csvFixedColumnCount = 9

var customCSVTagSeparator = ";"

// CustomCSVTransactionDataImporter implements the TransactionDataImporter interface
// for the user's specific CSV/XLSX format
type CustomCSVTransactionDataImporter struct{}

// CustomCSVImporter is the singleton instance
var CustomCSVImporter = &CustomCSVTransactionDataImporter{}

// csvRowTagInfo stores tag group info for a single data table row
type csvRowTagInfo struct {
	tagGroups map[string]string // tagGroupName → tagValue
}

// csvParsedRow holds a parsed row
type csvParsedRow struct {
	dateTime           string // "yyyy-mm-dd 00:00:00"
	transactionType    string // "Доход", "Расход", "Перевод"
	amount             int64  // amount in cents (always positive)
	isNegative         bool
	accountName        string
	currency           string
	counterpartyName   string
	categoryName       string // Статья
	parentCategoryName string // Род. статья
	description        string
	isTransfer         bool
	tagGroups          map[string]string // tagGroupName → tagValue
}

// isXlsxFile checks if data starts with PK (ZIP magic bytes = xlsx)
func isXlsxFile(data []byte) bool {
	return len(data) >= 4 && data[0] == 0x50 && data[1] == 0x4b && data[2] == 0x03 && data[3] == 0x04
}

// ParseImportedData parses CSV or XLSX format and returns imported transactions
func (c *CustomCSVTransactionDataImporter) ParseImportedData(ctx core.Context, user *models.User, data []byte, defaultTimezone *time.Location, additionalOptions converter.TransactionDataImporterOptions, accountMap map[string]*models.Account, expenseCategoryMap map[string]*models.TransactionCategory, incomeCategoryMap map[string]*models.TransactionCategory, transferCategoryMap map[string]*models.TransactionCategory, tagMap map[string]*models.TransactionTag) (models.ImportedTransactionSlice, []*models.Account, []*models.TransactionCategory, []*models.TransactionCategory, []*models.TransactionCategory, []*models.TransactionTag, error) {

	var allRecords [][]string
	var err error

	if isXlsxFile(data) {
		allRecords, err = c.parseXlsxData(data)
	} else {
		allRecords, err = c.parseCsvData(data)
	}

	if err != nil {
		log.Errorf(ctx, "[custom_csv_importer.ParseImportedData] failed to parse file for user \"uid:%d\", because %s", user.Uid, err.Error())
		return nil, nil, nil, nil, nil, nil, errs.ErrNotFoundTransactionDataInFile
	}

	if len(allRecords) < 2 {
		return nil, nil, nil, nil, nil, nil, errs.ErrNotFoundTransactionDataInFile
	}

	// Parse header row
	headerRow := allRecords[0]
	if len(headerRow) < csvFixedColumnCount {
		log.Errorf(ctx, "[custom_csv_importer.ParseImportedData] header row has only %d columns, expected at least %d", len(headerRow), csvFixedColumnCount)
		return nil, nil, nil, nil, nil, nil, errs.ErrMissingRequiredFieldInHeaderRow
	}

	// Find column indices and tag group columns
	colIndices := make(map[string]int)
	tagGroupColumns := make([]string, 0)

	for i, header := range headerRow {
		trimmed := strings.TrimSpace(header)
		colIndices[trimmed] = i
		if i >= csvFixedColumnCount {
			tagGroupColumns = append(tagGroupColumns, trimmed)
		}
	}

	// Parse all data rows
	allParsedRows := make([]csvParsedRow, 0, len(allRecords)-1)

	for rowIdx := 1; rowIdx < len(allRecords); rowIdx++ {
		record := allRecords[rowIdx]
		if len(record) < csvFixedColumnCount {
			continue
		}

		dateStr := strings.TrimSpace(c.getCell(record, colIndices, csvColDate))
		amountStr := strings.TrimSpace(c.getCell(record, colIndices, csvColAmount))
		accountName := strings.TrimSpace(c.getCell(record, colIndices, csvColAccount))
		currency := strings.TrimSpace(c.getCell(record, colIndices, csvColCurrency))
		counterpartyName := strings.TrimSpace(c.getCell(record, colIndices, csvColCounterparty))
		categoryName := strings.TrimSpace(c.getCell(record, colIndices, csvColCategory))
		parentCategoryName := strings.TrimSpace(c.getCell(record, colIndices, csvColParentCategory))
		description := strings.TrimSpace(c.getCell(record, colIndices, csvColDescription))

		if dateStr == "" || amountStr == "" || accountName == "" {
			continue
		}

		dateTime, err := c.parseDate(dateStr)
		if err != nil {
			log.Warnf(ctx, "[custom_csv_importer.ParseImportedData] cannot parse date \"%s\" in row %d: %s", dateStr, rowIdx+1, err.Error())
			continue
		}

		amount, isNegative, err := c.parseAmount(amountStr)
		if err != nil {
			log.Warnf(ctx, "[custom_csv_importer.ParseImportedData] cannot parse amount \"%s\" in row %d: %s", amountStr, rowIdx+1, err.Error())
			continue
		}

		transactionType := "Доход"
		if isNegative {
			transactionType = "Расход"
		}

		isTransfer := categoryName == "Конвертация валют" || categoryName == "Перевод между счетами"
		if isTransfer {
			transactionType = "Перевод"
		}

		// Collect tag group values
		tagGroups := make(map[string]string)
		for i, tgName := range tagGroupColumns {
			colIdx := csvFixedColumnCount + i
			if colIdx < len(record) {
				val := strings.TrimSpace(record[colIdx])
				if val != "" {
					tagGroups[tgName] = val
				}
			}
		}

		allParsedRows = append(allParsedRows, csvParsedRow{
			dateTime:           dateTime,
			transactionType:    transactionType,
			amount:             amount,
			isNegative:         isNegative,
			accountName:        accountName,
			currency:           currency,
			counterpartyName:   counterpartyName,
			categoryName:       categoryName,
			parentCategoryName: parentCategoryName,
			description:        description,
			isTransfer:         isTransfer,
			tagGroups:          tagGroups,
		})
	}

	// Now build data table rows, handling transfers as pairs
	columns := []datatable.TransactionDataTableColumn{
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME,
		datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE,
		datatable.TRANSACTION_DATA_TABLE_CATEGORY,
		datatable.TRANSACTION_DATA_TABLE_SUB_CATEGORY,
		datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME,
		datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY,
		datatable.TRANSACTION_DATA_TABLE_AMOUNT,
		datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME,
		datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_CURRENCY,
		datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT,
		datatable.TRANSACTION_DATA_TABLE_DESCRIPTION,
		datatable.TRANSACTION_DATA_TABLE_TAGS,
		datatable.TRANSACTION_DATA_TABLE_TAG_GROUP,
		datatable.TRANSACTION_DATA_TABLE_PAYEE,
	}
	mergedDataTable := datatable.CreateNewWritableTransactionDataTable(columns)

	// Separate transfers from non-transfers
	var transferRows []csvParsedRow
	var nonTransferRows []csvParsedRow
	var dataTableRowTagInfos []csvRowTagInfo

	for _, row := range allParsedRows {
		if row.isTransfer {
			transferRows = append(transferRows, row)
		} else {
			nonTransferRows = append(nonTransferRows, row)
		}
	}

	// Match transfer pairs
	matched := make([]bool, len(transferRows))
	for i := 0; i < len(transferRows); i++ {
		if matched[i] {
			continue
		}
		rowI := transferRows[i]

		bestMatch := -1
		for j := i + 1; j < len(transferRows); j++ {
			if matched[j] {
				continue
			}
			rowJ := transferRows[j]
			if rowI.dateTime == rowJ.dateTime && rowI.categoryName == rowJ.categoryName && rowI.isNegative != rowJ.isNegative {
				bestMatch = j
				break
			}
		}

		if bestMatch >= 0 {
			matched[i] = true
			matched[bestMatch] = true

			var sourceRow, destRow csvParsedRow
			if rowI.isNegative {
				sourceRow = rowI
				destRow = transferRows[bestMatch]
			} else {
				sourceRow = transferRows[bestMatch]
				destRow = rowI
			}

			rowMap := c.buildBaseRowMap(sourceRow)
			rowMap[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME] = destRow.accountName
			rowMap[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_CURRENCY] = destRow.currency
			rowMap[datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT] = fmt.Sprintf("%d", destRow.amount)

			srcDesc := sourceRow.description
			dstDesc := destRow.description
			if srcDesc != "" && dstDesc != "" && srcDesc != dstDesc {
				rowMap[datatable.TRANSACTION_DATA_TABLE_DESCRIPTION] = srcDesc + " → " + dstDesc
			} else if srcDesc != "" {
				rowMap[datatable.TRANSACTION_DATA_TABLE_DESCRIPTION] = srcDesc
			} else {
				rowMap[datatable.TRANSACTION_DATA_TABLE_DESCRIPTION] = dstDesc
			}

			mergedDataTable.Add(rowMap)
			dataTableRowTagInfos = append(dataTableRowTagInfos, csvRowTagInfo{tagGroups: sourceRow.tagGroups})
		} else {
			// Unmatched transfer — keep as transfer
			rowMap := c.buildBaseRowMap(rowI)
			rowMap[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME] = ""
			mergedDataTable.Add(rowMap)
			dataTableRowTagInfos = append(dataTableRowTagInfos, csvRowTagInfo{tagGroups: rowI.tagGroups})
		}
	}

	// Add non-transfer rows
	for _, row := range nonTransferRows {
		rowMap := c.buildBaseRowMap(row)
		mergedDataTable.Add(rowMap)
		dataTableRowTagInfos = append(dataTableRowTagInfos, csvRowTagInfo{tagGroups: row.tagGroups})
	}

	// Use standard importer
	fullTypeMapping := map[models.TransactionType]string{
		models.TRANSACTION_TYPE_INCOME:   "Доход",
		models.TRANSACTION_TYPE_EXPENSE:  "Расход",
		models.TRANSACTION_TYPE_TRANSFER: "Перевод",
	}

	dataTableImporter := converter.CreateNewImporterWithTypeNameMapping(fullTypeMapping, "", converter.TRANSACTION_GEO_LOCATION_ORDER_LONGITUDE_LATITUDE, customCSVTagSeparator)

	transactions, newAccounts, newSubExpenseCategories, newSubIncomeCategories, newSubTransferCategories, newTags, parseErr := dataTableImporter.ParseImportedData(ctx, user, mergedDataTable, defaultTimezone, additionalOptions, accountMap, expenseCategoryMap, incomeCategoryMap, transferCategoryMap, tagMap)

	if parseErr != nil {
		return nil, nil, nil, nil, nil, nil, parseErr
	}

	c.fixTagGroupAssignments(newTags, dataTableRowTagInfos)

	return transactions, newAccounts, newSubExpenseCategories, newSubIncomeCategories, newSubTransferCategories, newTags, nil
}

// parseCsvData reads CSV bytes into [][]string records
func (c *CustomCSVTransactionDataImporter) parseCsvData(data []byte) ([][]string, error) {
	reader := csv.NewReader(strings.NewReader(string(data)))
	reader.LazyQuotes = true
	return reader.ReadAll()
}

// parseXlsxData reads XLSX bytes into [][]string records (same shape as CSV)
func (c *CustomCSVTransactionDataImporter) parseXlsxData(data []byte) ([][]string, error) {
	xlsxDataTable, err := excel.CreateNewExcelOOXMLFileBasicDataTable(data, true)
	if err != nil {
		return nil, err
	}

	headerNames := xlsxDataTable.HeaderColumnNames()
	if len(headerNames) < 1 {
		return nil, fmt.Errorf("no header row found in xlsx")
	}

	var allRecords [][]string
	allRecords = append(allRecords, headerNames)

	rowIterator := xlsxDataTable.DataRowIterator()
	for rowIterator.HasNext() {
		basicRow := rowIterator.Next()
		if basicRow == nil {
			continue
		}

		row := make([]string, len(headerNames))
		for i := 0; i < len(headerNames); i++ {
			row[i] = basicRow.GetData(i)
		}
		allRecords = append(allRecords, row)
	}

	return allRecords, nil
}

// buildBaseRowMap creates a data table row map from a parsed row
func (c *CustomCSVTransactionDataImporter) buildBaseRowMap(row csvParsedRow) map[datatable.TransactionDataTableColumn]string {
	rowMap := make(map[datatable.TransactionDataTableColumn]string)
	rowMap[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME] = row.dateTime
	rowMap[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE] = row.transactionType
	rowMap[datatable.TRANSACTION_DATA_TABLE_AMOUNT] = fmt.Sprintf("%d", row.amount)

	if row.parentCategoryName != "" {
		rowMap[datatable.TRANSACTION_DATA_TABLE_CATEGORY] = row.parentCategoryName
		rowMap[datatable.TRANSACTION_DATA_TABLE_SUB_CATEGORY] = row.categoryName
	} else {
		rowMap[datatable.TRANSACTION_DATA_TABLE_CATEGORY] = ""
		rowMap[datatable.TRANSACTION_DATA_TABLE_SUB_CATEGORY] = row.categoryName
	}

	rowMap[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME] = row.accountName
	rowMap[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY] = row.currency
	rowMap[datatable.TRANSACTION_DATA_TABLE_DESCRIPTION] = row.description
	rowMap[datatable.TRANSACTION_DATA_TABLE_PAYEE] = row.counterpartyName

	rowMap[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME] = ""
	rowMap[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_CURRENCY] = ""
	rowMap[datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT] = ""

	var tagValues []string
	firstGroupName := ""
	for groupName, tagValue := range row.tagGroups {
		tagValues = append(tagValues, tagValue)
		if firstGroupName == "" {
			firstGroupName = groupName
		}
	}

	rowMap[datatable.TRANSACTION_DATA_TABLE_TAGS] = strings.Join(tagValues, customCSVTagSeparator)
	rowMap[datatable.TRANSACTION_DATA_TABLE_TAG_GROUP] = firstGroupName

	return rowMap
}

// fixTagGroupAssignments corrects ImportTagGroupName for tags
func (c *CustomCSVTransactionDataImporter) fixTagGroupAssignments(newTags []*models.TransactionTag, dataTableRowTagInfos []csvRowTagInfo) {
	tagValueToGroup := make(map[string]string)

	for _, info := range dataTableRowTagInfos {
		for groupName, tagValue := range info.tagGroups {
			if tagValue != "" && groupName != "" {
				tagValueToGroup[tagValue] = groupName
			}
		}
	}

	for _, tag := range newTags {
		if correctGroup, ok := tagValueToGroup[tag.Name]; ok {
			tag.ImportTagGroupName = correctGroup
		}
	}
}

// getCell safely gets a cell value from a record by column name
func (c *CustomCSVTransactionDataImporter) getCell(record []string, colIndices map[string]int, colName string) string {
	idx, ok := colIndices[colName]
	if !ok || idx >= len(record) {
		return ""
	}
	return record[idx]
}

// parseDate parses date strings in various formats and returns "yyyy-mm-dd 00:00:00"
func (c *CustomCSVTransactionDataImporter) parseDate(dateStr string) (string, error) {
	s := strings.TrimSpace(dateStr)
	if s == "" {
		return "", fmt.Errorf("invalid date format: %s", dateStr)
	}

	var day, month, year string

	if strings.Contains(s, ".") {
		parts := strings.Split(s, ".")
		if len(parts) != 3 {
			return "", fmt.Errorf("invalid date format: %s", dateStr)
		}
		day = strings.TrimSpace(parts[0])
		month = strings.TrimSpace(parts[1])
		year = strings.TrimSpace(parts[2])
	} else if strings.Contains(s, "/") {
		parts := strings.Split(s, "/")
		if len(parts) != 3 {
			return "", fmt.Errorf("invalid date format: %s", dateStr)
		}
		month = strings.TrimSpace(parts[0])
		day = strings.TrimSpace(parts[1])
		year = strings.TrimSpace(parts[2])
	} else if strings.Contains(s, "-") {
		parts := strings.Split(s, "-")
		if len(parts) != 3 {
			return "", fmt.Errorf("invalid date format: %s", dateStr)
		}
		first := strings.TrimSpace(parts[0])
		second := strings.TrimSpace(parts[1])
		third := strings.TrimSpace(parts[2])
		if len(first) == 4 {
			year = first
			month = second
			day = third
		} else {
			month = first
			day = second
			year = third
		}
	} else {
		return "", fmt.Errorf("invalid date format: %s", dateStr)
	}

	if len(year) == 2 {
		yearNum, err := strconv.Atoi(year)
		if err != nil {
			return "", fmt.Errorf("invalid date format: %s", dateStr)
		}
		if yearNum >= 70 {
			year = fmt.Sprintf("19%02s", year)
		} else {
			year = fmt.Sprintf("20%02s", year)
		}
	}

	return fmt.Sprintf("%s-%s-%s 00:00:00", padLeft(year, 4), padLeft(month, 2), padLeft(day, 2)), nil
}

func padLeft(s string, length int) string {
	for len(s) < length {
		s = "0" + s
	}
	return s
}

// parseAmount parses amount strings in various formats
func (c *CustomCSVTransactionDataImporter) parseAmount(amountStr string) (int64, bool, error) {
	amountStr = strings.TrimSpace(amountStr)
	if amountStr == "" {
		return 0, false, fmt.Errorf("empty amount")
	}

	isNegative := false

	if strings.HasPrefix(amountStr, "(") && strings.HasSuffix(amountStr, ")") {
		inner := amountStr[1 : len(amountStr)-1]
		if strings.HasPrefix(inner, "-") {
			isNegative = true
			inner = inner[1:]
		}
		amountStr = inner
	} else if strings.HasPrefix(amountStr, "-") {
		isNegative = true
		amountStr = amountStr[1:]
	}

	amountStr = strings.TrimSpace(amountStr)
	amountStr = strings.ReplaceAll(amountStr, " ", "")

	value, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, false, fmt.Errorf("cannot parse amount: %s", amountStr)
	}

	cents := int64(math.Round(value * 100))

	return cents, isNegative, nil
}
