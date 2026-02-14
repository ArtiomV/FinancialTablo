<template>
    <v-row class="match-height">
        <v-col cols="12">
            <!-- Page Header -->
            <div class="d-flex align-center mb-3">
                <h4 class="text-h5 font-weight-bold">{{ tt('Calendar') }}</h4>
                <v-spacer />
                <v-btn density="compact" color="default" variant="text" size="24"
                       class="ms-2" :icon="true" :loading="loading" @click="reload">
                    <template #loader>
                        <v-progress-circular indeterminate size="20"/>
                    </template>
                    <v-icon :icon="mdiRefresh" size="24" />
                    <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                </v-btn>
            </div>

            <v-card>
                <v-card-text>
                    <!-- Month navigation -->
                    <div class="d-flex align-center justify-center mb-4">
                        <v-btn icon size="small" variant="text" :disabled="loading" @click="navigateMonth(-1)">
                            <v-icon :icon="mdiChevronLeft" size="24" />
                        </v-btn>
                        <span class="text-h6 font-weight-bold mx-4" style="min-width: 200px; text-align: center">
                            {{ currentMonthLabel }}
                        </span>
                        <v-btn icon size="small" variant="text" :disabled="loading" @click="navigateMonth(1)">
                            <v-icon :icon="mdiChevronRight" size="24" />
                        </v-btn>
                    </div>

                    <!-- Calendar -->
                    <div class="calendar-page-container d-flex justify-center">
                        <transaction-calendar :key="calendarKey"
                                              calendar-class="justify-content-center"
                                              :all-days-clickable="true"
                                              :readonly="loading"
                                              :is-dark-mode="isDarkMode"
                                              :default-currency="defaultCurrency"
                                              :start-date="calendarStartDate"
                                              :daily-total-amounts="dailyTotalAmounts"
                                              day-has-transaction-class="calendar-day-has-transaction"
                                              v-model="selectedDate"
                                              @update:modelValue="onDateSelected" />
                    </div>

                    <!-- Selected day transactions -->
                    <div class="mt-4" v-if="selectedDayTransactions.length > 0">
                        <v-divider class="mb-3" />
                        <h6 class="text-subtitle-1 font-weight-bold mb-2">
                            {{ selectedDateLabel }}
                        </h6>
                        <v-table density="compact">
                            <thead>
                                <tr>
                                    <th style="min-width: 80px">{{ tt('Type') }}</th>
                                    <th style="min-width: 100px">{{ tt('Category') }}</th>
                                    <th style="min-width: 120px">{{ tt('Tags') }}</th>
                                    <th style="min-width: 120px">{{ tt('Counterparty') }}</th>
                                    <th style="min-width: 160px">{{ tt('Description') }}</th>
                                    <th style="min-width: 100px">{{ tt('Account') }}</th>
                                    <th class="text-right" style="min-width: 120px">{{ tt('Amount') }}</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr v-for="tx in selectedDayTransactions" :key="tx.id">
                                    <td>
                                        <span class="text-body-2" :class="getAmountClass(tx)">{{ getTransactionTypeName(tx) }}</span>
                                    </td>
                                    <td>
                                        <span class="text-body-2">{{ getCategoryName(tx) }}</span>
                                    </td>
                                    <td>
                                        <span class="text-body-2">{{ getTagNames(tx) }}</span>
                                    </td>
                                    <td>
                                        <span class="text-body-2">{{ getCounterpartyName(tx) }}</span>
                                    </td>
                                    <td>
                                        <span class="text-body-2">{{ tx.comment || '' }}</span>
                                    </td>
                                    <td>
                                        <span class="text-body-2">{{ getAccountName(tx) }}</span>
                                    </td>
                                    <td class="text-right">
                                        <span :class="getAmountClass(tx)">{{ getDisplayAmount(tx) }}</span>
                                    </td>
                                </tr>
                            </tbody>
                        </v-table>
                    </div>
                </v-card-text>
            </v-card>
        </v-col>
    </v-row>

    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import SnackBar from '@/components/desktop/SnackBar.vue';

import { ref, computed, onMounted, useTemplateRef } from 'vue';
import { useTheme } from 'vuetify';

