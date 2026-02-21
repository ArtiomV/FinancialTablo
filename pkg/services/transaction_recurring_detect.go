// transaction_recurring_detect.go detects recurring patterns among imported transactions
// and creates scheduled templates for them, linking future planned transactions.
package services

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// recurringPatternKey uniquely identifies a group of potentially recurring transactions
type recurringPatternKey struct {
	Type           models.TransactionDbType
	CategoryId     int64
	AccountId      int64
	Amount         int64
	CounterpartyId int64
}

// recurringPattern holds all transactions matching a pattern
type recurringPattern struct {
	Key          recurringPatternKey
	Transactions []*models.Transaction
}

// DetectedRecurrence represents a detected recurring pattern with its frequency info
type DetectedRecurrence struct {
	FrequencyType models.TransactionScheduleFrequencyType
	Frequency     string // comma-separated day values
	DayOfMonth    int    // most common day-of-month (for monthly)
	DayOfWeek     int    // most common day-of-week (for weekly)
}

const minRecurrenceCount = 3 // minimum occurrences to detect recurrence

// DetectAndCreateRecurringTemplates analyzes imported transactions for recurring patterns
// and creates scheduled templates. Future planned transactions are linked to these templates.
func (s *TransactionService) DetectAndCreateRecurringTemplates(
	c core.Context,
	uid int64,
	transactions []*models.Transaction,
	timezoneUtcOffset int16,
) (int, error) {
	if len(transactions) < minRecurrenceCount {
		return 0, nil
	}

	// Step 1: Group transactions by pattern key
	patternMap := make(map[recurringPatternKey]*recurringPattern)

	for _, txn := range transactions {
		key := recurringPatternKey{
			Type:           txn.Type,
			CategoryId:     txn.CategoryId,
			AccountId:      txn.AccountId,
			Amount:         txn.Amount,
			CounterpartyId: txn.CounterpartyId,
		}

		if _, ok := patternMap[key]; !ok {
			patternMap[key] = &recurringPattern{
				Key:          key,
				Transactions: make([]*models.Transaction, 0),
			}
		}
		patternMap[key].Transactions = append(patternMap[key].Transactions, txn)
	}

	// Step 2: For each pattern with enough occurrences, detect frequency
	templateCount := 0

	for _, pattern := range patternMap {
		if len(pattern.Transactions) < minRecurrenceCount {
			continue
		}

		// Sort by transaction time
		sort.Slice(pattern.Transactions, func(i, j int) bool {
			return pattern.Transactions[i].TransactionTime < pattern.Transactions[j].TransactionTime
		})

		recurrence := detectFrequency(pattern.Transactions, timezoneUtcOffset)
		if recurrence == nil {
			continue
		}

		// Step 3: Create a scheduled template
		err := s.createRecurringTemplate(c, uid, pattern, recurrence, timezoneUtcOffset)
		if err != nil {
			log.Warnf(c, "[transactions.DetectAndCreateRecurringTemplates] failed to create template for pattern (cat:%d, acc:%d, amt:%d), because %s",
				pattern.Key.CategoryId, pattern.Key.AccountId, pattern.Key.Amount, err.Error())
			continue
		}
		templateCount++
	}

	if templateCount > 0 {
		log.Infof(c, "[transactions.DetectAndCreateRecurringTemplates] created %d recurring templates for user \"uid:%d\"", templateCount, uid)
	}

	return templateCount, nil
}

