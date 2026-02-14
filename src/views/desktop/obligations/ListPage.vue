<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="title-and-toolbar d-flex align-center">
                        <span>{{ tt('Obligations') }}</span>
                        <v-btn class="ms-3" color="default" variant="outlined"
                               :disabled="loading || updating || hasEditing" @click="addObligation">{{ tt('Add') }}</v-btn>
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
                        <v-btn-toggle v-model="activeTab" mandatory density="compact" class="ms-4">
                            <v-btn value="receivable">{{ tt('Receivables') }}</v-btn>
                            <v-btn value="payable">{{ tt('Payables') }}</v-btn>
                        </v-btn-toggle>
                    </div>
                </template>

                <v-table class="obligations-table table-striped" :hover="!loading">
                    <thead>
                    <tr>
                        <th>
                            <div class="d-flex align-center">
                                <span>{{ tt('Counterparty') }}</span>
                                <span class="ms-4">{{ tt('Amount') }}</span>
                                <span class="ms-4">{{ tt('Currency') }}</span>
                                <span class="ms-4">{{ tt('Due Date') }}</span>
                                <span class="ms-4">{{ tt('Status') }}</span>
                                <span class="ms-4">{{ tt('Paid Amount') }}</span>
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

                    <tbody v-if="!loading && filteredObligations.length < 1">
                    <tr>
                        <td>
                            <span class="text-disabled">{{ tt('No available obligations') }}</span>
                        </td>
                    </tr>
                    </tbody>

                    <tbody v-if="!loading && filteredObligations.length > 0">
                    <tr v-for="obligation in filteredObligations" :key="obligation.id">
                        <td>
                            <div class="d-flex align-center" v-if="editing && editing.id === obligation.id">
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Counterparty')"
                                              class="me-2" style="max-width: 140px"
                                              v-model="editingCounterpartyName" readonly/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Amount')"
                                              class="me-2" style="max-width: 120px" type="number"
                                              v-model.number="editing.amount"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Currency')"
                                              class="me-2" style="max-width: 60px"
                                              v-model="editing.currency"/>
                                <v-text-field density="compact" variant="underlined" type="date"
                                              class="me-2" style="max-width: 130px"
                                              v-model="editingDueDateStr"/>
                                <v-select density="compact" variant="underlined" :items="statusOptions"
                                          item-title="text" item-value="value"
                                          class="me-2" style="max-width: 110px"
                                          v-model="editing.status"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Paid Amount')"
                                              class="me-2" style="max-width: 100px" type="number"
                                              v-model.number="editing.paidAmount"/>
                                <v-select density="compact" variant="underlined" :items="cfoOptions"
                                          item-title="text" item-value="value"
                                          class="me-2" style="max-width: 120px"
                                          v-model="editing.cfoId"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Comment')"
                                              class="me-2" style="max-width: 120px"
                                              v-model="editing.comment"/>
                                <v-spacer/>
                                <v-btn density="compact" color="primary" variant="text" size="24"
                                       class="ms-2" :icon="true" :disabled="updating"
                                       @click="saveObligation(obligation)">
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
                                <span class="me-4">{{ getCounterpartyName(obligation.counterpartyId) || '—' }}</span>
                                <span class="me-2">{{ formatAmount(obligation.amount) }}</span>
                                <span class="me-4 text-disabled">{{ obligation.currency }}</span>
                                <span class="me-4">{{ formatDate(obligation.dueDate) }}</span>
                                <v-chip size="small" :color="getStatusColor(obligation.status)" class="me-2">{{ getStatusText(obligation.status) }}</v-chip>
                                <span class="me-2" v-if="obligation.paidAmount">{{ tt('Paid Amount') }}: {{ formatAmount(obligation.paidAmount) }}</span>
                                <span class="me-2 text-disabled" v-if="getCfoName(obligation.cfoId)">{{ getCfoName(obligation.cfoId) }}</span>
                                <span class="text-disabled" v-if="obligation.comment">{{ obligation.comment }}</span>
                                <v-spacer/>
                                <v-btn density="compact" color="default" variant="text" size="24"
                                       class="ms-1" :icon="true" :disabled="updating || hasEditing"
                                       @click="editObligation(obligation)">
                                    <v-icon :icon="mdiSquareEditOutline" size="24" />
                                    <v-tooltip activator="parent">{{ tt('Edit') }}</v-tooltip>
                                </v-btn>
                                <v-btn density="compact" color="default" variant="text" size="24"
                                       class="ms-1" :icon="true" :disabled="updating || hasEditing"
                                       @click="confirmDelete(obligation)">
                                    <v-icon :icon="mdiDeleteOutline" size="24" />
                                    <v-tooltip activator="parent">{{ tt('Delete') }}</v-tooltip>
                                </v-btn>
                            </div>
                        </td>
                    </tr>
                    </tbody>

                    <!-- New obligation row -->
                    <tbody v-if="editing && !editing.id && !loading">
                    <tr>
                        <td>
                            <div class="d-flex align-center">
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Counterparty')"
                                              class="me-2" style="max-width: 140px"
                                              v-model="editingCounterpartyName" readonly/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Amount')"
                                              class="me-2" style="max-width: 120px" type="number"
                                              v-model.number="editing.amount"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Currency')"
                                              class="me-2" style="max-width: 60px"
                                              v-model="editing.currency"/>
                                <v-text-field density="compact" variant="underlined" type="date"
                                              class="me-2" style="max-width: 130px"
                                              v-model="editingDueDateStr"/>
                                <v-select density="compact" variant="underlined" :items="statusOptions"
                                          item-title="text" item-value="value"
                                          class="me-2" style="max-width: 110px"
                                          v-model="editing.status"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Paid Amount')"
                                              class="me-2" style="max-width: 100px" type="number"
                                              v-model.number="editing.paidAmount"/>
                                <v-select density="compact" variant="underlined" :items="cfoOptions"
                                          item-title="text" item-value="value"
                                          class="me-2" style="max-width: 120px"
                                          v-model="editing.cfoId"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Comment')"
                                              class="me-2" style="max-width: 120px"
                                              v-model="editing.comment"/>
                                <v-spacer/>
                                <v-btn density="compact" color="primary" variant="text" size="24"
                                       class="ms-2" :icon="true" :disabled="updating"
                                       @click="saveObligation(null)">
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
import { useCounterpartiesStore } from '@/stores/counterparty.ts';