import { useI18n } from '@/locales/helpers.ts';

import { useUserStore } from '@/stores/user.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionsStore } from '@/stores/transaction.ts';
import type { TransactionTotalAmount } from '@/stores/transaction.ts';
import { useCounterpartiesStore } from '@/stores/counterparty.ts';
import { useTransactionTagsStore } from '@/stores/transactionTag.ts';

import { TransactionType } from '@/core/transaction.ts';
import type { TextualYearMonthDay } from '@/core/datetime.ts';

import { Transaction } from '@/models/transaction.ts';


import {
    mdiRefresh,
    mdiChevronLeft,
    mdiChevronRight
} from '@mdi/js';

type SnackBarType = InstanceType<typeof SnackBar>;

const { tt, formatAmountToLocalizedNumeralsWithCurrency } = useI18n();
const theme = useTheme();

const userStore = useUserStore();
const accountsStore = useAccountsStore();
const transactionCategoriesStore = useTransactionCategoriesStore();
const transactionsStore = useTransactionsStore();
const counterpartiesStore = useCounterpartiesStore();
const transactionTagsStore = useTransactionTagsStore();

const snackbar = useTemplateRef<SnackBarType>('snackbar');

const loading = ref<boolean>(false);
const currentYear = ref<number>(new Date().getFullYear());
const currentMonth = ref<number>(new Date().getMonth() + 1); // 1-based
const selectedDate = ref<TextualYearMonthDay | ''>('');
const calendarKey = ref<number>(0);

const isDarkMode = computed<boolean>(() => theme.current.value.dark);
const defaultCurrency = computed<string>(() => userStore.currentUserDefaultCurrency);

const currentMonthLabel = computed<string>(() => {
    const monthName = tt('month_standalone_' + currentMonth.value);
    return `${monthName} ${currentYear.value}`;
});

const calendarStartDate = computed<Date>(() => {
    return new Date(currentYear.value, currentMonth.value - 1, 1);
});

const currentMonthTransactionData = computed(() => {
    const allTransactions = transactionsStore.transactions;
    if (!allTransactions || !allTransactions.length) {
        return null;
    }
    return allTransactions[0] || null;
});

const dailyTotalAmounts = computed<Record<string, TransactionTotalAmount> | undefined>(() => {
    return currentMonthTransactionData.value?.dailyTotalAmounts;
});

const selectedDateLabel = computed<string>(() => {
    if (!selectedDate.value) return '';
    const parts = selectedDate.value.split('-');
    if (parts.length !== 3) return selectedDate.value;
    const day = parseInt(parts[2]!);
    const month = parseInt(parts[1]!);
    const monthName = tt('month_standalone_' + month);
    return `${day} ${monthName} ${parts[0]}`;
});

const selectedDayTransactions = computed<Transaction[]>(() => {
    const data = currentMonthTransactionData.value;
    if (!data || !data.items || !selectedDate.value) {
        return [];
    }

    const result: Transaction[] = [];
    for (const tx of data.items) {
        if (tx.gregorianCalendarYearDashMonthDashDay === selectedDate.value) {
            result.push(tx);
        }
    }
    return result;
});

function getCategoryName(tx: Transaction): string {
    if (tx.categoryId) {
        const cat = transactionCategoriesStore.allTransactionCategoriesMap[tx.categoryId];
        if (cat) return cat.name;
    }
    return '';
}

function getAccountName(tx: Transaction): string {
    if (tx.sourceAccountId) {
        const acc = accountsStore.allAccountsMap[tx.sourceAccountId];
        if (acc) return acc.name;
    }
    return '';
}

function getCounterpartyName(tx: Transaction): string {
    if (tx.counterpartyId && tx.counterpartyId !== '0') {
        const cp = counterpartiesStore.allCounterpartiesMap[tx.counterpartyId];
        if (cp) return cp.name;
    }
    return '';
}

function getTagNames(tx: Transaction): string {
    if (!tx.tagIds || !tx.tagIds.length) {
        return '';
    }
    const names: string[] = [];
    for (const tagId of tx.tagIds) {
        const tag = transactionTagsStore.allTransactionTagsMap[tagId];
        if (tag) {
            names.push(tag.name);
        }
    }
    return names.join(', ');
}

