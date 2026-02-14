<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="title-and-toolbar d-flex align-center">
                        <span>{{ tt('Tax Records') }}</span>
                        <v-btn class="ms-3" color="default" variant="outlined"
                               :disabled="loading || updating || hasEditing" @click="addTaxRecord">{{ tt('Add') }}</v-btn>
                        <v-btn density="compact" color="default" variant="text" size="24"
                               class="ms-2" :icon="true" :disabled="loading || updating || hasEditing"
                               :loading="loading" @click="reload">
                            <template #loader>
                                <v-progress-circular indeterminate size="20"/>
                            </template>
                            <v-icon :icon="mdiRefresh" size="24" />
                            <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                        </v-btn>
                        <v-spacer/>
                    </div>
                </template>

                <v-table class="tax-records-table table-striped" :hover="!loading">
                    <thead>
                    <tr>
                        <th>
                            <div class="d-flex align-center">
                                <span>{{ tt('Tax Type') }}</span>
                                <span class="ms-4">{{ tt('Period') }}</span>
                                <span class="ms-4">{{ tt('Taxable Income') }}</span>
                                <span class="ms-4">{{ tt('Tax Amount') }}</span>
                                <span class="ms-4">{{ tt('Paid Amount') }}</span>
                                <span class="ms-4">{{ tt('Due Date') }}</span>
                                <span class="ms-4">{{ tt('Status') }}</span>
                                <v-spacer/>
                                <span>{{ tt('Operation') }}</span>
                            </div>
                        </th>
                    </tr>
                    </thead>

                    <tbody v-if="loading">
                    <tr>
                        <td>
                            <v-skeleton-loader type="text" :loading="true"/>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <v-skeleton-loader type="text" :loading="true"/>
                        </td>
                    </tr>
                    </tbody>

                    <tbody v-if="!loading && taxRecords.length < 1">
                    <tr>
                        <td>
                            <span class="text-disabled">{{ tt('No available tax records') }}</span>
                        </td>
                    </tr>
                    </tbody>

                    <tbody v-if="!loading && taxRecords.length > 0">
                    <tr v-for="record in taxRecords" :key="record.id">
                        <td>
                            <div class="d-flex align-center" v-if="editing && editing.id === record.id">
                                <v-select density="compact" variant="underlined" :items="taxTypeOptions"
                                          item-title="text" item-value="value"
                                          class="me-2" style="max-width: 130px"
                                          v-model="editing.taxType"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Year')"
                                              class="me-2" style="max-width: 70px" type="number"
                                              v-model.number="editing.periodYear"/>
                                <v-select density="compact" variant="underlined" :items="quarterOptions"
                                          item-title="text" item-value="value"
                                          class="me-2" style="max-width: 70px"
                                          v-model="editing.periodQuarter"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Taxable Income')"
                                              class="me-2" style="max-width: 110px" type="number"
                                              v-model.number="editing.taxableIncome"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Tax Amount')"
                                              class="me-2" style="max-width: 100px" type="number"
                                              v-model.number="editing.taxAmount"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Paid Amount')"
                                              class="me-2" style="max-width: 100px" type="number"
                                              v-model.number="editing.paidAmount"/>
                                <v-text-field density="compact" variant="underlined" type="date"
                                              class="me-2" style="max-width: 130px"
                                              v-model="editingDueDateStr"/>
                                <v-select density="compact" variant="underlined" :items="taxStatusOptions"
                                          item-title="text" item-value="value"
                                          class="me-2" style="max-width: 110px"
                                          v-model="editing.status"/>
                                <v-select density="compact" variant="underlined" :items="cfoOptions"
                                          item-title="text" item-value="value"
                                          class="me-2" style="max-width: 120px"
                                          v-model="editing.cfoId"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Comment')"
                                              class="me-2" style="max-width: 100px"
                                              v-model="editing.comment"/>
                                <v-spacer/>
                                <v-btn density="compact" color="primary" variant="text" size="24"
                                       class="ms-2" :icon="true" :disabled="updating"
                                       @click="saveTaxRecord(record)">
                                    <v-icon :icon="mdiCheck" size="24" />
                                    <v-tooltip activator="parent">{{ tt('Save') }}</v-tooltip>
                                </v-btn>
                                <v-btn density="compact" color="default" variant="text" size="24"
                                       class="ms-2" :icon="true" :disabled="updating"
                                       @click="cancelEdit">
                                    <v-icon :icon="mdiClose" size="24" />
                                    <v-tooltip activator="parent">{{ tt('Cancel') }}</v-tooltip>
                                </v-btn>
                            </div>
                            <div class="d-flex align-center" v-else>
                                <v-chip size="small" class="me-2">{{ getTaxTypeText(record.taxType) }}</v-chip>
                                <span class="me-4">{{ record.periodYear }} {{ tt('Quarter') }} {{ record.periodQuarter }}</span>
                                <span class="me-2">{{ tt('Taxable Income') }}: {{ formatAmount(record.taxableIncome) }}</span>
                                <span class="me-2">{{ tt('Tax Amount') }}: {{ formatAmount(record.taxAmount) }}</span>
                                <span class="me-2" v-if="record.paidAmount">{{ tt('Paid Amount') }}: {{ formatAmount(record.paidAmount) }}</span>
                                <span class="me-4">{{ formatDate(record.dueDate) }}</span>
                                <v-chip size="small" :color="getTaxStatusColor(record.status)" class="me-2">{{ getTaxStatusText(record.status) }}</v-chip>
                                <span class="me-2 text-disabled" v-if="getCfoName(record.cfoId)">{{ getCfoName(record.cfoId) }}</span>
                                <span class="text-disabled" v-if="record.comment">{{ record.comment }}</span>
                                <v-spacer/>
                                <v-btn density="compact" color="default" variant="text" size="24"
                                       class="ms-1" :icon="true" :disabled="updating || hasEditing"
                                       @click="editTaxRecord(record)">
                                    <v-icon :icon="mdiSquareEditOutline" size="24" />
                                    <v-tooltip activator="parent">{{ tt('Edit') }}</v-tooltip>
                                </v-btn>
                                <v-btn density="compact" color="default" variant="text" size="24"
                                       class="ms-1" :icon="true" :disabled="updating || hasEditing"
                                       @click="confirmDelete(record)">
                                    <v-icon :icon="mdiDeleteOutline" size="24" />
                                    <v-tooltip activator="parent">{{ tt('Delete') }}</v-tooltip>
                                </v-btn>
                            </div>
                        </td>
                    </tr>
                    </tbody>

                    <!-- New tax record row -->
                    <tbody v-if="editing && !editing.id && !loading">
                    <tr>
                        <td>
                            <div class="d-flex align-center">
                                <v-select density="compact" variant="underlined" :items="taxTypeOptions"
                                          item-title="text" item-value="value"
                                          class="me-2" style="max-width: 130px"
                                          v-model="editing.taxType"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Year')"
                                              class="me-2" style="max-width: 70px" type="number"
                                              v-model.number="editing.periodYear"/>
                                <v-select density="compact" variant="underlined" :items="quarterOptions"
                                          item-title="text" item-value="value"
                                          class="me-2" style="max-width: 70px"
                                          v-model="editing.periodQuarter"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Taxable Income')"
                                              class="me-2" style="max-width: 110px" type="number"
                                              v-model.number="editing.taxableIncome"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Tax Amount')"
                                              class="me-2" style="max-width: 100px" type="number"
                                              v-model.number="editing.taxAmount"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Paid Amount')"
                                              class="me-2" style="max-width: 100px" type="number"
                                              v-model.number="editing.paidAmount"/>
                                <v-text-field density="compact" variant="underlined" type="date"
                                              class="me-2" style="max-width: 130px"
                                              v-model="editingDueDateStr"/>
                                <v-select density="compact" variant="underlined" :items="taxStatusOptions"
                                          item-title="text" item-value="value"
                                          class="me-2" style="max-width: 110px"
                                          v-model="editing.status"/>
                                <v-select density="compact" variant="underlined" :items="cfoOptions"
                                          item-title="text" item-value="value"
                                          class="me-2" style="max-width: 120px"
                                          v-model="editing.cfoId"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Comment')"
                                              class="me-2" style="max-width: 100px"
                                              v-model="editing.comment"/>
                                <v-spacer/>
                                <v-btn density="compact" color="primary" variant="text" size="24"
                                       class="ms-2" :icon="true" :disabled="updating"
                                       @click="saveTaxRecord(null)">
                                    <v-icon :icon="mdiCheck" size="24" />
                                    <v-tooltip activator="parent">{{ tt('Save') }}</v-tooltip>
                                </v-btn>
                                <v-btn density="compact" color="default" variant="text" size="24"
                                       class="ms-2" :icon="true" :disabled="updating"
                                       @click="cancelEdit">
                                    <v-icon :icon="mdiClose" size="24" />
                                    <v-tooltip activator="parent">{{ tt('Cancel') }}</v-tooltip>
                                </v-btn>
                            </div>
                        </td>
                    </tr>
                    </tbody>
                </v-table>
            </v-card>
        </v-col>
    </v-row>

    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackBar"/>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';

