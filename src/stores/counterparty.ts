import { ref, computed } from 'vue';
import { defineStore } from 'pinia';

import { type BeforeResolveFunction, itemAndIndex } from '@/core/base.ts';

import {
    type CounterpartyInfoResponse,
    type CounterpartyNewDisplayOrderRequest,
    Counterparty
} from '@/models/counterparty.ts';

import { isEquals } from '@/lib/common.ts';

import logger from '@/lib/logger.ts';
import services, { type ApiResponsePromise } from '@/lib/services.ts';

export const useCounterpartiesStore = defineStore('counterparties', () => {
    const allCounterparties = ref<Counterparty[]>([]);
    const allCounterpartiesMap = ref<Record<string, Counterparty>>({});
    const counterpartyListStateInvalid = ref<boolean>(true);

    const allVisibleCounterparties = computed<Counterparty[]>(() => {
        const visibleCounterparties: Counterparty[] = [];

        for (const counterparty of allCounterparties.value) {
            if (!counterparty.hidden) {
                visibleCounterparties.push(counterparty);
            }
        }

        return visibleCounterparties;
    });

    const allAvailableCounterpartiesCount = computed<number>(() => allCounterparties.value.length);

    function loadCounterpartyList(counterparties: Counterparty[]): void {
        allCounterparties.value = counterparties;
        allCounterpartiesMap.value = {};

        for (const counterparty of counterparties) {
            allCounterpartiesMap.value[counterparty.id] = counterparty;
        }
    }

    function addCounterpartyToList(counterparty: Counterparty): void {
        allCounterparties.value.push(counterparty);
        allCounterpartiesMap.value[counterparty.id] = counterparty;
    }

    function updateCounterpartyInList(currentCounterparty: Counterparty): void {
        for (const [counterparty, index] of itemAndIndex(allCounterparties.value)) {
            if (counterparty.id === currentCounterparty.id) {
                allCounterparties.value.splice(index, 1, currentCounterparty);
                break;
            }
        }

        allCounterpartiesMap.value[currentCounterparty.id] = currentCounterparty;
    }

    function updateCounterpartyDisplayOrderInList({ from, to }: { from: number, to: number }): void {
        allCounterparties.value.splice(to, 0, allCounterparties.value.splice(from, 1)[0] as Counterparty);
    }

    function updateCounterpartyVisibilityInList({ counterparty, hidden }: { counterparty: Counterparty, hidden: boolean }): void {
        if (allCounterpartiesMap.value[counterparty.id]) {
            allCounterpartiesMap.value[counterparty.id]!.hidden = hidden;
        }
    }

    function removeCounterpartyFromList(currentCounterparty: Counterparty): void {
        for (const [counterparty, index] of itemAndIndex(allCounterparties.value)) {
            if (counterparty.id === currentCounterparty.id) {
                allCounterparties.value.splice(index, 1);
                break;
            }
        }

        if (allCounterpartiesMap.value[currentCounterparty.id]) {
            delete allCounterpartiesMap.value[currentCounterparty.id];
        }
    }

    function updateCounterpartyListInvalidState(invalidState: boolean): void {
        counterpartyListStateInvalid.value = invalidState;
    }

    function resetCounterparties(): void {
        allCounterparties.value = [];
        allCounterpartiesMap.value = {};
        counterpartyListStateInvalid.value = true;
    }

    function loadAllCounterparties({ force }: { force?: boolean }): Promise<Counterparty[]> {
        if (!force && !counterpartyListStateInvalid.value) {
            return new Promise((resolve) => {
                resolve(allCounterparties.value);
            });
        }

        return new Promise((resolve, reject) => {
            services.getAllCounterparties().then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve counterparty list' });
                    return;
                }

                if (counterpartyListStateInvalid.value) {
                    updateCounterpartyListInvalidState(false);
                }

                const counterparties = Counterparty.ofMulti(data.result);

                if (force && data.result && isEquals(allCounterparties.value, counterparties)) {
                    reject({ message: 'Counterparty list is up to date', isUpToDate: true });
                    return;
                }

                loadCounterpartyList(counterparties);

                resolve(counterparties);
            }).catch(error => {
                if (force) {
                    logger.error('failed to force load counterparty list', error);
                } else {
                    logger.error('failed to load counterparty list', error);
                }

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve counterparty list' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function saveCounterparty({ counterparty, beforeResolve }: { counterparty: Counterparty, beforeResolve?: BeforeResolveFunction }): Promise<Counterparty> {
        return new Promise((resolve, reject) => {
            let promise: ApiResponsePromise<CounterpartyInfoResponse>;

            if (!counterparty.id) {
                promise = services.addCounterparty(counterparty.toCreateRequest());
            } else {
                promise = services.modifyCounterparty(counterparty.toModifyRequest());
            }

            promise.then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (!counterparty.id) {
                        reject({ message: 'Unable to add counterparty' });
                    } else {
                        reject({ message: 'Unable to save counterparty' });
                    }
                    return;
                }

                const newCounterparty = Counterparty.of(data.result);

                if (beforeResolve) {
                    beforeResolve(() => {
                        if (!counterparty.id) {
                            addCounterpartyToList(newCounterparty);
                        } else {
                            updateCounterpartyInList(newCounterparty);
                        }
                    });
                } else {
                    if (!counterparty.id) {
                        addCounterpartyToList(newCounterparty);
                    } else {
                        updateCounterpartyInList(newCounterparty);
                    }
                }

                resolve(newCounterparty);
            }).catch(error => {
                logger.error('failed to save counterparty', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (!counterparty.id) {
                        reject({ message: 'Unable to add counterparty' });
                    } else {
                        reject({ message: 'Unable to save counterparty' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function hideCounterparty({ counterparty, hidden }: { counterparty: Counterparty, hidden: boolean }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.hideCounterparty({
                id: counterparty.id,
                hidden: hidden
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (hidden) {
                        reject({ message: 'Unable to hide this counterparty' });
                    } else {
                        reject({ message: 'Unable to unhide this counterparty' });
                    }
                    return;
                }

                updateCounterpartyVisibilityInList({ counterparty, hidden });

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to change counterparty visibility', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (hidden) {
                        reject({ message: 'Unable to hide this counterparty' });
                    } else {
                        reject({ message: 'Unable to unhide this counterparty' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function changeCounterpartyDisplayOrder({ counterpartyId, from, to }: { counterpartyId: string, from: number, to: number }): Promise<void> {
        return new Promise((resolve, reject) => {
            const currentCounterparty = allCounterpartiesMap.value[counterpartyId];

            if (!currentCounterparty || !allCounterparties.value[to]) {
                reject({ message: 'Unable to move counterparty' });
                return;
            }

            if (!counterpartyListStateInvalid.value) {
                updateCounterpartyListInvalidState(true);
            }

            updateCounterpartyDisplayOrderInList({ from, to });

            resolve();
        });
    }

    function updateCounterpartyDisplayOrders(): Promise<boolean> {
        const newDisplayOrders: CounterpartyNewDisplayOrderRequest[] = [];

        for (const [counterparty, index] of itemAndIndex(allCounterparties.value)) {
            newDisplayOrders.push({
                id: counterparty.id,
                displayOrder: index + 1
            });
        }

        return new Promise((resolve, reject) => {
            services.moveCounterparty({
                newDisplayOrders: newDisplayOrders
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to move counterparty' });
                    return;
                }

                if (counterpartyListStateInvalid.value) {
                    updateCounterpartyListInvalidState(false);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to save counterparties display order', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to move counterparty' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function deleteCounterparty({ counterparty, beforeResolve }: { counterparty: Counterparty, beforeResolve?: BeforeResolveFunction }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.deleteCounterparty({
                id: counterparty.id
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to delete this counterparty' });
                    return;
                }

                if (beforeResolve) {
                    beforeResolve(() => {
                        removeCounterpartyFromList(counterparty);
                    });
                } else {
                    removeCounterpartyFromList(counterparty);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to delete counterparty', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete this counterparty' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return {
        // states
        allCounterparties,
        allCounterpartiesMap,
        counterpartyListStateInvalid,
        // computed states
        allVisibleCounterparties,
        allAvailableCounterpartiesCount,
        // functions
        updateCounterpartyListInvalidState,
        resetCounterparties,
        loadAllCounterparties,
        saveCounterparty,
        hideCounterparty,
        changeCounterpartyDisplayOrder,
        updateCounterpartyDisplayOrders,
        deleteCounterparty
    }
});
