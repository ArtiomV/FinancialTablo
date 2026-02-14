<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="title-and-toolbar d-flex align-center">
                        <span>{{ tt('Investors') }}</span>
                        <v-btn class="ms-3" color="default" variant="outlined"
                               :disabled="loading || updating || hasEditingDeal" @click="addDeal">{{ tt('Add') }}</v-btn>
                        <v-btn density="compact" color="default" variant="text" size="24"
                               class="ms-2" :icon="true" :disabled="loading || updating || hasEditingDeal"
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

                <v-table class="investors-table table-striped" :hover="!loading">
                    <thead>
                    <tr>
                        <th>
                            <div class="d-flex align-center">
                                <span>{{ tt('Investor Name') }}</span>
                                <span class="ms-4">{{ tt('Deal Type') }}</span>
                                <span class="ms-4">{{ tt('Amount') }}</span>
                                <span class="ms-4">{{ tt('Currency') }}</span>
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

                    <tbody v-if="!loading && deals.length < 1">
                    <tr>
                        <td>
                            <span class="text-disabled">{{ tt('No available investor deal') }}</span>
                        </td>
                    </tr>
                    </tbody>

                    <tbody v-if="!loading && deals.length > 0">
                    <template v-for="deal in deals" :key="deal.id">
                    <tr :class="{ 'expanded-row': expandedDealId === deal.id }">
                        <td>
                            <div class="d-flex align-center" v-if="editingDeal && editingDeal.id === deal.id">
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Investor Name')"
                                              class="me-2" style="max-width: 160px"
                                              v-model="editingDeal.investorName"/>
                                <v-select density="compact" variant="underlined" :items="dealTypeOptions"
                                          item-title="text" item-value="value"
                                          class="me-2" style="max-width: 140px"
                                          v-model="editingDeal.dealType"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Amount')"
                                              class="me-2" style="max-width: 120px" type="number"
                                              v-model.number="editingDeal.investmentAmount"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Currency')"
                                              class="me-2" style="max-width: 60px"
                                              v-model="editingDeal.currency"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Annual Rate %')"
                                              class="me-2" style="max-width: 80px" type="number"
                                              v-model.number="editingDeal.annualRate"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Total to Repay')"
                                              class="me-2" style="max-width: 120px" type="number"
                                              v-model.number="editingDeal.totalToRepay"/>
                                <v-select density="compact" variant="underlined" :items="cfoOptions"
                                          item-title="text" item-value="value"
                                          class="me-2" style="max-width: 140px"
                                          v-model="editingDeal.cfoId"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Comment')"
                                              class="me-2" style="max-width: 120px"
                                              v-model="editingDeal.comment"/>
                                <v-spacer/>
                                <v-btn density="compact" color="primary" variant="text" size="24"
                                       class="ms-2" :icon="true" :disabled="updating"
                                       @click="saveDeal(deal)">
                                    <v-icon :icon="mdiCheck" size="24" />
                                    <v-tooltip activator="parent">{{ tt('Save') }}</v-tooltip>
                                </v-btn>
                                <v-btn density="compact" color="default" variant="text" size="24"
                                       class="ms-2" :icon="true" :disabled="updating"
                                       @click="cancelEditDeal">
                                    <v-icon :icon="mdiClose" size="24" />
                                    <v-tooltip activator="parent">{{ tt('Cancel') }}</v-tooltip>
                                </v-btn>
                            </div>
                            <div class="d-flex align-center" v-else>
                                <span class="me-4">{{ deal.investorName }}</span>
                                <v-chip size="small" class="me-2">{{ getDealTypeText(deal.dealType) }}</v-chip>
                                <span class="me-2">{{ formatAmount(deal.investmentAmount) }}</span>
                                <span class="me-4 text-disabled">{{ deal.currency }}</span>
                                <v-chip size="x-small" variant="outlined" class="me-2" v-if="deal.annualRate">{{ deal.annualRate }}%</v-chip>
                                <span class="me-2 text-disabled" v-if="deal.totalToRepay">{{ tt('Repay') }}: {{ formatAmount(deal.totalToRepay) }}</span>
                                <span class="me-2 text-disabled" v-if="getCfoName(deal.cfoId)">{{ getCfoName(deal.cfoId) }}</span>
                                <span class="text-disabled" v-if="deal.comment">{{ deal.comment }}</span>
                                <v-spacer/>
                                <v-btn density="compact" color="default" variant="text" size="24"
                                       class="ms-1" :icon="true" :disabled="updating || hasEditingDeal"
                                       @click="togglePayments(deal)">
                                    <v-icon :icon="expandedDealId === deal.id ? mdiChevronUp : mdiChevronDown" size="24" />
                                    <v-tooltip activator="parent">{{ tt('Payments') }}</v-tooltip>
                                </v-btn>
                                <v-btn density="compact" color="default" variant="text" size="24"
                                       class="ms-1" :icon="true" :disabled="updating || hasEditingDeal"
                                       @click="editDeal(deal)">
                                    <v-icon :icon="mdiSquareEditOutline" size="24" />
                                    <v-tooltip activator="parent">{{ tt('Edit') }}</v-tooltip>
                                </v-btn>
                                <v-btn density="compact" color="default" variant="text" size="24"
                                       class="ms-1" :icon="true" :disabled="updating || hasEditingDeal"
                                       @click="confirmDeleteDeal(deal)">
                                    <v-icon :icon="mdiDeleteOutline" size="24" />
                                    <v-tooltip activator="parent">{{ tt('Delete') }}</v-tooltip>
                                </v-btn>
                            </div>
                        </td>
                    </tr>
                    <!-- Payments sub-table -->
                    <tr v-if="expandedDealId === deal.id" class="payments-row">
                        <td class="pa-0">
                            <div class="payments-container pa-4">
                                <div class="d-flex align-center mb-2">
                                    <span class="text-subtitle-2">{{ tt('Payments') }}</span>
                                    <v-btn class="ms-2" color="default" variant="outlined" size="small"
                                           :disabled="updating || hasEditingPayment" @click="addPayment(deal)">{{ tt('Add Payment') }}</v-btn>
                                    <v-spacer/>
                                    <span class="text-body-2 text-disabled">{{ tt('Total Paid') }}: {{ formatAmount(getTotalPaid(deal.id)) }}</span>
                                    <span class="text-body-2 ms-4" :class="getDebtColor(deal)">{{ tt('Remaining') }}: {{ formatAmount(getRemaining(deal)) }}</span>
                                </div>
                                <v-table density="compact" v-if="(dealPayments[deal.id] || []).length > 0">
                                    <thead>
                                    <tr>
                                        <th style="width: 100px">{{ tt('Date') }}</th>
                                        <th style="width: 100px">{{ tt('Amount') }}</th>
                                        <th style="width: 120px">{{ tt('Payment Type') }}</th>
                                        <th>{{ tt('Comment') }}</th>
                                        <th style="width: 80px">{{ tt('Operation') }}</th>
                                    </tr>
                                    </thead>
                                    <tbody>
                                    <tr v-for="payment in dealPayments[deal.id]" :key="payment.id">
                                        <td v-if="editingPayment && editingPayment.id === payment.id">
                                            <v-text-field density="compact" variant="underlined" type="date"
                                                          v-model="editingPaymentDateStr" style="max-width: 120px"/>
                                        </td>
                                        <td v-else>{{ formatPaymentDate(payment.paymentDate) }}</td>

                                        <td v-if="editingPayment && editingPayment.id === payment.id">
                                            <v-text-field density="compact" variant="underlined" type="number"
                                                          v-model.number="editingPayment.amount" style="max-width: 100px"/>
                                        </td>
                                        <td v-else>{{ formatAmount(payment.amount) }}</td>

                                        <td v-if="editingPayment && editingPayment.id === payment.id">
                                            <v-select density="compact" variant="underlined" :items="paymentTypeOptions"
                                                      item-title="text" item-value="value"
                                                      v-model="editingPayment.paymentType" style="max-width: 120px"/>
                                        </td>
                                        <td v-else>{{ getPaymentTypeText(payment.paymentType) }}</td>

                                        <td v-if="editingPayment && editingPayment.id === payment.id">
                                            <v-text-field density="compact" variant="underlined"
                                                          v-model="editingPayment.comment" style="max-width: 150px"/>
                                        </td>
                                        <td v-else>{{ payment.comment }}</td>

                                        <td>
                                            <div class="d-flex" v-if="editingPayment && editingPayment.id === payment.id">
                                                <v-btn density="compact" color="primary" variant="text" size="20"
                                                       :icon="true" :disabled="updating"
                                                       @click="savePaymentEdit(deal.id)">
                                                    <v-icon :icon="mdiCheck" size="20" />
                                                </v-btn>
                                                <v-btn density="compact" color="default" variant="text" size="20"
                                                       :icon="true" :disabled="updating"
                                                       @click="cancelEditPayment">
                                                    <v-icon :icon="mdiClose" size="20" />
                                                </v-btn>
                                            </div>
                                            <div class="d-flex" v-else>
                                                <v-btn density="compact" color="default" variant="text" size="20"
                                                       :icon="true" :disabled="updating || hasEditingPayment"
                                                       @click="editPayment(payment)">
                                                    <v-icon :icon="mdiSquareEditOutline" size="20" />
                                                </v-btn>
                                                <v-btn density="compact" color="default" variant="text" size="20"
                                                       :icon="true" :disabled="updating || hasEditingPayment"
                                                       @click="confirmDeletePayment(deal.id, payment)">
                                                    <v-icon :icon="mdiDeleteOutline" size="20" />
                                                </v-btn>
                                            </div>
                                        </td>
                                    </tr>
                                    <!-- New payment row -->
                                    <tr v-if="editingPayment && !editingPayment.id && editingPayment.dealId === deal.id">
                                        <td>
                                            <v-text-field density="compact" variant="underlined" type="date"
                                                          v-model="editingPaymentDateStr" style="max-width: 120px"/>
                                        </td>
                                        <td>
                                            <v-text-field density="compact" variant="underlined" type="number"
                                                          v-model.number="editingPayment.amount" style="max-width: 100px"/>
                                        </td>
                                        <td>
                                            <v-select density="compact" variant="underlined" :items="paymentTypeOptions"
                                                      item-title="text" item-value="value"
                                                      v-model="editingPayment.paymentType" style="max-width: 120px"/>
                                        </td>
                                        <td>
                                            <v-text-field density="compact" variant="underlined"
                                                          v-model="editingPayment.comment" style="max-width: 150px"/>
                                        </td>
                                        <td>
                                            <div class="d-flex">
                                                <v-btn density="compact" color="primary" variant="text" size="20"
                                                       :icon="true" :disabled="updating"
                                                       @click="savePaymentEdit(deal.id)">
                                                    <v-icon :icon="mdiCheck" size="20" />
                                                </v-btn>
                                                <v-btn density="compact" color="default" variant="text" size="20"
                                                       :icon="true" :disabled="updating"
                                                       @click="cancelEditPayment">
                                                    <v-icon :icon="mdiClose" size="20" />
                                                </v-btn>
                                            </div>
                                        </td>
                                    </tr>
                                    </tbody>
                                </v-table>
                                <div v-else-if="!loadingPayments" class="text-disabled text-body-2">{{ tt('No payments yet') }}</div>
                                <v-progress-linear v-if="loadingPayments" indeterminate color="primary"/>
                            </div>
                        </td>
                    </tr>
                    </template>
                    </tbody>

                    <!-- New deal row -->
                    <tbody v-if="editingDeal && !editingDeal.id && !loading">
                    <tr>
                        <td>
                            <div class="d-flex align-center">
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Investor Name')"
                                              class="me-2" style="max-width: 160px"
                                              v-model="editingDeal.investorName"/>
                                <v-select density="compact" variant="underlined" :items="dealTypeOptions"
                                          item-title="text" item-value="value"
                                          class="me-2" style="max-width: 140px"
                                          v-model="editingDeal.dealType"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Amount')"
                                              class="me-2" style="max-width: 120px" type="number"
                                              v-model.number="editingDeal.investmentAmount"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Currency')"
                                              class="me-2" style="max-width: 60px"
                                              v-model="editingDeal.currency"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Annual Rate %')"
                                              class="me-2" style="max-width: 80px" type="number"
                                              v-model.number="editingDeal.annualRate"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Total to Repay')"
                                              class="me-2" style="max-width: 120px" type="number"
                                              v-model.number="editingDeal.totalToRepay"/>
                                <v-select density="compact" variant="underlined" :items="cfoOptions"
                                          item-title="text" item-value="value"
                                          class="me-2" style="max-width: 140px"
                                          v-model="editingDeal.cfoId"/>
                                <v-text-field density="compact" variant="underlined" :placeholder="tt('Comment')"
                                              class="me-2" style="max-width: 120px"
                                              v-model="editingDeal.comment"/>
                                <v-spacer/>
                                <v-btn density="compact" color="primary" variant="text" size="24"
                                       class="ms-2" :icon="true" :disabled="updating"
                                       @click="saveDeal(null)">
                                    <v-icon :icon="mdiCheck" size="24" />
                                    <v-tooltip activator="parent">{{ tt('Save') }}</v-tooltip>
                                </v-btn>
                                <v-btn density="compact" color="default" variant="text" size="24"
                                       class="ms-2" :icon="true" :disabled="updating"
                                       @click="cancelEditDeal">
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
    mdiDeleteOutline,
    mdiChevronUp,
    mdiChevronDown
} from '@mdi/js';