import {
    mdiRefresh,
    mdiCheck,
    mdiClose,
    mdiSquareEditOutline,
    mdiDeleteOutline
} from '@mdi/js';

import { useI18n } from '@/locales/helpers.ts';
import { useCFOsStore } from '@/stores/cfo.ts';

import { TaxRecord } from '@/models/tax_record.ts';

import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';

import services from '@/lib/services.ts';

type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;
type SnackBarType = InstanceType<typeof SnackBar>;

const { tt } = useI18n();
const cfosStore = useCFOsStore();

const loading = ref<boolean>(true);
const updating = ref<boolean>(false);

const taxRecords = ref<TaxRecord[]>([]);
const editing = ref<TaxRecord | null>(null);
const editingDueDateStr = ref<string>('');

const confirmDialog = ref<ConfirmDialogType | null>(null);
const snackBar = ref<SnackBarType | null>(null);

const hasEditing = computed(() => editing.value !== null);

const taxTypeOptions = computed(() => [
    { text: tt('Income Tax'), value: 1 },
    { text: tt('VAT'), value: 2 },
    { text: tt('Property Tax'), value: 3 },
    { text: tt('Other Tax'), value: 4 }
]);

const quarterOptions = computed(() => [
    { text: 'Q1', value: 1 },
    { text: 'Q2', value: 2 },
    { text: 'Q3', value: 3 },
    { text: 'Q4', value: 4 }
]);

