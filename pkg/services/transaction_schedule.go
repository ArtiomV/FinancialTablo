// transaction_schedule.go handles scheduled and planned transaction generation.
// Runs via cron to create transactions at their scheduled times.
package services

import (
	"strconv"
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// GeneratePlannedTransactions creates planned future transactions based on a repeatable transaction
func (s *TransactionService) GeneratePlannedTransactions(c core.Context, baseTransaction *models.Transaction, tagIds []int64, frequencyType models.TransactionScheduleFrequencyType, frequency string, templateId int64, splitRequests []models.TransactionSplitCreateRequest) (int, error) {
	if baseTransaction.Uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	// Parse the frequency string (comma-separated day numbers)
	freqParts := strings.Split(frequency, ",")
	freqDays := make([]int, 0, len(freqParts))
	for _, part := range freqParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		day, err := strconv.Atoi(part)
		if err != nil {
			return 0, errs.ErrFormatInvalid
		}
		freqDays = append(freqDays, day)
	}

	if len(freqDays) == 0 {
		return 0, errs.ErrFormatInvalid
	}

	// Calculate the base date from the transaction time
	baseUnixTime := utils.GetUnixTimeFromTransactionTime(baseTransaction.TransactionTime)
	tz := time.FixedZone("Transaction Timezone", int(baseTransaction.TimezoneUtcOffset)*60)
	baseDate := time.Unix(baseUnixTime, 0).In(tz)

	// Calculate end date: January 31 of next year
	nextYear := baseDate.Year() + 1
	endDate := time.Date(nextYear, time.January, 31, 23, 59, 59, 0, tz)

	// Generate all future dates
	var futureDates []time.Time

	switch frequencyType {
	case models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_WEEKLY:
		// freqDays contains weekday numbers (0=Sun to 6=Sat)
		current := baseDate.AddDate(0, 0, 1) // start from the day after the base date
		for current.Before(endDate) || current.Equal(endDate) {
			weekday := int(current.Weekday())
			for _, d := range freqDays {
				if weekday == d {
					futureDates = append(futureDates, current)
					break
				}
			}
			current = current.AddDate(0, 0, 1)
		}

	case models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY:
		// freqDays contains day-of-month numbers
		current := baseDate
		// Start from current month, then iterate month by month
		for {
			for _, day := range freqDays {
				// Create date for this day in the current month
				candidate := time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, tz)
				// Get last day of this month
				lastDay := candidate.AddDate(0, 1, -1).Day()
				actualDay := day
				if actualDay > lastDay {
					actualDay = lastDay // clamp to last day of month (e.g., 31 -> 28 for Feb)
				}
				candidate = time.Date(current.Year(), current.Month(), actualDay, baseDate.Hour(), baseDate.Minute(), baseDate.Second(), 0, tz)
				if candidate.After(baseDate) && (candidate.Before(endDate) || candidate.Equal(endDate)) {
					futureDates = append(futureDates, candidate)
				}
			}
			current = time.Date(current.Year(), current.Month()+1, 1, 0, 0, 0, 0, tz)
			if current.After(endDate) {
				break
			}
		}

	case models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_BIMONTHLY,
		models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_QUARTERLY,
		models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_SEMIANNUALLY,
		models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_ANNUALLY:
		// Determine the month interval
		var monthInterval int
		switch frequencyType {
		case models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_BIMONTHLY:
			monthInterval = 2
		case models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_QUARTERLY:
			monthInterval = 3
		case models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_SEMIANNUALLY:
			monthInterval = 6
		case models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_ANNUALLY:
			monthInterval = 12
		}

		for _, day := range freqDays {
			// Start from the base month + interval
			current := time.Date(baseDate.Year(), baseDate.Month(), 1, 0, 0, 0, 0, tz)
			// Move to the next interval from base
			current = current.AddDate(0, monthInterval, 0)

			for current.Before(endDate) || current.Equal(endDate) {
				lastDay := current.AddDate(0, 1, -1).Day()
				actualDay := day
				if actualDay > lastDay {
					actualDay = lastDay
				}
				candidate := time.Date(current.Year(), current.Month(), actualDay, baseDate.Hour(), baseDate.Minute(), baseDate.Second(), 0, tz)
				if candidate.Before(endDate) || candidate.Equal(endDate) {
					futureDates = append(futureDates, candidate)
				}
				current = current.AddDate(0, monthInterval, 0)
			}
		}
	}

	// Create a planned transaction for each future date
	count := 0
	for _, futureDate := range futureDates {
		futureUnixTime := futureDate.Unix()
		futureTransactionTime := utils.GetMinTransactionTimeFromUnixTime(futureUnixTime)

		plannedTransaction := &models.Transaction{
			Uid:                  baseTransaction.Uid,
			Type:                 baseTransaction.Type,
			CategoryId:           baseTransaction.CategoryId,
			TransactionTime:      futureTransactionTime,
			TimezoneUtcOffset:    baseTransaction.TimezoneUtcOffset,
			AccountId:            baseTransaction.AccountId,
			Amount:               baseTransaction.Amount,
			RelatedAccountId:     baseTransaction.RelatedAccountId,
			RelatedAccountAmount: baseTransaction.RelatedAccountAmount,
			HideAmount:           baseTransaction.HideAmount,
			Comment:              baseTransaction.Comment,
			CounterpartyId:       baseTransaction.CounterpartyId,
			GeoLongitude:         baseTransaction.GeoLongitude,
			GeoLatitude:          baseTransaction.GeoLatitude,
			CreatedIp:            baseTransaction.CreatedIp,
			Planned:              true,
			SourceTemplateId:     templateId,
		}

		err := s.CreateTransaction(c, plannedTransaction, tagIds, nil)
		if err != nil {
			log.Warnf(c, "[transactions.GeneratePlannedTransactions] failed to create planned transaction for user \"uid:%d\", because %s", baseTransaction.Uid, err.Error())
			return count, err
		}

		// Copy splits to the planned transaction
		if len(splitRequests) > 0 {
			splitErr := TransactionSplits.CreateSplits(c, baseTransaction.Uid, plannedTransaction.TransactionId, splitRequests)
			if splitErr != nil {
				log.Warnf(c, "[transactions.GeneratePlannedTransactions] failed to create splits for planned transaction \"id:%d\" for user \"uid:%d\", because %s", plannedTransaction.TransactionId, baseTransaction.Uid, splitErr.Error())
			}
		}

		count++
	}

	return count, nil
}

