<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="title-and-toolbar d-flex align-center">
                        <span>{{ tt('Budgets') }}</span>
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
                        <v-btn-toggle v-model="activeTab" mandatory density="compact" class="ms-4">
                            <v-btn value="budget">{{ tt('Budget Entry') }}</v-btn>
                            <v-btn value="planfact">{{ tt('Plan-Fact') }}</v-btn>
                        </v-btn-toggle>
                    </div>

                    <!-- Budget Entry Tab -->
                    <div v-if="activeTab === 'budget'">
                        <v-table density="compact" v-if="!loading">
                            <thead>
                            <tr>
                                <th>{{ tt('Category') }}</th>
                                <th style="width: 150px">{{ tt('Planned Amount') }}</th>
                                <th style="width: 150px">{{ tt('Comment') }}</th>
                            </tr>
                            </thead>
                            <tbody>
                            <tr v-for="line in budgetLines" :key="line.categoryId">
                                <td>{{ line.categoryName }}</td>
                                <td>
                                    <v-text-field density="compact" variant="underlined" type="number"
                                                  v-model.number="line.plannedAmount" hide-details
                                                  style="max-width: 140px"/>
                                </td>
                                <td>
                                    <v-text-field density="compact" variant="underlined"
                                                  v-model="line.comment" hide-details
                                                  style="max-width: 140px"/>
                                </td>
                            </tr>
                            </tbody>
                        </v-table>
                        <v-progress-linear v-if="loading" indeterminate color="primary"/>
                        <div class="d-flex mt-4" v-if="!loading && budgetLines.length > 0">
                            <v-btn color="primary" variant="tonal" :disabled="saving" :loading="saving"
                                   @click="saveBudgets">{{ tt('Save') }}</v-btn>
                        </div>
                        <div v-if="!loading && budgetLines.length === 0" class="text-disabled">
                            {{ tt('No categories available for budgeting') }}
                        </div>
                    </div>

                    <!-- Plan-Fact Tab -->
                    <div v-if="activeTab === 'planfact'">
                        <v-table density="compact" v-if="!loadingPlanFact && planFactLines.length > 0">
                            <thead>
                            <tr>
                                <th>{{ tt('Category') }}</th>
                                <th style="width: 100px">{{ tt('Type') }}</th>
                                <th style="width: 120px">{{ tt('Plan') }}</th>
                                <th style="width: 120px">{{ tt('Fact') }}</th>
                                <th style="width: 120px">{{ tt('Deviation') }}</th>
                                <th style="width: 80px">{{ tt('Deviation %') }}</th>
                            </tr>
                            </thead>
                            <tbody>
                            <tr v-for="line in planFactLines" :key="line.categoryId"
                                :class="getDeviationClass(line)">
                                <td>{{ line.categoryName }}</td>
                                <td>{{ getCategoryTypeText(line.categoryType) }}</td>
                                <td>{{ formatAmount(line.plannedAmount) }}</td>
                                <td>{{ formatAmount(line.factAmount) }}</td>
                                <td :class="line.deviation > 0 ? 'text-success' : line.deviation < 0 ? 'text-error' : ''">
                                    {{ formatAmount(line.deviation) }}
                                </td>
                                <td>{{ line.deviationPct !== null ? line.deviationPct + '%' : '—' }}</td>
                            </tr>
                            </tbody>
                        </v-table>
                        <v-progress-linear v-if="loadingPlanFact" indeterminate color="primary"/>
                        <div v-if="!loadingPlanFact && planFactLines.length === 0" class="text-disabled">
                            {{ tt('No plan-fact data available') }}
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
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';

import type { BudgetInfoResponse, BudgetItemRequest, PlanFactLineResponse } from '@/models/budget.ts';

import SnackBar from '@/components/desktop/SnackBar.vue';

import services from '@/lib/services.ts';

type SnackBarType = InstanceType<typeof SnackBar>;

