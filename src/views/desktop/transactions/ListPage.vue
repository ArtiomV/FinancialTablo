<template>
    <v-row class="match-height">
        <v-col cols="12">
            <!-- Page Header -->
            <div class="d-flex align-center mb-3">
                <h4 class="text-h5 font-weight-bold">{{ tt('Operations') }}</h4>
                <v-spacer />
                <v-btn class="ms-2" color="success" variant="outlined" size="small" :prepend-icon="mdiPlus"
                       :disabled="loading || !canAddTransaction" @click="addWithType(TransactionType.Income)">
                    {{ tt('Add Income Transaction') }}
                </v-btn>
                <v-btn class="ms-2" color="error" variant="outlined" size="small" :prepend-icon="mdiMinus"
                       :disabled="loading || !canAddTransaction" @click="addWithType(TransactionType.Expense)">
                    {{ tt('Add Expense Transaction') }}
                </v-btn>
                <v-btn class="ms-2" color="info" variant="outlined" size="small" :prepend-icon="mdiSwapHorizontal"
                       :disabled="loading || !canAddTransaction" @click="addWithType(TransactionType.Transfer)">
                    {{ tt('Add Transfer Transaction') }}
                </v-btn>
                <v-btn density="compact" color="default" variant="text" size="24"
                       class="ms-2" :icon="true" :loading="loading" @click="reload(true, false)">
                    <template #loader>
                        <v-progress-circular indeterminate size="20"/>
                    </template>
                    <v-icon :icon="mdiRefresh" size="24" />
                    <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                </v-btn>
            </div>

            <v-card>
                    <v-window class="d-flex flex-grow-1 disable-tab-transition w-100-window-container" v-model="activeTab">
                            <v-window-item value="transactionPage">
                                <v-card variant="flat" min-height="920">
                                    <template #title>
                                        <div class="title-and-toolbar d-flex align-center flex-wrap" style="row-gap: 0.5rem">
                                            <!-- Period filter with outline — arrows outside -->
                                            <div class="d-flex align-center">
                                                <v-btn icon size="x-small" variant="text"
                                                       :disabled="loading || query.dateType === DateRange.All.type"
                                                       @click="navigatePeriod(-1)">
                                                    <v-icon :icon="mdiChevronLeft" size="18" />
                                                </v-btn>
                                                <v-menu :close-on-content-click="false">
                                                    <template #activator="{ props: menuProps }">
                                                        <v-btn variant="outlined" v-bind="menuProps" size="small" class="text-none font-weight-bold">
                                                            {{ currentPeriodLabel }}
                                                        </v-btn>
                                                    </template>
                                                        <v-list density="compact">
                                                            <v-list-item @click="changeDateFilter(DateRange.ThisWeek.type)">
                                                                <v-list-item-title>{{ tt('This week filter') }}</v-list-item-title>
                                                            </v-list-item>
                                                            <v-list-item @click="changeDateFilter(DateRange.ThisMonth.type)">
                                                                <v-list-item-title>{{ tt('This month filter') }}</v-list-item-title>
                                                            </v-list-item>
                                                            <v-list-item @click="changeDateFilter(DateRange.ThisQuarter.type)">
                                                                <v-list-item-title>{{ tt('This quarter filter') }}</v-list-item-title>
                                                            </v-list-item>
                                                            <v-list-item @click="changeDateFilter(DateRange.ThisYear.type)">
                                                                <v-list-item-title>{{ tt('This year filter') }}</v-list-item-title>
                                                            </v-list-item>
                                                            <v-list-item @click="changeDateFilter(DateRange.All.type)">
                                                                <v-list-item-title>{{ tt('All time') }}</v-list-item-title>
                                                            </v-list-item>
                                                            <v-divider />
                                                            <div class="px-3 py-2">
                                                                <div class="d-flex align-center ga-2">
                                                                    <v-text-field type="date" density="compact" hide-details variant="outlined"
                                                                                  :label="tt('From date')" v-model="customDateFrom" style="min-width: 110px" />
                                                                    <v-text-field type="date" density="compact" hide-details variant="outlined"
                                                                                  :label="tt('To date')" v-model="customDateTo" style="min-width: 110px" />
                                                                    <v-btn size="small" color="primary" variant="tonal"
                                                                           @click="applyCustomDateRange">{{ tt('Apply') }}</v-btn>
                                                                </div>
                                                            </div>
                                                        </v-list>
                                                </v-menu>
                                                <v-btn icon size="x-small" variant="text"
                                                       :disabled="loading || query.dateType === DateRange.All.type"
                                                       @click="navigatePeriod(1)">
                                                    <v-icon :icon="mdiChevronRight" size="18" />
                                                </v-btn>
                                            </div>

                                            <!-- Type filter — individual outlined buttons -->
                                            <div class="ms-3 d-flex align-center ga-1">
                                                <v-btn variant="outlined" size="small"
                                                       :color="queryType === 0 ? 'primary' : 'default'"
                                                       :disabled="loading" @click="queryType = 0">
                                                    {{ tt('All Filter') }}
                                                </v-btn>
                                                <v-btn variant="outlined" size="small"
                                                       :color="queryType === TransactionType.Income ? 'primary' : 'default'"
                                                       :disabled="loading" @click="queryType = TransactionType.Income">
                                                    {{ tt('Income Filter') }}
                                                </v-btn>
                                                <v-btn variant="outlined" size="small"
                                                       :color="queryType === TransactionType.Expense ? 'primary' : 'default'"
                                                       :disabled="loading" @click="queryType = TransactionType.Expense">
                                                    {{ tt('Expense Filter') }}
                                                </v-btn>
                                                <v-btn variant="outlined" size="small"
                                                       :color="queryType === TransactionType.Transfer ? 'primary' : 'default'"
                                                       :disabled="loading" @click="queryType = TransactionType.Transfer">
                                                    {{ tt('Transfer Filter') }}
                                                </v-btn>
                                            </div>

                                            <v-spacer/>

                                            <!-- Totals (income/expense/balance) — near Filter button -->
                                            <div class="me-3 d-flex align-center flex-wrap text-caption" style="row-gap: 0.25rem"
                                                 v-if="showTotalAmountInTransactionListPage && currentMonthTotalAmount">
                                                <span class="text-medium-emphasis">{{ currentMonthTotalAmount.incomeCount }} {{ tt('Inflows label') }}</span>
                                                <span class="text-income ms-1" v-if="!loading">+{{ currentMonthTotalAmount.income }}</span>
                                                <span class="text-medium-emphasis ms-2">{{ currentMonthTotalAmount.expenseCount }} {{ tt('Outflows label') }}</span>
                                                <span class="text-expense ms-1" v-if="!loading">–{{ currentMonthTotalAmount.expense }}</span>
                                                <span class="text-medium-emphasis ms-2">{{ tt('Balance label') }}</span>
                                                <span :class="currentMonthTotalAmount.balancePositive ? 'text-income' : 'text-expense'" class="ms-1" v-if="!loading">
                                                    {{ currentMonthTotalAmount.balancePositive ? '+' : '–' }}{{ currentMonthTotalAmount.balanceAmount }}
                                                </span>
                                            </div>

                                            <!-- Filter button -->
                                            <v-menu v-model="showFilterPanel" :close-on-content-click="false" location="bottom end">
                                                <template #activator="{ props: filterProps }">
                                                    <v-btn variant="outlined" size="small" :prepend-icon="mdiFilterVariant"
                                                           v-bind="filterProps">
                                                        {{ tt('Filters') }}
                                                        <v-badge v-if="activeFilterCount > 0" :content="activeFilterCount"
                                                                 color="primary" floating />
                                                    </v-btn>
                                                </template>
                                                <v-card width="380" class="pa-4">
                                                    <div class="text-subtitle-2 mb-2">{{ tt('Filters') }}</div>
                                                    <div class="d-flex ga-2 mb-2">
                                                        <v-text-field density="compact" hide-details type="number"
                                                                      :label="tt('Min Amount')" v-model.number="filterAmountMin" />
                                                        <v-text-field density="compact" hide-details type="number"
                                                                      :label="tt('Max Amount')" v-model.number="filterAmountMax" />
                                                    </div>
                                                    <v-autocomplete density="compact" hide-details class="mb-2"
                                                                    item-title="name" item-value="id" clearable
                                                                    :label="tt('Transaction Categories')"
                                                                    :items="allCategoryList"
                                                                    :model-value="filterCategoryId"
                                                                    @update:model-value="filterCategoryId = $event || ''" />
                                                    <v-autocomplete density="compact" hide-details class="mb-2"
                                                                    item-title="name" item-value="id" clearable
                                                                    :label="tt('Account')"
                                                                    :items="allAccounts"
                                                                    :model-value="filterAccountId"
                                                                    @update:model-value="filterAccountId = $event || ''" />
                                                    <v-autocomplete density="compact" hide-details class="mb-2"
                                                                    item-title="name" item-value="id" clearable
                                                                    :label="tt('Counterparty')"
                                                                    :items="counterpartiesStore.allVisibleCounterparties"
                                                                    :model-value="filterCounterpartyId"
                                                                    @update:model-value="filterCounterpartyId = $event || ''" />
                                                    <v-text-field density="compact" hide-details class="mb-2"
                                                                  :prepend-inner-icon="mdiMagnify"
                                                                  :placeholder="tt('Search transaction description')"
                                                                  v-model="searchKeyword"
                                                                  @keyup.enter="applyKeywordFilter" />
                                                    <div class="d-flex justify-end ga-2 mt-3">
                                                        <v-btn variant="text" size="small" @click="clearAllFilters">{{ tt('Clear All Filters') }}</v-btn>
                                                        <v-btn color="primary" variant="tonal" size="small" @click="applyAllFilters">{{ tt('Apply') }}</v-btn>
                                                    </div>
                                                </v-card>
                                            </v-menu>
                                        </div>
                                    </template>

                                    <div class="px-4 pt-0 pb-2" v-if="dailyBalanceForecast && dailyBalanceForecast.length > 0">
                                        <daily-balance-forecast-card :data="dailyBalanceForecast"
                                                                     :loading="loadingForecast" :disabled="loadingForecast"
                                                                     :is-dark-mode="isDarkMode" />
                                    </div>

                                    <div class="px-4 pb-2" v-if="plannedTransactionsCount > 0">
                                        <a class="text-body-2 cursor-pointer text-primary"
                                           @click="showPlannedTransactions = !showPlannedTransactions">
                                            {{ showPlannedTransactions
                                                ? `${tt('Hide Future Planned Transactions')} (${plannedTransactionsCount})`
                                                : `${tt('Show Future Planned Transactions')} (${plannedTransactionsCount})` }}
                                        </a>
                                    </div>

                                    <v-table class="transaction-table" :hover="!loading">
                                        <thead>
                                        <tr>
                                            <th class="transaction-table-column-amount text-no-wrap">
                                                <span>{{ tt('Amount') }}</span>
                                            </th>
                                            <th class="transaction-table-column-counterparty text-no-wrap">
                                                <span>{{ tt('Counterparty') }}</span>
                                            </th>
                                            <th class="transaction-table-column-category text-no-wrap">
                                                <span>{{ queryCategoryName }}</span>
                                            </th>
                                            <th class="transaction-table-column-actions text-no-wrap text-right">
                                            </th>
                                        </tr>
                                        </thead>

                                        <tbody v-if="loading && (!displayTransactions || !displayTransactions.length || displayTransactions.length < 1)">
                                        <tr :key="itemIdx" v-for="itemIdx in skeletonData">
                                            <td class="px-0" :colspan="4">
                                                <v-skeleton-loader type="text" :loading="true"></v-skeleton-loader>
                                            </td>
                                        </tr>
                                        </tbody>

                                        <tbody v-if="!loading && (!displayTransactions || !displayTransactions.length || displayTransactions.length < 1)">
                                        <tr>
                                            <td :colspan="4">{{ tt('No transaction data') }}</td>
                                        </tr>
                                        </tbody>

                                        <tbody :key="transaction.id"
                                               :class="{ 'disabled': loading, 'has-bottom-border': idx < displayTransactions.length - 1 }"
                                               v-for="(transaction, idx) in displayTransactions">
                                            <tr class="transaction-list-row-date no-hover text-sm"
                                                v-if="idx === 0 || (idx > 0 && (transaction.gregorianCalendarYearDashMonthDashDay !== displayTransactions[idx - 1]!.gregorianCalendarYearDashMonthDashDay))">
                                                <td :colspan="4" class="font-weight-bold">
                                                    <div class="d-flex align-center">
                                                        <span>{{ getDisplayLongDate(transaction) }}</span>
                                                        <v-chip class="ms-1" color="default" size="x-small"
                                                                v-if="transaction.displayDayOfWeek">
                                                            {{ getWeekdayLongName(transaction.displayDayOfWeek) }}
                                                        </v-chip>
                                                    </div>
                                                </td>
                                            </tr>
                                            <tr class="transaction-table-row-data text-sm cursor-pointer"
                                                :style="transaction.planned ? { opacity: 0.6 } : undefined"
                                                @click="show(transaction)">
                                                <td class="transaction-table-column-amount" :class="{ 'text-expense': transaction.type === TransactionType.Expense, 'text-income': transaction.type === TransactionType.Income }">
                                                    <div class="d-flex align-center" v-if="transaction.sourceAccount">
                                                        <v-btn v-if="transaction.splits && transaction.splits.length > 0"
                                                               icon variant="text" size="x-small" class="me-1"
                                                               @click.stop="toggleSplitExpand(transaction.id)">
                                                            <v-icon :icon="expandedSplitIds.has(transaction.id) ? mdiChevronUp : mdiChevronDown" size="18" />
                                                        </v-btn>
                                                        <span>{{ getDisplayAmount(transaction) }}</span>
                                                    </div>
                                                    <div class="text-caption text-medium-emphasis" v-if="transaction.sourceAccount" style="color: rgba(var(--v-theme-on-background), 0.5) !important">
                                                        {{ transaction.sourceAccount.name }}
                                                    </div>
                                                </td>
                                                <td class="transaction-table-column-counterparty">
                                                    <div>
                                                        <span v-if="transaction.type === TransactionType.Transfer && transaction.sourceAccount && transaction.destinationAccount">
                                                            {{ transaction.sourceAccount.name }}
                                                            <v-icon class="icon-with-direction mx-1" size="13" :icon="mdiArrowRight" />
                                                            {{ transaction.destinationAccount.name }}
                                                        </span>
                                                        <span v-else-if="transaction.counterpartyId && transaction.counterpartyId !== '0' && counterpartiesStore.allCounterpartiesMap[transaction.counterpartyId]">
                                                            {{ counterpartiesStore.allCounterpartiesMap[transaction.counterpartyId]!.name }}
                                                        </span>
                                                    </div>
                                                    <div class="text-caption text-medium-emphasis text-truncate" v-if="transaction.comment" style="max-width: 250px">
                                                        {{ transaction.comment }}
                                                    </div>
                                                </td>
                                                <td class="transaction-table-column-category">
                                                    <div>
                                                        <span v-if="transaction.splits && transaction.splits.length > 0">
                                                            {{ tt('Split') }} ({{ transaction.splits.length }})
                                                        </span>
                                                        <span v-else-if="transaction.type === TransactionType.ModifyBalance">
                                                            {{ tt('Modify Balance') }}
                                                        </span>
                                                        <span v-else-if="transaction.type !== TransactionType.ModifyBalance && transaction.category">
                                                            {{ transaction.category.name }}
                                                        </span>
                                                        <span v-else-if="transaction.type !== TransactionType.ModifyBalance && !transaction.category">
                                                            {{ getTransactionTypeName(transaction.type, 'Transaction') }}
                                                        </span>
                                                    </div>
                                                    <div class="text-caption text-medium-emphasis" v-if="transaction.tagIds && transaction.tagIds.length">
                                                        <v-icon size="12" :icon="mdiBriefcaseOutline" class="me-1" />
                                                        <span :key="tagId" v-for="(tagId, tIdx) in transaction.tagIds">{{ tIdx > 0 ? ', ' : '' }}{{ allTransactionTags[tagId]?.name }}</span>
                                                    </div>
                                                </td>
                                                <td class="transaction-table-column-actions text-right">
                                                    <div class="d-flex align-center justify-end">
                                                        <div class="transaction-row-actions d-flex align-center">
                                                            <v-btn v-if="transaction.planned" color="primary" variant="tonal" size="x-small"
                                                                   :prepend-icon="mdiCheckCircleOutline" class="me-1"
                                                                   :disabled="confirmingPlannedTransaction"
                                                                   @click.stop="confirmPlannedTransaction(transaction)">
                                                                {{ tt('Confirm') }}
                                                            </v-btn>
                                                            <v-btn icon variant="text" size="x-small" color="default"
                                                                   @click.stop="show(transaction)">
                                                                <v-icon :icon="mdiPencilOutline" size="18" />
                                                                <v-tooltip activator="parent">{{ tt('Edit') }}</v-tooltip>
                                                            </v-btn>
                                                            <v-btn icon variant="text" size="x-small" color="default" class="ms-1"
                                                                   @click.stop="duplicateTransaction(transaction)">
                                                                <v-icon :icon="mdiContentDuplicate" size="18" />
                                                                <v-tooltip activator="parent">{{ tt('Duplicate') }}</v-tooltip>
                                                            </v-btn>
                                                            <v-btn v-if="!transaction.planned" icon variant="text" size="x-small" color="default" class="ms-1"
                                                                   :disabled="settingTransactionPlanned"
                                                                   @click.stop="convertToPlanned(transaction)">
                                                                <v-icon :icon="mdiCalendarClock" size="18" />
                                                                <v-tooltip activator="parent">{{ tt('Convert to Planned') }}</v-tooltip>
                                                            </v-btn>
                                                            <v-btn icon variant="text" size="x-small" color="error" class="ms-1"
                                                                   @click.stop="deleteTransaction(transaction)">
                                                                <v-icon :icon="mdiDeleteOutline" size="18" />
                                                                <v-tooltip activator="parent">{{ tt('Delete') }}</v-tooltip>
                                                            </v-btn>
                                                        </div>
                                                        <v-icon v-if="transaction.sourceTemplateId && transaction.sourceTemplateId !== '0'"
                                                                :icon="mdiAutorenew" size="16" class="ms-2" color="primary" />
                                                    </div>
                                                </td>
                                            </tr>
                                            <!-- Split child rows -->
                                            <template v-if="transaction.splits && transaction.splits.length > 0 && expandedSplitIds.has(transaction.id)">
                                                <tr v-for="(split, splitIdx) in transaction.splits"
                                                    :key="'split-' + transaction.id + '-' + splitIdx"
                                                    class="transaction-split-row text-sm">
                                                    <td class="transaction-table-column-amount"
                                                        :class="{ 'text-expense': (split.splitType || transaction.type) === TransactionType.Expense, 'text-income': (split.splitType || transaction.type) === TransactionType.Income }">
                                                        <span class="ps-7">{{ formatAmountToLocalizedNumeralsWithCurrency(split.amount, transaction.sourceAccount ? transaction.sourceAccount.currency : undefined) }}</span>
                                                    </td>
                                                    <td class="transaction-table-column-counterparty">
                                                        <div v-if="split.tagIds && split.tagIds.length > 0" class="text-caption text-medium-emphasis">
                                                            <v-icon size="12" :icon="mdiBriefcaseOutline" class="me-1" />
                                                            <span :key="tagId" v-for="(tagId, tIdx) in split.tagIds">{{ tIdx > 0 ? ', ' : '' }}{{ allTransactionTags[tagId]?.name }}</span>
                                                        </div>
                                                    </td>
                                                    <td class="transaction-table-column-category">
                                                        <span>{{ getSplitCategoryName(split.categoryId) }}</span>
                                                    </td>
                                                    <td class="transaction-table-column-actions"></td>
                                                </tr>
                                            </template>
                                        </tbody>
                                    </v-table>

                                    <div class="mt-2 mb-4 d-flex justify-center" v-if="hasMorePages" ref="loadMoreTrigger">
                                        <v-btn variant="text" color="primary" :loading="loadingMore" @click="loadMore">
                                            {{ tt('Load More') }}
                                        </v-btn>
                                    </div>
                                    <div class="mt-2 mb-4 d-flex justify-center text-body-2 text-medium-emphasis" v-else-if="!loading && displayTransactions.length > 0">
                                        {{ tt('All transactions loaded') }}
                                    </div>
                                </v-card>
                            </v-window-item>
                        </v-window>
            </v-card>
        </v-col>
    </v-row>

    <date-range-selection-dialog :title="tt('Custom Date Range')"
                                 :min-time="customMinDatetime"
                                 :max-time="customMaxDatetime"
                                 v-model:show="showCustomDateRangeDialog"
                                 @dateRange:change="changeCustomDateFilter"
                                 @error="onShowDateRangeError" />

    <edit-dialog ref="editDialog" :type="TransactionEditPageType.Transaction" />
    <a-i-image-recognition-dialog ref="aiImageRecognitionDialog" />
    <import-dialog ref="importDialog" :persistent="true" />

    <v-dialog width="800" v-model="showFilterAccountDialog">
        <account-filter-settings-card type="transactionListCurrent" :dialog-mode="true"
                                      @settings:change="changeMultipleAccountsFilter" />
    </v-dialog>

    <v-dialog width="800" v-model="showFilterCategoryDialog">
        <category-filter-settings-card type="transactionListCurrent" :dialog-mode="true" :category-types="allowCategoryTypes"
                                       @settings:change="changeMultipleCategoriesFilter" />
    </v-dialog>

    <v-dialog width="800" v-model="showFilterTagDialog">
        <transaction-tag-filter-settings-card type="transactionListCurrent" :dialog-mode="true"
                                       @settings:change="changeMultipleTagsFilter" />
    </v-dialog>

    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />

    <v-dialog persistent min-width="360" width="auto" v-model="showDeletePlannedDialog">
        <v-card>
            <v-toolbar color="error">
                <v-toolbar-title>{{ tt('Delete Planned Transaction') }}</v-toolbar-title>
            </v-toolbar>
            <v-card-text class="pa-4 pb-6">{{ tt('This is a planned transaction. What do you want to delete?') }}</v-card-text>
            <v-card-actions class="px-4 pb-4 d-flex flex-wrap justify-end ga-2">
                <v-btn color="gray" @click="showDeletePlannedDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn color="error" variant="tonal" @click="showDeletePlannedDialog = false; doDeleteOnePlanned()">{{ tt('Delete Only This') }}</v-btn>
                <v-btn color="error" @click="showDeletePlannedDialog = false; doDeleteAllFuturePlanned()">{{ tt('Delete All Future') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>

<script setup lang="ts">
import { VMenu } from 'vuetify/components/VMenu';
// PaginationButtons replaced with infinite scroll
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import EditDialog from './list/dialogs/EditDialog.vue';
import AIImageRecognitionDialog from './list/dialogs/AIImageRecognitionDialog.vue';
import ImportDialog from './import/ImportDialog.vue';
import AccountFilterSettingsCard from '@/views/desktop/common/cards/AccountFilterSettingsCard.vue';
import CategoryFilterSettingsCard from '@/views/desktop/common/cards/CategoryFilterSettingsCard.vue';
import DailyBalanceForecastCard from '@/views/desktop/overview/cards/DailyBalanceForecastCard.vue';
import TransactionTagFilterSettingsCard from '@/views/desktop/common/cards/TransactionTagFilterSettingsCard.vue';
import { TransactionEditPageType } from '@/views/base/transactions/TransactionEditPageBase.ts';

import { ref, computed, useTemplateRef, watch, nextTick, onBeforeUnmount } from 'vue';
import { useRouter, onBeforeRouteUpdate } from 'vue-router';
import { useTheme } from 'vuetify';

import { useI18n } from '@/locales/helpers.ts';
import { TransactionListPageType, useTransactionListPageBase } from '@/views/base/transactions/TransactionListPageBase.ts';
import { useTransactionList } from '@/composables/useTransactionList.ts';

import { useSettingsStore } from '@/stores/setting.ts';
import { useUserStore } from '@/stores/user.ts';
// accountsStore, transactionTagsStore — now used internally by composable
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionsStore } from '@/stores/transaction.ts';
import { useTransactionTemplatesStore } from '@/stores/transactionTemplate.ts';
import { useCounterpartiesStore } from '@/stores/counterparty.ts';
import { useDesktopPageStore } from '@/stores/desktopPage.ts';
import { useOverviewStore } from '@/stores/overview.ts';
import { useHomePageBase } from '@/views/base/HomePageBase.ts';

// @ts-ignore — keys no longer directly used, delegated to composable
import {
    // @ts-ignore
    keys
} from '@/core/base.ts';
import {
    type TimeRangeAndDateType,
    // @ts-ignore — DateRangeScene now handled by composable
    DateRangeScene,
    DateRange
} from '@/core/datetime.ts';
import { AmountFilterType } from '@/core/numeral.ts';
import { ThemeType } from '@/core/theme.ts';
import { TransactionType } from '@/core/transaction.ts';
import { TemplateType }  from '@/core/template.ts';
import { type Transaction } from '@/models/transaction.ts';
import type { TransactionTemplate } from '@/models/transaction_template.ts';

import {
    isObject,
    isString,
    isNumber,
    objectFieldWithValueToArrayItem
} from '@/lib/common.ts';
import {
    getCurrentUnixTime,
    parseDateTimeFromUnixTime,
    getDayFirstDateTimeBySpecifiedUnixTime,
    // @ts-ignore — getDateTypeByDateRange still used in init
    getDateTypeByDateRange,
    // @ts-ignore — getDateTypeByBillingCycleDateRange now handled by composable
    getDateTypeByBillingCycleDateRange,
    getDateRangeByDateType,
    // @ts-ignore — getDateRangeByBillingCycleDateType now handled by composable
    getDateRangeByBillingCycleDateType,
    getUnixTimeBeforeUnixTime,
    getUnixTimeAfterUnixTime,
    getYearFirstUnixTime,
    getYearLastUnixTime,
    getQuarterFirstUnixTime,
    getQuarterLastUnixTime,
    getTodayFirstUnixTime,
    getYearMonthFirstUnixTime,
    getYearMonthLastUnixTime
} from '@/lib/datetime.ts';
import {
    // @ts-ignore
    categoryTypeToTransactionType,
    transactionTypeToCategoryType
} from '@/lib/category.ts';
// @ts-ignore
import { isDataExportingEnabled, isDataImportingEnabled } from '@/lib/server_settings.ts';
import { scrollToSelectedItem, startDownloadFile } from '@/lib/ui/common.ts';
// services import removed — confirmPlannedTransaction now delegated to composable
import services from '@/lib/services.ts';
import logger from '@/lib/logger.ts';

import {
    mdiMagnify,
    mdiRefresh,
    mdiBriefcaseOutline,
    mdiArrowRight,
    mdiCheckCircleOutline,
    mdiChevronLeft,
    mdiChevronRight,
    mdiPlus,
    mdiMinus,
    mdiSwapHorizontal,
    mdiFilterVariant,
    mdiPencilOutline,
    mdiDeleteOutline,
    mdiContentDuplicate,
    mdiAutorenew,
    mdiCalendarClock,
    mdiChevronDown,
    mdiChevronUp
} from '@mdi/js';

interface TransactionListProps {
    initPageType?: string;
    initDateType?: string,
    initMaxTime?: string,
    initMinTime?: string,
    initType?: string,
    initCategoryIds?: string,
    initAccountIds?: string,
    initTagFilter?: string,
    initAmountFilter?: string,
    initKeyword?: string
}

const props = defineProps<TransactionListProps>();

type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;
type SnackBarType = InstanceType<typeof SnackBar>;
type EditDialogType = InstanceType<typeof EditDialog>;
type ImportDialogType = InstanceType<typeof ImportDialog>;

interface TransactionListDisplayTotalAmount {
    income: string;
    expense: string;
    incomeCount: number;
    expenseCount: number;
    balanceAmount: string;
    balancePositive: boolean;
}

const router = useRouter();
const theme = useTheme();

const {
    tt,
    getWeekdayLongName,
    formatAmountToLocalizedNumeralsWithCurrency,
    // @ts-ignore
    formatDateTimeToGregorianLikeLongYearMonth,
    // @ts-ignore
    formatDateTimeToShortDate
} = useI18n();

const {
    pageType,
    loading,
    customMinDatetime,
    customMaxDatetime,
    // @ts-ignore
    currentCalendarDate,
    firstDayOfWeek,
    fiscalYearStart,
    defaultCurrency,
    showTotalAmountInTransactionListPage,
    allAccounts,
    // @ts-ignore
    allAccountsMap,
    // @ts-ignore
    allAvailableAccountsCount,
    allCategories,
    // @ts-ignore
    allPrimaryCategories,
    // @ts-ignore
    allAvailableCategoriesCount,
    allTransactionTags,
    query,
    // @ts-ignore
    queryMinTime,
    // @ts-ignore
    queryMaxTime,
    queryMonthlyData,
    // @ts-ignore
    queryMonth,
    queryAllFilterCategoryIds,
    // @ts-ignore
    queryAllFilterAccountIds,
    queryAllFilterTagIds,
    queryAllFilterCategoryIdsCount,
    queryAllFilterAccountIdsCount,
    queryCategoryName,
    // @ts-ignore
    queryAmount,
    // @ts-ignore
    transactionCalendarMinDate,
    // @ts-ignore
    transactionCalendarMaxDate,
    currentMonthTransactionData,
    // @ts-ignore
    isSameAsDefaultTimezoneOffsetMinutes,
    canAddTransaction,
    // @ts-ignore
    getDisplayTime,
    getDisplayLongDate,
    // @ts-ignore
    getDisplayTimezone,
    // @ts-ignore
    getDisplayTimeInDefaultTimezone,
    getDisplayAmount,
    getDisplayMonthTotalAmount,
    getTransactionTypeName,
} = useTransactionListPageBase();

const settingsStore = useSettingsStore();
const userStore = useUserStore();
// accountsStore, transactionTagsStore — now used internally by composable
const transactionCategoriesStore = useTransactionCategoriesStore();
const transactionsStore = useTransactionsStore();
const transactionTemplatesStore = useTransactionTemplatesStore();
const counterpartiesStore = useCounterpartiesStore();
const desktopPageStore = useDesktopPageStore();
const overviewStore = useOverviewStore();

const {
    dailyBalanceForecast
} = useHomePageBase();

const loadingForecast = ref<boolean>(false);

const categoryFilterMenu = useTemplateRef<VMenu>('categoryFilterMenu');
const amountFilterMenu = useTemplateRef<VMenu>('amountFilterMenu');
const accountFilterMenu = useTemplateRef<VMenu>('accountFilterMenu');

const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');
const snackbar = useTemplateRef<SnackBarType>('snackbar');
const editDialog = useTemplateRef<EditDialogType>('editDialog');
const importDialog = useTemplateRef<ImportDialogType>('importDialog');

// Stub refs required by useTransactionList deps but not used in desktop UI
const showCustomDateRangeSheet = ref<boolean>(false);
const showCustomMonthSheet = ref<boolean>(false);

// Delete planned transaction dialog
const showDeletePlannedDialog = ref<boolean>(false);
const plannedTransactionToDelete = ref<Transaction | null>(null);

// Convert to planned state
const settingTransactionPlanned = ref<boolean>(false);

// Split transactions expand/collapse state
const expandedSplitIds = ref<Set<string>>(new Set());

const {
    loadingMore,
    showPlannedTransactions,
    confirmingPlannedTransaction,
    // @ts-ignore
    loadingError,
    // @ts-ignore
    transactionToDelete,
    reload: composableReload,
    loadMore: composableLoadMore,
    changeDateFilter: composableChangeDateFilter,
    changeCustomDateFilter: composableChangeCustomDateFilter,
    changeTypeFilter: composableChangeTypeFilter,
    changeCategoryFilter: composableChangeCategoryFilter,
    changeAccountFilter: composableChangeAccountFilter,
    changeKeywordFilter: composableChangeKeywordFilter,
    confirmPlannedTransaction: composableConfirmPlannedTransaction,
    remove: composableRemove,
    removeAllFuture: composableRemoveAllFuture,
    // @ts-ignore
    changePageType,
    // @ts-ignore
    shiftDateRange,
    // @ts-ignore
    changeCustomMonthDateFilter,
    // @ts-ignore
    changeAmountFilter: composableChangeAmountFilter,
    // @ts-ignore
    changeTagFilter: composableChangeTagFilter
} = useTransactionList(
    {
        showToast: (msg: string) => snackbar.value?.showMessage(msg),
        showAlert: (msg: string) => snackbar.value?.showMessage(msg),
        showLoading: () => { /* desktop doesn't show loading overlay */ },
        hideLoading: () => { /* desktop doesn't show loading overlay */ },
        onSwipeoutDeleted: (_domId: string, done: () => void) => { done(); },
        getTransactionDomId: (transaction: Transaction) => 'transaction_' + transaction.id
    },
    {
        pageType,
        loading,
        customMinDatetime,
        customMaxDatetime,
        currentCalendarDate,
        firstDayOfWeek,
        fiscalYearStart,
        defaultCurrency,
        queryMonthlyData,
        query,
        queryAllFilterCategoryIds,
        allCategories,
        showCustomDateRangeSheet,
        showCustomMonthSheet
    }
);

// Initialize showPlannedTransactions to true for desktop (composable defaults to false)
showPlannedTransactions.value = true;

const activeTab = ref<string>('transactionPage');
// currentPage, temporaryCountPerPage, totalCount removed — pagination replaced by cursor-based loading via composable
const searchKeyword = ref<string>('');
const currentAmountFilterType = ref<string>('');
const currentAmountFilterValue1 = ref<number>(0);
const currentAmountFilterValue2 = ref<number>(0);
// currentPageTransactions and allLoadedTransactions removed — data now comes from composable/store
const categoryMenuState = ref<boolean>(false);
const amountMenuState = ref<boolean>(false);
const exportingData = ref<boolean>(false);
const showCustomDateRangeDialog = ref<boolean>(false);
const showFilterAccountDialog = ref<boolean>(false);
const showFilterCategoryDialog = ref<boolean>(false);
const showFilterTagDialog = ref<boolean>(false);
const loadMoreTrigger = ref<HTMLElement | null>(null);
const showFilterPanel = ref<boolean>(false);
const filterAccountId = ref<string>('');
const filterCounterpartyId = ref<string>('');
const customDateFrom = ref<string>('');
const customDateTo = ref<string>('');
const filterAmountMin = ref<number | null>(null);
const filterAmountMax = ref<number | null>(null);
const filterCategoryId = ref<string>('');

const isDarkMode = computed<boolean>(() => theme.global.name.value === ThemeType.Dark);

const activeFilterCount = computed<number>(() => {
    let count = 0;
    if (searchKeyword.value) count++;
    if (filterAccountId.value) count++;
    if (filterCounterpartyId.value) count++;
    if (filterCategoryId.value) count++;
    if (filterAmountMin.value !== null || filterAmountMax.value !== null) count++;
    if (query.value.amountFilter) count++;
    return count;
});

const allCategoryList = computed(() => {
    const result: { id: string; name: string }[] = [];
    for (const catType in allCategories.value) {
        const cats = allCategories.value[catType];
        if (cats && Array.isArray(cats)) {
            for (const cat of cats) {
                if (cat && cat.id && cat.name) {
                    result.push({ id: cat.id, name: cat.name });
                }
            }
        } else if (cats && cats.id && cats.name) {
            // Fallback for single object
            result.push({ id: cats.id, name: cats.name });
        }
    }
    return result;
});

const allowCategoryTypes = computed<string>(() => {
    if (TransactionType.Income <= query.value.type && query.value.type <= TransactionType.Transfer) {
        return transactionTypeToCategoryType(query.value.type)?.toString() ?? '';
    }

    return '';
});

const transactions = computed<Transaction[]>(() => {
    if (queryMonthlyData.value) {
        const transactionData = currentMonthTransactionData.value;

        if (!transactionData || !transactionData.items) {
            return [];
        }

        return transactionData.items;
    } else {
        // Flatten the store's TransactionMonthList[] into a flat Transaction[]
        const result: Transaction[] = [];
        for (const monthList of transactionsStore.transactions) {
            for (const transaction of monthList.items) {
                result.push(transaction);
            }
        }
        return result;
    }
});

const displayTransactions = computed<Transaction[]>(() => {
    const all = transactions.value;

    if (showPlannedTransactions.value) {
        const planned = all.filter(t => t.planned);
        const confirmed = all.filter(t => !t.planned);
        return [...planned, ...confirmed];
    }

    return all.filter(t => !t.planned);
});

// Auto-expand all split transactions when list changes
watch(displayTransactions, (txns) => {
    for (const t of txns) {
        if (t.splits && t.splits.length > 0) {
            expandedSplitIds.value.add(t.id);
        }
    }
}, { immediate: true });

const plannedTransactionsCount = computed<number>(() => {
    return transactions.value.filter(t => t.planned).length;
});

// @ts-ignore
const isWeekPeriod = computed<boolean>(() => {
    return query.value.dateType === DateRange.ThisWeek.type || query.value.dateType === DateRange.LastWeek.type;
});

// @ts-ignore
const isMonthPeriod = computed<boolean>(() => {
    return query.value.dateType === DateRange.ThisMonth.type || query.value.dateType === DateRange.LastMonth.type;
});

// @ts-ignore
const isYearPeriod = computed<boolean>(() => {
    return query.value.dateType === DateRange.ThisYear.type || query.value.dateType === DateRange.LastYear.type;
});

function formatShortDateWithMonthName(unixTime: number): string {
    const dt = parseDateTimeFromUnixTime(unixTime);
    const ymd = dt.toGregorianCalendarYearMonthDay();
    const day = ymd.day < 10 ? '0' + ymd.day : String(ymd.day);
    const monthShort = tt('month_short_' + ymd.month);
    return `${day} ${monthShort} ${ymd.year}`;
}

const currentPeriodLabel = computed<string>(() => {
    const dt = query.value.dateType;
    const minTime = query.value.minTime;
    const maxTime = query.value.maxTime;

    if (dt === DateRange.All.type) {
        return tt('All time');
    }

    // Use navigationMode or dateType to determine label format
    const mode = navigationMode.value ||
        ((dt === DateRange.ThisMonth.type || dt === DateRange.LastMonth.type) ? 'month' :
        (dt === DateRange.ThisYear.type || dt === DateRange.LastYear.type) ? 'year' :
        (dt === DateRange.ThisQuarter.type) ? 'quarter' : '');

    if (mode === 'month' && minTime) {
        const minDateTime = parseDateTimeFromUnixTime(minTime);
        const ymd = minDateTime.toGregorianCalendarYearMonthDay();
        return tt('month_standalone_' + ymd.month) + ' ' + ymd.year;
    }

    if (mode === 'year' && minTime) {
        const minDateTime = parseDateTimeFromUnixTime(minTime);
        const ymd = minDateTime.toGregorianCalendarYearMonthDay();
        return String(ymd.year);
    }

    if (mode === 'quarter' && minTime) {
        const minDateTime = parseDateTimeFromUnixTime(minTime);
        const ymd = minDateTime.toGregorianCalendarYearMonthDay();
        const q = Math.ceil(ymd.month / 3);
        return `Q${q} ${ymd.year}`;
    }

    if (minTime && maxTime) {
        return `${formatShortDateWithMonthName(minTime)} – ${formatShortDateWithMonthName(maxTime)}`;
    }

    return tt('Custom Date');
});

const queryType = computed<number>({
    get: () => query.value.type,
    set: (value) => changeTypeFilter(value)
});

// @ts-ignore
const queryAllSelectedFilterCategoryIds = computed<string>(() => {
    if (queryAllFilterCategoryIdsCount.value === 0) {
        return '';
    } else if (queryAllFilterCategoryIdsCount.value === 1) {
        return query.value.categoryIds;
    } else { // queryAllFilterCategoryIdsCount.value > 1
        return 'multiple';
    }
});

// @ts-ignore
const queryAllSelectedFilterAccountIds = computed<string>(() => {
    if (queryAllFilterAccountIdsCount.value === 0) {
        return '';
    } else if (queryAllFilterAccountIdsCount.value === 1) {
        return query.value.accountIds;
    } else { // queryAllFilterAccountIdsCount.value > 1
        return 'multiple';
    }
});

const hasMorePages = computed<boolean>(() => {
    if (queryMonthlyData.value) {
        return false; // Monthly data loads all transactions at once, no pagination needed
    }
    return transactionsStore.hasMoreTransaction;
});

const skeletonData = computed<number[]>(() => {
    const data: number[] = [];
    const skeletonCount = (pageType.value === TransactionListPageType.List.type
        ? settingsStore.appSettings.itemsCountInTransactionListPage
        : 3);

    for (let i = 0; i < skeletonCount; i++) {
        data.push(i);
    }

    return data;
});

const currentMonthTotalAmount = computed<TransactionListDisplayTotalAmount | null>(() => {
    // Count income/expense transactions
    const allTxns = transactions.value;
    let incomeCount = 0;
    let expenseCount = 0;
    for (const t of allTxns) {
        if (!t.planned) {
            if (t.type === TransactionType.Income) incomeCount++;
            else if (t.type === TransactionType.Expense) expenseCount++;
        }
    }

    if (queryMonthlyData.value) {
        const transactionData = currentMonthTransactionData.value;

        if (!transactionData) {
            return null;
        }

        const rawBalance = transactionData.totalAmount.income - transactionData.totalAmount.expense;
        const balancePositive = rawBalance >= 0;
        const balanceAbs = Math.abs(rawBalance);
        const incompleteBalance = transactionData.totalAmount.incompleteIncome || transactionData.totalAmount.incompleteExpense;

        return {
            income: getDisplayMonthTotalAmount(transactionData.totalAmount.income, false, '', transactionData.totalAmount.incompleteIncome),
            expense: getDisplayMonthTotalAmount(transactionData.totalAmount.expense, false, '', transactionData.totalAmount.incompleteExpense),
            incomeCount,
            expenseCount,
            balanceAmount: getDisplayMonthTotalAmount(balanceAbs, false, '', incompleteBalance),
            balancePositive
        };
    } else {
        const grandTotal = transactionsStore.grandTotalAmount;

        if (grandTotal.income === 0 && grandTotal.expense === 0) {
            return null;
        }

        const rawBalance = grandTotal.income - grandTotal.expense;
        const balancePositive = rawBalance >= 0;
        const balanceAbs = Math.abs(rawBalance);
        const incompleteBalance = grandTotal.incompleteIncome || grandTotal.incompleteExpense;

        return {
            income: getDisplayMonthTotalAmount(grandTotal.income, false, '', grandTotal.incompleteIncome),
            expense: getDisplayMonthTotalAmount(grandTotal.expense, false, '', grandTotal.incompleteExpense),
            incomeCount,
            expenseCount,
            balanceAmount: getDisplayMonthTotalAmount(balanceAbs, false, '', incompleteBalance),
            balancePositive
        };
    }
});

function getAmountFilterParameterCount(filterType: string): number {
    const amountFilterType = AmountFilterType.valueOf(filterType);
    return amountFilterType ? amountFilterType.paramCount : 0;
}

let skipNextRouteUpdate = false;

function updateUrlOnly(): void {
    skipNextRouteUpdate = true;
    router.replace(`/transaction/list?${transactionsStore.getTransactionListPageParams(pageType.value)}`);
}

function updateUrlWhenChanged(changed: boolean): void {
    if (changed) {
        loading.value = true;
        transactionsStore.clearTransactions();
        router.push(`/transaction/list?${transactionsStore.getTransactionListPageParams(pageType.value)}`);
    }
}

function init(initProps: TransactionListProps): void {
    // Desktop-specific: handle pageType
    if (initProps.initPageType) {
        const type = TransactionListPageType.valueOf(parseInt(initProps.initPageType));

        if (type) {
            pageType.value = type.type;
        }
    }

    searchKeyword.value = initProps.initKeyword || '';
    currentAmountFilterType.value = '';

    // Initialize filter (includes amountFilter which composable's init doesn't support)
    // TODO: once composable's InitQuery supports amountFilter, delegate to composable's init
    let dateRange: TimeRangeAndDateType | null = getDateRangeByDateType(initProps.initDateType ? parseInt(initProps.initDateType) : undefined, firstDayOfWeek.value, fiscalYearStart.value);

    if (!dateRange && initProps.initDateType && initProps.initMaxTime && initProps.initMinTime &&
        (DateRange.isBillingCycle(parseInt(initProps.initDateType)) || initProps.initDateType === DateRange.Custom.type.toString()) &&
        parseInt(initProps.initMaxTime) > 0 && parseInt(initProps.initMinTime) > 0) {
        dateRange = {
            dateType: parseInt(initProps.initDateType),
            maxTime: parseInt(initProps.initMaxTime),
            minTime: parseInt(initProps.initMinTime)
        };
    }

    transactionsStore.initTransactionListFilter({
        dateType: dateRange ? dateRange.dateType : undefined,
        maxTime: dateRange ? dateRange.maxTime : undefined,
        minTime: dateRange ? dateRange.minTime : undefined,
        type: initProps.initType && parseInt(initProps.initType) > 0 ? parseInt(initProps.initType) : undefined,
        categoryIds: initProps.initCategoryIds,
        accountIds: initProps.initAccountIds,
        tagFilter: initProps.initTagFilter,
        amountFilter: initProps.initAmountFilter || '',
        keyword: initProps.initKeyword || ''
    });

    // Use composable's reload for core data loading
    reload(false, true);

    // Desktop-specific: load templates
    transactionTemplatesStore.loadAllTemplates({
        templateType: TemplateType.Normal.type,
        force: false
    });
}

function reload(force: boolean, init: boolean): void {
    // Desktop-specific: check for add-dialog trigger on init
    if (init) {
        if (desktopPageStore.showAddTransactionDialogInTransactionList) {
            desktopPageStore.resetShowAddTransactionDialogInTransactionList();
            add();
        }
    }

    // Delegate core data loading to composable
    composableReload(force ? () => { /* done callback signals forced refresh */ } : undefined);

    // Desktop-specific: load counterparties in parallel
    counterpartiesStore.loadAllCounterparties({ force: false });

    // Desktop-specific extras (forecast, infinite scroll)
    reloadDesktopExtras(force);
}

function reloadDesktopExtras(force: boolean): void {
    // Load forecast data in background
    // Expand time window: if viewed period is in the future, load from today
    // so that planned transactions between today and period start are included
    const todayStart = getTodayFirstUnixTime();
    const forecastStart = Math.min(todayStart, query.value.minTime);
    const forecastEnd = Math.max(todayStart, query.value.maxTime);
    loadingForecast.value = true;
    overviewStore.loadMonthlyTransactionsForBalanceForecast({
        force: force,
        startTime: forecastStart,
        endTime: forecastEnd,
        displayStartTime: query.value.minTime,
        displayEndTime: query.value.maxTime
    }).then(() => {
        loadingForecast.value = false;
    }).catch(() => {
        loadingForecast.value = false;
    });

    // Set up infinite scroll observer after data loads
    nextTick(() => setupInfiniteScroll());
}

function loadMore(): void {
    composableLoadMore(true);
    // Re-observe for next load after composable finishes
    nextTick(() => setupInfiniteScroll());
}

let infiniteScrollObserver: IntersectionObserver | null = null;

function setupInfiniteScroll(): void {
    // Clean up previous observer
    if (infiniteScrollObserver) {
        infiniteScrollObserver.disconnect();
        infiniteScrollObserver = null;
    }

    if (!hasMorePages.value || !loadMoreTrigger.value) {
        return;
    }

    infiniteScrollObserver = new IntersectionObserver((entries) => {
        if (entries.length > 0 && entries[0]!.isIntersecting && !loadingMore.value) {
            loadMore();
        }
    }, { threshold: 0.1 });

    infiniteScrollObserver.observe(loadMoreTrigger.value);
}

// Navigation mode: preserved across navigations so we know how to shift and label
const navigationMode = ref<string>(''); // 'week', 'month', 'quarter', 'year', ''

function navigatePeriod(direction: number): void {
    const dt = query.value.dateType;
    const currentMin = query.value.minTime;
    const currentMax = query.value.maxTime;

    if (!currentMin || !currentMax || dt === DateRange.All.type) {
        return;
    }

    let newMin: number;
    let newMax: number;

    // Parse the start of current period to determine year/month
    const minDt = parseDateTimeFromUnixTime(currentMin);
    const ymd = minDt.toGregorianCalendarYearMonthDay();

    // Determine navigation mode from dateType or preserved mode
    const mode = navigationMode.value ||
        ((dt === DateRange.ThisWeek.type || dt === DateRange.LastWeek.type) ? 'week' :
        (dt === DateRange.ThisMonth.type || dt === DateRange.LastMonth.type) ? 'month' :
        (dt === DateRange.ThisQuarter.type) ? 'quarter' :
        (dt === DateRange.ThisYear.type || dt === DateRange.LastYear.type) ? 'year' : '');

    if (mode === 'week') {
        if (direction > 0) {
            newMin = getUnixTimeAfterUnixTime(currentMin, 7, 'days');
            newMax = getUnixTimeAfterUnixTime(currentMax, 7, 'days');
        } else {
            newMin = getUnixTimeBeforeUnixTime(currentMin, 7, 'days');
            newMax = getUnixTimeBeforeUnixTime(currentMax, 7, 'days');
        }
        navigationMode.value = 'week';
    } else if (mode === 'month') {
        let targetMonth = ymd.month + direction;
        let targetYear = ymd.year;
        if (targetMonth > 12) { targetMonth -= 12; targetYear++; }
        if (targetMonth < 1) { targetMonth += 12; targetYear--; }
        newMin = getYearMonthFirstUnixTime({ year: targetYear, month1base: targetMonth });
        newMax = getYearMonthLastUnixTime({ year: targetYear, month1base: targetMonth });
        navigationMode.value = 'month';
    } else if (mode === 'quarter') {
        const currentQuarter = Math.ceil(ymd.month / 3);
        let targetQuarter = currentQuarter + direction;
        let targetYear = ymd.year;
        if (targetQuarter > 4) { targetQuarter = 1; targetYear++; }
        if (targetQuarter < 1) { targetQuarter = 4; targetYear--; }
        newMin = getQuarterFirstUnixTime({ year: targetYear, quarter: targetQuarter });
        newMax = getQuarterLastUnixTime({ year: targetYear, quarter: targetQuarter });
        navigationMode.value = 'quarter';
    } else if (mode === 'year') {
        const targetYear = ymd.year + direction;
        newMin = getYearFirstUnixTime(targetYear);
        newMax = getYearLastUnixTime(targetYear);
        navigationMode.value = 'year';
    } else {
        const duration = currentMax - currentMin;
        if (direction > 0) {
            newMin = currentMax + 1;
            newMax = newMin + duration;
        } else {
            newMax = currentMin - 1;
            newMin = newMax - duration;
        }
    }

    changeCustomDateFilter(newMin, newMax);
}

function changeDateFilter(dateRange: TimeRangeAndDateType | number | null): void {
    navigationMode.value = ''; // reset navigation mode when user picks a period from menu

    // Handle custom date range dialog (desktop-specific)
    if (dateRange === DateRange.Custom.type || (isObject(dateRange) && dateRange.dateType === DateRange.Custom.type && !dateRange.minTime && !dateRange.maxTime)) {
        if (!query.value.minTime || !query.value.maxTime) {
            customMaxDatetime.value = getCurrentUnixTime();
            customMinDatetime.value = getDayFirstDateTimeBySpecifiedUnixTime(customMaxDatetime.value).getUnixTime();
        } else {
            customMaxDatetime.value = query.value.maxTime;
            customMinDatetime.value = query.value.minTime;
        }

        showCustomDateRangeDialog.value = true;
        return;
    }

    // Delegate to composable for numeric date types
    if (isNumber(dateRange)) {
        composableChangeDateFilter(dateRange);
        updateUrlOnly();
        // Reload desktop-specific data (forecast, pagination reset)
        reloadDesktopExtras(false);
    }
}

function changeCustomDateFilter(minTime: number, maxTime: number): void {
    showCustomDateRangeDialog.value = false;
    composableChangeCustomDateFilter(minTime, maxTime);
    updateUrlOnly();
    reloadDesktopExtras(false);
}

function changeTypeFilter(type: number): void {
    composableChangeTypeFilter(type);
    updateUrlOnly();
    reloadDesktopExtras(false);
}

// @ts-ignore
function changeCategoryFilter(categoryIds: string): void {
    categoryMenuState.value = false;
    composableChangeCategoryFilter(categoryIds);
    updateUrlOnly();
    reloadDesktopExtras(false);
}

function changeMultipleCategoriesFilter(changed: boolean): void {
    categoryMenuState.value = false;
    showFilterCategoryDialog.value = false;
    updateUrlWhenChanged(changed);
}

// @ts-ignore
function changeAccountFilter(accountIds: string): void {
    composableChangeAccountFilter(accountIds);
    updateUrlOnly();
    reloadDesktopExtras(false);
}

function changeMultipleAccountsFilter(changed: boolean): void {
    showFilterAccountDialog.value = false;
    updateUrlWhenChanged(changed);
}

function changeMultipleTagsFilter(changed: boolean): void {
    showFilterTagDialog.value = false;

    updateUrlWhenChanged(changed);
}

function changeKeywordFilter(keyword: string): void {
    composableChangeKeywordFilter(keyword);
    updateUrlOnly();
    reloadDesktopExtras(false);
}

// @ts-ignore
function changeAmountFilter(filterType: string): void {
    currentAmountFilterType.value = '';
    amountMenuState.value = false;

    if (query.value.amountFilter === filterType) {
        return;
    }

    let amountFilter = filterType;

    if (filterType) {
        const amountCount = getAmountFilterParameterCount(filterType);

        if (!amountCount) {
            return;
        }

        if (amountCount === 1) {
            amountFilter += ':' + currentAmountFilterValue1.value;
        } else if (amountCount === 2) {
            if (currentAmountFilterValue2.value < currentAmountFilterValue1.value) {
                snackbar.value?.showMessage(tt('Incorrect amount range'));
                return;
            }

            amountFilter += ':' + currentAmountFilterValue1.value + ':' + currentAmountFilterValue2.value;
        } else {
            return;
        }
    }

    if (query.value.amountFilter === amountFilter) {
        return;
    }

    const changed = transactionsStore.updateTransactionListFilter({
        amountFilter: amountFilter
    });

    updateUrlWhenChanged(changed);
}

function add(template?: TransactionTemplate): void {
    const currentUnixTime = getCurrentUnixTime();

    let newTransactionTime: number | undefined = undefined;

    if (query.value.maxTime && query.value.minTime) {
        if (query.value.maxTime < currentUnixTime) {
            newTransactionTime = query.value.maxTime;
        } else if (currentUnixTime < query.value.minTime) {
            newTransactionTime = query.value.minTime;
        }
    }

    editDialog.value?.open({
        time: newTransactionTime,
        type: query.value.type,
        categoryId: queryAllFilterCategoryIdsCount.value === 1 ? query.value.categoryIds : '',
        accountId: queryAllFilterAccountIdsCount.value === 1 ? query.value.accountIds : '',
        tagIds: objectFieldWithValueToArrayItem(queryAllFilterTagIds.value, true).join(',') || '',
        template: template
    }).then(result => {
        if (result && result.message) {
            snackbar.value?.showMessage(result.message);
        }

        reload(false, false);
    }).catch(error => {
        if (error) {
            snackbar.value?.showError(error);
        }
    });
}

function addWithType(type: number, template?: TransactionTemplate): void {
    const currentUnixTime = getCurrentUnixTime();

    let newTransactionTime: number | undefined = undefined;

    if (query.value.maxTime && query.value.minTime) {
        if (query.value.maxTime < currentUnixTime) {
            newTransactionTime = query.value.maxTime;
        } else if (currentUnixTime < query.value.minTime) {
            newTransactionTime = query.value.minTime;
        }
    }

    editDialog.value?.open({
        time: newTransactionTime,
        type: type,
        categoryId: queryAllFilterCategoryIdsCount.value === 1 ? query.value.categoryIds : '',
        accountId: queryAllFilterAccountIdsCount.value === 1 ? query.value.accountIds : '',
        tagIds: objectFieldWithValueToArrayItem(queryAllFilterTagIds.value, true).join(',') || '',
        template: template
    }).then(result => {
        if (result && result.message) {
            snackbar.value?.showMessage(result.message);
        }

        reload(false, false);
    }).catch(error => {
        if (error) {
            snackbar.value?.showError(error);
        }
    });
}

// @ts-ignore
function exportTransactions(fileExtension: string): void {
    if (exportingData.value) {
        return;
    }

    const nickname = userStore.currentUserNickname;
    let exportFileName = '';

    if (nickname) {
        exportFileName = tt('dataExport.exportFilename', {
            nickname: nickname
        }) + '.' + fileExtension;
    } else {
        exportFileName = tt('dataExport.defaultExportFilename') + '.' + fileExtension;
    }

    const exportTransactionReq = transactionsStore.getExportTransactionDataRequestByTransactionFilter();

    exportingData.value = true;

    userStore.getExportedUserData(fileExtension, exportTransactionReq).then(data => {
        startDownloadFile(exportFileName, data);
        exportingData.value = false;
    }).catch(error => {
        exportingData.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function show(transaction: Transaction): void {
    editDialog.value?.open({
        id: transaction.id,
        currentTransaction: transaction
    }).then(result => {
        if (result && result.message) {
            snackbar.value?.showMessage(result.message);
        }

        reload(false, false);
    }).catch(error => {
        if (error) {
            snackbar.value?.showError(error);
        }
    });
}

function confirmPlannedTransaction(transaction: Transaction): void {
    confirmDialog.value?.open(tt('Confirm Planned Transaction'), tt('Are you sure you want to confirm this planned transaction? The transaction date will be set to today.')).then(() => {
        composableConfirmPlannedTransaction(transaction);
    }).catch(() => {
        // User cancelled the confirmation dialog
    });
}

// @ts-ignore
function scrollCategoryMenuToSelectedItem(opened: boolean): void {
    if (opened) {
        scrollMenuToSelectedItem(categoryFilterMenu.value);
    }
}

// @ts-ignore
function scrollAmountMenuToSelectedItem(opened: boolean): void {
    if (opened) {
        currentAmountFilterType.value = '';

        let amount1 = 0, amount2 = 0;

        if (isString(query.value.amountFilter)) {
            try {
                const filterItems = query.value.amountFilter.split(':');
                const amountCount = getAmountFilterParameterCount(filterItems[0] as string);

                if (filterItems.length === 2 && amountCount === 1) {
                    amount1 = parseInt(filterItems[1] as string);
                } else if (filterItems.length === 3 && amountCount === 2) {
                    amount1 = parseInt(filterItems[1] as string);
                    amount2 = parseInt(filterItems[2] as string);
                }
            } catch (ex) {
                logger.warn('cannot parse amount from filter value, original value is ' + query.value.amountFilter, ex);
            }
        }

        currentAmountFilterValue1.value = amount1;
        currentAmountFilterValue2.value = amount2;

        scrollMenuToSelectedItem(amountFilterMenu.value);
    }
}

// @ts-ignore
function scrollAccountMenuToSelectedItem(opened: boolean): void {
    if (opened) {
        scrollMenuToSelectedItem(accountFilterMenu.value);
    }
}

function scrollMenuToSelectedItem(menu: VMenu | null): void {
    nextTick(() => {
        scrollToSelectedItem(menu?.contentEl, 'div.v-list', 'div.v-list', 'div.v-list-item.list-item-selected');
    });
}

function applyCustomDateRange(): void {
    if (!customDateFrom.value || !customDateTo.value) {
        return;
    }
    const fromDate = new Date(customDateFrom.value);
    const toDate = new Date(customDateTo.value);
    if (isNaN(fromDate.getTime()) || isNaN(toDate.getTime())) {
        return;
    }
    const minTime = Math.floor(fromDate.getTime() / 1000);
    const maxTime = Math.floor(toDate.getTime() / 1000) + 86399; // end of day
    changeCustomDateFilter(minTime, maxTime);
}

function applyKeywordFilter(): void {
    changeKeywordFilter(searchKeyword.value);
}

function clearAllFilters(): void {
    searchKeyword.value = '';
    filterAccountId.value = '';
    filterCounterpartyId.value = '';
    filterCategoryId.value = '';
    filterAmountMin.value = null;
    filterAmountMax.value = null;
    showFilterPanel.value = false;

    let changed = false;
    if (query.value.keyword) {
        changed = transactionsStore.updateTransactionListFilter({ keyword: '' }) || changed;
    }
    if (query.value.accountIds) {
        changed = transactionsStore.updateTransactionListFilter({ accountIds: '' }) || changed;
    }
    if (query.value.categoryIds) {
        changed = transactionsStore.updateTransactionListFilter({ categoryIds: '' }) || changed;
    }
    if (query.value.amountFilter) {
        changed = transactionsStore.updateTransactionListFilter({ amountFilter: '' }) || changed;
    }
    updateUrlWhenChanged(changed);
}

function applyAllFilters(): void {
    showFilterPanel.value = false;

    let changed = false;
    if (query.value.keyword !== searchKeyword.value) {
        changed = transactionsStore.updateTransactionListFilter({ keyword: searchKeyword.value }) || changed;
    }
    if (query.value.accountIds !== filterAccountId.value) {
        changed = transactionsStore.updateTransactionListFilter({ accountIds: filterAccountId.value }) || changed;
    }
    if (query.value.categoryIds !== filterCategoryId.value) {
        changed = transactionsStore.updateTransactionListFilter({ categoryIds: filterCategoryId.value }) || changed;
    }
    if (query.value.counterpartyId !== filterCounterpartyId.value) {
        changed = transactionsStore.updateTransactionListFilter({ counterpartyId: filterCounterpartyId.value }) || changed;
    }
    // Amount filter
    if (filterAmountMin.value !== null || filterAmountMax.value !== null) {
        const min = filterAmountMin.value || 0;
        const max = filterAmountMax.value || 0;
        let amountFilter = '';
        if (min > 0 && max > 0) {
            amountFilter = `bt:${min}:${max}`;
        } else if (min > 0) {
            amountFilter = `gte:${min}`;
        } else if (max > 0) {
            amountFilter = `lte:${max}`;
        }
        if (query.value.amountFilter !== amountFilter) {
            changed = transactionsStore.updateTransactionListFilter({ amountFilter: amountFilter }) || changed;
        }
    }
    updateUrlWhenChanged(changed);
}

function duplicateTransaction(transaction: Transaction): void {
    editDialog.value?.open({
        time: undefined,
        type: transaction.type,
        categoryId: transaction.category ? transaction.category.id : '',
        accountId: transaction.sourceAccount ? transaction.sourceAccount.id : '',
        tagIds: transaction.tagIds ? transaction.tagIds.join(',') : ''
    }).then(result => {
        if (result && result.message) {
            snackbar.value?.showMessage(result.message);
        }
        reload(false, false);
    }).catch(error => {
        if (error) {
            snackbar.value?.showError(error);
        }
    });
}

function deleteTransaction(transaction: Transaction): void {
    if (transaction.planned) {
        plannedTransactionToDelete.value = transaction;
        showDeletePlannedDialog.value = true;
    } else {
        confirmDialog.value?.open(tt('Are you sure you want to delete this transaction?')).then(() => {
            composableRemove(transaction, true, () => { /* no-op, already confirmed via dialog */ });
        }).catch(() => {
            // User cancelled
        });
    }
}

function doDeleteOnePlanned(): void {
    if (plannedTransactionToDelete.value) {
        composableRemove(plannedTransactionToDelete.value, true, () => { /* no-op */ });
        plannedTransactionToDelete.value = null;
    }
}

function doDeleteAllFuturePlanned(): void {
    if (plannedTransactionToDelete.value) {
        composableRemoveAllFuture(plannedTransactionToDelete.value);
        plannedTransactionToDelete.value = null;
    }
}

function toggleSplitExpand(transactionId: string): void {
    if (expandedSplitIds.value.has(transactionId)) {
        expandedSplitIds.value.delete(transactionId);
    } else {
        expandedSplitIds.value.add(transactionId);
    }
}

function getSplitCategoryName(categoryId: string): string {
    const category = transactionCategoriesStore.allTransactionCategoriesMap[categoryId];
    return category ? category.name : tt('Unknown Category');
}

function convertToPlanned(transaction: Transaction): void {
    confirmDialog.value?.open(tt('Convert to Planned'), tt('Are you sure you want to convert this actual transaction to a planned transaction? The transaction will no longer affect your account balances.')).then(() => {
        settingTransactionPlanned.value = true;
        services.setTransactionPlanned({ id: transaction.id, planned: true }).then(response => {
            settingTransactionPlanned.value = false;
            if (response.data && response.data.result) {
                snackbar.value?.showMessage(tt('Transaction has been converted to planned'));
            }
            reload(false, false);
        }).catch(error => {
            settingTransactionPlanned.value = false;
            if (error) {
                snackbar.value?.showError(error);
            }
        });
    }).catch(() => {
        // User cancelled
    });
}

function onShowDateRangeError(message: string): void {
    snackbar.value?.showError(message);
}

onBeforeRouteUpdate((to) => {
    if (skipNextRouteUpdate) {
        skipNextRouteUpdate = false;
        return;
    }

    if (to.query) {
        init({
            initDateType: (to.query['dateType'] as string | null) || undefined,
            initMinTime: (to.query['minTime'] as string | null) || undefined,
            initMaxTime: (to.query['maxTime'] as string | null) || undefined,
            initType: (to.query['type'] as string | null) || undefined,
            initCategoryIds: (to.query['categoryIds'] as string | null) || undefined,
            initAccountIds: (to.query['accountIds'] as string | null) || undefined,
            initTagFilter: (to.query['tagFilter'] as string | null) || undefined,
            initAmountFilter: (to.query['amountFilter'] as string | null) || undefined,
            initKeyword: (to.query['keyword'] as string | null) || undefined
        });
    } else {
        init({});
    }
});


watch(() => desktopPageStore.showAddTransactionDialogInTransactionList, (newValue) => {
    if (newValue) {
        desktopPageStore.resetShowAddTransactionDialogInTransactionList();
        add();
    }
});

onBeforeUnmount(() => {
    if (infiniteScrollObserver) {
        infiniteScrollObserver.disconnect();
        infiniteScrollObserver = null;
    }
});

init(props);
</script>

<style>
.transaction-type-inline-filter {
    max-width: 180px;
    flex-shrink: 0;
}

.transaction-type-inline-filter .v-input--density-compact {
    --v-input-control-height: 38px !important;
    --v-input-padding-top: 5px !important;
}

.transaction-keyword-filter .v-input--density-compact {
    --v-input-control-height: 38px !important;
    --v-input-padding-top: 5px !important;
    --v-input-padding-bottom: 5px !important;
    --v-input-chips-margin-top: 0px !important;
    --v-input-chips-margin-bottom: 0px !important;
    inline-size: 20rem;

    .v-field__input {
        min-block-size: 38px !important;
    }
}

.transaction-list-datetime-range {
    min-height: 28px;
    flex-wrap: wrap;
    row-gap: 1rem;
}

.transaction-list-custom-datetime-range {
    line-height: 1rem;
}

.transaction-list-datetime-range .transaction-list-datetime-range-text {
    color: rgba(var(--v-theme-on-background), var(--v-medium-emphasis-opacity)) !important;
}

.v-table.transaction-table > .v-table__wrapper > table {
    th, td {
        white-space: nowrap;
    }
}

.v-table.transaction-table .transaction-table-row-data > td {
    padding-top: 8px;
    padding-bottom: 8px;
}

.v-table.transaction-table .transaction-list-row-date > td {
    height: 40px !important;
}

.transaction-table .transaction-table-column-actions {
    min-width: 100px;
    width: 100px;
}

.transaction-table .transaction-table-row-data .transaction-row-actions {
    opacity: 0;
    transition: opacity 0.15s ease;
}

.transaction-table .transaction-table-row-data:hover .transaction-row-actions {
    opacity: 1;
}

.transaction-table .transaction-split-row > td {
    padding-top: 2px;
    padding-bottom: 2px;
    background-color: rgba(var(--v-theme-on-background), 0.02);
    border-top: none !important;
    font-size: 0.8rem;
}

.transaction-table .transaction-split-row > td:first-child {
    padding-left: 40px !important;
}

.transaction-table .transaction-table-column-amount {
    min-width: 120px;
}

.transaction-table .transaction-table-column-counterparty {
    min-width: 300px;
    width: 30%;
    padding-left: 3% !important;
    white-space: normal !important;
}

.transaction-table .transaction-table-column-category {
    min-width: 280px;
    padding-left: 30% !important;
    width: 100% !important;
    white-space: normal !important;
}

.transaction-table .transaction-table-column-category .v-btn,
.transaction-table .transaction-table-column-account .v-btn {
    font-size: 0.75rem;
}

.transaction-table .transaction-table-column-category .v-btn .v-btn__append,
.transaction-table .transaction-table-column-account .v-btn .v-btn__append {
    margin-inline-start: 0in;
}


.transaction-time-menu .item-icon,
.transaction-category-menu .item-icon,
.transaction-amount-menu .item-icon,
.transaction-account-menu .item-icon,
.transaction-tag-menu .item-icon,
.transaction-table .item-icon {
    padding-bottom: 3px;
}

.transaction-amount-filter-value {
    width: 100px;
}

.transaction-amount-filter-value input.v-field__input {
    min-height: 32px !important;
    padding: 0 8px 0 8px;
}

.transaction-category-menu .has-children-item-selected span,
.transaction-category-menu .item-in-multiple-selection span,
.transaction-account-menu .item-in-multiple-selection span,
.transaction-tag-menu .item-in-multiple-selection span {
    font-weight: bold;
}

.transaction-calendar-container .dp__main .dp__menu {
    --dp-border-radius: 6px;
    --dp-menu-border-color: rgba(var(--v-border-color), var(--v-border-opacity));
}

.transaction-calendar-container .dp__main .dp__calendar {
    --dp-border-color: rgba(var(--v-border-color), var(--v-border-opacity));
}

.transaction-calendar-container .dp__main .dp__calendar .dp__calendar_row {
    --dp-cell-size: 80px;
    --dp-primary-color: rgba(var(--v-theme-primary), var(--v-activated-opacity));
    --dp-primary-text-color: rgb(var(--v-theme-primary));
}

.transaction-calendar-container .dp__main.transaction-calendar-with-alternate-date .dp__calendar .dp__calendar_row {
    --dp-cell-size: 100px;
}

.transaction-calendar-container .dp__main .dp__calendar .dp__calendar_row > .dp__calendar_item {
    overflow: hidden;
}

.transaction-calendar-container .dp__main .dp__calendar .dp__calendar_row > .dp__calendar_item .transaction-calendar-daily-amounts > span.transaction-calendar-alternate-date {
    font-size: 0.9rem;
}

.transaction-calendar-container .dp__main .dp__calendar .dp__calendar_row > .dp__calendar_item .transaction-calendar-daily-amounts > span.transaction-calendar-daily-amount {
    font-size: 0.95rem;
}

</style>
