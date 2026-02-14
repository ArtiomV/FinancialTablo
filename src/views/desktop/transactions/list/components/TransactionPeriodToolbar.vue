<template>
    <div class="d-flex align-center">
        <!-- Period navigation arrows and selector -->
        <v-btn icon size="x-small" variant="text"
               :disabled="disabled || isAllTime"
               @click="$emit('navigate', -1)">
            <v-icon :icon="mdiChevronLeft" size="18" />
        </v-btn>
        <v-menu :close-on-content-click="false">
            <template #activator="{ props: menuProps }">
                <v-btn variant="outlined" v-bind="menuProps" size="small" class="text-none font-weight-bold">
                    {{ periodLabel }}
                </v-btn>
            </template>
            <v-list density="compact">
                <v-list-item @click="$emit('changePeriod', 'thisWeek')">
                    <v-list-item-title>{{ tt('This week filter') }}</v-list-item-title>
                </v-list-item>
                <v-list-item @click="$emit('changePeriod', 'thisMonth')">
                    <v-list-item-title>{{ tt('This month filter') }}</v-list-item-title>
                </v-list-item>
                <v-list-item @click="$emit('changePeriod', 'thisQuarter')">
                    <v-list-item-title>{{ tt('This quarter filter') }}</v-list-item-title>
                </v-list-item>
                <v-list-item @click="$emit('changePeriod', 'thisYear')">
                    <v-list-item-title>{{ tt('This year filter') }}</v-list-item-title>
                </v-list-item>
                <v-list-item @click="$emit('changePeriod', 'all')">
                    <v-list-item-title>{{ tt('All time') }}</v-list-item-title>
                </v-list-item>
                <v-divider />
                <div class="px-3 py-2">
                    <div class="d-flex align-center ga-2">
                        <v-text-field type="date" density="compact" hide-details variant="outlined"
                                      :label="tt('From date')" v-model="localDateFrom" style="min-width: 110px" />
                        <v-text-field type="date" density="compact" hide-details variant="outlined"
                                      :label="tt('To date')" v-model="localDateTo" style="min-width: 110px" />
                        <v-btn size="small" color="primary" variant="tonal"
                               @click="applyCustomRange">{{ tt('Apply') }}</v-btn>
                    </div>
                </div>
            </v-list>
        </v-menu>
        <v-btn icon size="x-small" variant="text"
               :disabled="disabled || isAllTime"
               @click="$emit('navigate', 1)">
            <v-icon :icon="mdiChevronRight" size="18" />
        </v-btn>

        <!-- Type filter buttons -->
        <div class="ms-3 d-flex align-center ga-1">
            <v-btn variant="outlined" size="small"
                   :color="activeType === 0 ? 'primary' : 'default'"
                   :disabled="disabled" @click="$emit('changeType', 0)">
                {{ tt('All Filter') }}
            </v-btn>
            <v-btn variant="outlined" size="small"
                   :color="activeType === TransactionType.Income ? 'primary' : 'default'"
                   :disabled="disabled" @click="$emit('changeType', TransactionType.Income)">
                {{ tt('Income Filter') }}
            </v-btn>
            <v-btn variant="outlined" size="small"
                   :color="activeType === TransactionType.Expense ? 'primary' : 'default'"
                   :disabled="disabled" @click="$emit('changeType', TransactionType.Expense)">
                {{ tt('Expense Filter') }}
            </v-btn>
            <v-btn variant="outlined" size="small"
                   :color="activeType === TransactionType.Transfer ? 'primary' : 'default'"
                   :disabled="disabled" @click="$emit('changeType', TransactionType.Transfer)">
                {{ tt('Transfer Filter') }}
            </v-btn>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import { TransactionType } from '@/core/transaction.ts';
import { mdiChevronLeft, mdiChevronRight } from '@mdi/js';

interface TransactionPeriodToolbarProps {
    periodLabel: string;
    activeType: number;
    disabled: boolean;
    isAllTime: boolean;
}

defineProps<TransactionPeriodToolbarProps>();

const emit = defineEmits<{
    navigate: [direction: number];
    changePeriod: [period: string];
    changeType: [type: number];
    applyCustomDateRange: [minTime: number, maxTime: number];
}>();

const { tt } = useI18n();

const localDateFrom = ref<string>('');
const localDateTo = ref<string>('');

function applyCustomRange(): void {
    if (!localDateFrom.value || !localDateTo.value) {
        return;
    }
    const fromDate = new Date(localDateFrom.value);
    const toDate = new Date(localDateTo.value);
    if (isNaN(fromDate.getTime()) || isNaN(toDate.getTime())) {
        return;
    }
    const minTime = Math.floor(fromDate.getTime() / 1000);
    const maxTime = Math.floor(toDate.getTime() / 1000) + 86399;
    emit('applyCustomDateRange', minTime, maxTime);
}
</script>
