<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="title-and-toolbar d-flex align-center">
                        <span>{{ tt('Profit & Loss') }}</span>
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

                    <v-table density="compact" v-if="!loading && report">
                        <tbody>
                        <tr>
                            <td class="font-weight-bold">{{ tt('Revenue') }}</td>
                            <td class="text-right font-weight-bold">{{ formatAmount(report.revenue) }}</td>
                        </tr>
                        <tr>
                            <td class="ps-8">{{ tt('Cost of Goods') }}</td>
                            <td class="text-right text-error">- {{ formatAmount(report.costOfGoods) }}</td>
                        </tr>
                        <tr class="bg-grey-lighten-4">
                            <td class="font-weight-bold">{{ tt('Gross Profit') }}</td>
                            <td class="text-right font-weight-bold" :class="report.grossProfit >= 0 ? 'text-success' : 'text-error'">{{ formatAmount(report.grossProfit) }}</td>
                        </tr>
                        <tr>
                            <td class="ps-8">{{ tt('Operating Expenses') }}</td>
                            <td class="text-right text-error">- {{ formatAmount(report.operatingExpense) }}</td>
                        </tr>
                        <tr>
                            <td class="ps-8">{{ tt('Depreciation') }}</td>
                            <td class="text-right text-error">- {{ formatAmount(report.depreciation) }}</td>
                        </tr>
                        <tr class="bg-grey-lighten-4">
                            <td class="font-weight-bold">{{ tt('Operating Profit') }}</td>
                            <td class="text-right font-weight-bold" :class="report.operatingProfit >= 0 ? 'text-success' : 'text-error'">{{ formatAmount(report.operatingProfit) }}</td>
                        </tr>
                        <tr>
                            <td class="ps-8">{{ tt('Financial Expenses') }}</td>
                            <td class="text-right text-error">- {{ formatAmount(report.financialExpense) }}</td>
                        </tr>
                        <tr>
                            <td class="ps-8">{{ tt('Tax Expense') }}</td>
                            <td class="text-right text-error">- {{ formatAmount(report.taxExpense) }}</td>
                        </tr>
                        <tr class="bg-primary" style="color: white;">
                            <td class="font-weight-bold">{{ tt('Net Profit') }}</td>
                            <td class="text-right font-weight-bold">{{ formatAmount(report.netProfit) }}</td>
                        </tr>
                        </tbody>
                    </v-table>
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

import type { PnLResponse } from '@/models/report.ts';

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

const report = ref<PnLResponse | null>(null);

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
        const response = await services.getPnL({
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
