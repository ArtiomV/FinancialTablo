import { computed } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

import { useSettingsStore } from '@/stores/setting.ts';
import { useUserStore } from '@/stores/user.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useOverviewStore } from '@/stores/overview.ts';
import { useExchangeRatesStore } from '@/stores/exchangeRates.ts';

import type { HiddenAmount, NumberWithSuffix } from '@/core/numeral.ts';
import { DISPLAY_HIDDEN_AMOUNT } from '@/consts/numeral.ts';
import { TransactionType } from '@/core/transaction.ts';

import { Account } from '@/models/account.ts';
import { Transaction } from '@/models/transaction.ts';
import type {
    TransactionOverviewResponse,
    TransactionOverviewDisplayTime,
    TransactionOverviewResponseItem
} from '@/models/transaction.ts';

import {
    parseDateTimeFromUnixTime,
    getTodayFirstUnixTime
} from '@/lib/datetime.ts';
import { getAllFilteredAccountsBalance } from '@/lib/account.ts';
import { isNumber } from '@/lib/common.ts';

export function useHomePageBase() {
    const {
        formatDateTimeToLongDate,
        formatDateTimeToLongMonthDay,
        formatDateTimeToGregorianLikeLongYear,
        formatDateTimeToGregorianLikeLongMonth,
        formatAmountToLocalizedNumeralsWithCurrency
    } = useI18n();

    const settingsStore = useSettingsStore();
    const userStore = useUserStore();
    const accountsStore = useAccountsStore();
    const overviewStore = useOverviewStore();

    const showAmountInHomePage = computed<boolean>({
        get: () => settingsStore.appSettings.showAmountInHomePage,
        set: (value) => settingsStore.setShowAmountInHomePage(value)
    });

    const defaultCurrency = computed<string>(() => userStore.currentUserDefaultCurrency);
    const allAccounts = computed<Account[]>(() => accountsStore.allAccounts);

    const netAssets = computed<string>(() => {
        const netAssets: number | HiddenAmount | NumberWithSuffix = accountsStore.getNetAssets(showAmountInHomePage.value);
        return formatAmountToLocalizedNumeralsWithCurrency(netAssets, defaultCurrency.value);
    });

    const totalAssets = computed<string>(() => {
        const totalAssets: number | HiddenAmount | NumberWithSuffix = accountsStore.getTotalAssets(showAmountInHomePage.value);
        return formatAmountToLocalizedNumeralsWithCurrency(totalAssets, defaultCurrency.value);
    });

    const totalLiabilities = computed<string>(() => {
        const totalLiabilities: number | HiddenAmount | NumberWithSuffix = accountsStore.getTotalLiabilities(showAmountInHomePage.value);
        return formatAmountToLocalizedNumeralsWithCurrency(totalLiabilities, defaultCurrency.value);
    });

    const displayDateRange = computed<TransactionOverviewDisplayTime>(() => {
        return {
            today: {
                displayTime: formatDateTimeToLongDate(parseDateTimeFromUnixTime(overviewStore.transactionDataRange.today.startTime)),
            },
            thisWeek: {
                startTime: formatDateTimeToLongMonthDay(parseDateTimeFromUnixTime(overviewStore.transactionDataRange.thisWeek.startTime)),
                endTime: formatDateTimeToLongMonthDay(parseDateTimeFromUnixTime(overviewStore.transactionDataRange.thisWeek.endTime))
            },
            thisMonth: {
                displayTime: formatDateTimeToGregorianLikeLongMonth(parseDateTimeFromUnixTime(overviewStore.transactionDataRange.thisMonth.startTime)),
                startTime: formatDateTimeToLongMonthDay(parseDateTimeFromUnixTime(overviewStore.transactionDataRange.thisMonth.startTime)),
                endTime: formatDateTimeToLongMonthDay(parseDateTimeFromUnixTime(overviewStore.transactionDataRange.thisMonth.endTime))
            },
            thisYear: {
                displayTime: formatDateTimeToGregorianLikeLongYear(parseDateTimeFromUnixTime(overviewStore.transactionDataRange.thisYear.startTime))
            }
        };
    });

    const transactionOverview = computed<TransactionOverviewResponse>(() => overviewStore.transactionOverview);

    const exchangeRatesStore = useExchangeRatesStore();

    // Daily balance forecast: compute cumulative balance per day for the loaded period
    const dailyBalanceForecast = computed<{ date: string; dateLabel: string; balance: number; isFuture: boolean; dailyIncome: number; dailyExpense: number }[]>(() => {
        const transactions = overviewStore.monthlyTransactionsForForecast;
        const periodStart = overviewStore.forecastStartTime;
        const periodEnd = overviewStore.forecastEndTime;

        if (!transactions || !periodStart || !periodEnd || !accountsStore.allAccounts || accountsStore.allAccounts.length === 0) {
            return [];
        }

        // Calculate current net balance across all visible accounts (same logic as getNetAssets but returns raw number)
        const currentDefaultCurrency = userStore.currentUserDefaultCurrency;
        const accountsBalance = getAllFilteredAccountsBalance(
            accountsStore.allCategorizedAccountsMap,
            settingsStore.appSettings.accountCategoryOrders,
            (account: Account) => !settingsStore.appSettings.totalAmountExcludeAccountIds[account.id]
        );

        let currentBalance = 0;
        for (const ab of accountsBalance) {
            if (ab.isLiability) {
                if (ab.currency === currentDefaultCurrency) {
                    currentBalance -= ab.balance;
                } else {
                    const exchanged = exchangeRatesStore.getExchangedAmount(ab.balance, ab.currency, currentDefaultCurrency);
                    if (isNumber(exchanged)) {
                        currentBalance -= Math.trunc(exchanged as number);
                    }
                }
            } else {
                if (ab.currency === currentDefaultCurrency) {
                    currentBalance += ab.balance;
                } else {
                    const exchanged = exchangeRatesStore.getExchangedAmount(ab.balance, ab.currency, currentDefaultCurrency);
                    if (isNumber(exchanged)) {
                        currentBalance += Math.trunc(exchanged as number);
                    }
                }
            }
        }

        // Compute total number of days in the display period
        const todayStart = getTodayFirstUnixTime();
        const totalDays = Math.max(1, Math.ceil((periodEnd - periodStart) / 86400) + 1);

        // Build daily deltas for the display period, plus post-period and pre-period deltas
        // for correct balance calculation
        const actualDeltas: Record<number, number> = {};
        const plannedDeltas: Record<number, number> = {};
        const dailyIncomeMap: Record<number, number> = {};
        const dailyExpenseMap: Record<number, number> = {};

        // Sum of actual deltas for transactions AFTER the display period (up to today)
        let postPeriodActualDelta = 0;
        // Sum of PLANNED deltas for transactions BEFORE the display period (between today and period start)
        // Only planned transactions matter here — actual ones are already in currentBalance
        let prePeriodDelta = 0;

        const accountsMap = accountsStore.allAccountsMap;

        for (const txResponse of transactions) {
            const tx = Transaction.of(txResponse);

            // Skip transfers and ModifyBalance — net effect is 0 for transfers, skip modify balance
            if (tx.type === TransactionType.Transfer || tx.type === TransactionType.ModifyBalance) {
                continue;
            }

            // Skip planned transactions for actual balance calculations
            if (tx.planned) {
                const dayIdx = Math.floor((tx.time - periodStart) / 86400);
                let amount = tx.sourceAmount;
                const account = accountsMap[tx.sourceAccountId];
                if (account && account.currency !== currentDefaultCurrency) {
                    const exchanged = exchangeRatesStore.getExchangedAmount(amount, account.currency, currentDefaultCurrency);
                    if (isNumber(exchanged)) {
                        amount = Math.trunc(exchanged as number);
                    }
                }
                let delta = 0;
                if (tx.type === TransactionType.Income) {
                    delta = amount;
                } else if (tx.type === TransactionType.Expense) {
                    delta = -amount;
                }

                if (dayIdx >= 0 && dayIdx < totalDays) {
                    // Planned transaction within the display period
                    plannedDeltas[dayIdx] = (plannedDeltas[dayIdx] || 0) + delta;
                } else if (dayIdx < 0) {
                    // Planned transaction BEFORE the display period (between today and period start)
                    prePeriodDelta += delta;
                }
                continue;
            }

            // Convert amount to default currency if needed
            let amount = tx.sourceAmount;
            const account = accountsMap[tx.sourceAccountId];
            if (account && account.currency !== currentDefaultCurrency) {
                const exchanged = exchangeRatesStore.getExchangedAmount(amount, account.currency, currentDefaultCurrency);
                if (isNumber(exchanged)) {
                    amount = Math.trunc(exchanged as number);
                }
            }

            // Compute delta for this transaction
            let delta = 0;
            if (tx.type === TransactionType.Income) {
                delta = amount;
            } else if (tx.type === TransactionType.Expense) {
                delta = -amount;
            }

            // Determine which day index this transaction belongs to
            const dayIdx = Math.floor((tx.time - periodStart) / 86400);

            if (dayIdx >= 0 && dayIdx < totalDays) {
                // Transaction is within the display period
                actualDeltas[dayIdx] = (actualDeltas[dayIdx] || 0) + delta;

                // Track income/expense for tooltip (only actual, not planned)
                if (tx.type === TransactionType.Income) {
                    dailyIncomeMap[dayIdx] = (dailyIncomeMap[dayIdx] || 0) + amount;
                } else if (tx.type === TransactionType.Expense) {
                    dailyExpenseMap[dayIdx] = (dailyExpenseMap[dayIdx] || 0) + amount;
                }
            } else if (dayIdx >= totalDays) {
                // Transaction is AFTER the display period (between periodEnd and today)
                postPeriodActualDelta += delta;
            }
            // NOTE: actual (non-planned) transactions with dayIdx < 0 are NOT added to prePeriodDelta
            // because they are already reflected in currentBalance (account balances include them).
            // Only planned transactions contribute to prePeriodDelta (handled above).
        }

        // Today's day index within the period
        const todayDayIdx = Math.floor((todayStart - periodStart) / 86400);
        // Clamp to valid range: if today is before period, use -1; if after, use totalDays
        const effectiveTodayIdx = Math.min(Math.max(todayDayIdx, -1), totalDays);

        // currentBalance reflects all actual (non-planned) transactions up to now
        const balances: Record<number, number> = {};

        if (effectiveTodayIdx >= 0 && effectiveTodayIdx < totalDays) {
            // Today is within the display period
            balances[effectiveTodayIdx] = currentBalance;

            // Go backwards from today to start of period
            for (let d = effectiveTodayIdx - 1; d >= 0; d--) {
                balances[d] = (balances[d + 1] || 0) - (actualDeltas[d + 1] || 0);
            }

            // Go forwards from today to end of period
            for (let d = effectiveTodayIdx + 1; d < totalDays; d++) {
                balances[d] = (balances[d - 1] || 0) + (actualDeltas[d] || 0) + (plannedDeltas[d] || 0);
            }
        } else if (effectiveTodayIdx >= totalDays) {
            // Today is AFTER this period — all days are in the past
            // Correct balance at end of period = currentBalance - postPeriodActualDelta
            // (postPeriodActualDelta is the sum of changes between periodEnd and today)
            balances[totalDays - 1] = currentBalance - postPeriodActualDelta;
            for (let d = totalDays - 2; d >= 0; d--) {
                balances[d] = (balances[d + 1] || 0) - (actualDeltas[d + 1] || 0);
            }
        } else {
            // Today is BEFORE this period — all days are in the future
            // Balance at start of period = currentBalance + all deltas between today and period start
            // prePeriodDelta includes both actual and planned transactions between today and periodStart
            balances[0] = currentBalance + prePeriodDelta;
            for (let d = 1; d < totalDays; d++) {
                balances[d] = (balances[d - 1] || 0) + (actualDeltas[d] || 0) + (plannedDeltas[d] || 0);
            }
        }

        // Build result array
        const result: { date: string; dateLabel: string; balance: number; isFuture: boolean; dailyIncome: number; dailyExpense: number }[] = [];
        for (let d = 0; d < totalDays; d++) {
            const dayUnixTime = periodStart + d * 86400;
            const dayDateTime = parseDateTimeFromUnixTime(dayUnixTime);
            const longLabel = formatDateTimeToLongMonthDay(dayDateTime);
            const ymd = dayDateTime.toGregorianCalendarYearMonthDay();
            const dayStr = ymd.day < 10 ? '0' + ymd.day : String(ymd.day);
            const monthNum = ymd.month;
            const monthStr = monthNum < 10 ? '0' + monthNum : String(monthNum);

            result.push({
                date: dayStr + '.' + monthStr,
                dateLabel: longLabel,
                balance: balances[d] || 0,
                isFuture: d > effectiveTodayIdx,
                dailyIncome: dailyIncomeMap[d] || 0,
                dailyExpense: dailyExpenseMap[d] || 0
            });
        }

        return result;
    });

    function getDisplayAmount(amount: number): string {
        if (!showAmountInHomePage.value) {
            return formatAmountToLocalizedNumeralsWithCurrency(DISPLAY_HIDDEN_AMOUNT, defaultCurrency.value);
        }

        return formatAmountToLocalizedNumeralsWithCurrency(amount, defaultCurrency.value);
    }

    function getDisplayIncomeAmount(category: TransactionOverviewResponseItem): string {
        return getDisplayAmount(category.incomeAmount);
    }

    function getDisplayExpenseAmount(category: TransactionOverviewResponseItem): string {
        return getDisplayAmount(category.expenseAmount);
    }

    return {
        // computed states
        showAmountInHomePage,
        defaultCurrency,
        allAccounts,
        netAssets,
        totalAssets,
        totalLiabilities,
        displayDateRange,
        transactionOverview,
        dailyBalanceForecast,
        // functions
        getDisplayAmount,
        getDisplayIncomeAmount,
        getDisplayExpenseAmount
    };
}
