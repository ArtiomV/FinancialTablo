package dengioperacii

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/converters/datatable"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// testDataRow implements datatable.CommonDataTableRow for testing
type testDataRow struct {
	data map[string]string
}

func (r *testDataRow) ColumnCount() int {
	return len(r.data)
}

func (r *testDataRow) HasData(columnName string) bool {
	_, exists := r.data[columnName]
	return exists
}

func (r *testDataRow) GetData(columnName string) string {
	return r.data[columnName]
}

// allColumns returns the standard set of header column names used by the parser
func allColumns() []string {
	return []string{
		"Дата",
		"Сумма",
		"Счет",
		"Валюта",
		"Контрагент",
		"Статья",
		"Описание",
		"Направление",
		"Субнаправление",
		"Род. статья",
	}
}

func newTestUser() *models.User {
	return &models.User{
		Uid:             1,
		DefaultCurrency: "MDL",
	}
}

// --- Parse: income rows ---

func TestParse_IncomeRow(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":           "09.02.2026",
		"Сумма":          "5000",
		"Счет":           "Кошелек",
		"Валюта":         "MDL",
		"Контрагент":     "Работодатель",
		"Статья":         "Зарплата",
		"Описание":       "Февральская зарплата",
		"Направление":    "Работа",
		"Субнаправление": "Основная",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "1")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "2026-02-09 00:00:00", data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME])
	assert.Equal(t, "Доход", data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE])
	assert.Equal(t, "5000", data[datatable.TRANSACTION_DATA_TABLE_AMOUNT])
	assert.Equal(t, "Зарплата", data[datatable.TRANSACTION_DATA_TABLE_CATEGORY])
	assert.Equal(t, "Кошелек", data[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME])
	assert.Equal(t, "MDL", data[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY])
	assert.Equal(t, "Февральская зарплата", data[datatable.TRANSACTION_DATA_TABLE_DESCRIPTION])
	assert.Equal(t, "Работодатель", data[datatable.TRANSACTION_DATA_TABLE_PAYEE])
	assert.Equal(t, "Работа", data[datatable.TRANSACTION_DATA_TABLE_TAG_GROUP])
	assert.Equal(t, "Основная", data[datatable.TRANSACTION_DATA_TABLE_TAGS])
	assert.Equal(t, "", data[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME])
	assert.False(t, parser.lastIsNegative)
}

func TestParse_IncomeRow_ParenthesesPositive(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":   "01.01.2026",
		"Сумма":  "(1500)",
		"Счет":   "Банк",
		"Валюта": "USD",
		"Статья": "Подарок",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "2")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "Доход", data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE])
	assert.Equal(t, "1500", data[datatable.TRANSACTION_DATA_TABLE_AMOUNT])
	assert.False(t, parser.lastIsNegative)
}

func TestParse_IncomeRow_PlusSign(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":   "15.06.2025",
		"Сумма":  "+250.50",
		"Счет":   "Карта",
		"Валюта": "EUR",
		"Статья": "Бонус",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "3")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "Доход", data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE])
	assert.Equal(t, "250.5", data[datatable.TRANSACTION_DATA_TABLE_AMOUNT])
	assert.False(t, parser.lastIsNegative)
}

// --- Parse: expense rows ---

func TestParse_ExpenseRow(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":           "10.02.2026",
		"Сумма":          "-559.08",
		"Счет":           "Карта",
		"Валюта":         "MDL",
		"Контрагент":     "Магазин",
		"Статья":         "Продукты",
		"Описание":       "Покупки в магазине",
		"Направление":    "Еда",
		"Субнаправление": "Продукты питания",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "4")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "2026-02-10 00:00:00", data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME])
	assert.Equal(t, "Расход", data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE])
	assert.Equal(t, "559.08", data[datatable.TRANSACTION_DATA_TABLE_AMOUNT])
	assert.Equal(t, "Продукты", data[datatable.TRANSACTION_DATA_TABLE_CATEGORY])
	assert.Equal(t, "Карта", data[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME])
	assert.Equal(t, "MDL", data[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_CURRENCY])
	assert.Equal(t, "Покупки в магазине", data[datatable.TRANSACTION_DATA_TABLE_DESCRIPTION])
	assert.Equal(t, "Магазин", data[datatable.TRANSACTION_DATA_TABLE_PAYEE])
	assert.Equal(t, "Еда", data[datatable.TRANSACTION_DATA_TABLE_TAG_GROUP])
	assert.Equal(t, "Продукты питания", data[datatable.TRANSACTION_DATA_TABLE_TAGS])
	assert.Equal(t, "", data[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME])
	assert.True(t, parser.lastIsNegative)
}