import { useI18n } from '@/locales/helpers.ts';
import { useInvestorDealsStore } from '@/stores/investorDeal.ts';
import { useCFOsStore } from '@/stores/cfo.ts';

import { InvestorDeal } from '@/models/investor_deal.ts';
import { InvestorPayment } from '@/models/investor_payment.ts';

import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';

type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;
type SnackBarType = InstanceType<typeof SnackBar>;

const { tt } = useI18n();
const investorDealsStore = useInvestorDealsStore();
const cfosStore = useCFOsStore();

const loading = ref<boolean>(true);
const updating = ref<boolean>(false);
const loadingPayments = ref<boolean>(false);

const editingDeal = ref<InvestorDeal | null>(null);
const editingPayment = ref<InvestorPayment | null>(null);
const editingPaymentDateStr = ref<string>('');
const expandedDealId = ref<string>('');
const dealPayments = ref<Record<string, InvestorPayment[]>>({});

const confirmDialog = ref<ConfirmDialogType | null>(null);
const snackBar = ref<SnackBarType | null>(null);

const deals = computed(() => investorDealsStore.allDeals);
const hasEditingDeal = computed(() => editingDeal.value !== null);
const hasEditingPayment = computed(() => editingPayment.value !== null);

const dealTypeOptions = computed(() => [
    { text: tt('Loan'), value: 1 },
    { text: tt('Equity'), value: 2 },
    { text: tt('Revenue Share'), value: 3 },
    { text: tt('Other'), value: 4 }
]);

