import { ref, type Ref, type ComputedRef } from 'vue';

import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';
import { useTransactionsStore } from '@/stores/transaction.ts';

import { keys } from '@/core/base.ts';
import {
    type TimeRangeAndDateType,
    type Year0BasedMonth,
    type WeekDayValue,
    DateRangeScene,
    DateRange
} from '@/core/datetime.ts';
import type { Transaction } from '@/models/transaction.ts';

import services from '@/lib/services.ts';
import {
    getCurrentUnixTime,
    parseDateTimeFromUnixTime,
    getDayFirstDateTimeBySpecifiedUnixTime,
    getYearMonthFirstUnixTime,
    getYearMonthLastUnixTime,
    getShiftedDateRangeAndDateType,
    getShiftedDateRangeAndDateTypeForBillingCycle,
    getDateTypeByDateRange,
    getDateTypeByBillingCycleDateRange,
    getDateRangeByDateType,
    getDateRangeByBillingCycleDateType,
    getFullMonthDateRange,
    getValidMonthDayOrCurrentDayShortDate
} from '@/lib/datetime.ts';
import {
    transactionTypeToCategoryType
} from '@/lib/category.ts';

import { TransactionListPageType } from '@/views/base/transactions/TransactionListPageBase.ts';

export interface UseTransactionListOptions {
    showToast: (message: string) => void;
    showAlert: (message: string) => void;
    showLoading: (condition?: (() => boolean)) => void;
    hideLoading: () => void;
    onSwipeoutDeleted: (domId: string, done: () => void) => void;
    getTransactionDomId: (transaction: Transaction) => string;
    onBeforeReload?: () => void;
    onAfterReload?: () => void;
    onAfterLoadMore?: () => void;
}

export interface UseTransactionListDeps {
    pageType: Ref<number>;
    loading: Ref<boolean>;
    customMinDatetime: Ref<number>;
    customMaxDatetime: Ref<number>;
    currentCalendarDate: Ref<string>;
    firstDayOfWeek: ComputedRef<WeekDayValue>;
    fiscalYearStart: ComputedRef<number>;
    defaultCurrency: ComputedRef<string>;
    queryMonthlyData: ComputedRef<boolean>;
    query: ComputedRef<{ dateType: number; minTime: number; maxTime: number; type: number; categoryIds: string; accountIds: string; tagFilter: string; keyword: string; amountFilter: string }>;
    queryAllFilterCategoryIds: ComputedRef<Record<string, boolean>>;
    allCategories: ComputedRef<Record<string, { type: number }>>;
    showCustomDateRangeSheet: Ref<boolean>;
    showCustomMonthSheet: Ref<boolean>;
}

export interface InitQuery {
    dateType?: string;
    maxTime?: string;
    minTime?: string;
    type?: string;
    categoryIds?: string;
    accountIds?: string;
    tagFilter?: string;
    keyword?: string;
}

