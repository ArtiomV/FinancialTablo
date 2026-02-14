<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="title-and-toolbar d-flex align-center">
                        <span>{{ tt('Balance Sheet') }}</span>
                        <v-spacer/>
                    </div>
                </template>

                <v-card-text>
                    <div class="d-flex align-center mb-4">
                        <v-select density="compact" variant="outlined" :items="cfoOptions"
                                  item-title="text" item-value="value"
                                  style="max-width: 180px" class="me-2"
                                  v-model="selectedCfoId" @update:model-value="loadData"/>
                    </div>

                    <v-progress-linear v-if="loading" indeterminate color="primary"/>

                    <v-row v-if="!loading && report">
                        <v-col cols="6">
                            <div class="text-h6 mb-2">{{ tt('Assets') }}</div>
                            <v-table density="compact">
                                <tbody>
                                <tr v-for="line in report.assetLines" :key="line.label">
                                    <td>{{ tt(line.label) }}</td>
                                    <td class="text-right">{{ formatAmount(line.amount) }}</td>
                                </tr>
                                <tr class="bg-success" style="color: white;">
                                    <td class="font-weight-bold">{{ tt('Total Assets') }}</td>
                                    <td class="text-right font-weight-bold">{{ formatAmount(report.totalAssets) }}</td>
                                </tr>
                                </tbody>
                            </v-table>
                        </v-col>
                        <v-col cols="6">
                            <div class="text-h6 mb-2">{{ tt('Liabilities') }} + {{ tt('Equity') }}</div>
                            <v-table density="compact">
                                <tbody>
                                <tr v-for="line in report.liabilityLines" :key="line.label">
                                    <td>{{ tt(line.label) }}</td>
                                    <td class="text-right">{{ formatAmount(line.amount) }}</td>
                                </tr>
                                <tr class="bg-error" style="color: white;">
                                    <td class="font-weight-bold">{{ tt('Total Liabilities') }}</td>
                                    <td class="text-right font-weight-bold">{{ formatAmount(report.totalLiability) }}</td>
                                </tr>
                                <tr class="bg-primary" style="color: white;">
                                    <td class="font-weight-bold">{{ tt('Equity') }}</td>
                                    <td class="text-right font-weight-bold">{{ formatAmount(report.equity) }}</td>
                                </tr>
                                </tbody>
                            </v-table>
                        </v-col>
                    </v-row>
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

import type { BalanceResponse } from '@/models/report.ts';

import SnackBar from '@/components/desktop/SnackBar.vue';

import services from '@/lib/services.ts';

type SnackBarType = InstanceType<typeof SnackBar>;

const { tt } = useI18n();
const cfosStore = useCFOsStore();

const snackBar = ref<SnackBarType | null>(null);

const loading = ref<boolean>(true);

const selectedCfoId = ref<string>('0');

const report = ref<BalanceResponse | null>(null);

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

async function loadData(): Promise<void> {
    loading.value = true;

    try {
        const response = await services.getBalance({
            cfoId: selectedCfoId.value
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