const paymentTypeOptions = computed(() => [
    { text: tt('Principal'), value: 1 },
    { text: tt('Interest'), value: 2 },
    { text: tt('Mixed'), value: 3 }
]);

const cfoOptions = computed(() => {
    const options = [{ text: tt('No CFO'), value: '0' }];

    for (const cfo of cfosStore.allCFOs) {
        options.push({ text: cfo.name, value: cfo.id });
    }

    return options;
});

function getDealTypeText(type: number): string {
    const option = dealTypeOptions.value.find(o => o.value === type);
    return option ? option.text : '';
}

function getPaymentTypeText(type: number): string {
    const option = paymentTypeOptions.value.find(o => o.value === type);
    return option ? option.text : '';
}

function getCfoName(cfoId: string): string {
    if (!cfoId || cfoId === '0') return '';
    const cfo = cfosStore.allCFOsMap[cfoId];
    return cfo ? cfo.name : '';
}

function formatAmount(amount: number): string {
    return (amount / 100).toLocaleString('ru-RU', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function formatPaymentDate(timestamp: number): string {
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

function getTotalPaid(dealId: string): number {
    const payments = dealPayments.value[dealId];
    if (!payments) return 0;
    return payments.reduce((sum, p) => sum + p.amount, 0);
}

function getRemaining(deal: InvestorDeal): number {
    if (!deal.totalToRepay) return 0;
    return deal.totalToRepay - getTotalPaid(deal.id);
}

function getDebtColor(deal: InvestorDeal): string {
    const remaining = getRemaining(deal);
    if (remaining <= 0) return 'text-success';
    return 'text-warning';
}

function addDeal(): void {
    editingDeal.value = InvestorDeal.createNew();
}

function editDeal(deal: InvestorDeal): void {
    editingDeal.value = deal.clone();
}

function cancelEditDeal(): void {
    editingDeal.value = null;
}

async function saveDeal(originalDeal: InvestorDeal | null): Promise<void> {
    if (!editingDeal.value) return;

    if (!editingDeal.value.investorName) {
        snackBar.value?.showMessage('Investor name cannot be empty');
        return;
    }

    updating.value = true;

    try {
        await investorDealsStore.saveDeal({ deal: editingDeal.value });
        editingDeal.value = null;
        snackBar.value?.showMessage('Investor deal saved successfully');
    } catch (error: any) {
        if (!error.processed) {
            snackBar.value?.showError(error);
        }
    } finally {
        updating.value = false;
    }
}

function confirmDeleteDeal(deal: InvestorDeal): void {
    confirmDialog.value?.open('Are you sure you want to delete this investor deal?').then(() => {
        deleteDeal(deal);
    });
}

async function deleteDeal(deal: InvestorDeal): Promise<void> {
    updating.value = true;

    try {
        await investorDealsStore.deleteDeal({ deal });
        if (expandedDealId.value === deal.id) {
            expandedDealId.value = '';
        }
        snackBar.value?.showMessage('Investor deal deleted successfully');
    } catch (error: any) {
        if (!error.processed) {
            snackBar.value?.showError(error);
        }
    } finally {
        updating.value = false;
    }
}

async function togglePayments(deal: InvestorDeal): Promise<void> {
    if (expandedDealId.value === deal.id) {
        expandedDealId.value = '';
        return;
    }

    expandedDealId.value = deal.id;

    if (!dealPayments.value[deal.id]) {
        loadingPayments.value = true;

        try {
            const payments = await investorDealsStore.loadPaymentsByDealId(deal.id);
            dealPayments.value[deal.id] = payments;
        } catch (error: any) {
            // ignore
        } finally {
            loadingPayments.value = false;
        }
    }
}

function addPayment(deal: InvestorDeal): void {
    editingPayment.value = InvestorPayment.createNew(deal.id);
    editingPaymentDateStr.value = '';
}

function editPayment(payment: InvestorPayment): void {
    editingPayment.value = payment.clone();
    editingPaymentDateStr.value = unixToDateStr(payment.paymentDate);
}

function cancelEditPayment(): void {
    editingPayment.value = null;
    editingPaymentDateStr.value = '';
}

async function savePaymentEdit(dealId: string): Promise<void> {
    if (!editingPayment.value) return;

    editingPayment.value.paymentDate = dateStrToUnix(editingPaymentDateStr.value);

    updating.value = true;

    try {
        await investorDealsStore.savePayment({ payment: editingPayment.value });
        // Reload payments
        const payments = await investorDealsStore.loadPaymentsByDealId(dealId);
        dealPayments.value[dealId] = payments;
        editingPayment.value = null;
        editingPaymentDateStr.value = '';
        snackBar.value?.showMessage('Payment saved successfully');
    } catch (error: any) {
        if (!error.processed) {
            snackBar.value?.showError(error);
        }
    } finally {
        updating.value = false;
    }
}

function confirmDeletePayment(dealId: string, payment: InvestorPayment): void {
    confirmDialog.value?.open('Are you sure you want to delete this payment?').then(() => {
        deletePayment(dealId, payment);
    });
}

async function deletePayment(dealId: string, payment: InvestorPayment): Promise<void> {
    updating.value = true;

    try {
        await investorDealsStore.deletePayment({ paymentId: payment.id });
        // Reload payments
        const payments = await investorDealsStore.loadPaymentsByDealId(dealId);
        dealPayments.value[dealId] = payments;
        snackBar.value?.showMessage('Payment deleted successfully');
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
        await investorDealsStore.loadAllDeals({ force: true });
    } catch (error: any) {
        if (!error.isUpToDate) {
            if (!error.processed) {
                snackBar.value?.showError(error);
            }
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
        await investorDealsStore.loadAllDeals({ force: false });
    } catch (error: any) {
        if (!error.processed) {
            snackBar.value?.showError(error);
        }
    } finally {
        loading.value = false;
    }
});
</script>

<style scoped>
.payments-container {
    background-color: rgba(var(--v-theme-surface-variant), 0.05);
    border-top: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}

.expanded-row {
    background-color: rgba(var(--v-theme-surface-variant), 0.03);
}
</style>
