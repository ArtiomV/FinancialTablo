import { ref } from 'vue';
import { defineStore } from 'pinia';

import { type BeforeResolveFunction, itemAndIndex } from '@/core/base.ts';

import {
    type InvestorDealInfoResponse,
    InvestorDeal
} from '@/models/investor_deal.ts';

import {
    type InvestorPaymentInfoResponse,
    InvestorPayment
} from '@/models/investor_payment.ts';

import { isEquals } from '@/lib/common.ts';

import logger from '@/lib/logger.ts';
import services, { type ApiResponsePromise } from '@/lib/services.ts';

export const useInvestorDealsStore = defineStore('investorDeals', () => {
    const allDeals = ref<InvestorDeal[]>([]);
    const allDealsMap = ref<Record<string, InvestorDeal>>({});
    const dealListStateInvalid = ref<boolean>(true);

    function loadDealList(deals: InvestorDeal[]): void {
        allDeals.value = deals;
        allDealsMap.value = {};

        for (const deal of deals) {
            allDealsMap.value[deal.id] = deal;
        }
    }

    function addDealToList(deal: InvestorDeal): void {
        allDeals.value.push(deal);
        allDealsMap.value[deal.id] = deal;
    }

    function updateDealInList(currentDeal: InvestorDeal): void {
        for (const [deal, index] of itemAndIndex(allDeals.value)) {
            if (deal.id === currentDeal.id) {
                allDeals.value.splice(index, 1, currentDeal);
                break;
            }
        }

        allDealsMap.value[currentDeal.id] = currentDeal;
    }

    function removeDealFromList(currentDeal: InvestorDeal): void {
        for (const [deal, index] of itemAndIndex(allDeals.value)) {
            if (deal.id === currentDeal.id) {
                allDeals.value.splice(index, 1);
                break;
            }
        }

        if (allDealsMap.value[currentDeal.id]) {
            delete allDealsMap.value[currentDeal.id];
        }
    }

    function updateDealListInvalidState(invalidState: boolean): void {
        dealListStateInvalid.value = invalidState;
    }

    function resetDeals(): void {
        allDeals.value = [];
        allDealsMap.value = {};
        dealListStateInvalid.value = true;
    }

    function loadAllDeals({ force }: { force?: boolean }): Promise<InvestorDeal[]> {
        if (!force && !dealListStateInvalid.value) {
            return new Promise((resolve) => {
                resolve(allDeals.value);
            });
        }

        return new Promise((resolve, reject) => {
            services.getAllInvestorDeals().then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve investor deal list' });
                    return;
                }

                if (dealListStateInvalid.value) {
                    updateDealListInvalidState(false);
                }

                const deals = InvestorDeal.ofMulti(data.result);

                if (force && data.result && isEquals(allDeals.value, deals)) {
                    reject({ message: 'Investor deal list is up to date', isUpToDate: true });
                    return;
                }

                loadDealList(deals);

                resolve(deals);
            }).catch(error => {
                if (force) {
                    logger.error('failed to force load investor deal list', error);
                } else {
                    logger.error('failed to load investor deal list', error);
                }

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve investor deal list' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function saveDeal({ deal, beforeResolve }: { deal: InvestorDeal, beforeResolve?: BeforeResolveFunction }): Promise<InvestorDeal> {
        return new Promise((resolve, reject) => {
            let promise: ApiResponsePromise<InvestorDealInfoResponse>;

            if (!deal.id) {
                promise = services.addInvestorDeal(deal.toCreateRequest());
            } else {
                promise = services.modifyInvestorDeal(deal.toModifyRequest());
            }

            promise.then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (!deal.id) {
                        reject({ message: 'Unable to add investor deal' });
                    } else {
                        reject({ message: 'Unable to save investor deal' });
                    }
                    return;
                }

                const newDeal = InvestorDeal.of(data.result);

                if (beforeResolve) {
                    beforeResolve(() => {
                        if (!deal.id) {
                            addDealToList(newDeal);
                        } else {
                            updateDealInList(newDeal);
                        }
                    });
                } else {
                    if (!deal.id) {
                        addDealToList(newDeal);
                    } else {
                        updateDealInList(newDeal);
                    }
                }

                resolve(newDeal);
            }).catch(error => {
                logger.error('failed to save investor deal', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (!deal.id) {
                        reject({ message: 'Unable to add investor deal' });
                    } else {
                        reject({ message: 'Unable to save investor deal' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function deleteDeal({ deal, beforeResolve }: { deal: InvestorDeal, beforeResolve?: BeforeResolveFunction }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.deleteInvestorDeal({
                id: deal.id
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to delete this investor deal' });
                    return;
                }

                if (beforeResolve) {
                    beforeResolve(() => {
                        removeDealFromList(deal);
                    });
                } else {
                    removeDealFromList(deal);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to delete investor deal', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete this investor deal' });
                } else {
                    reject(error);
                }
            });
        });
    }

    // Payments
    function loadPaymentsByDealId(dealId: string): Promise<InvestorPayment[]> {
        return new Promise((resolve, reject) => {
            services.getInvestorPaymentsByDeal({ dealId }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve investor payment list' });
                    return;
                }

                resolve(InvestorPayment.ofMulti(data.result));
            }).catch(error => {
                logger.error('failed to load investor payments', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve investor payment list' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function savePayment({ payment }: { payment: InvestorPayment }): Promise<InvestorPayment> {
        return new Promise((resolve, reject) => {
            let promise: ApiResponsePromise<InvestorPaymentInfoResponse>;

            if (!payment.id) {
                promise = services.addInvestorPayment(payment.toCreateRequest());
            } else {
                promise = services.modifyInvestorPayment(payment.toModifyRequest());
            }

            promise.then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (!payment.id) {
                        reject({ message: 'Unable to add investor payment' });
                    } else {
                        reject({ message: 'Unable to save investor payment' });
                    }
                    return;
                }

                resolve(InvestorPayment.of(data.result));
            }).catch(error => {
                logger.error('failed to save investor payment', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (!payment.id) {
                        reject({ message: 'Unable to add investor payment' });
                    } else {
                        reject({ message: 'Unable to save investor payment' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function deletePayment({ paymentId }: { paymentId: string }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.deleteInvestorPayment({
                id: paymentId
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to delete this investor payment' });
                    return;
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to delete investor payment', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete this investor payment' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return {
        allDeals,
        allDealsMap,
        dealListStateInvalid,
        updateDealListInvalidState,
        resetDeals,
        loadAllDeals,
        saveDeal,
        deleteDeal,
        loadPaymentsByDealId,
        savePayment,
        deletePayment
    }
});