func TestParse_ExpenseRow_ParenthesesNegative(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":   "05.03.2026",
		"Сумма":  "(-264)",
		"Счет":   "Кошелек",
		"Валюта": "MDL",
		"Статья": "Транспорт",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "5")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "Расход", data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE])
	assert.Equal(t, "264", data[datatable.TRANSACTION_DATA_TABLE_AMOUNT])
	assert.True(t, parser.lastIsNegative)
}

// --- Parse: transfer rows ---

func TestParse_TransferRow_ConvertaciyaValyut(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":   "12.02.2026",
		"Сумма":  "-14000",
		"Счет":   "Кошелек MDL",
		"Валюта": "MDL",
		"Статья": "Конвертация валют",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "6")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "Перевод", data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE])
	assert.Equal(t, "14000", data[datatable.TRANSACTION_DATA_TABLE_AMOUNT])
	assert.Equal(t, "Конвертация валют", data[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME])
	assert.Equal(t, "Кошелек MDL", data[datatable.TRANSACTION_DATA_TABLE_ACCOUNT_NAME])
	assert.True(t, parser.lastIsNegative)
}

func TestParse_TransferRow_PerevodMezhduSchetami(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":   "20.01.2026",
		"Сумма":  "3000",
		"Счет":   "Сберегательный",
		"Валюта": "MDL",
		"Статья": "Перевод между счетами",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "7")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "Перевод", data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE])
	assert.Equal(t, "3000", data[datatable.TRANSACTION_DATA_TABLE_AMOUNT])
	assert.Equal(t, "Перевод между счетами", data[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME])
	assert.False(t, parser.lastIsNegative)
}

// --- Parse: empty / nil input handling ---

func TestParse_EmptyDate_ReturnsNilAndFalse(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":   "",
		"Сумма":  "100",
		"Счет":   "Банк",
		"Валюта": "MDL",
		"Статья": "Прочее",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "8")

	assert.Nil(t, err)
	assert.False(t, valid)
	assert.Nil(t, data)
}

func TestParse_EmptyRow_ReturnsNilAndFalse(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата": "",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "9")

	assert.Nil(t, err)
	assert.False(t, valid)
	assert.Nil(t, data)
}

func TestParse_MissingDateColumn_SkipsDateParsing(t *testing.T) {
	// Parser created without the date column in headers
	parser := createDengioperaciiTransactionDataRowParserInternal([]string{"Сумма", "Счет", "Валюта", "Статья"})
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Сумма":  "200",
		"Счет":   "Банк",
		"Валюта": "USD",
		"Статья": "Прочее",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "10")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "", data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME])
	assert.Equal(t, "200", data[datatable.TRANSACTION_DATA_TABLE_AMOUNT])
}

// --- Parse: zero amount handling ---

func TestParse_ZeroAmount(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":   "01.01.2026",
		"Сумма":  "0",
		"Счет":   "Банк",
		"Валюта": "MDL",
		"Статья": "Прочее",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "11")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "0", data[datatable.TRANSACTION_DATA_TABLE_AMOUNT])
	assert.Equal(t, "Доход", data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE])
	assert.False(t, parser.lastIsNegative)
}

func TestParse_EmptyAmount_TreatedAsZero(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":   "01.01.2026",
		"Сумма":  "",
		"Счет":   "Банк",
		"Валюта": "MDL",
		"Статья": "Прочее",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "12")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "0", data[datatable.TRANSACTION_DATA_TABLE_AMOUNT])
	assert.False(t, parser.lastIsNegative)
}

// --- Parse: row without a category ---

func TestParse_EmptyCategory(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":   "15.01.2026",
		"Сумма":  "-100",
		"Счет":   "Кошелек",
		"Валюта": "MDL",
		"Статья": "",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "13")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "", data[datatable.TRANSACTION_DATA_TABLE_CATEGORY])
	// Empty category is not a transfer category, so type is determined by sign
	assert.Equal(t, "Расход", data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE])
	assert.Equal(t, "", data[datatable.TRANSACTION_DATA_TABLE_RELATED_ACCOUNT_NAME])
}

func TestParse_MissingCategoryColumn(t *testing.T) {
	// Parser created without category column
	parser := createDengioperaciiTransactionDataRowParserInternal([]string{"Дата", "Сумма", "Счет", "Валюта"})
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":   "15.01.2026",
		"Сумма":  "500",
		"Счет":   "Банк",
		"Валюта": "EUR",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "14")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "", data[datatable.TRANSACTION_DATA_TABLE_CATEGORY])
	assert.Equal(t, "Доход", data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TYPE])
}

// --- Parse: invalid date handling ---

func TestParse_InvalidDate_NoSeparator_ReturnsError(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	// A date string with no recognized separator (., /, -) triggers an error
	row := &testDataRow{data: map[string]string{
		"Дата":   "20260209",
		"Сумма":  "100",
		"Счет":   "Банк",
		"Валюта": "MDL",
		"Статья": "Прочее",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "15")

	assert.NotNil(t, err)
	assert.False(t, valid)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "invalid date format")
}

