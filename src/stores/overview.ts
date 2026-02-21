import { ref, computed } from 'vue';
import { defineStore } from 'pinia';

import { useSettingsStore } from './setting.ts';
import { useUserStore } from './user.ts';
import { useAccountsStore } from './account.ts';
import { useTransactionCategoriesStore } from './transactionCategory.ts';
import { useExchangeRatesStore } from './exchangeRates.ts';

import { type WritableStartEndTime, DateRange } from '@/core/datetime.ts';
import { TimezoneTypeForStatistics } from '@/core/timezone.ts';
import type { TransactionType } from '@/core/transaction.ts';

import type {
    TransactionAmountsRequestType,
    TransactionAmountsRequestParams,
    TransactionAmountsResponse,
    TransactionOverviewResponse,
    TransactionInfoResponse
} from '@/models/transaction.ts';
import { ALL_TRANSACTION_AMOUNTS_REQUEST_TYPE } from '@/models/transaction.ts';

import {
    isDefined,
    isNumber,
    isEquals,
    isObjectEmpty,
    objectFieldWithValueToArrayItem
} from '@/lib/common.ts';
import {
    getUnixTimeBeforeUnixTime,
    getTodayFirstUnixTime,
    getTodayLastUnixTime,
    getThisWeekFirstUnixTime,
    getThisWeekLastUnixTime,
    getThisMonthFirstUnixTime,
    getThisMonthLastUnixTime,
    getThisYearFirstUnixTime,
    getThisYearLastUnixTime
} from '@/lib/datetime.ts';
import { getFinalAccountIdsByFilteredAccountIds } from '@/lib/account.ts';
import { getFinalCategoryIdsByFilteredCategoryIds } from '@/lib/category.ts';
import logger from '@/lib/logger.ts';
import services from '@/lib/services.ts';

interface TransactionDataRange extends Record<TransactionAmountsRequestType, WritableStartEndTime> {
    today: {
        startTime: number;
        endTime: number;
    };
    thisWeek: {
        startTime: number;
        endTime: number;
    };
    thisMonth: {
        startTime: number;
        endTime: number;
    };
    thisYear: {
        startTime: number;
        endTime: number;
    };
    lastMonth: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLastMonth: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast2Months: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast3Months: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast4Months: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast5Months: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast6Months: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast7Months: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast8Months: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast9Months: {
        startTime: number;
        endTime: number;
    };
    monthBeforeLast10Months: {
        startTime: number;
        endTime: number;
    };
}

interface TransactionOverviewOptions {
    loadLast11Months: boolean;
}

