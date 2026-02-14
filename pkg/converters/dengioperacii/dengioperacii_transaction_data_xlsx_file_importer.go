package dengioperacii

import (
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

// dengioperaciiTransactionDataXlsxFileImporter defines the structure of the xlsx importer for "Деньги-операции" format
type dengioperaciiTransactionDataXlsxFileImporter struct {
}

// Initialize singleton instance
var (
	DengioperaciiTransactionDataXlsxFileImporter = &dengioperaciiTransactionDataXlsxFileImporter{}
)

// transferRowData holds the parsed row data along with the original sign of the amount
type transferRowData struct {
	data       map[datatable.TransactionDataTableColumn]string
	isNegative bool // true = debit (source), false = credit (destination)
}

// ParseImportedData returns the imported data by parsing the "Деньги-операции" xlsx data
func (c *dengioperaciiTransactionDataXlsxFileImporter) ParseImportedData(ctx core.Context, user *models.User, data []byte, defaultTimezone *time.Location, additionalOptions converter.TransactionDataImporterOptions, accountMap map[string]*models.Account, expenseCategoryMap map[string]*models.TransactionCategory, incomeCategoryMap map[string]*models.TransactionCategory, transferCategoryMap map[string]*models.TransactionCategory, tagMap map[string]*models.TransactionTag) (models.ImportedTransactionSlice, []*models.Account, []*models.TransactionCategory, []*models.TransactionCategory, []*models.TransactionCategory, []*models.TransactionTag, error) {
	xlsxDataTable, err := excel.CreateNewExcelOOXMLFileBasicDataTable(data, true)

	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	commonDataTable := datatable.CreateNewCommonDataTableFromBasicDataTable(xlsxDataTable)

	if !commonDataTable.HasColumn(dengioperaciiTransactionDateColumnName) ||
		!commonDataTable.HasColumn(dengioperaciiTransactionAmountColumnName) ||
		!commonDataTable.HasColumn(dengioperaciiTransactionAccountColumnName) ||
		!commonDataTable.HasColumn(dengioperaciiTransactionCategoryColumnName) {
		log.Errorf(ctx, "[dengioperacii_transaction_data_xlsx_file_importer.ParseImportedData] cannot parse data, because missing essential columns in header row")
		return nil, nil, nil, nil, nil, nil, errs.ErrMissingRequiredFieldInHeaderRow
	}

	// Parse all rows using the row parser, keeping track of sign for transfers
	transactionRowParser := createDengioperaciiTransactionDataRowParserInternal(xlsxDataTable.HeaderColumnNames())

	// Collect all parsed rows
	type parsedRow struct {
		data       map[datatable.TransactionDataTableColumn]string
		isNegative bool
		isTransfer bool
	}

	var allRows []parsedRow
	rowIterator := commonDataTable.DataRowIterator()

	for rowIterator.HasNext() {
		commonRow := rowIterator.Next()
		if commonRow == nil {
			continue
		}

		rowId := rowIterator.CurrentRowId()
		rowData, rowDataValid, err := transactionRowParser.ParseWithSign(ctx, user, commonRow, rowId)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, err
		}
		if !rowDataValid || rowData == nil {
			continue
		}

		isTransfer := rowData[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE] == dengioperaciiTransactionTypeTransferName
		isNeg := transactionRowParser.lastIsNegative

		allRows = append(allRows, parsedRow{
			data:       rowData,
			isNegative: isNeg,
			isTransfer: isTransfer,
		})
	}

	// Build a WritableTransactionDataTable with merged transfers
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

	// Separate transfer rows and non-transfer rows
	var transferRows []parsedRow
	var nonTransferRows []parsedRow

	for _, row := range allRows {
		if row.isTransfer {
			transferRows = append(transferRows, row)
		} else {
			nonTransferRows = append(nonTransferRows, row)
		}
	}

	// Match transfer pairs: same date + same category (Конвертация валют / Перевод между счетами)
	// Negative amount = debit (source account), Positive amount = credit (destination account)
	matched := make([]bool, len(transferRows))

	for i := 0; i < len(transferRows); i++ {
		if matched[i] {
			continue
		}

		rowI := transferRows[i]
		dateI := rowI.data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME]
		catI := rowI.data[datatable.TRANSACTION_DATA_TABLE_CATEGORY] // "Статья" stored in CATEGORY

		// Find a matching partner
		bestMatch := -1
		for j := i + 1; j < len(transferRows); j++ {
			if matched[j] {
				continue
			}

			rowJ := transferRows[j]
			dateJ := rowJ.data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME]
			catJ := rowJ.data[datatable.TRANSACTION_DATA_TABLE_CATEGORY]

			// Same date and same category, one negative and one positive
			if dateI == dateJ && catI == catJ && rowI.isNegative != rowJ.isNegative {
				bestMatch = j
				break
			}
		}

		if bestMatch >= 0 {
			matched[i] = true
			matched[bestMatch] = true

			var sourceRow, destRow parsedRow
			if rowI.isNegative {
				sourceRow = rowI
				destRow = transferRows[bestMatch]
			} else {
				sourceRow = transferRows[bestMatch]
				destRow = rowI
			}

			// Create merged transfer row
			mergedRow := make(map[datatable.TransactionDataTableColumn]string)
			mergedRow[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME] = sourceRow.data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME]
			mergedRow[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE] = dengioperaciiTransactionTypeTransferName
			mergedRow[datatable.TRANSACTION_DATA_TABLE_CATEGORY] = sourceRow.data[datatable.TRANSACTION_DATA_TABLE_CATEGORY]
			mergedRow[datatable.TRANSACTION_DATA_TABLE_SUB_CATEGORY] = sourceRow.data[datatable.TRANSACTION_DATA_TABLE_SUB_CATEGORY]

			// Source account info
			mergedRow[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME] = sourceRow.data[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME]
			mergedRow[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY] = sourceRow.data[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY]
			mergedRow[datatable.TRANSACTION_DATA_TABLE_AMOUNT] = sourceRow.data[datatable.TRANSACTION_DATA_TABLE_AMOUNT]

			// Destination account info
			mergedRow[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME] = destRow.data[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME]
			mergedRow[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_CURRENCY] = destRow.data[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY]
			mergedRow[datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT] = destRow.data[datatable.TRANSACTION_DATA_TABLE_AMOUNT]

			// Combine descriptions if both have them
			srcDesc := sourceRow.data[datatable.TRANSACTION_DATA_TABLE_DESCRIPTION]
			dstDesc := destRow.data[datatable.TRANSACTION_DATA_TABLE_DESCRIPTION]
			if srcDesc != "" && dstDesc != "" && srcDesc != dstDesc {
				mergedRow[datatable.TRANSACTION_DATA_TABLE_DESCRIPTION] = srcDesc + " → " + dstDesc
			} else if srcDesc != "" {
				mergedRow[datatable.TRANSACTION_DATA_TABLE_DESCRIPTION] = srcDesc
			} else {
				mergedRow[datatable.TRANSACTION_DATA_TABLE_DESCRIPTION] = dstDesc
			}

			// Tag Group and Tags from source (if any)
			mergedRow[datatable.TRANSACTION_DATA_TABLE_TAG_GROUP] = sourceRow.data[datatable.TRANSACTION_DATA_TABLE_TAG_GROUP]
			mergedRow[datatable.TRANSACTION_DATA_TABLE_TAGS] = sourceRow.data[datatable.TRANSACTION_DATA_TABLE_TAGS]

			// Counterparty (payee) from source
			mergedRow[datatable.TRANSACTION_DATA_TABLE_PAYEE] = sourceRow.data[datatable.TRANSACTION_DATA_TABLE_PAYEE]

			mergedDataTable.Add(mergedRow)
		} else {
			// Unmatched transfer row — keep as-is (will be treated as standalone transfer)
			// Set RELATED_ACCOUNT_NAME to empty to avoid using category name as account
			rowI.data[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME] = ""
			mergedDataTable.Add(rowI.data)
		}
	}

	// Add non-transfer rows as-is
	for _, row := range nonTransferRows {
		mergedDataTable.Add(row.data)
	}

	// Use the standard importer with the merged data table
	dataTableImporter := converter.CreateNewImporterWithTypeNameMapping(dengioperaciiTransactionTypeNameMapping, "", converter.TRANSACTION_GEO_LOCATION_ORDER_LONGITUDE_LATITUDE, dengioperaciiTagSeparator)

	return dataTableImporter.ParseImportedData(ctx, user, mergedDataTable, defaultTimezone, additionalOptions, accountMap, expenseCategoryMap, incomeCategoryMap, transferCategoryMap, tagMap)
}

// buildTransferKey returns a key for matching transfer pairs
func buildTransferKey(dateStr, categoryName string) string {
	return strings.TrimSpace(dateStr) + "|" + strings.TrimSpace(categoryName)
}