function getTransactionTypeName(tx: Transaction): string {
    switch (tx.type) {
        case TransactionType.Income:
            return tt('Income');
        case TransactionType.Expense:
            return tt('Expense');
        case TransactionType.Transfer:
            return tt('Transfer');
        case TransactionType.ModifyBalance:
            return tt('Modify Balance');
        default:
            return '';
    }
}

function getAmountClass(tx: Transaction): string {
    if (tx.type === TransactionType.Income) return 'text-success font-weight-medium';
    if (tx.type === TransactionType.Expense) return 'text-error font-weight-medium';
    return 'text-info font-weight-medium';
}

function getDisplayAmount(tx: Transaction): string {
    const prefix = tx.type === TransactionType.Income ? '+' : (tx.type === TransactionType.Expense ? '\u2013' : '');
    const currency = tx.sourceAccount?.currency ?? defaultCurrency.value;
    return prefix + formatAmountToLocalizedNumeralsWithCurrency(tx.sourceAmount, currency);
}

function onDateSelected(value: string): void {
    if (!value) return;
    const parts = value.split('-');
    if (parts.length !== 3) return;
    const selectedYear = parseInt(parts[0]!);
    const selectedMonth = parseInt(parts[1]!);

    // If the selected date is in a different month, navigate to that month
    if (selectedYear !== currentYear.value || selectedMonth !== currentMonth.value) {
        currentYear.value = selectedYear;
        currentMonth.value = selectedMonth;
        calendarKey.value++;
        loadTransactions();
    }
}

function navigateMonth(direction: number): void {
    let newMonth = currentMonth.value + direction;
    let newYear = currentYear.value;

    if (newMonth < 1) {
        newMonth = 12;
        newYear--;
    } else if (newMonth > 12) {
        newMonth = 1;
        newYear++;
    }

    currentYear.value = newYear;
    currentMonth.value = newMonth;
    selectedDate.value = '';
    calendarKey.value++;
    loadTransactions();
}

function reload(): void {
    loadTransactions();
}

function loadTransactions(): void {
    loading.value = true;

    Promise.all([
        accountsStore.loadAllAccounts({ force: false }),
        transactionCategoriesStore.loadAllCategories({ force: false }),
        counterpartiesStore.loadAllCounterparties({ force: false }),
        transactionTagsStore.loadAllTags({ force: false })
    ]).then(() => {
        return transactionsStore.loadMonthlyAllTransactions({
            year: currentYear.value,
            month: currentMonth.value,
            autoExpand: true,
            defaultCurrency: defaultCurrency.value
        });
    }).then(() => {
        loading.value = false;
    }).catch(error => {
        loading.value = false;
        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

onMounted(() => {
    loadTransactions();
});
</script>

<style>
.calendar-page-container .dp__main .dp__menu {
    --dp-border-radius: 6px;
    --dp-menu-border-color: rgba(var(--v-border-color), var(--v-border-opacity));
}

.calendar-page-container .dp__main .dp__calendar {
    --dp-border-color: rgba(var(--v-border-color), var(--v-border-opacity));
}

.calendar-page-container .dp__main .dp__calendar .dp__calendar_row {
    --dp-cell-size: 90px;
    --dp-primary-color: rgba(var(--v-theme-primary), var(--v-activated-opacity));
    --dp-primary-text-color: rgb(var(--v-theme-on-surface));
}

.calendar-page-container .dp__main.transaction-calendar-with-alternate-date .dp__calendar .dp__calendar_row {
    --dp-cell-size: 110px;
}

.calendar-page-container .dp__main .dp__calendar .dp__calendar_row > .dp__calendar_item {
    overflow: hidden;
}

.calendar-page-container .dp__main .dp__calendar .dp__calendar_row > .dp__calendar_item .transaction-calendar-daily-amounts > span.transaction-calendar-alternate-date {
    font-size: 0.9rem;
}

.calendar-page-container .dp__main .dp__calendar .dp__calendar_row > .dp__calendar_item .transaction-calendar-daily-amounts > span.transaction-calendar-daily-amount {
    font-size: 0.95rem;
}

.calendar-day-has-transaction {
    font-weight: bold;
}
</style>
