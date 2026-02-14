package dengioperacii

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/mayswind/ezbookkeeping/pkg/converters/datatable"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// Excel column names
const dengioperaciiTransactionDateColumnName = "Дата"
const dengioperaciiTransactionAmountColumnName = "Сумма"
const dengioperaciiTransactionAccountColumnName = "Счет"
const dengioperaciiTransactionCurrencyColumnName = "Валюта"
const dengioperaciiTransactionCounterpartyColumnName = "Контрагент"
const dengioperaciiTransactionCategoryColumnName = "Статья"
const dengioperaciiTransactionDescriptionColumnName = "Описание"
const dengioperaciiTransactionDirectionColumnName = "Направление"
const dengioperaciiTransactionSubDirectionColumnName = "Субнаправление"
const dengioperaciiTransactionParentCategoryColumnName = "Род. статья"

// Transaction type name constants used in the internal mapping
const dengioperaciiTransactionTypeIncomeName = "Доход"
const dengioperaciiTransactionTypeExpenseName = "Расход"
const dengioperaciiTransactionTypeTransferName = "Перевод"

// Categories that represent transfers (not income/expense)
var dengioperaciiTransferCategories = map[string]bool{
	"Конвертация валют":     true,
	"Перевод между счетами": true,
}

// Tag separator used to join multiple tags in the TAGS column
const dengioperaciiTagSeparator = "||"

// Supported columns that the transaction data table will expose
var dengioperaciiTransactionSupportedColumns = map[datatable.TransactionDataTableColumn]bool{
	datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME:         true,
	datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE:         true,
	datatable.TRANSACTION_DATA_TABLE_CATEGORY:                 true,
	datatable.TRANSACTION_DATA_TABLE_SUB_CATEGORY:             true,
	datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME:             true,
	datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY:         true,
	datatable.TRANSACTION_DATA_TABLE_AMOUNT:                   true,
	datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME:     true,
	datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_CURRENCY: true,
	datatable.TRANSACTION_DATA_TABLE_RELATED_AMOUNT:           true,
	datatable.TRANSACTION_DATA_TABLE_DESCRIPTION:              true,
	datatable.TRANSACTION_DATA_TABLE_TAGS:                     true,
	datatable.TRANSACTION_DATA_TABLE_TAG_GROUP:                true,
	datatable.TRANSACTION_DATA_TABLE_PAYEE:                    true,
}

// Transaction type name mapping for DataTableTransactionDataImporter
var dengioperaciiTransactionTypeNameMapping = map[models.TransactionType]string{
	models.TRANSACTION_TYPE_INCOME:   dengioperaciiTransactionTypeIncomeName,
	models.TRANSACTION_TYPE_EXPENSE:  dengioperaciiTransactionTypeExpenseName,
	models.TRANSACTION_TYPE_TRANSFER: dengioperaciiTransactionTypeTransferName,
}

// dengioperaciiTransactionDataRowParser defines the structure of the row parser
type dengioperaciiTransactionDataRowParser struct {
	existedOriginalDataColumns map[string]bool
	lastIsNegative             bool // tracks the sign of the last parsed amount (used for transfer merging)
}

// ParseWithSign is the same as Parse but used internally to track the amount sign for transfer merging
func (p *dengioperaciiTransactionDataRowParser) ParseWithSign(ctx core.Context, user *models.User, dataRow datatable.CommonDataTableRow, rowId string) (rowData map[datatable.TransactionDataTableColumn]string, rowDataValid bool, err error) {
	return p.Parse(ctx, user, dataRow, rowId)
}