func TestParse_InvalidDate_TooFewParts(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":   "09.02",
		"Сумма":  "100",
		"Счет":   "Банк",
		"Валюта": "MDL",
		"Статья": "Прочее",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "16")

	assert.NotNil(t, err)
	assert.False(t, valid)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "invalid date format")
}

func TestParse_InvalidAmount_ReturnsError(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":   "01.01.2026",
		"Сумма":  "abc",
		"Счет":   "Банк",
		"Валюта": "MDL",
		"Статья": "Прочее",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "17")

	assert.NotNil(t, err)
	assert.False(t, valid)
	assert.Nil(t, data)
	assert.Contains(t, err.Error(), "invalid amount")
}

// --- convertDateFormat unit tests ---

func TestConvertDateFormat_RussianFull(t *testing.T) {
	result, err := convertDateFormat("09.02.2026")
	assert.Nil(t, err)
	assert.Equal(t, "2026-02-09 00:00:00", result)
}

func TestConvertDateFormat_RussianShort(t *testing.T) {
	result, err := convertDateFormat("09.02.26")
	assert.Nil(t, err)
	assert.Equal(t, "2026-02-09 00:00:00", result)
}

func TestConvertDateFormat_RussianShort_Pre2000(t *testing.T) {
	result, err := convertDateFormat("15.06.95")
	assert.Nil(t, err)
	assert.Equal(t, "1995-06-15 00:00:00", result)
}

func TestConvertDateFormat_ISO(t *testing.T) {
	result, err := convertDateFormat("2026-02-09")
	assert.Nil(t, err)
	assert.Equal(t, "2026-02-09 00:00:00", result)
}

func TestConvertDateFormat_USSlash(t *testing.T) {
	result, err := convertDateFormat("02/09/2026")
	assert.Nil(t, err)
	assert.Equal(t, "2026-02-09 00:00:00", result)
}

func TestConvertDateFormat_USSlashShort(t *testing.T) {
	result, err := convertDateFormat("02/09/26")
	assert.Nil(t, err)
	assert.Equal(t, "2026-02-09 00:00:00", result)
}

func TestConvertDateFormat_USDash(t *testing.T) {
	result, err := convertDateFormat("02-09-26")
	assert.Nil(t, err)
	assert.Equal(t, "2026-02-09 00:00:00", result)
}

func TestConvertDateFormat_Empty_ReturnsError(t *testing.T) {
	_, err := convertDateFormat("")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid date format")
}

func TestConvertDateFormat_Whitespace_ReturnsError(t *testing.T) {
	_, err := convertDateFormat("   ")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid date format")
}

func TestConvertDateFormat_NoSeparator_ReturnsError(t *testing.T) {
	_, err := convertDateFormat("20260209")
	assert.NotNil(t, err)
}

func TestConvertDateFormat_TooFewDotParts_ReturnsError(t *testing.T) {
	_, err := convertDateFormat("09.02")
	assert.NotNil(t, err)
}

func TestConvertDateFormat_TooManyDotParts_ReturnsError(t *testing.T) {
	_, err := convertDateFormat("09.02.20.26")
	assert.NotNil(t, err)
}

func TestConvertDateFormat_InvalidYearChars(t *testing.T) {
	_, err := convertDateFormat("09.02.XX")
	assert.NotNil(t, err)
}

// --- parseAmount unit tests ---

func TestParseAmount_BarePositive(t *testing.T) {
	result, isNeg, err := parseAmount("5149930")
	assert.Nil(t, err)
	assert.Equal(t, "5149930", result)
	assert.False(t, isNeg)
}

func TestParseAmount_BareNegative(t *testing.T) {
	result, isNeg, err := parseAmount("-14000")
	assert.Nil(t, err)
	assert.Equal(t, "14000", result)
	assert.True(t, isNeg)
}

func TestParseAmount_ParensPositive(t *testing.T) {
	result, isNeg, err := parseAmount("(264)")
	assert.Nil(t, err)
	assert.Equal(t, "264", result)
	assert.False(t, isNeg)
}

func TestParseAmount_ParensNegative(t *testing.T) {
	result, isNeg, err := parseAmount("(-559.08)")
	assert.Nil(t, err)
	assert.Equal(t, "559.08", result)
	assert.True(t, isNeg)
}

func TestParseAmount_WithPlus(t *testing.T) {
	result, isNeg, err := parseAmount("+100.50")
	assert.Nil(t, err)
	assert.Equal(t, "100.5", result)
	assert.False(t, isNeg)
}

