<template>
    <div class="d-flex align-center flex-wrap text-caption" style="row-gap: 0.25rem" v-if="totals">
        <span class="text-medium-emphasis">{{ totals.incomeCount }} {{ tt('Inflows label') }}</span>
        <span class="text-income ms-1" v-if="!loading">+{{ totals.income }}</span>
        <span class="text-medium-emphasis ms-2">{{ totals.expenseCount }} {{ tt('Outflows label') }}</span>
        <span class="text-expense ms-1" v-if="!loading">&ndash;{{ totals.expense }}</span>
        <span class="text-medium-emphasis ms-2">{{ tt('Balance label') }}</span>
        <span :class="totals.balancePositive ? 'text-income' : 'text-expense'" class="ms-1" v-if="!loading">
            {{ totals.balancePositive ? '+' : '–' }}{{ totals.balanceAmount }}
        </span>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from '@/locales/helpers.ts';

interface TotalAmountDisplay {
    income: string;
    expense: string;
    incomeCount: number;
    expenseCount: number;
    balanceAmount: string;
    balancePositive: boolean;
}

interface TransactionTotalsBarProps {
    totals: TotalAmountDisplay | null;
    loading: boolean;
}

defineProps<TransactionTotalsBarProps>();
const { tt } = useI18n();
</script>