// Parse returns the converted transaction data row
func (p *dengioperaciiTransactionDataRowParser) Parse(ctx core.Context, user *models.User, dataRow datatable.CommonDataTableRow, rowId string) (rowData map[datatable.TransactionDataTableColumn]string, rowDataValid bool, err error) {
	data := make(map[datatable.TransactionDataTableColumn]string, len(dengioperaciiTransactionSupportedColumns))
	p.lastIsNegative = false

	// 1. Parse date from DD.MM.YYYY to YYYY-MM-DD HH:MM:SS format
	if p.hasOriginalColumn(dengioperaciiTransactionDateColumnName) {
		dateStr := dataRow.GetData(dengioperaciiTransactionDateColumnName)
		if dateStr == "" {
			return nil, false, nil // skip empty rows
		}
		convertedDate, err := convertDateFormat(dateStr)
		if err != nil {
			return nil, false, err
		}
		data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME] = convertedDate
	}

	// 2. Read "Статья" (article) — used for both Category and transfer detection
	categoryName := ""
	if p.hasOriginalColumn(dengioperaciiTransactionCategoryColumnName) {
		categoryName = strings.TrimSpace(dataRow.GetData(dengioperaciiTransactionCategoryColumnName))
	}
	data[datatable.TRANSACTION_DATA_TABLE_CATEGORY] = categoryName
	data[datatable.TRANSACTION_DATA_TABLE_SUB_CATEGORY] = ""

	// 3. Parse amount (always positive for ezbookkeeping)
	if p.hasOriginalColumn(dengioperaciiTransactionAmountColumnName) {
		amountStr := dataRow.GetData(dengioperaciiTransactionAmountColumnName)
		parsedAmount, isNegative, err := parseAmount(amountStr)
		if err != nil {
			return nil, false, err
		}
		data[datatable.TRANSACTION_DATA_TABLE_AMOUNT] = parsedAmount
		p.lastIsNegative = isNegative

		// Determine transaction type from transfer categories or amount sign
		if dengioperaciiTransferCategories[categoryName] {
			data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE] = dengioperaciiTransactionTypeTransferName
			data[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME] = categoryName
		} else if isNegative {
			data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE] = dengioperaciiTransactionTypeExpenseName
			data[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME] = ""
		} else {
			data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE] = dengioperaciiTransactionTypeIncomeName
			data[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME] = ""
		}
	}

	// 4. "Направление" → Tag Group, "Субнаправление" → Tags
	direction := ""
	if p.hasOriginalColumn(dengioperaciiTransactionDirectionColumnName) {
		direction = strings.TrimSpace(dataRow.GetData(dengioperaciiTransactionDirectionColumnName))
		direction = strings.TrimPrefix(direction, "- ")
		direction = strings.TrimPrefix(direction, "-")
		direction = strings.TrimSpace(direction)
	}
	data[datatable.TRANSACTION_DATA_TABLE_TAG_GROUP] = direction

	subDirection := ""
	if p.hasOriginalColumn(dengioperaciiTransactionSubDirectionColumnName) {
		subDirection = strings.TrimSpace(dataRow.GetData(dengioperaciiTransactionSubDirectionColumnName))
		subDirection = strings.TrimPrefix(subDirection, "- ")
		subDirection = strings.TrimPrefix(subDirection, "-")
		subDirection = strings.TrimSpace(subDirection)
	}
	data[datatable.TRANSACTION_DATA_TABLE_TAGS] = subDirection

	// 5. Account
	if p.hasOriginalColumn(dengioperaciiTransactionAccountColumnName) {
		data[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME] = dataRow.GetData(dengioperaciiTransactionAccountColumnName)
	}

	// 6. Currency
	if p.hasOriginalColumn(dengioperaciiTransactionCurrencyColumnName) {
		data[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY] = dataRow.GetData(dengioperaciiTransactionCurrencyColumnName)
	}

	// 7. Description (only from description column, counterparty is separate)
	description := ""
	if p.hasOriginalColumn(dengioperaciiTransactionDescriptionColumnName) {
		description = dataRow.GetData(dengioperaciiTransactionDescriptionColumnName)
	}
	data[datatable.TRANSACTION_DATA_TABLE_DESCRIPTION] = description

	// 8. Counterparty (stored in PAYEE column for the importer to handle)
	if p.hasOriginalColumn(dengioperaciiTransactionCounterpartyColumnName) {
		counterparty := strings.TrimSpace(dataRow.GetData(dengioperaciiTransactionCounterpartyColumnName))
		if counterparty != "" {
			data[datatable.TRANSACTION_DATA_TABLE_PAYEE] = counterparty
		}
	}

	return data, true, nil
}

func (p *dengioperaciiTransactionDataRowParser) hasOriginalColumn(columnName string) bool {
	_, exists := p.existedOriginalDataColumns[columnName]
	return exists
}