const taxStatusOptions = computed(() => [
    { text: tt('Pending'), value: 1 },
    { text: tt('Paid'), value: 2 },
    { text: tt('Overdue'), value: 3 }
]);

const cfoOptions = computed(() => {
    const options = [{ text: tt('No CFO'), value: '0' }];
    for (const cfo of cfosStore.allCFOs) {
        options.push({ text: cfo.name, value: cfo.id });
    }
    return options;
});

function getTaxTypeText(taxType: number): string {
    const option = taxTypeOptions.value.find(o => o.value === taxType);
    return option ? option.text : '';
}

function getTaxStatusText(status: number): string {
    const option = taxStatusOptions.value.find(o => o.value === status);
    return option ? option.text : '';
}

function getTaxStatusColor(status: number): string {
    if (status === 1) return 'warning';
    if (status === 2) return 'success';
    if (status === 3) return 'error';
    return 'default';
}

function getCfoName(cfoId: string): string {
    if (!cfoId || cfoId === '0') return '';
    const cfo = cfosStore.allCFOsMap[cfoId];
    return cfo ? cfo.name : '';
}

function formatAmount(amount: number): string {
    return (amount / 100).toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function formatDate(timestamp: number): string {
    if (!timestamp) return '';
    const d = new Date(timestamp * 1000);
    return d.toLocaleDateString('ru-RU');
}

function dateStrToUnix(dateStr: string): number {
    if (!dateStr) return 0;
    const d = new Date(dateStr);
    return Math.floor(d.getTime() / 1000);
}

function unixToDateStr(timestamp: number): string {
    if (!timestamp) return '';
    const d = new Date(timestamp * 1000);
    return d.toISOString().split('T')[0] as string;
}

function addTaxRecord(): void {
    editing.value = TaxRecord.createNew();
    editingDueDateStr.value = '';
}

function editTaxRecord(record: TaxRecord): void {
    editing.value = record.clone();
    editingDueDateStr.value = unixToDateStr(record.dueDate);
}

function cancelEdit(): void {
    editing.value = null;
    editingDueDateStr.value = '';
}

async function saveTaxRecord(_original: TaxRecord | null): Promise<void> {
    if (!editing.value) return;

    editing.value.dueDate = dateStrToUnix(editingDueDateStr.value);

    updating.value = true;

    try {
        if (editing.value.id) {
            const response = await services.modifyTaxRecord(editing.value.toModifyRequest());
            const updated = TaxRecord.of(response.data.result);
            const idx = taxRecords.value.findIndex(r => r.id === updated.id);
            if (idx >= 0) {
                taxRecords.value[idx] = updated;
            }
        } else {
            const response = await services.addTaxRecord(editing.value.toCreateRequest());
            taxRecords.value.push(TaxRecord.of(response.data.result));
        }
        editing.value = null;
        editingDueDateStr.value = '';
        snackBar.value?.showMessage('Tax record saved successfully');
    } catch (error: any) {
        if (!error.processed) {
            snackBar.value?.showError(error);
        }
    } finally {
        updating.value = false;
    }
}

function confirmDelete(record: TaxRecord): void {
    confirmDialog.value?.open('Are you sure you want to delete this tax record?').then(() => {
        deleteTaxRecord(record);
    });
}

async function deleteTaxRecord(record: TaxRecord): Promise<void> {
    updating.value = true;

    try {
        await services.deleteTaxRecord({ id: record.id });
        taxRecords.value = taxRecords.value.filter(r => r.id !== record.id);
        snackBar.value?.showMessage('Tax record deleted successfully');
    } catch (error: any) {
        if (!error.processed) {
            snackBar.value?.showError(error);
        }
    } finally {
        updating.value = false;
    }
}

async function reload(): Promise<void> {
    loading.value = true;

    try {
        const response = await services.getAllTaxRecords();
        taxRecords.value = TaxRecord.ofMulti(response.data.result || []);
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

    await reload();
});
</script>