// detectFrequency analyzes sorted transactions to determine recurrence frequency
func detectFrequency(transactions []*models.Transaction, timezoneUtcOffset int16) *DetectedRecurrence {
	if len(transactions) < minRecurrenceCount {
		return nil
	}

	tz := time.FixedZone("User Timezone", int(timezoneUtcOffset)*60)

	// Convert transaction times to dates
	dates := make([]time.Time, len(transactions))
	for i, txn := range transactions {
		unixTime := utils.GetUnixTimeFromTransactionTime(txn.TransactionTime)
		dates[i] = time.Unix(unixTime, 0).In(tz)
	}

	// Calculate intervals between consecutive dates
	intervals := make([]int, 0, len(dates)-1)
	for i := 1; i < len(dates); i++ {
		delta := int(dates[i].Sub(dates[i-1]).Hours() / 24)
		if delta > 0 {
			intervals = append(intervals, delta)
		}
	}

	if len(intervals) == 0 {
		return nil
	}

	avgInterval := 0.0
	for _, iv := range intervals {
		avgInterval += float64(iv)
	}
	avgInterval /= float64(len(intervals))

	// Analyze day-of-month distribution
	domCounts := make(map[int]int)
	dowCounts := make(map[int]int) // day of week
	for _, d := range dates {
		domCounts[d.Day()]++
		dowCounts[int(d.Weekday())]++
	}

	// Find most common day-of-month
	bestDOM := 0
	bestDOMCount := 0
	for dom, cnt := range domCounts {
		if cnt > bestDOMCount {
			bestDOM = dom
			bestDOMCount = cnt
		}
	}

	// Find most common day-of-week
	bestDOW := 0
	bestDOWCount := 0
	for dow, cnt := range dowCounts {
		if cnt > bestDOWCount {
			bestDOW = dow
			bestDOWCount = cnt
		}
	}

	// Determine frequency type
	// Weekly: avg interval 5-9 days, or consistent day-of-week with 80%+ hits
	// Monthly: avg interval 25-35 days
	// Bimonthly: avg interval 55-65 days
	// Quarterly: avg interval 85-95 days

	// Check if it's weekly (consistent day of week)
	if avgInterval >= 5 && avgInterval <= 9 {
		weekdayConsistency := float64(bestDOWCount) / float64(len(dates))
		if weekdayConsistency >= 0.7 {
			return &DetectedRecurrence{
				FrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_WEEKLY,
				Frequency:     strconv.Itoa(bestDOW),
				DayOfWeek:     bestDOW,
			}
		}
	}

	// Check monthly patterns
	if avgInterval >= 25 && avgInterval <= 35 {
		// Check if most transactions land on the same day or near end of month
		domConsistency := float64(bestDOMCount) / float64(len(dates))

		// Also check for end-of-month pattern (28-31)
		endOfMonthCount := 0
		for _, d := range dates {
			if d.Day() >= 28 {
				endOfMonthCount++
			}
		}
		endOfMonthConsistency := float64(endOfMonthCount) / float64(len(dates))

		if domConsistency >= 0.6 {
			return &DetectedRecurrence{
				FrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY,
				Frequency:     strconv.Itoa(bestDOM),
				DayOfMonth:    bestDOM,
			}
		}

		if endOfMonthConsistency >= 0.7 {
			// End-of-month pattern — use day 31 (will clamp to last day)
			return &DetectedRecurrence{
				FrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY,
				Frequency:     "31",
				DayOfMonth:    31,
			}
		}

		// Even if day varies a bit, still monthly if interval is consistent
		intervalStdDev := calcStdDev(intervals)
		if intervalStdDev < 5.0 && len(transactions) >= 5 {
			return &DetectedRecurrence{
				FrequencyType: models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY,
				Frequency:     strconv.Itoa(bestDOM),
				DayOfMonth:    bestDOM,
			}
		}
	}

	return nil
}

func calcStdDev(values []int) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := 0.0
	for _, v := range values {
		mean += float64(v)
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, v := range values {
		diff := float64(v) - mean
		variance += diff * diff
	}
	variance /= float64(len(values))
	return math.Sqrt(variance)
}