export function useTransactionList(options: UseTransactionListOptions, deps: UseTransactionListDeps) {
    const accountsStore = useAccountsStore();
    const transactionCategoriesStore = useTransactionCategoriesStore();
    const transactionTagsStore = useTransactionTagsStore();
    const transactionsStore = useTransactionsStore();

    const loadingError = ref<unknown | null>(null);
    const loadingMore = ref<boolean>(false);
    const transactionToDelete = ref<Transaction | null>(null);
    const showPlannedTransactions = ref<boolean>(false);
    const confirmingPlannedTransaction = ref<boolean>(false);

    function init(initQuery: InitQuery): void {
        let dateRange: TimeRangeAndDateType | null = getDateRangeByDateType(
            initQuery.dateType ? parseInt(initQuery.dateType) : undefined,
            deps.firstDayOfWeek.value,
            deps.fiscalYearStart.value
        );

        if (!dateRange && initQuery.dateType && initQuery.maxTime && initQuery.minTime &&
            (DateRange.isBillingCycle(parseInt(initQuery.dateType)) || initQuery.dateType === DateRange.Custom.type.toString()) &&
            parseInt(initQuery.maxTime) > 0 && parseInt(initQuery.minTime) > 0) {
            dateRange = {
                dateType: parseInt(initQuery.dateType),
                maxTime: parseInt(initQuery.maxTime),
                minTime: parseInt(initQuery.minTime)
            };
        }

        transactionsStore.initTransactionListFilter({
            dateType: dateRange ? dateRange.dateType : undefined,
            maxTime: dateRange ? dateRange.maxTime : undefined,
            minTime: dateRange ? dateRange.minTime : undefined,
            type: initQuery.type && parseInt(initQuery.type) > 0 ? parseInt(initQuery.type) : undefined,
            categoryIds: initQuery.categoryIds,
            accountIds: initQuery.accountIds,
            tagFilter: initQuery.tagFilter,
            keyword: initQuery.keyword
        });

        reload();
    }

    function reload(done?: () => void): void {
        const force = !!done;

        if (!done) {
            deps.loading.value = true;
        }

        options.onBeforeReload?.();

        Promise.all([
            accountsStore.loadAllAccounts({ force: false }),
            transactionCategoriesStore.loadAllCategories({ force: false }),
            transactionTagsStore.loadAllTags({ force: false })
        ]).then(() => {
            if (deps.queryMonthlyData.value) {
                const currentMonthMinDate = parseDateTimeFromUnixTime(deps.query.value.minTime);
                const currentYear = currentMonthMinDate.getGregorianCalendarYear();
                const currentMonth = currentMonthMinDate.getGregorianCalendarMonth();

                return transactionsStore.loadMonthlyAllTransactions({
                    year: currentYear,
                    month: currentMonth,
                    autoExpand: true,
                    defaultCurrency: deps.defaultCurrency.value
                });
            } else {
                return transactionsStore.loadTransactions({
                    reload: true,
                    autoExpand: true,
                    defaultCurrency: deps.defaultCurrency.value
                });
            }
        }).then(() => {
            done?.();

            if (force) {
                options.showToast('Data has been updated');
            }

            deps.loading.value = false;
            options.onAfterReload?.();
        }).catch(error => {
            if (error.processed || done) {
                deps.loading.value = false;
            }

            done?.();

            if (!error.processed) {
                if (!done) {
                    loadingError.value = error;
                }

                options.showToast(error.message || error);
            }
        });
    }

    function loadMore(autoExpand: boolean): void {
        if (!transactionsStore.hasMoreTransaction) {
            return;
        }

        if (loadingMore.value || deps.loading.value) {
            return;
        }

        loadingMore.value = true;

        transactionsStore.loadTransactions({
            reload: false,
            autoExpand: autoExpand,
            defaultCurrency: deps.defaultCurrency.value
        }).then(() => {
            loadingMore.value = false;
            options.onAfterLoadMore?.();
        }).catch(error => {
            loadingMore.value = false;

            if (!error.processed) {
                options.showToast(error.message || error);
            }
        });
    }

    function changePageType(type: number): void {
        deps.pageType.value = type;
        deps.currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(deps.query.value.minTime, deps.currentCalendarDate.value);

        if (deps.pageType.value === TransactionListPageType.Calendar.type) {
            const dateRange = getFullMonthDateRange(deps.query.value.minTime, deps.query.value.maxTime, deps.firstDayOfWeek.value, deps.fiscalYearStart.value);

            if (dateRange) {
                const changed = transactionsStore.updateTransactionListFilter({
                    dateType: dateRange.dateType,
                    maxTime: dateRange.maxTime,
                    minTime: dateRange.minTime
                });

                if (changed) {
                    deps.currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(deps.query.value.minTime, deps.currentCalendarDate.value);
                    reload();
                }
            }
        }
    }

    function changeDateFilter(dateType: number): void {
        if (dateType === DateRange.Custom.type) {
            if (!deps.query.value.minTime || !deps.query.value.maxTime) {
                deps.customMaxDatetime.value = getCurrentUnixTime();
                deps.customMinDatetime.value = getDayFirstDateTimeBySpecifiedUnixTime(deps.customMaxDatetime.value).getUnixTime();
            } else {
                deps.customMaxDatetime.value = deps.query.value.maxTime;
                deps.customMinDatetime.value = deps.query.value.minTime;
            }

            if (deps.pageType.value === TransactionListPageType.Calendar.type) {
                deps.showCustomMonthSheet.value = true;
            } else {
                deps.showCustomDateRangeSheet.value = true;
            }

            return;
        } else if (deps.query.value.dateType === dateType) {
            return;
        }

        let dateRange: TimeRangeAndDateType | null = null;

        if (DateRange.isBillingCycle(dateType)) {
            dateRange = getDateRangeByBillingCycleDateType(dateType, deps.firstDayOfWeek.value, deps.fiscalYearStart.value, accountsStore.getAccountStatementDate(deps.query.value.accountIds));
        } else {
            dateRange = getDateRangeByDateType(dateType, deps.firstDayOfWeek.value, deps.fiscalYearStart.value);
        }

        if (!dateRange) {
            return;
        }

        if (deps.pageType.value === TransactionListPageType.Calendar.type) {
            deps.currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(dateRange.minTime, deps.currentCalendarDate.value);
            const fullMonthDateRange = getFullMonthDateRange(dateRange.minTime, dateRange.maxTime, deps.firstDayOfWeek.value, deps.fiscalYearStart.value);

            if (fullMonthDateRange) {
                dateRange = fullMonthDateRange;
                deps.currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(dateRange.minTime, deps.currentCalendarDate.value);
            }
        }

        const changed = transactionsStore.updateTransactionListFilter({
            dateType: dateRange.dateType,
            maxTime: dateRange.maxTime,
            minTime: dateRange.minTime
        });

        if (changed) {
            reload();
        }
    }

    function changeCustomDateFilter(minTime: number, maxTime: number): void {
        if (!minTime || !maxTime) {
            return;
        }

        let dateType: number | null = getDateTypeByBillingCycleDateRange(minTime, maxTime, deps.firstDayOfWeek.value, deps.fiscalYearStart.value, DateRangeScene.Normal, accountsStore.getAccountStatementDate(deps.query.value.accountIds));

        if (!dateType) {
            dateType = getDateTypeByDateRange(minTime, maxTime, deps.firstDayOfWeek.value, deps.fiscalYearStart.value, DateRangeScene.Normal);
        }

        if (deps.pageType.value === TransactionListPageType.Calendar.type) {
            deps.currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(minTime, deps.currentCalendarDate.value);
            const dateRange = getFullMonthDateRange(minTime, maxTime, deps.firstDayOfWeek.value, deps.fiscalYearStart.value);

            if (dateRange) {
                minTime = dateRange.minTime;
                maxTime = dateRange.maxTime;
                dateType = dateRange.dateType;
                deps.currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(minTime, deps.currentCalendarDate.value);
            }
        }

        const changed = transactionsStore.updateTransactionListFilter({
            dateType: dateType,
            maxTime: maxTime,
            minTime: minTime
        });

        deps.showCustomDateRangeSheet.value = false;

        if (changed) {
            reload();
        }
    }

    function changeCustomMonthDateFilter(yearMonth: Year0BasedMonth): void {
        if (!yearMonth) {
            return;
        }

        const minTime = getYearMonthFirstUnixTime(yearMonth);
        const maxTime = getYearMonthLastUnixTime(yearMonth);
        const dateType = getDateTypeByDateRange(minTime, maxTime, deps.firstDayOfWeek.value, deps.fiscalYearStart.value, DateRangeScene.Normal);

        if (deps.pageType.value === TransactionListPageType.Calendar.type) {
            deps.currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(minTime, deps.currentCalendarDate.value);
        }

        const changed = transactionsStore.updateTransactionListFilter({
            dateType: dateType,
            maxTime: maxTime,
            minTime: minTime
        });

        deps.showCustomMonthSheet.value = false;

        if (changed) {
            reload();
        }
    }

    function shiftDateRange(minTime: number, maxTime: number, scale: number): void {
        if (deps.query.value.dateType === DateRange.All.type) {
            return;
        }

        let newDateRange: TimeRangeAndDateType | null = null;

        if (DateRange.isBillingCycle(deps.query.value.dateType) || deps.query.value.dateType === DateRange.Custom.type) {
            newDateRange = getShiftedDateRangeAndDateTypeForBillingCycle(minTime, maxTime, scale, deps.firstDayOfWeek.value, deps.fiscalYearStart.value, DateRangeScene.Normal, accountsStore.getAccountStatementDate(deps.query.value.accountIds));
        }

        if (!newDateRange) {
            newDateRange = getShiftedDateRangeAndDateType(minTime, maxTime, scale, deps.firstDayOfWeek.value, deps.fiscalYearStart.value, DateRangeScene.Normal);
        }

        if (deps.pageType.value === TransactionListPageType.Calendar.type) {
            deps.currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(newDateRange.minTime, deps.currentCalendarDate.value);
            const fullMonthDateRange = getFullMonthDateRange(newDateRange.minTime, newDateRange.maxTime, deps.firstDayOfWeek.value, deps.fiscalYearStart.value);

            if (fullMonthDateRange) {
                newDateRange = fullMonthDateRange;
                deps.currentCalendarDate.value = getValidMonthDayOrCurrentDayShortDate(newDateRange.minTime, deps.currentCalendarDate.value);
            }
        }

        const changed = transactionsStore.updateTransactionListFilter({
            dateType: newDateRange.dateType,
            maxTime: newDateRange.maxTime,
            minTime: newDateRange.minTime
        });

        if (changed) {
            reload();
        }
    }

    function changeTypeFilter(type: number): void {
        if (deps.query.value.type === type) {
            return;
        }

        let newCategoryFilter = undefined;

        if (type && deps.query.value.categoryIds) {
            newCategoryFilter = '';

            for (const categoryId of keys(deps.queryAllFilterCategoryIds.value)) {
                const category = deps.allCategories.value[categoryId];

                if (category && category.type === transactionTypeToCategoryType(type)) {
                    if (newCategoryFilter.length > 0) {
                        newCategoryFilter += ',';
                    }

                    newCategoryFilter += categoryId;
                }
            }
        }

        const changed = transactionsStore.updateTransactionListFilter({
            type: type,
            categoryIds: newCategoryFilter
        });

        if (changed) {
            reload();
        }
    }

    function changeCategoryFilter(categoryIds: string): void {
        if (deps.query.value.categoryIds === categoryIds) {
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            categoryIds: categoryIds
        });

        if (changed) {
            reload();
        }
    }

    function changeAccountFilter(accountIds: string): void {
        if (deps.query.value.accountIds === accountIds) {
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            accountIds: accountIds
        });

        if (changed) {
            reload();
        }
    }

    function changeKeywordFilter(keyword: string): void {
        if (deps.query.value.keyword === keyword) {
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            keyword: keyword
        });

        if (changed) {
            reload();
        }
    }

    function changeAmountFilter(filterType: string, navigateToAmountFilter: (filterType: string, currentValue: string) => void): void {
        if (deps.query.value.amountFilter === filterType) {
            return;
        }

        if (filterType) {
            navigateToAmountFilter(filterType, deps.query.value.amountFilter);
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            amountFilter: filterType
        });

        if (changed) {
            reload();
        }
    }

    function changeTagFilter(tagFilter: string): void {
        if (deps.query.value.tagFilter === tagFilter) {
            return;
        }

        const changed = transactionsStore.updateTransactionListFilter({
            tagFilter: tagFilter
        });

        if (changed) {
            reload();
        }
    }

    function remove(transaction: Transaction | null, confirm: boolean, showDeleteSheet: () => void): void {
        if (!transaction) {
            options.showAlert('An error occurred');
            return;
        }

        if (!confirm) {
            transactionToDelete.value = transaction;
            showDeleteSheet();
            return;
        }

        transactionToDelete.value = null;
        options.showLoading();

        transactionsStore.deleteTransaction({
            transaction: transaction,
            defaultCurrency: deps.defaultCurrency.value,
            beforeResolve: (done) => {
                options.onSwipeoutDeleted(options.getTransactionDomId(transaction), done);
            }
        }).then(() => {
            options.hideLoading();
        }).catch(error => {
            options.hideLoading();

            if (!error.processed) {
                options.showToast(error.message || error);
            }
        });
    }

    function removeAllFuture(transaction: Transaction | null): void {
        if (!transaction) {
            options.showAlert('An error occurred');
            return;
        }

        transactionToDelete.value = null;
        options.showLoading();

        services.deleteAllFuturePlannedTransactions({
            id: transaction.id
        }).then(() => {
            options.hideLoading();
            transactionsStore.updateTransactionListInvalidState(true);
            reload();
        }).catch(error => {
            options.hideLoading();

            if (!error.processed) {
                options.showToast(error.message || error);
            }
        });
    }

    function confirmPlannedTransaction(transaction: Transaction): void {
        confirmingPlannedTransaction.value = true;
        options.showLoading(() => confirmingPlannedTransaction.value);

        services.confirmPlannedTransaction({ id: transaction.id }).then(() => {
            confirmingPlannedTransaction.value = false;
            options.hideLoading();
            options.showToast('Transaction confirmed successfully');
            reload();
        }).catch(error => {
            confirmingPlannedTransaction.value = false;
            options.hideLoading();

            if (!error.processed) {
                options.showToast('Unable to confirm planned transaction');
            }
        });
    }

    return {
        // state
        loadingError,
        loadingMore,
        transactionToDelete,
        showPlannedTransactions,
        confirmingPlannedTransaction,
        // functions
        init,
        reload,
        loadMore,
        changePageType,
        changeDateFilter,
        changeCustomDateFilter,
        changeCustomMonthDateFilter,
        shiftDateRange,
        changeTypeFilter,
        changeCategoryFilter,
        changeAccountFilter,
        changeKeywordFilter,
        changeAmountFilter,
        changeTagFilter,
        remove,
        removeAllFuture,
        confirmPlannedTransaction
    };
}