// convertDateFormat converts various date formats to YYYY-MM-DD HH:MM:SS
// Supported formats:
//   - DD.MM.YYYY (Russian: 09.02.2026)
//   - DD.MM.YY   (Russian short: 09.02.26)
//   - MM-DD-YY   (Excel US short: 02-12-26)
//   - MM/DD/YY   (Excel US: 02/12/26)
//   - YYYY-MM-DD (ISO: 2026-02-09)
//   - MM/DD/YYYY (US full: 02/09/2026)
func convertDateFormat(dateStr string) (string, error) {
	s := strings.TrimSpace(dateStr)
	if s == "" {
		return "", fmt.Errorf("invalid date format: %s", dateStr)
	}

	var day, month, year string

	if strings.Contains(s, ".") {
		// Dot-separated: DD.MM.YYYY or DD.MM.YY (Russian format)
		parts := strings.Split(s, ".")
		if len(parts) != 3 {
			return "", fmt.Errorf("invalid date format: %s", dateStr)
		}
		day = strings.TrimSpace(parts[0])
		month = strings.TrimSpace(parts[1])
		year = strings.TrimSpace(parts[2])
	} else if strings.Contains(s, "/") {
		// Slash-separated: MM/DD/YYYY or MM/DD/YY (US format)
		parts := strings.Split(s, "/")
		if len(parts) != 3 {
			return "", fmt.Errorf("invalid date format: %s", dateStr)
		}
		month = strings.TrimSpace(parts[0])
		day = strings.TrimSpace(parts[1])
		year = strings.TrimSpace(parts[2])
	} else if strings.Contains(s, "-") {
		// Dash-separated: could be YYYY-MM-DD (ISO) or MM-DD-YY (US short from Excel)
		parts := strings.Split(s, "-")
		if len(parts) != 3 {
			return "", fmt.Errorf("invalid date format: %s", dateStr)
		}
		first := strings.TrimSpace(parts[0])
		second := strings.TrimSpace(parts[1])
		third := strings.TrimSpace(parts[2])

		if len(first) == 4 {
			// YYYY-MM-DD (ISO format)
			year = first
			month = second
			day = third
		} else {
			// MM-DD-YY (US short format from Excel)
			month = first
			day = second
			year = third
		}
	} else {
		return "", fmt.Errorf("invalid date format: %s", dateStr)
	}

	if len(day) < 1 || len(month) < 1 || len(year) < 1 {
		return "", fmt.Errorf("invalid date format: %s", dateStr)
	}

	// Handle 2-digit years
	if len(year) == 2 {
		yearNum, err := strconv.Atoi(year)
		if err != nil {
			return "", fmt.Errorf("invalid date format: %s", dateStr)
		}
		if yearNum >= 70 {
			year = fmt.Sprintf("19%s", padLeft(year, 2))
		} else {
			year = fmt.Sprintf("20%s", padLeft(year, 2))
		}
	}

	return fmt.Sprintf("%s-%s-%s 00:00:00", padLeft(year, 4), padLeft(month, 2), padLeft(day, 2)), nil
}

// padLeft pads a string to the given length with leading zeros
func padLeft(s string, length int) string {
	for len(s) < length {
		s = "0" + s
	}
	return s
}

// parseAmount parses amount strings in various formats:
// "(264)" -> "264", false (positive in parens = income)
// "(-559.08)" -> "559.08", true (negative in parens = expense)
// "5149930" -> "5149930", false (bare positive = income)
// "-14000" -> "14000", true (bare negative = expense)
// Returns the absolute amount string and whether it's negative
func parseAmount(amountStr string) (string, bool, error) {
	s := strings.TrimSpace(amountStr)
	if s == "" {
		return "0", false, nil
	}

	// Remove outer parentheses if present
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		s = s[1 : len(s)-1]
	}

	// Check for negative sign
	isNegative := false
	if strings.HasPrefix(s, "-") {
		isNegative = true
		s = s[1:]
	} else if strings.HasPrefix(s, "+") {
		s = s[1:]
	}

	// Validate it's a number
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return "", false, fmt.Errorf("invalid amount: %s", amountStr)
	}

	// Round to 2 decimal places for ezbookkeeping (amounts are in cents)
	f = math.Round(f*100) / 100

	// Format without trailing zeros but with max 2 decimal places
	result := strconv.FormatFloat(f, 'f', -1, 64)

	return result, isNegative, nil
}

// createDengioperaciiTransactionDataRowParser returns the row parser (as interface)
func createDengioperaciiTransactionDataRowParser(headerColumnNames []string) datatable.CommonTransactionDataRowParser {
	return createDengioperaciiTransactionDataRowParserInternal(headerColumnNames)
}

// createDengioperaciiTransactionDataRowParserInternal returns the row parser (concrete type)
func createDengioperaciiTransactionDataRowParserInternal(headerColumnNames []string) *dengioperaciiTransactionDataRowParser {
	existedOriginalDataColumns := make(map[string]bool, len(headerColumnNames))

	for i := 0; i < len(headerColumnNames); i++ {
		existedOriginalDataColumns[headerColumnNames[i]] = true
	}

	return &dengioperaciiTransactionDataRowParser{
		existedOriginalDataColumns: existedOriginalDataColumns,
	}
}