export const useOverviewStore = defineStore('overview', () => {
    const settingsStore = useSettingsStore();
    const userStore = useUserStore();
    const accountsStore = useAccountsStore();
    const transactionCategoriesStore = useTransactionCategoriesStore();
    const exchangeRatesStore = useExchangeRatesStore();

    const transactionDataRange = ref<TransactionDataRange>(getTransactionDateRange());

    const transactionOverviewOptions = ref<TransactionOverviewOptions>({
        loadLast11Months: false
    });

    const transactionOverviewData = ref<TransactionAmountsResponse>({});
    const transactionOverviewStateInvalid = ref<boolean>(true);

    const monthlyTransactionsForForecast = ref<TransactionInfoResponse[]>([]);
    const monthlyTransactionsForForecastLoaded = ref<boolean>(false);
    const forecastStartTime = ref<number>(0);
    const forecastEndTime = ref<number>(0);
    const forecastDisplayStartTime = ref<number>(0);
    const forecastDisplayEndTime = ref<number>(0);

    const transactionOverview = computed<TransactionOverviewResponse>(() => {
        const overviewData = transactionOverviewData.value;

        if (!overviewData || !overviewData.thisMonth) {
            return {
                thisMonth: {
                    valid: false,
                    incomeAmount: 0,
                    expenseAmount: 0,
                    incompleteIncomeAmount: false,
                    incompleteExpenseAmount: false
                }
            } as TransactionOverviewResponse;
        }

        const finalOverviewData: TransactionOverviewResponse = {};
        const defaultCurrency = userStore.currentUserDefaultCurrency;

        ALL_TRANSACTION_AMOUNTS_REQUEST_TYPE.forEach(field => {
            const item = overviewData[field];

            if (!item) {
                return;
            }

            let totalIncomeAmount = 0;
            let totalExpenseAmount = 0;

            if (item.amounts) {
                for (const amount of item.amounts) {
                    if (amount.currency !== defaultCurrency) {
                        const incomeAmount = exchangeRatesStore.getExchangedAmount(amount.incomeAmount, amount.currency, defaultCurrency);
                        const expenseAmount = exchangeRatesStore.getExchangedAmount(amount.expenseAmount, amount.currency, defaultCurrency);

                        if (isNumber(incomeAmount)) {
                            totalIncomeAmount += Math.trunc(incomeAmount);
                        }

                        if (isNumber(expenseAmount)) {
                            totalExpenseAmount += Math.trunc(expenseAmount);
                        }
                    } else {
                        totalIncomeAmount += amount.incomeAmount;
                        totalExpenseAmount += amount.expenseAmount;
                    }
                }
            }

            finalOverviewData[field] = {
                valid: true,
                incomeAmount: totalIncomeAmount,
                expenseAmount: totalExpenseAmount,
                incompleteIncomeAmount: false,
                incompleteExpenseAmount: false,
                amounts: item.amounts || []
            };
        });

        return finalOverviewData;
    });

    function getTransactionDateRange(): TransactionDataRange {
        const dateRange: TransactionDataRange = {
            today: { startTime: 0, endTime: 0 },
            thisWeek: { startTime: 0, endTime: 0 },
            thisMonth: { startTime: 0, endTime: 0 },
            thisYear: { startTime: 0, endTime: 0 },
            lastMonth: { startTime: 0, endTime: 0 },
            monthBeforeLastMonth: { startTime: 0, endTime: 0 },
            monthBeforeLast2Months: { startTime: 0, endTime: 0 },
            monthBeforeLast3Months: { startTime: 0, endTime: 0 },
            monthBeforeLast4Months: { startTime: 0, endTime: 0 },
            monthBeforeLast5Months: { startTime: 0, endTime: 0 },
            monthBeforeLast6Months: { startTime: 0, endTime: 0 },
            monthBeforeLast7Months: { startTime: 0, endTime: 0 },
            monthBeforeLast8Months: { startTime: 0, endTime: 0 },
            monthBeforeLast9Months: { startTime: 0, endTime: 0 },
            monthBeforeLast10Months: { startTime: 0, endTime: 0 }
        };

        initTransactionDateRange(dateRange);
        return dateRange;
    }

    function initTransactionDateRange(dateRange: TransactionDataRange): void {
        dateRange.today.startTime = getTodayFirstUnixTime();
        dateRange.today.endTime = getTodayLastUnixTime();

        dateRange.thisWeek.startTime = getThisWeekFirstUnixTime(userStore.currentUserFirstDayOfWeek);
        dateRange.thisWeek.endTime = getThisWeekLastUnixTime(userStore.currentUserFirstDayOfWeek);

        dateRange.thisMonth.startTime = getThisMonthFirstUnixTime();
        dateRange.thisMonth.endTime = getThisMonthLastUnixTime();

        dateRange.thisYear.startTime = getThisYearFirstUnixTime();
        dateRange.thisYear.endTime = getThisYearLastUnixTime();

        dateRange.lastMonth.startTime = getUnixTimeBeforeUnixTime(getThisMonthFirstUnixTime(), 1, 'months');
        dateRange.lastMonth.endTime = getUnixTimeBeforeUnixTime(getThisMonthFirstUnixTime(), 1, 'seconds');

        dateRange.monthBeforeLastMonth.startTime = getUnixTimeBeforeUnixTime(dateRange.lastMonth.startTime, 1, 'months');
        dateRange.monthBeforeLastMonth.endTime = getUnixTimeBeforeUnixTime(dateRange.lastMonth.startTime, 1, 'seconds');

        dateRange.monthBeforeLast2Months.startTime = getUnixTimeBeforeUnixTime(dateRange.monthBeforeLastMonth.startTime, 1, 'months');
        dateRange.monthBeforeLast2Months.endTime = getUnixTimeBeforeUnixTime(dateRange.monthBeforeLastMonth.startTime, 1, 'seconds');

        for (let i = 3; i <= 10; i++) {
            dateRange[`monthBeforeLast${i}Months` as TransactionAmountsRequestType].startTime = getUnixTimeBeforeUnixTime(dateRange[`monthBeforeLast${i - 1}Months` as TransactionAmountsRequestType].startTime, 1, 'months');
            dateRange[`monthBeforeLast${i}Months` as TransactionAmountsRequestType].endTime = getUnixTimeBeforeUnixTime(dateRange[`monthBeforeLast${i - 1}Months` as TransactionAmountsRequestType].startTime, 1, 'seconds');
        }
    }

    function updateTransactionDateRange(): void {
        initTransactionDateRange(transactionDataRange.value);
    }

    function updateTransactionOverviewInvalidState(invalidState: boolean): void {
        transactionOverviewStateInvalid.value = invalidState;
    }

    function resetTransactionOverview(): void {
        updateTransactionDateRange();
        transactionOverviewOptions.value.loadLast11Months = false;
        transactionOverviewData.value = {};
        transactionOverviewStateInvalid.value = true;
        monthlyTransactionsForForecast.value = [];
        monthlyTransactionsForForecastLoaded.value = false;
        forecastStartTime.value = 0;
        forecastEndTime.value = 0;
        forecastDisplayStartTime.value = 0;
        forecastDisplayEndTime.value = 0;
    }

    function loadTransactionOverview({ force, loadLast11Months }: { force: boolean, loadLast11Months?: boolean }): Promise<TransactionAmountsResponse> {
        let dateChanged = false;
        let rangeChanged = false;

        if (transactionDataRange.value.today.startTime !== getTodayFirstUnixTime()) {
            dateChanged = true;
            updateTransactionDateRange();
        }

        if (loadLast11Months && !transactionOverviewOptions.value.loadLast11Months) {
            rangeChanged = true;
        }

        if (!dateChanged && !rangeChanged && !force && !transactionOverviewStateInvalid.value) {
            return new Promise((resolve) => {
                resolve(transactionOverviewData.value);
            });
        }

        const requestParams: TransactionAmountsRequestParams = {
            useTransactionTimezone: settingsStore.appSettings.timezoneUsedForStatisticsInHomePage === TimezoneTypeForStatistics.TransactionTimezone.type,
            today: transactionDataRange.value.today,
            thisWeek: transactionDataRange.value.thisWeek,
            thisMonth: transactionDataRange.value.thisMonth,
            thisYear: transactionDataRange.value.thisYear
        };

        if (loadLast11Months) {
            requestParams.lastMonth = transactionDataRange.value.lastMonth;
            requestParams.monthBeforeLastMonth = transactionDataRange.value.monthBeforeLastMonth;
            requestParams.monthBeforeLast2Months = transactionDataRange.value.monthBeforeLast2Months;
            requestParams.monthBeforeLast3Months = transactionDataRange.value.monthBeforeLast3Months;
            requestParams.monthBeforeLast4Months = transactionDataRange.value.monthBeforeLast4Months;
            requestParams.monthBeforeLast5Months = transactionDataRange.value.monthBeforeLast5Months;
            requestParams.monthBeforeLast6Months = transactionDataRange.value.monthBeforeLast6Months;
            requestParams.monthBeforeLast7Months = transactionDataRange.value.monthBeforeLast7Months;
            requestParams.monthBeforeLast8Months = transactionDataRange.value.monthBeforeLast8Months;
            requestParams.monthBeforeLast9Months = transactionDataRange.value.monthBeforeLast9Months;
            requestParams.monthBeforeLast10Months = transactionDataRange.value.monthBeforeLast10Months;
        }

        const excludeAccountIds: string[] = objectFieldWithValueToArrayItem(settingsStore.appSettings.overviewAccountFilterInHomePage, true);
        const excludeCategoryIds: string[] = objectFieldWithValueToArrayItem(settingsStore.appSettings.overviewTransactionCategoryFilterInHomePage, true);

        return new Promise((resolve, reject) => {
            services.getTransactionAmounts(requestParams, excludeAccountIds, excludeCategoryIds).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve transaction overview' });
                    return;
                }

                if (transactionOverviewStateInvalid.value) {
                    updateTransactionOverviewInvalidState(false);
                }

                if (force && data.result && isEquals(transactionOverviewData.value, data.result)) {
                    reject({ message: 'Data is up to date', isUpToDate: true });
                    return;
                }

                transactionOverviewData.value = data.result;
                transactionOverviewOptions.value.loadLast11Months = !!loadLast11Months;

                resolve(data.result);
            }).catch(error => {
                if (force) {
                    logger.error('failed to force load transaction overview', error);
                } else {
                    logger.error('failed to load transaction overview', error);
                }

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve transaction overview' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function loadMonthlyTransactionsForBalanceForecast({ force, startTime: customStartTime, endTime: customEndTime, displayStartTime, displayEndTime }: { force: boolean, startTime?: number, endTime?: number, displayStartTime?: number, displayEndTime?: number }): Promise<TransactionInfoResponse[]> {
        const startTime = customStartTime || getThisMonthFirstUnixTime();
        const endTime = customEndTime || getThisMonthLastUnixTime();

        if (!force && monthlyTransactionsForForecastLoaded.value
            && forecastStartTime.value === startTime && forecastEndTime.value === endTime) {
            // Still update display range even if data is cached
            forecastDisplayStartTime.value = displayStartTime || startTime;
            forecastDisplayEndTime.value = displayEndTime || endTime;
            return Promise.resolve(monthlyTransactionsForForecast.value);
        }

        return new Promise((resolve, reject) => {
            services.getAllTransactions({ startTime, endTime }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve monthly transactions for forecast' });
                    return;
                }

                monthlyTransactionsForForecast.value = data.result;
                monthlyTransactionsForForecastLoaded.value = true;
                forecastStartTime.value = startTime;
                forecastEndTime.value = endTime;
                forecastDisplayStartTime.value = displayStartTime || startTime;
                forecastDisplayEndTime.value = displayEndTime || endTime;
                resolve(data.result);
            }).catch(error => {
                logger.error('failed to load monthly transactions for balance forecast', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve monthly transactions for forecast' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function getTransactionListPageParams({ type, dateType, minTime, maxTime }: { type?: TransactionType, dateType?: number, minTime?: number, maxTime?: number }): string {
        const querys: string[] = [];

        if (isDefined(type)) {
            querys.push('type=' + type);
        }

        if (isDefined(dateType)) {
            querys.push('dateType=' + dateType);

            if (dateType === DateRange.Custom.type) {
                if (isNumber(minTime) && minTime > 0) {
                    querys.push('minTime=' + minTime);
                }

                if (isNumber(maxTime) && maxTime > 0) {
                    querys.push('maxTime=' + maxTime);
                }
            }
        }

        if (!isObjectEmpty(settingsStore.appSettings.overviewTransactionCategoryFilterInHomePage)) {
            querys.push('categoryIds=' + getFinalCategoryIdsByFilteredCategoryIds(transactionCategoriesStore.allTransactionCategoriesMap, settingsStore.appSettings.overviewTransactionCategoryFilterInHomePage));
        }

        if (!isObjectEmpty(settingsStore.appSettings.overviewAccountFilterInHomePage)) {
            querys.push('accountIds=' + getFinalAccountIdsByFilteredAccountIds(accountsStore.allAccountsMap, settingsStore.appSettings.overviewAccountFilterInHomePage));
        }

        return querys.join('&');
    }

    return {
        // states
        transactionDataRange,
        transactionOverviewOptions,
        transactionOverviewData,
        transactionOverviewStateInvalid,
        monthlyTransactionsForForecast,
        monthlyTransactionsForForecastLoaded,
        forecastStartTime,
        forecastEndTime,
        forecastDisplayStartTime,
        forecastDisplayEndTime,
        // computed states,
        transactionOverview,
        // functions
        updateTransactionOverviewInvalidState,
        resetTransactionOverview,
        loadTransactionOverview,
        loadMonthlyTransactionsForBalanceForecast,
        getTransactionListPageParams
    };
});