// matchesFrequencyDay checks if the current day matches any frequency day,
// with last-day-of-month clamping (e.g., Feb 28 matches frequency=31)
func matchesFrequencyDay(transactionTime time.Time, frequencyValueSet map[int64]bool, tz *time.Location) bool {
	todayDay := int64(transactionTime.Day())
	if frequencyValueSet[todayDay] {
		return true
	}
	// Check if today is the last day of the month and any frequency day exceeds it
	lastDayOfMonth := time.Date(transactionTime.Year(), transactionTime.Month()+1, 0, 0, 0, 0, 0, tz).Day()
	if int(todayDay) == lastDayOfMonth {
		for fd := range frequencyValueSet {
			if int(fd) > lastDayOfMonth {
				return true
			}
		}
	}
	return false
}

// CreateScheduledTransactions saves all scheduled transactions that should be created now
func (s *TransactionService) CreateScheduledTransactions(c core.Context, currentUnixTime int64, interval time.Duration) error {
	var allTemplates []*models.TransactionTemplate
	intervalMinute := int(interval / time.Minute)
	currentTime := time.Unix(currentUnixTime, 0)
	currentMinute := (currentTime.Minute() / intervalMinute) * intervalMinute

	startTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), currentTime.Hour(), currentMinute, 0, 0, time.Local)
	startTimeInUTC := startTime.In(time.UTC)

	minutesElapsedOfDayInUtc := startTimeInUTC.Hour()*60 + startTimeInUTC.Minute()
	secondsElapsedOfDayInUtc := minutesElapsedOfDayInUtc * 60
	todayFirstTimeInUTC := startTimeInUTC.Add(time.Duration(-secondsElapsedOfDayInUtc) * time.Second)
	todayFirstUnixTimeInUTC := todayFirstTimeInUTC.Unix()

	minScheduledAt := minutesElapsedOfDayInUtc
	maxScheduledAt := minScheduledAt + intervalMinute

	for i := 0; i < s.UserDataDBCount(); i++ {
		var templates []*models.TransactionTemplate
		err := s.UserDataDBByIndex(i).NewSession(c).Where("deleted=? AND template_type=? AND scheduled_frequency_type>=? AND scheduled_frequency_type<=? AND (scheduled_start_time IS NULL OR scheduled_start_time<=?) AND (scheduled_end_time IS NULL OR scheduled_end_time>=?) AND scheduled_at>=? AND scheduled_at<?", false, models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE, models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_WEEKLY, models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_ANNUALLY, startTime.Unix(), startTime.Unix(), minScheduledAt, maxScheduledAt).Find(&templates)

		if err != nil {
			return err
		}

		allTemplates = append(allTemplates, templates...)
	}

	if len(allTemplates) < 1 {
		return nil
	}

	log.Infof(c, "[transactions.CreateScheduledTransactions] should process %d scheduled transaction templates now (scheduled at from %d to %d)", len(allTemplates), minScheduledAt, maxScheduledAt)

	successCount := 0
	skipCount := 0
	failedCount := 0

	for i := 0; i < len(allTemplates); i++ {
		template := allTemplates[i]

		if template.ScheduledFrequencyType == models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_DISABLED {
			skipCount++
			log.Warnf(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" disabled scheduled transaction frequency", template.TemplateId)
			continue
		}

		if template.ScheduledFrequencyType < models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_WEEKLY ||
			template.ScheduledFrequencyType > models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_ANNUALLY ||
			template.ScheduledFrequency == "" {
			skipCount++
			log.Warnf(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" has invalid scheduled transaction frequency", template.TemplateId)
			continue
		}

		frequencyValues, err := utils.StringArrayToInt64Array(strings.Split(template.ScheduledFrequency, ","))

		if err != nil {
			skipCount++
			log.Warnf(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" has invalid scheduled transaction frequency, because %s", template.TemplateId, err.Error())
			continue
		}

		frequencyValueSet := utils.ToSet(frequencyValues)
		templateTimeZone := time.FixedZone("Template Timezone", int(template.ScheduledTimezoneUtcOffset)*60)
		transactionUnixTime := todayFirstUnixTimeInUTC + int64(template.ScheduledAt)*60
		transactionTime := time.Unix(transactionUnixTime, 0).In(templateTimeZone)

		shouldSkip := false

		switch template.ScheduledFrequencyType {
		case models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_WEEKLY:
			if !frequencyValueSet[int64(transactionTime.Weekday())] {
				shouldSkip = true
			}
		case models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY:
			if !matchesFrequencyDay(transactionTime, frequencyValueSet, templateTimeZone) {
				shouldSkip = true
			}
		case models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_BIMONTHLY:
			if !matchesFrequencyDay(transactionTime, frequencyValueSet, templateTimeZone) {
				shouldSkip = true
			} else if template.ScheduledStartTime != nil {
				scheduleStartTime := time.Unix(*template.ScheduledStartTime, 0).In(templateTimeZone)
				monthDiff := (int(transactionTime.Year())-int(scheduleStartTime.Year()))*12 + int(transactionTime.Month()) - int(scheduleStartTime.Month())
				if monthDiff%2 != 0 {
					shouldSkip = true
				}
			}
		case models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_QUARTERLY:
			if !matchesFrequencyDay(transactionTime, frequencyValueSet, templateTimeZone) {
				shouldSkip = true
			} else if template.ScheduledStartTime != nil {
				scheduleStartTime := time.Unix(*template.ScheduledStartTime, 0).In(templateTimeZone)
				monthDiff := (int(transactionTime.Year())-int(scheduleStartTime.Year()))*12 + int(transactionTime.Month()) - int(scheduleStartTime.Month())
				if monthDiff%3 != 0 {
					shouldSkip = true
				}
			}
		case models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_SEMIANNUALLY:
			if !matchesFrequencyDay(transactionTime, frequencyValueSet, templateTimeZone) {
				shouldSkip = true
			} else if template.ScheduledStartTime != nil {
				scheduleStartTime := time.Unix(*template.ScheduledStartTime, 0).In(templateTimeZone)
				monthDiff := (int(transactionTime.Year())-int(scheduleStartTime.Year()))*12 + int(transactionTime.Month()) - int(scheduleStartTime.Month())
				if monthDiff%6 != 0 {
					shouldSkip = true
				}
			}
		case models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_ANNUALLY:
			if !matchesFrequencyDay(transactionTime, frequencyValueSet, templateTimeZone) {
				shouldSkip = true
			} else if template.ScheduledStartTime != nil {
				scheduleStartTime := time.Unix(*template.ScheduledStartTime, 0).In(templateTimeZone)
				if transactionTime.Month() != scheduleStartTime.Month() {
					shouldSkip = true
				}
			}
		}

		if shouldSkip {
			skipCount++
			log.Infof(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" does not need to create transaction at this time", template.TemplateId)
			continue
		}

		if template.ScheduledStartTime != nil && *template.ScheduledStartTime > transactionUnixTime {
			skipCount++
			log.Infof(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" does not need to create transaction, now is earlier than the start time %d", template.TemplateId, *template.ScheduledStartTime)
			continue
		}

		if template.ScheduledEndTime != nil && *template.ScheduledEndTime < transactionUnixTime {
			skipCount++
			log.Infof(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" does not need to create transaction, now is later than the end time %d", template.TemplateId, *template.ScheduledEndTime)
			continue
		}

		var transactionDbType models.TransactionDbType

		switch template.Type {
		case models.TRANSACTION_TYPE_EXPENSE:
			transactionDbType = models.TRANSACTION_DB_TYPE_EXPENSE
		case models.TRANSACTION_TYPE_INCOME:
			transactionDbType = models.TRANSACTION_DB_TYPE_INCOME
		case models.TRANSACTION_TYPE_TRANSFER:
			transactionDbType = models.TRANSACTION_DB_TYPE_TRANSFER_OUT
		default:
			skipCount++
			log.Warnf(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" has invalid transaction type", template.TemplateId)
			continue
		}

		transaction := &models.Transaction{
			Uid:               template.Uid,
			Type:              transactionDbType,
			CategoryId:        template.CategoryId,
			TransactionTime:   utils.GetMinTransactionTimeFromUnixTime(transactionTime.Unix()),
			TimezoneUtcOffset: template.ScheduledTimezoneUtcOffset,
			AccountId:         template.AccountId,
			Amount:            template.Amount,
			HideAmount:        template.HideAmount,
			Comment:           template.Comment,
			CreatedIp:         "127.0.0.1",
			ScheduledCreated:  true,
			SourceTemplateId:  template.TemplateId,
		}

		if template.Type == models.TRANSACTION_TYPE_TRANSFER {
			transaction.RelatedAccountId = template.RelatedAccountId
			transaction.RelatedAccountAmount = template.RelatedAccountAmount
		}

		tagIds := template.GetTagIds()
		err = s.CreateTransaction(c, transaction, tagIds, nil)

		if err == nil {
			successCount++
			log.Infof(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" has created a new transaction \"id:%d\"", template.TemplateId, transaction.TransactionId)

			// Copy splits from a recent transaction of this template
			recentTxns, rtErr := s.GetTransactionsByTemplateId(c, template.Uid, template.TemplateId, 1)
			if rtErr == nil && len(recentTxns) > 0 && recentTxns[0].TransactionId != transaction.TransactionId {
				txSplits, tsErr := TransactionSplits.GetSplitsByTransactionId(c, template.Uid, recentTxns[0].TransactionId)
				if tsErr == nil && len(txSplits) > 0 {
					splitReqs := make([]models.TransactionSplitCreateRequest, len(txSplits))
					for si, sp := range txSplits {
						splitReqs[si] = models.TransactionSplitCreateRequest{
							CategoryId: sp.CategoryId,
							Amount:     sp.Amount,
							TagIds:     sp.GetTagIdStringSlice(),
						}
					}
					splitErr := TransactionSplits.CreateSplits(c, template.Uid, transaction.TransactionId, splitReqs)
					if splitErr != nil {
						log.Warnf(c, "[transactions.CreateScheduledTransactions] failed to create splits for scheduled transaction \"id:%d\", because %s", transaction.TransactionId, splitErr.Error())
					}
				}
			}
		} else {
			failedCount++
			log.Errorf(c, "[transactions.CreateScheduledTransactions] transaction template \"id:%d\" failed to create new transaction, because %s", template.TemplateId, err.Error())
		}
	}

	log.Infof(c, "[transactions.CreateScheduledTransactions] %d transactions has been created successfully, %d templates does not need to create transactions and %d transactions failed to create", successCount, skipCount, failedCount)

	return nil
}
