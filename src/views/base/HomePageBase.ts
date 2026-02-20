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
        const dataStart = overviewStore.forecastStartTime;
        const dataEnd = overviewStore.forecastEndTime;
        // Display period may be narrower than data period (e.g., viewing March while data starts from today in Feb)
        const displayStart = overviewStore.forecastDisplayStartTime || dataStart;
        const displayEnd = overviewStore.forecastDisplayEndTime || dataEnd;

        if (!transactions || !dataStart || !dataEnd || !accountsStore.allAccounts || accountsStore.allAccounts.length === 0) {
            return [];
        }

        // Calculate current net balance across all visible accounts
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

        // Use the FULL data range for delta calculation
        const todayStart = getTodayFirstUnixTime();
        const totalDataDays = Math.max(1, Math.ceil((dataEnd - dataStart) / 86400) + 1);

        // Build daily deltas over the FULL data range
        const actualDeltas: Record<number, number> = {};
        const plannedDeltas: Record<number, number> = {};
        const dailyIncomeMap: Record<number, number> = {};
        const dailyExpenseMap: Record<number, number> = {};

        let firstActualTransactionDayIdx = totalDataDays;

        const accountsMap = accountsStore.allAccountsMap;

        for (const txResponse of transactions) {
            const tx = Transaction.of(txResponse);

            if (tx.type === TransactionType.ModifyBalance) {
                continue;
            }

            const dayIdx = Math.floor((tx.time - dataStart) / 86400);
            if (dayIdx < 0 || dayIdx >= totalDataDays) {
                continue;
            }

            let delta = 0;

            if (tx.type === TransactionType.Transfer) {
                // For transfers: compute net effect in default currency
                // Source account loses sourceAmount, destination gains destinationAmount
                const srcAccount = accountsMap[tx.sourceAccountId];
                const dstAccount = accountsMap[tx.destinationAccountId];

                let srcAmountInDefault = tx.sourceAmount;
                if (srcAccount && srcAccount.currency !== currentDefaultCurrency) {
                    const exchanged = exchangeRatesStore.getExchangedAmount(tx.sourceAmount, srcAccount.currency, currentDefaultCurrency);
                    if (isNumber(exchanged)) {
                        srcAmountInDefault = Math.trunc(exchanged as number);
                    }
                }

                let dstAmountInDefault = tx.destinationAmount;
                if (dstAccount && dstAccount.currency !== currentDefaultCurrency) {
                    const exchanged = exchangeRatesStore.getExchangedAmount(tx.destinationAmount, dstAccount.currency, currentDefaultCurrency);
                    if (isNumber(exchanged)) {
                        dstAmountInDefault = Math.trunc(exchanged as number);
                    }
                }

                // Net effect: destination gains - source loses
                delta = dstAmountInDefault - srcAmountInDefault;

                // Show exchange rate difference as income/expense
                if (delta > 0) {
                    dailyIncomeMap[dayIdx] = (dailyIncomeMap[dayIdx] || 0) + delta;
                } else if (delta < 0) {
                    dailyExpenseMap[dayIdx] = (dailyExpenseMap[dayIdx] || 0) + Math.abs(delta);
                }
            } else {
                // Income or Expense
                let amount = tx.sourceAmount;
                const account = accountsMap[tx.sourceAccountId];
                if (account && account.currency !== currentDefaultCurrency) {
                    const exchanged = exchangeRatesStore.getExchangedAmount(amount, account.currency, currentDefaultCurrency);
                    if (isNumber(exchanged)) {
                        amount = Math.trunc(exchanged as number);
                    }
                }

                if (tx.type === TransactionType.Income) {
                    delta = amount;
                    dailyIncomeMap[dayIdx] = (dailyIncomeMap[dayIdx] || 0) + amount;
                } else if (tx.type === TransactionType.Expense) {
                    delta = -amount;
                    dailyExpenseMap[dayIdx] = (dailyExpenseMap[dayIdx] || 0) + amount;
                }
            }

            if (delta !== 0) {
                if (tx.planned) {
                    plannedDeltas[dayIdx] = (plannedDeltas[dayIdx] || 0) + delta;
                } else {
                    actualDeltas[dayIdx] = (actualDeltas[dayIdx] || 0) + delta;
                    if (dayIdx < firstActualTransactionDayIdx) {
                        firstActualTransactionDayIdx = dayIdx;
                    }
                }
            }
        }

        // Today's day index within the DATA range
        const todayDayIdx = Math.floor((todayStart - dataStart) / 86400);
        const effectiveTodayIdx = Math.min(Math.max(todayDayIdx, -1), totalDataDays);

        const hasAnyActualTransactions = firstActualTransactionDayIdx < totalDataDays;

        // Calculate balances over the FULL data range
        const balances: Record<number, number> = {};

        if (effectiveTodayIdx >= 0 && effectiveTodayIdx < totalDataDays) {
            balances[effectiveTodayIdx] = currentBalance;

            for (let d = effectiveTodayIdx - 1; d >= 0; d--) {
                balances[d] = (balances[d + 1] || 0) - (actualDeltas[d + 1] || 0);
            }

            if (hasAnyActualTransactions && firstActualTransactionDayIdx > 0) {
                for (let d = 0; d < firstActualTransactionDayIdx; d++) {
                    balances[d] = 0;
                }
            }

            for (let d = effectiveTodayIdx + 1; d < totalDataDays; d++) {
                balances[d] = (balances[d - 1] || 0) + (actualDeltas[d] || 0) + (plannedDeltas[d] || 0);
            }
        } else if (effectiveTodayIdx >= totalDataDays) {
            balances[totalDataDays - 1] = currentBalance;
            for (let d = totalDataDays - 2; d >= 0; d--) {
                balances[d] = (balances[d + 1] || 0) - (actualDeltas[d + 1] || 0);
            }
        } else {
            balances[0] = currentBalance;
            for (let d = 1; d < totalDataDays; d++) {
                balances[d] = (balances[d - 1] || 0) + (actualDeltas[d] || 0) + (plannedDeltas[d] || 0);
            }
        }

        // Now extract only the DISPLAY range from the full balances
        const displayStartDayIdx = Math.max(0, Math.floor((displayStart - dataStart) / 86400));
        const displayTotalDays = Math.max(1, Math.ceil((displayEnd - displayStart) / 86400) + 1);

        const result: { date: string; dateLabel: string; balance: number; isFuture: boolean; dailyIncome: number; dailyExpense: number }[] = [];
        for (let d = 0; d < displayTotalDays; d++) {
            const fullIdx = displayStartDayIdx + d;
            const dayUnixTime = displayStart + d * 86400;
            const dayDateTime = parseDateTimeFromUnixTime(dayUnixTime);
            const longLabel = formatDateTimeToLongMonthDay(dayDateTime);
            const ymd = dayDateTime.toGregorianCalendarYearMonthDay();
            const dayStr = ymd.day < 10 ? '0' + ymd.day : String(ymd.day);
            const monthNum = ymd.month;
            const monthStr = monthNum < 10 ? '0' + monthNum : String(monthNum);

            result.push({
                date: dayStr + '.' + monthStr,
                dateLabel: longLabel,
                balance: balances[fullIdx] || 0,
                isFuture: fullIdx > effectiveTodayIdx,
                dailyIncome: dailyIncomeMap[fullIdx] || 0,
                dailyExpense: dailyExpenseMap[fullIdx] || 0
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