import { Obligation } from '@/models/obligation.ts';

import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';

import services from '@/lib/services.ts';

type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;
type SnackBarType = InstanceType<typeof SnackBar>;

const { tt } = useI18n();
const cfosStore = useCFOsStore();
const counterpartiesStore = useCounterpartiesStore();

const loading = ref<boolean>(true);
const updating = ref<boolean>(false);
const activeTab = ref<string>('receivable');

const obligations = ref<Obligation[]>([]);
const editing = ref<Obligation | null>(null);
const editingDueDateStr = ref<string>('');
const editingCounterpartyName = ref<string>('');

const confirmDialog = ref<ConfirmDialogType | null>(null);
const snackBar = ref<SnackBarType | null>(null);

const hasEditing = computed(() => editing.value !== null);

const filteredObligations = computed(() => {
    const type = activeTab.value === 'receivable' ? 1 : 2;
    return obligations.value.filter(o => o.obligationType === type);
});

const statusOptions = computed(() => [
    { text: tt('Active'), value: 1 },
    { text: tt('Partial'), value: 2 },
    { text: tt('Paid'), value: 3 }
]);

const cfoOptions = computed(() => {
    const options = [{ text: tt('No CFO'), value: '0' }];
    for (const cfo of cfosStore.allCFOs) {
        options.push({ text: cfo.name, value: cfo.id });
    }
    return options;
});

function getStatusText(status: number): string {
    const option = statusOptions.value.find(o => o.value === status);
    return option ? option.text : '';
}

function getStatusColor(status: number): string {
    if (status === 1) return 'info';
    if (status === 2) return 'warning';
    if (status === 3) return 'success';
    return 'default';
}

function getCounterpartyName(counterpartyId: string): string {
    if (!counterpartyId || counterpartyId === '0') return '';
    const cp = counterpartiesStore.allCounterpartiesMap[counterpartyId];
    return cp ? cp.name : '';
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

function addObligation(): void {
    const type = activeTab.value === 'receivable' ? 1 : 2;
    editing.value = Obligation.createNew(type);
    editingDueDateStr.value = '';
    editingCounterpartyName.value = '';
}

function editObligation(obligation: Obligation): void {
    editing.value = obligation.clone();
    editingDueDateStr.value = unixToDateStr(obligation.dueDate);
    editingCounterpartyName.value = getCounterpartyName(obligation.counterpartyId);
}

function cancelEdit(): void {
    editing.value = null;
    editingDueDateStr.value = '';
    editingCounterpartyName.value = '';
}

async function saveObligation(_original: Obligation | null): Promise<void> {
    if (!editing.value) return;

    editing.value.dueDate = dateStrToUnix(editingDueDateStr.value);

    updating.value = true;

    try {
        if (editing.value.id) {
            const response = await services.modifyObligation(editing.value.toModifyRequest());
            const updated = Obligation.of(response.data.result);
            const idx = obligations.value.findIndex(o => o.id === updated.id);
            if (idx >= 0) {
                obligations.value[idx] = updated;
            }
        } else {
            const response = await services.addObligation(editing.value.toCreateRequest());
            obligations.value.push(Obligation.of(response.data.result));
        }
        editing.value = null;
        editingDueDateStr.value = '';
        editingCounterpartyName.value = '';
        snackBar.value?.showMessage('Obligation saved successfully');
    } catch (error: any) {
        if (!error.processed) {
            snackBar.value?.showError(error);
        }
    } finally {
        updating.value = false;
    }
}

function confirmDelete(obligation: Obligation): void {
    confirmDialog.value?.open('Are you sure you want to delete this obligation?').then(() => {
        deleteObligation(obligation);
    });
}

async function deleteObligation(obligation: Obligation): Promise<void> {
    updating.value = true;

    try {
        await services.deleteObligation({ id: obligation.id });
        obligations.value = obligations.value.filter(o => o.id !== obligation.id);
        snackBar.value?.showMessage('Obligation deleted successfully');
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
        const response = await services.getAllObligations();
        obligations.value = Obligation.ofMulti(response.data.result || []);
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

    try {
        await counterpartiesStore.loadAllCounterparties({ force: false });
    } catch {
        // ignore
    }

    await reload();
});
</script>