// createRecurringTemplate creates a scheduled template for a detected recurring pattern
func (s *TransactionService) createRecurringTemplate(
	c core.Context,
	uid int64,
	pattern *recurringPattern,
	recurrence *DetectedRecurrence,
	timezoneUtcOffset int16,
) error {
	// Determine the transaction type for the template
	var templateType models.TransactionType
	switch pattern.Key.Type {
	case models.TRANSACTION_DB_TYPE_EXPENSE:
		templateType = models.TRANSACTION_TYPE_EXPENSE
	case models.TRANSACTION_DB_TYPE_INCOME:
		templateType = models.TRANSACTION_TYPE_INCOME
	case models.TRANSACTION_DB_TYPE_TRANSFER_OUT, models.TRANSACTION_DB_TYPE_TRANSFER_IN:
		templateType = models.TRANSACTION_TYPE_TRANSFER
	default:
		return fmt.Errorf("unsupported transaction type: %d", pattern.Key.Type)
	}

	// Build a descriptive name
	name := "Repeat: auto"

	// Calculate ScheduledAt (minutes elapsed in UTC for midnight in user's timezone)
	templateTimeZone := time.FixedZone("Template Timezone", int(timezoneUtcOffset)*60)
	transactionTimeUTC := time.Date(2020, 1, 1, 0, 0, 0, 0, templateTimeZone).In(time.UTC)
	minutesElapsedOfDayInUtc := transactionTimeUTC.Hour()*60 + transactionTimeUTC.Minute()
	scheduledAt := int16(minutesElapsedOfDayInUtc)

	// Build frequency string with ordered values
	frequency := recurrence.Frequency
	freqParts := strings.Split(frequency, ",")
	freqValues := make([]int, 0, len(freqParts))
	for _, part := range freqParts {
		part = strings.TrimSpace(part)
		if v, err := strconv.Atoi(part); err == nil {
			freqValues = append(freqValues, v)
		}
	}
	sort.Ints(freqValues)
	sortedParts := make([]string, len(freqValues))
	for i, v := range freqValues {
		sortedParts[i] = strconv.Itoa(v)
	}
	frequency = strings.Join(sortedParts, ",")

	template := &models.TransactionTemplate{
		Uid:                        uid,
		TemplateType:               models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE,
		Name:                       name,
		Type:                       templateType,
		CategoryId:                 pattern.Key.CategoryId,
		AccountId:                  pattern.Key.AccountId,
		Amount:                     pattern.Key.Amount,
		ScheduledFrequencyType:     recurrence.FrequencyType,
		ScheduledFrequency:         frequency,
		ScheduledAt:                scheduledAt,
		ScheduledTimezoneUtcOffset: timezoneUtcOffset,
		TagIds:                     "",
		Comment:                    "",
		HideAmount:                 false,
		DisplayOrder:               0,
		Hidden:                     false,
	}

	// Handle transfer-related fields
	if templateType == models.TRANSACTION_TYPE_TRANSFER && len(pattern.Transactions) > 0 {
		template.RelatedAccountId = pattern.Transactions[0].RelatedAccountId
		template.RelatedAccountAmount = pattern.Transactions[0].RelatedAccountAmount
	}

	// Create the template
	err := TransactionTemplates.CreateTemplate(c, template)
	if err != nil {
		return err
	}

	log.Infof(c, "[transactions.createRecurringTemplate] created template \"id:%d\" freq_type=%d freq=%s for %d transactions",
		template.TemplateId, recurrence.FrequencyType, frequency, len(pattern.Transactions))

	// Step 4: Link planned (future) transactions to this template
	linkedCount := 0
	for _, txn := range pattern.Transactions {
		if txn.Planned {
			err := s.linkTransactionToTemplate(c, uid, txn.TransactionId, template.TemplateId)
			if err != nil {
				log.Warnf(c, "[transactions.createRecurringTemplate] failed to link transaction \"id:%d\" to template \"id:%d\": %s",
					txn.TransactionId, template.TemplateId, err.Error())
				continue
			}
			linkedCount++
		}
	}

	if linkedCount > 0 {
		log.Infof(c, "[transactions.createRecurringTemplate] linked %d planned transactions to template \"id:%d\"",
			linkedCount, template.TemplateId)
	}

	return nil
}

// linkTransactionToTemplate sets the SourceTemplateId on a transaction
func (s *TransactionService) linkTransactionToTemplate(c core.Context, uid int64, transactionId int64, templateId int64) error {
	now := time.Now().Unix()

	updateModel := &models.Transaction{
		SourceTemplateId: templateId,
		UpdatedUnixTime:  now,
	}

	updatedRows, err := s.UserDataDB(uid).NewSession(c).
		ID(transactionId).
		Cols("source_template_id", "updated_unix_time").
		Where("uid=? AND deleted=?", uid, false).
		Update(updateModel)

	if err != nil {
		return err
	}

	if updatedRows < 1 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}
