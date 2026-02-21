<template>
    <tbody :class="{ 'disabled': disabled, 'has-bottom-border': showBottomBorder }">
        <!-- Date separator row -->
        <tr class="transaction-list-row-date no-hover text-sm" v-if="showDateHeader">
            <td :colspan="4" class="font-weight-bold">
                <div class="d-flex align-center">
                    <span>{{ displayDate }}</span>
                    <v-chip class="ms-1" color="default" size="x-small"
                            v-if="transaction.displayDayOfWeek">
                        {{ weekdayName }}
                    </v-chip>
                </div>
            </td>
        </tr>
        <!-- Transaction data row -->
        <tr class="transaction-table-row-data text-sm cursor-pointer"
            :style="transaction.planned ? { opacity: 0.6 } : undefined"
            @click="$emit('show', transaction)">
            <td class="transaction-table-column-amount"
                :class="{ 'text-expense': transaction.type === TransactionType.Expense, 'text-income': transaction.type === TransactionType.Income }">
                <div v-if="transaction.sourceAccount">
                    <span>{{ displayAmount }}</span>
                </div>
                <div class="text-caption text-medium-emphasis" v-if="transaction.sourceAccount"
                     style="color: rgba(var(--v-theme-on-background), 0.5) !important">
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
                    <span v-else-if="counterpartyName">
                        {{ counterpartyName }}
                    </span>
                </div>
                <div class="text-caption text-medium-emphasis text-truncate" v-if="transaction.comment" style="max-width: 250px">
                    {{ transaction.comment }}
                </div>
            </td>
            <td class="transaction-table-column-category">
                <div>
                    <span v-if="transaction.type === TransactionType.ModifyBalance">
                        {{ tt('Modify Balance') }}
                    </span>
                    <span v-else-if="transaction.category">
                        {{ transaction.category.name }}
                    </span>
                    <span v-else>
                        {{ transactionTypeName }}
                    </span>
                </div>
                <div class="text-caption text-medium-emphasis" v-if="transaction.tagIds && transaction.tagIds.length">
                    <v-icon size="12" :icon="mdiBriefcaseOutline" class="me-1" />
                    <span :key="tagId" v-for="(tagId, tIdx) in transaction.tagIds">{{ tIdx > 0 ? ', ' : '' }}{{ tagsMap[tagId]?.name }}</span>
                </div>
            </td>
            <td class="transaction-table-column-actions text-right">
                <div class="transaction-row-actions d-flex align-center justify-end">
                    <v-btn v-if="transaction.planned" color="primary" variant="tonal" size="x-small"
                           :prepend-icon="mdiCheckCircleOutline" class="me-1"
                           :disabled="confirmingPlanned"
                           @click.stop="$emit('confirm', transaction)">
                        {{ tt('Confirm') }}
                    </v-btn>
                    <v-btn icon variant="text" size="x-small" color="default"
                           @click.stop="$emit('show', transaction)">
                        <v-icon :icon="mdiPencilOutline" size="18" />
                        <v-tooltip activator="parent">{{ tt('Edit') }}</v-tooltip>
                    </v-btn>
                    <v-btn icon variant="text" size="x-small" color="default" class="ms-1"
                           @click.stop="$emit('duplicate', transaction)">
                        <v-icon :icon="mdiContentDuplicate" size="18" />
                        <v-tooltip activator="parent">{{ tt('Duplicate') }}</v-tooltip>
                    </v-btn>
                    <v-btn icon variant="text" size="x-small" color="error" class="ms-1"
                           @click.stop="$emit('delete', transaction)">
                        <v-icon :icon="mdiDeleteOutline" size="18" />
                        <v-tooltip activator="parent">{{ tt('Delete') }}</v-tooltip>
                    </v-btn>
                    <v-icon v-if="transaction.sourceTemplateId && transaction.sourceTemplateId !== '0'"
                            :icon="mdiAutorenew" size="16" class="ms-2" color="primary"
                            style="opacity: 1; flex-shrink: 0;" />
                </div>
            </td>
        </tr>
    </tbody>
</template>

<script setup lang="ts">
import { useI18n } from '@/locales/helpers.ts';
import { TransactionType } from '@/core/transaction.ts';
import type { Transaction } from '@/models/transaction.ts';

import {
    mdiArrowRight,
    mdiBriefcaseOutline,
    mdiCheckCircleOutline,
    mdiPencilOutline,
    mdiDeleteOutline,
    mdiContentDuplicate,
    mdiAutorenew
} from '@mdi/js';

interface TransactionTableRowProps {
    transaction: Transaction;
    showDateHeader: boolean;
    showBottomBorder: boolean;
    disabled: boolean;
    confirmingPlanned: boolean;
    displayAmount: string;
    displayDate: string;
    weekdayName: string;
    transactionTypeName: string;
    counterpartyName: string;
    tagsMap: Record<string, { name: string } | undefined>;
}

defineProps<TransactionTableRowProps>();

defineEmits<{
    show: [transaction: Transaction];
    confirm: [transaction: Transaction];
    duplicate: [transaction: Transaction];
    delete: [transaction: Transaction];
}>();

const { tt } = useI18n();
</script>
