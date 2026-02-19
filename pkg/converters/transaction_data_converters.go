package converters

import (
	"github.com/mayswind/ezbookkeeping/pkg/converters/converter"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
)

// GetTransactionDataExporter returns the transaction data exporter according to the file type
func GetTransactionDataExporter(fileType string) converter.TransactionDataExporter {
	return nil
}

// GetTransactionDataImporter returns the transaction data importer according to the file type
func GetTransactionDataImporter(fileType string) (converter.TransactionDataImporter, error) {
	if fileType == "custom_csv" {
		return CustomCSVImporter, nil
	}

	return nil, errs.ErrImportFileTypeNotSupported
}

// IsCustomDelimiterSeparatedValuesFileType returns whether the file type is the delimiter-separated values file type
func IsCustomDelimiterSeparatedValuesFileType(fileType string) bool {
	return false
}

// IsCustomExcelFileType returns whether the file type is the custom excel file type
func IsCustomExcelFileType(fileType string) bool {
	return false
}