const { tt } = useI18n();
const cfosStore = useCFOsStore();
const categoriesStore = useTransactionCategoriesStore();

const snackBar = ref<SnackBarType | null>(null);

const loading = ref<boolean>(true);
const saving = ref<boolean>(false);
const loadingPlanFact = ref<boolean>(false);
const activeTab = ref<string>('budget');

const now = new Date();
const selectedYear = ref<number>(now.getFullYear());
const selectedMonth = ref<number>(now.getMonth() + 1);
const selectedCfoId = ref<string>('0');

interface BudgetLine {
    categoryId: string;
    categoryName: string;
    plannedAmount: number;
    comment: string;
}

const budgetLines = ref<BudgetLine[]>([]);
const planFactLines = ref<PlanFactLineResponse[]>([]);

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

function getCategoryTypeText(type: number): string {
    if (type === 1) return tt('Income');
    if (type === 2) return tt('Expense');
    return '';
}

function getDeviationClass(line: PlanFactLineResponse): string {
    if (line.deviation > 0 && line.categoryType === 1) return '';
    if (line.deviation < 0 && line.categoryType === 2) return '';
    return '';
}

async function loadData(): Promise<void> {
    if (activeTab.value === 'budget') {
        await loadBudgets();
    } else {
        await loadPlanFact();
    }
}

async function loadBudgets(): Promise<void> {
    loading.value = true;

    try {
        // Load categories
        await categoriesStore.loadAllCategories({ force: false });

        // Get existing budgets
        const response = await services.getBudgets({
            year: selectedYear.value,
            month: selectedMonth.value,
            cfoId: selectedCfoId.value
        });

        const existingBudgets = response.data?.result || [];
        const budgetMap: Record<string, BudgetInfoResponse> = {};
        for (const b of existingBudgets) {
            budgetMap[b.categoryId] = b;
        }

        // Build lines from all categories (income + expense, visible only)
        const lines: BudgetLine[] = [];
        const allCats = categoriesStore.allTransactionCategories;

        // Type 1 = Income, Type 2 = Expense
        for (const typeKey of [1, 2]) {
            const cats = allCats[typeKey] || [];
            for (const cat of cats) {
                if (!cat.visible) continue;

                const existing = budgetMap[cat.id];
                lines.push({
                    categoryId: cat.id,
                    categoryName: cat.name,
                    plannedAmount: existing ? existing.plannedAmount : 0,
                    comment: existing ? existing.comment : ''
                });
            }
        }

        budgetLines.value = lines;
    } catch (error: any) {
        if (!error.processed) {
            snackBar.value?.showError(error);
        }
    } finally {
        loading.value = false;
    }
}

async function saveBudgets(): Promise<void> {
    saving.value = true;

    try {
        const items: BudgetItemRequest[] = budgetLines.value
            .filter(l => l.plannedAmount !== 0 || l.comment)
            .map(l => ({
                categoryId: l.categoryId,
                plannedAmount: l.plannedAmount,
                comment: l.comment
            }));

        await services.saveBudgets({
            year: selectedYear.value,
            month: selectedMonth.value,
            cfoId: selectedCfoId.value,
            budgets: items
        });

        snackBar.value?.showMessage('Budgets saved successfully');
    } catch (error: any) {
        if (!error.processed) {
            snackBar.value?.showError(error);
        }
    } finally {
        saving.value = false;
    }
}

async function loadPlanFact(): Promise<void> {
    loadingPlanFact.value = true;

    try {
        const response = await services.getPlanFact({
            year: selectedYear.value,
            month: selectedMonth.value,
            cfoId: selectedCfoId.value
        });

        planFactLines.value = response.data?.result?.lines || [];
    } catch (error: any) {
        if (!error.processed) {
            snackBar.value?.showError(error);
        }
    } finally {
        loadingPlanFact.value = false;
    }
}

onMounted(async () => {
    try {
        await cfosStore.loadAllCFOs({ force: false });
    } catch {
        // ignore
    }

    await loadBudgets();
});
</script>
