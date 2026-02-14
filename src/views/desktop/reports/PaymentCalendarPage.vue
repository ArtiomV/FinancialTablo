<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="title-and-toolbar d-flex align-center">
                        <span>{{ tt('Payment Calendar') }}</span>
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
                    </div>

                    <v-progress-linear v-if="loading" indeterminate color="primary"/>

                    <v-table density="compact" v-if="!loading && items.length > 0">
                        <thead>
                        <tr>
                            <th style="width: 100px">{{ tt('Date') }}</th>
                            <th style="width: 100px">{{ tt('Type') }}</th>
                            <th style="width: 120px" class="text-right">{{ tt('Amount') }}</th>
                            <th style="width: 60px">{{ tt('Currency') }}</th>
                            <th>{{ tt('Description') }}</th>
                            <th style="width: 120px" class="text-right">{{ tt('Cumulative') }}</th>
                        </tr>
                        </thead>
                        <tbody>
                        <tr v-for="(item, idx) in items" :key="idx"
                            :class="getTypeClass(item.type)">
                            <td>{{ formatDate(item.date) }}</td>
                            <td>
                                <v-chip size="x-small" :color="getTypeColor(item.type)">{{ tt(item.type) }}</v-chip>
                            </td>
                            <td class="text-right">{{ formatAmount(item.amount) }}</td>
                            <td class="text-disabled">{{ item.currency }}</td>
                            <td>{{ item.description }}</td>
                            <td class="text-right" :class="cumulativeAt(idx) >= 0 ? 'text-success' : 'text-error'">{{ formatAmount(cumulativeAt(idx)) }}</td>
                        </tr>
                        </tbody>
                    </v-table>

                    <div v-if="!loading && items.length === 0" class="text-disabled">
                        {{ tt('No upcoming payments') }}
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

import type { PaymentCalendarItem } from '@/models/report.ts';

import SnackBar from '@/components/desktop/SnackBar.vue';

import services from '@/lib/services.ts';

type SnackBarType = InstanceType<typeof SnackBar>;

const { tt } = useI18n();

const snackBar = ref<SnackBarType | null>(null);

const loading = ref<boolean>(true);

const now = new Date();
const selectedYear = ref<number>(now.getFullYear());
const selectedMonth = ref<number>(now.getMonth() + 1);

const items = ref<PaymentCalendarItem[]>([]);

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

function formatAmount(amount: number): string {
    return (amount / 100).toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function formatDate(timestamp: number): string {
    if (!timestamp) return '';
    const d = new Date(timestamp * 1000);
    return d.toLocaleDateString('ru-RU');
}

function getTypeColor(type: string): string {
    if (type === 'Receivable') return 'success';
    if (type === 'Payable') return 'error';
    if (type === 'Tax') return 'warning';
    if (type === 'Planned') return 'info';
    return 'default';
}

function getTypeClass(type: string): string {
    if (type === 'Payable' || type === 'Tax') return '';
    return '';
}

function cumulativeAt(idx: number): number {
    let sum = 0;
    for (let i = 0; i <= idx; i++) {
        const item = items.value[i]!;
        if (item.type === 'Receivable') {
            sum += item.amount;
        } else {
            sum -= item.amount;
        }
    }
    return sum;
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
        const response = await services.getPaymentCalendar({
            startTime,
            endTime
        });
        items.value = response.data.result?.items || [];
    } catch (error: any) {
        if (!error.processed) {
            snackBar.value?.showError(error);
        }
    } finally {
        loading.value = false;
    }
}

onMounted(async () => {
    await loadData();
});
</script>
