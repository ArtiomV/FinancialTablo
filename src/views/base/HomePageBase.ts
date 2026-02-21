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
        const dataStart = overviewStore.forecastStartTime;
        const dataEnd = overviewStore.forecastEndTime;
        // Display period may be narrower than data period (e.g., viewing March while data starts from today in Feb)
        const displayStart = overviewStore.forecastDisplayStartTime || dataStart;
        const displayEnd = overviewStore.forecastDisplayEndTime || dataEnd;

        if (!transactions || !dataStart || !dataEnd || !accountsStore.allAccounts || accountsStore.allAccounts.length === 0) {
            return [];
        }

        const currentDefaultCurrency = userStore.currentUserDefaultCurrency;
        const accountsMap = accountsStore.allAccountsMap;
        const excludedAccountIds = settingsStore.appSettings.totalAmountExcludeAccountIds || {};

        // Use the FULL data range for delta calculation
        const todayStart = getTodayFirstUnixTime();
        const totalDataDays = Math.max(1, Math.ceil((dataEnd - dataStart) / 86400) + 1);

        // ====================================================================
        // PER-ACCOUNT running balances in NATIVE currencies.
        // This avoids the distortion caused by converting all historical
        // transactions at current exchange rates. Each account accumulates
        // income/expense/transfers in its own currency, and only at display
        // time do we convert the per-account balance to the default currency.
        // ====================================================================

        // Per-account daily deltas: accountId → dayIdx → delta (in native currency cents)
        const accountDeltas: Record<string, Record<number, number>> = {};
        // Daily income/expense in default currency for tooltip display
        const dailyIncomeMap: Record<number, number> = {};
        const dailyExpenseMap: Record<number, number> = {};

        // Helper to get or create account delta map
        const getAccountDeltas = (accountId: string): Record<number, number> => {
            if (!accountDeltas[accountId]) {
                accountDeltas[accountId] = {};
            }
            return accountDeltas[accountId]!;
        };

        for (const txResponse of transactions) {
            const tx = Transaction.of(txResponse);

            if (tx.type === TransactionType.ModifyBalance) {
                continue;
            }

            // Skip transactions from excluded accounts
            if (tx.type === TransactionType.Transfer) {
                if (excludedAccountIds[tx.sourceAccountId] && excludedAccountIds[tx.destinationAccountId]) {
                    continue;
                }
            } else {
                if (excludedAccountIds[tx.sourceAccountId]) {
                    continue;
                }
            }

            const dayIdx = Math.floor((tx.time - dataStart) / 86400);
            if (dayIdx < 0 || dayIdx >= totalDataDays) {
                continue;
            }

            if (tx.type === TransactionType.Transfer) {
                const srcExcluded = !!excludedAccountIds[tx.sourceAccountId];
                const dstExcluded = !!excludedAccountIds[tx.destinationAccountId];

                // Source account loses sourceAmount (in source account's native currency)
                if (!srcExcluded) {
                    const srcDeltas = getAccountDeltas(tx.sourceAccountId);
                    srcDeltas[dayIdx] = (srcDeltas[dayIdx] || 0) - tx.sourceAmount;
                }
                // Destination account gains destinationAmount (in destination account's native currency)
                if (!dstExcluded) {
                    const dstDeltas = getAccountDeltas(tx.destinationAccountId);
                    dstDeltas[dayIdx] = (dstDeltas[dayIdx] || 0) + tx.destinationAmount;
                }

                // For tooltip: convert to default currency for income/expense display
                const srcAccount = accountsMap[tx.sourceAccountId];
                const dstAccount = accountsMap[tx.destinationAccountId];
                let srcInDefault = tx.sourceAmount;
                let dstInDefault = tx.destinationAmount;
                if (srcAccount && srcAccount.currency !== currentDefaultCurrency) {
                    const ex = exchangeRatesStore.getExchangedAmount(tx.sourceAmount, srcAccount.currency, currentDefaultCurrency);
                    if (isNumber(ex)) srcInDefault = Math.trunc(ex as number);
                }
                if (dstAccount && dstAccount.currency !== currentDefaultCurrency) {
                    const ex = exchangeRatesStore.getExchangedAmount(tx.destinationAmount, dstAccount.currency, currentDefaultCurrency);
                    if (isNumber(ex)) dstInDefault = Math.trunc(ex as number);
                }
                let netDefault = 0;
                if (srcExcluded) netDefault = dstInDefault;
                else if (dstExcluded) netDefault = -srcInDefault;
                else netDefault = dstInDefault - srcInDefault;

                if (netDefault > 0) {
                    dailyIncomeMap[dayIdx] = (dailyIncomeMap[dayIdx] || 0) + netDefault;
                } else if (netDefault < 0) {
                    dailyExpenseMap[dayIdx] = (dailyExpenseMap[dayIdx] || 0) + Math.abs(netDefault);
                }
            } else {
                // Income or Expense — affects the source account in its native currency
                const acctDeltas = getAccountDeltas(tx.sourceAccountId);
                if (tx.type === TransactionType.Income) {
                    acctDeltas[dayIdx] = (acctDeltas[dayIdx] || 0) + tx.sourceAmount;
                } else if (tx.type === TransactionType.Expense) {
                    acctDeltas[dayIdx] = (acctDeltas[dayIdx] || 0) - tx.sourceAmount;
                }

                // For tooltip: convert to default currency
                const account = accountsMap[tx.sourceAccountId];
                let amountDefault = tx.sourceAmount;
                if (account && account.currency !== currentDefaultCurrency) {
                    const ex = exchangeRatesStore.getExchangedAmount(tx.sourceAmount, account.currency, currentDefaultCurrency);
                    if (isNumber(ex)) amountDefault = Math.trunc(ex as number);
                }
                if (tx.type === TransactionType.Income) {
                    dailyIncomeMap[dayIdx] = (dailyIncomeMap[dayIdx] || 0) + amountDefault;
                } else if (tx.type === TransactionType.Expense) {
                    dailyExpenseMap[dayIdx] = (dailyExpenseMap[dayIdx] || 0) + amountDefault;
                }
            }
        }

        // Build per-account cumulative balances, then convert each day to default currency
        // and sum across all accounts to get the total balance per day.
        const balances: Record<number, number> = {};

        // Collect all account IDs that have deltas
        const allAccountIds = Object.keys(accountDeltas);

        // For each account, compute cumulative running balance (native currency)
        // then convert to default currency and add to total per-day balance
        for (const accountId of allAccountIds) {
            const deltas = accountDeltas[accountId]!;
            const account = accountsMap[accountId];
            const currency = account?.currency || currentDefaultCurrency;
            const isLiability = account?.isLiability ?? false;

            let cumulativeNative = 0;
            for (let d = 0; d < totalDataDays; d++) {
                cumulativeNative += deltas[d] || 0;

                // Convert native balance to default currency
                let balanceInDefault: number;
                if (currency === currentDefaultCurrency) {
                    balanceInDefault = cumulativeNative;
                } else {
                    const exchanged = exchangeRatesStore.getExchangedAmount(Math.abs(cumulativeNative), currency, currentDefaultCurrency);
                    if (isNumber(exchanged)) {
                        balanceInDefault = cumulativeNative >= 0 ? Math.trunc(exchanged as number) : -Math.trunc(exchanged as number);
                    } else {
                        balanceInDefault = cumulativeNative; // fallback: treat as default currency
                    }
                }

                // Liabilities subtract from net worth
                if (isLiability) {
                    balances[d] = (balances[d] ?? 0) - balanceInDefault;
                } else {
                    balances[d] = (balances[d] ?? 0) + balanceInDefault;
                }
            }
        }

        // Today's day index within the DATA range
        const todayDayIdx = Math.floor((todayStart - dataStart) / 86400);
        const effectiveTodayIdx = Math.min(Math.max(todayDayIdx, -1), totalDataDays - 1);

        // Now extract only the DISPLAY range from the full balances
        const displayTotalDays = Math.max(1, Math.floor((displayEnd - displayStart) / 86400) + 1);
        const displayStartDayIdx = Math.max(0, Math.floor((displayStart - dataStart) / 86400));

        // For long display periods (> 90 days), aggregate by month for a cleaner chart.
        // Show the balance at the END of each month (last day), with monthly income/expense totals.
        const useMonthlyAggregation = displayTotalDays > 90;

        const result: { date: string; dateLabel: string; balance: number; isFuture: boolean; dailyIncome: number; dailyExpense: number }[] = [];

        if (useMonthlyAggregation) {
            // Monthly aggregation: one data point per month (last day of each month)
            let currentMonth = -1;
            let currentYear = -1;
            let monthIncome = 0;
            let monthExpense = 0;
            let lastFullIdxInMonth = displayStartDayIdx;
            let lastDayUnixInMonth = displayStart;
            let lastIsFutureInMonth = false;

            for (let d = 0; d < displayTotalDays; d++) {
                const fullIdx = displayStartDayIdx + d;
                const dayUnixTime = displayStart + d * 86400;
                const dayDateTime = parseDateTimeFromUnixTime(dayUnixTime);
                const ymd = dayDateTime.toGregorianCalendarYearMonthDay();

                // When month changes, emit the previous month's data point
                if ((ymd.month !== currentMonth || ymd.year !== currentYear) && currentMonth !== -1) {
                    const lastDayDateTime = parseDateTimeFromUnixTime(lastDayUnixInMonth);
                    const longLabel = formatDateTimeToLongMonthDay(lastDayDateTime);
                    const lastYmd = lastDayDateTime.toGregorianCalendarYearMonthDay();
                    const monthStr = lastYmd.month < 10 ? '0' + lastYmd.month : String(lastYmd.month);
                    const clampedMonthIdx = Math.min(lastFullIdxInMonth, totalDataDays - 1);

                    result.push({
                        date: monthStr + '.' + String(lastYmd.year).slice(2),
                        dateLabel: longLabel,
                        balance: balances[clampedMonthIdx] ?? 0,
                        isFuture: lastIsFutureInMonth,
                        dailyIncome: monthIncome,
                        dailyExpense: monthExpense
                    });

                    monthIncome = 0;
                    monthExpense = 0;
                }

                currentMonth = ymd.month;
                currentYear = ymd.year;
                lastFullIdxInMonth = fullIdx;
                lastDayUnixInMonth = dayUnixTime;
                lastIsFutureInMonth = fullIdx > effectiveTodayIdx;
                monthIncome += dailyIncomeMap[fullIdx] || 0;
                monthExpense += dailyExpenseMap[fullIdx] || 0;
            }

            // Flush the last accumulated month (handles mid-month range end)
            if (currentMonth !== -1) {
                const lastDayDateTime = parseDateTimeFromUnixTime(lastDayUnixInMonth);
                const longLabel = formatDateTimeToLongMonthDay(lastDayDateTime);
                const lastYmd = lastDayDateTime.toGregorianCalendarYearMonthDay();
                const monthStr = lastYmd.month < 10 ? '0' + lastYmd.month : String(lastYmd.month);
                const clampedMonthIdx = Math.min(lastFullIdxInMonth, totalDataDays - 1);

                result.push({
                    date: monthStr + '.' + String(lastYmd.year).slice(2),
                    dateLabel: longLabel,
                    balance: balances[clampedMonthIdx] ?? 0,
                    isFuture: lastIsFutureInMonth,
                    dailyIncome: monthIncome,
                    dailyExpense: monthExpense
                });
            }
        } else {
            // Daily: one data point per day (original behavior for short periods)
            for (let d = 0; d < displayTotalDays; d++) {
                const fullIdx = displayStartDayIdx + d;
                const dayUnixTime = displayStart + d * 86400;
                const dayDateTime = parseDateTimeFromUnixTime(dayUnixTime);
                const longLabel = formatDateTimeToLongMonthDay(dayDateTime);
                const ymd = dayDateTime.toGregorianCalendarYearMonthDay();
                const dayStr = ymd.day < 10 ? '0' + ymd.day : String(ymd.day);
                const monthNum = ymd.month;
                const monthStr = monthNum < 10 ? '0' + monthNum : String(monthNum);

                // Clamp fullIdx to valid balance range to avoid reading undefined → 0
                const clampedIdx = Math.min(fullIdx, totalDataDays - 1);
                result.push({
                    date: dayStr + '.' + monthStr,
                    dateLabel: longLabel,
                    balance: balances[clampedIdx] ?? 0,
                    isFuture: fullIdx > effectiveTodayIdx,
                    dailyIncome: dailyIncomeMap[fullIdx] || 0,
                    dailyExpense: dailyExpenseMap[fullIdx] || 0
                });
            }
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