func TestParseAmount_Empty_ReturnsZero(t *testing.T) {
	result, isNeg, err := parseAmount("")
	assert.Nil(t, err)
	assert.Equal(t, "0", result)
	assert.False(t, isNeg)
}

func TestParseAmount_Whitespace_ReturnsZero(t *testing.T) {
	result, isNeg, err := parseAmount("   ")
	assert.Nil(t, err)
	assert.Equal(t, "0", result)
	assert.False(t, isNeg)
}

func TestParseAmount_Zero(t *testing.T) {
	result, isNeg, err := parseAmount("0")
	assert.Nil(t, err)
	assert.Equal(t, "0", result)
	assert.False(t, isNeg)
}

func TestParseAmount_DecimalRounding(t *testing.T) {
	result, isNeg, err := parseAmount("99.999")
	assert.Nil(t, err)
	assert.Equal(t, "100", result)
	assert.False(t, isNeg)
}

func TestParseAmount_InvalidText_ReturnsError(t *testing.T) {
	_, _, err := parseAmount("abc")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid amount")
}

func TestParseAmount_InvalidParens_ReturnsError(t *testing.T) {
	_, _, err := parseAmount("(abc)")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid amount")
}

// --- Direction / SubDirection trimming ---

func TestParse_DirectionTrimsDashPrefix(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":           "01.01.2026",
		"Сумма":          "100",
		"Счет":           "Банк",
		"Валюта":         "MDL",
		"Статья":         "Прочее",
		"Направление":    "- Категория",
		"Субнаправление": "- Подкатегория",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "18")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "Категория", data[datatable.TRANSACTION_DATA_TABLE_TAG_GROUP])
	assert.Equal(t, "Подкатегория", data[datatable.TRANSACTION_DATA_TABLE_TAGS])
}

func TestParse_DirectionTrimsDashWithoutSpace(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":           "01.01.2026",
		"Сумма":          "100",
		"Счет":           "Банк",
		"Валюта":         "MDL",
		"Статья":         "Прочее",
		"Направление":    "-Тест",
		"Субнаправление": "-Суб",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "19")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "Тест", data[datatable.TRANSACTION_DATA_TABLE_TAG_GROUP])
	assert.Equal(t, "Суб", data[datatable.TRANSACTION_DATA_TABLE_TAGS])
}

// --- Counterparty / Payee handling ---

func TestParse_EmptyCounterparty_NoPayee(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":       "01.01.2026",
		"Сумма":      "100",
		"Счет":       "Банк",
		"Валюта":     "MDL",
		"Статья":     "Прочее",
		"Контрагент": "",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "20")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "", data[datatable.TRANSACTION_DATA_TABLE_PAYEE])
}

func TestParse_WhitespaceCounterparty_NoPayee(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":       "01.01.2026",
		"Сумма":      "100",
		"Счет":       "Банк",
		"Валюта":     "MDL",
		"Статья":     "Прочее",
		"Контрагент": "   ",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "21")

	assert.Nil(t, err)
	assert.True(t, valid)
	// Whitespace-only counterparty is trimmed to empty, so no payee is set
	assert.Equal(t, "", data[datatable.TRANSACTION_DATA_TABLE_PAYEE])
}

// --- createDengioperaciiTransactionDataRowParser tests ---

func TestCreateParser_ReturnsNonNil(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParser(allColumns())
	assert.NotNil(t, parser)
}

func TestCreateParser_EmptyHeaders(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal([]string{})
	assert.NotNil(t, parser)
	assert.Equal(t, 0, len(parser.existedOriginalDataColumns))
}

// --- Sub category is always empty ---

func TestParse_SubCategoryAlwaysEmpty(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":   "01.01.2026",
		"Сумма":  "100",
		"Счет":   "Банк",
		"Валюта": "MDL",
		"Статья": "Зарплата",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "22")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "", data[datatable.TRANSACTION_DATA_TABLE_SUB_CATEGORY])
}

// --- Multiple date formats through Parse ---

func TestParse_ISODateFormat(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":   "2026-02-09",
		"Сумма":  "100",
		"Счет":   "Банк",
		"Валюта": "MDL",
		"Статья": "Прочее",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "23")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "2026-02-09 00:00:00", data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME])
}

func TestParse_USSlashDateFormat(t *testing.T) {
	parser := createDengioperaciiTransactionDataRowParserInternal(allColumns())
	ctx := core.NewNullContext()
	user := newTestUser()

	row := &testDataRow{data: map[string]string{
		"Дата":   "02/09/2026",
		"Сумма":  "100",
		"Счет":   "Банк",
		"Валюта": "MDL",
		"Статья": "Прочее",
	}}

	data, valid, err := parser.Parse(ctx, user, row, "24")

	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "2026-02-09 00:00:00", data[datatable.TRANSACTION_DATA_TABLE_TRANSACTION_TIME])
}
