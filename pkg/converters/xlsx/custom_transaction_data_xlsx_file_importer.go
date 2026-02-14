package xlsx

import (
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/converters/converter"
	"github.com/mayswind/ezbookkeeping/pkg/converters/datatable"
	"github.com/mayswind/ezbookkeeping/pkg/converters/dsv"
	"github.com/mayswind/ezbookkeeping/pkg/converters/excel"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

var customTransactionTypeNameMapping = map[models.TransactionType]string{
	models.TRANSACTION_TYPE_MODIFY_BALANCE: "1",
	models.TRANSACTION_TYPE_INCOME:         "2",
	models.TRANSACTION_TYPE_EXPENSE:        "3",
	models.TRANSACTION_TYPE_TRANSFER:       "4",
}

// customTransactionDataXlsxFileImporter defines the structure of custom xlsx importer for transaction data
type customTransactionDataXlsxFileImporter struct {
	columnIndexMapping         map[datatable.TransactionDataTableColumn]int
	transactionTypeNameMapping map[string]models.TransactionType
	hasHeaderLine              bool
	timeFormat                 string
	timezoneFormat             string
	amountDecimalSeparator     string
	amountDigitGroupingSymbol  string
	geoLocationSeparator       string
	geoLocationOrder           converter.TransactionGeoLocationOrder
	transactionTagSeparator    string
}

// ParseImportedData returns the imported data by parsing the custom transaction xlsx data
func (c *customTransactionDataXlsxFileImporter) ParseImportedData(ctx core.Context, user *models.User, data []byte, defaultTimezone *time.Location, additionalOptions converter.TransactionDataImporterOptions, accountMap map[string]*models.Account, expenseCategoryMap map[string]*models.TransactionCategory, incomeCategoryMap map[string]*models.TransactionCategory, transferCategoryMap map[string]*models.TransactionCategory, tagMap map[string]*models.TransactionTag) (models.ImportedTransactionSlice, []*models.Account, []*models.TransactionCategory, []*models.TransactionCategory, []*models.TransactionCategory, []*models.TransactionTag, error) {
	dataTable, err := excel.CreateNewExcelOOXMLFileBasicDataTable(data, c.hasHeaderLine)

	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	transactionDataTable := dsv.CreateNewCustomPlainTextDataTable(dataTable, c.columnIndexMapping, c.transactionTypeNameMapping, c.timeFormat, c.timezoneFormat, c.amountDecimalSeparator, c.amountDigitGroupingSymbol)
	dataTableImporter := converter.CreateNewImporterWithTypeNameMapping(customTransactionTypeNameMapping, c.geoLocationSeparator, c.geoLocationOrder, c.transactionTagSeparator)

	return dataTableImporter.ParseImportedData(ctx, user, transactionDataTable, defaultTimezone, additionalOptions, accountMap, expenseCategoryMap, incomeCategoryMap, transferCategoryMap, tagMap)
}

// CreateNewCustomTransactionDataXlsxFileImporter returns a new custom xlsx importer for transaction data
func CreateNewCustomTransactionDataXlsxFileImporter(columnIndexMapping map[datatable.TransactionDataTableColumn]int, transactionTypeNameMapping map[string]models.TransactionType, hasHeaderLine bool, timeFormat string, timezoneFormat string, amountDecimalSeparator string, amountDigitGroupingSymbol string, geoLocationSeparator string, geoLocationOrder string, transactionTagSeparator string) (converter.TransactionDataImporter, error) {
	if geoLocationOrder == "" {
		geoLocationOrder = string(converter.TRANSACTION_GEO_LOCATION_ORDER_LONGITUDE_LATITUDE)
	} else if geoLocationOrder != string(converter.TRANSACTION_GEO_LOCATION_ORDER_LONGITUDE_LATITUDE) &&
		geoLocationOrder != string(converter.TRANSACTION_GEO_LOCATION_ORDER_LATITUDE_LONGITUDE) {
		return nil, errs.ErrImportFileTypeNotSupported
	}

	if _, exists := columnIndexMapping[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME]; !exists {
		return nil, errs.ErrMissingRequiredFieldInHeaderRow
	}

	if _, exists := columnIndexMapping[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE]; !exists {
		return nil, errs.ErrMissingRequiredFieldInHeaderRow
	}

	if _, exists := columnIndexMapping[datatable.TRANSACTION_DATA_TABLE_AMOUNT]; !exists {
		return nil, errs.ErrMissingRequiredFieldInHeaderRow
	}

	return &customTransactionDataXlsxFileImporter{
		columnIndexMapping:         columnIndexMapping,
		transactionTypeNameMapping: transactionTypeNameMapping,
		hasHeaderLine:              hasHeaderLine,
		timeFormat:                 timeFormat,
		timezoneFormat:             timezoneFormat,
		amountDecimalSeparator:     amountDecimalSeparator,
		amountDigitGroupingSymbol:  amountDigitGroupingSymbol,
		geoLocationSeparator:       geoLocationSeparator,
		geoLocationOrder:           converter.TransactionGeoLocationOrder(geoLocationOrder),
		transactionTagSeparator:    transactionTagSeparator,
	}, nil
}
