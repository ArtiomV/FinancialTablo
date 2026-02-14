<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="title-and-toolbar d-flex align-center">
                        <span>{{ tt('Cash Flow Statement') }}</span>
                        <v-spacer/>
                    </div>
                </template>

                <v-card-text>
                    <div class="d-flex align-center mb-4">
                        <v-select density="compact" variant="outlined" :items="yearOptions"
                                  item-title="text" item-value="value"
                                  style="max-width: 120px" class="me-2"
                                  v-model="selectedYear" @update:model-value="loadData"/>
                        <v-select density="compact" variant="outlined" :items="monthOptions"
                                  item-title="text" item-value="value"
                                  style="max-width: 140px" class="me-2"
                                  v-model="selectedMonth" @update:model-value="loadData"/>
                        <v-select density="compact" variant="outlined" :items="cfoOptions"
                                  item-title="text" item-value="value"
                                  style="max-width: 180px" class="me-2"
                                  v-model="selectedCfoId" @update:model-value="loadData"/>
                    </div>

                    <v-progress-linear v-if="loading" indeterminate color="primary"/>

                    <div v-if="!loading && report">
                        <div v-for="activity in report.activities" :key="activity.activityType" class="mb-6">
                            <div class="text-h6 mb-2">{{ getActivityName(activity.activityType) }}</div>
                            <v-table density="compact" v-if="activity.lines.length > 0">
                                <thead>
                                <tr>
                                    <th>{{ tt('Category') }}</th>
                                    <th style="width: 120px" class="text-right">{{ tt('Income') }}</th>
                                    <th style="width: 120px" class="text-right">{{ tt('Expense') }}</th>
                                    <th style="width: 120px" class="text-right">{{ tt('Net') }}</th>
                                </tr>
                                </thead>
                                <tbody>
                                <tr v-for="line in activity.lines" :key="line.categoryId">
                                    <td>{{ line.categoryName }}</td>
                                    <td class="text-right">{{ formatAmount(line.income) }}</td>
                                    <td class="text-right">{{ formatAmount(line.expense) }}</td>
                                    <td class="text-right" :class="line.net >= 0 ? 'text-success' : 'text-error'">{{ formatAmount(line.net) }}</td>
                                </tr>
                                <tr class="font-weight-bold">
                                    <td>{{ tt('Total') }}</td>
                                    <td class="text-right">{{ formatAmount(activity.totalIncome) }}</td>
                                    <td class="text-right">{{ formatAmount(activity.totalExpense) }}</td>
                                    <td class="text-right" :class="activity.totalNet >= 0 ? 'text-success' : 'text-error'">{{ formatAmount(activity.totalNet) }}</td>
                                </tr>
                                </tbody>
                            </v-table>
                            <div v-else class="text-disabled">{{ tt('No data') }}</div>
                        </div>

                        <v-divider class="my-4"/>
                        <div class="d-flex justify-end">
                            <span class="text-h6">{{ tt('Total Net Cash Flow') }}: </span>
                            <span class="text-h6 ms-2" :class="report.totalNet >= 0 ? 'text-success' : 'text-error'">{{ formatAmount(report.totalNet) }}</span>
                        </div>
                    </div>
                </v-card-text>
            </v-card>
        </v-col>
    </v-row>

    <snack-bar ref="snackBar"/>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { useCFOsStore } from '@/stores/cfo.ts';

import type { CashFlowResponse } from '@/models/report.ts';

import SnackBar from '@/components/desktop/SnackBar.vue';

import services from '@/lib/services.ts';

type SnackBarType = InstanceType<typeof SnackBar>;

const { tt } = useI18n();
const cfosStore = useCFOsStore();

const snackBar = ref<SnackBarType | null>(null);

const loading = ref<boolean>(true);

const now = new Date();
const selectedYear = ref<number>(now.getFullYear());
const selectedMonth = ref<number>(now.getMonth() + 1);
const selectedCfoId = ref<string>('0');

const report = ref<CashFlowResponse | null>(null);

const yearOptions = computed(() => {
    const currentYear = now.getFullYear();
    const options = [];
    for (let y = currentYear - 2; y <= currentYear + 2; y++) {
        options.push({ text: String(y), value: y });
    }
    return options;
});

const monthOptions = computed(() => {
    const months = [
        'January', 'February', 'March', 'April', 'May', 'June',
        'July', 'August', 'September', 'October', 'November', 'December'
    ];
    return months.map((m, i) => ({ text: tt(m), value: i + 1 }));
});

const cfoOptions = computed(() => {
    const options = [{ text: tt('All CFOs'), value: '0' }];
    for (const cfo of cfosStore.allCFOs) {
        options.push({ text: cfo.name, value: cfo.id });
    }
    return options;
});

function getActivityName(type: number): string {
    if (type === 1) return tt('Operating Activities');
    if (type === 2) return tt('Investing Activities');
    if (type === 3) return tt('Financing Activities');
    return '';
}

function formatAmount(amount: number): string {
    return (amount / 100).toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function getPeriodRange(): { startTime: number, endTime: number } {
    const start = new Date(selectedYear.value, selectedMonth.value - 1, 1);
    const end = new Date(selectedYear.value, selectedMonth.value, 1);
    return {
        startTime: Math.floor(start.getTime() / 1000),
        endTime: Math.floor(end.getTime() / 1000)
    };
}

async function loadData(): Promise<void> {
    loading.value = true;

    try {
        const { startTime, endTime } = getPeriodRange();
        const response = await services.getCashFlow({
            cfoId: selectedCfoId.value,
            startTime,
            endTime
        });
        report.value = response.data.result;
    } catch (error: any) {
        if (!error.processed) {
            snackBar.value?.showError(error);
        }
    } finally {
        loading.value = false;
    }
}

onMounted(async () => {
    try {
        await cfosStore.loadAllCFOs({ force: false });
    } catch {
        // ignore
    }
    await loadData();
});
</script>
