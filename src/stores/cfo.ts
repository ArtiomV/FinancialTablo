import { ref, computed } from 'vue';
import { defineStore } from 'pinia';

import { type BeforeResolveFunction, itemAndIndex } from '@/core/base.ts';

import {
    type CFOInfoResponse,
    type CFONewDisplayOrderRequest,
    CFO
} from '@/models/cfo.ts';

import { isEquals } from '@/lib/common.ts';

import logger from '@/lib/logger.ts';
import services, { type ApiResponsePromise } from '@/lib/services.ts';

export const useCFOsStore = defineStore('cfos', () => {
    const allCFOs = ref<CFO[]>([]);
    const allCFOsMap = ref<Record<string, CFO>>({});
    const cfoListStateInvalid = ref<boolean>(true);

    const allVisibleCFOs = computed<CFO[]>(() => {
        const visibleCFOs: CFO[] = [];

        for (const cfo of allCFOs.value) {
            if (!cfo.hidden) {
                visibleCFOs.push(cfo);
            }
        }

        return visibleCFOs;
    });

    const allAvailableCFOsCount = computed<number>(() => allCFOs.value.length);

    function loadCFOList(cfos: CFO[]): void {
        allCFOs.value = cfos;
        allCFOsMap.value = {};

        for (const cfo of cfos) {
            allCFOsMap.value[cfo.id] = cfo;
        }
    }

    function addCFOToList(cfo: CFO): void {
        allCFOs.value.push(cfo);
        allCFOsMap.value[cfo.id] = cfo;
    }

    function updateCFOInList(currentCFO: CFO): void {
        for (const [cfo, index] of itemAndIndex(allCFOs.value)) {
            if (cfo.id === currentCFO.id) {
                allCFOs.value.splice(index, 1, currentCFO);
                break;
            }
        }

        allCFOsMap.value[currentCFO.id] = currentCFO;
    }

    function updateCFODisplayOrderInList({ from, to }: { from: number, to: number }): void {
        allCFOs.value.splice(to, 0, allCFOs.value.splice(from, 1)[0] as CFO);
    }

    function updateCFOVisibilityInList({ cfo, hidden }: { cfo: CFO, hidden: boolean }): void {
        if (allCFOsMap.value[cfo.id]) {
            allCFOsMap.value[cfo.id]!.hidden = hidden;
        }
    }

    function removeCFOFromList(currentCFO: CFO): void {
        for (const [cfo, index] of itemAndIndex(allCFOs.value)) {
            if (cfo.id === currentCFO.id) {
                allCFOs.value.splice(index, 1);
                break;
            }
        }

        if (allCFOsMap.value[currentCFO.id]) {
            delete allCFOsMap.value[currentCFO.id];
        }
    }

    function updateCFOListInvalidState(invalidState: boolean): void {
        cfoListStateInvalid.value = invalidState;
    }

    function resetCFOs(): void {
        allCFOs.value = [];
        allCFOsMap.value = {};
        cfoListStateInvalid.value = true;
    }

    function loadAllCFOs({ force }: { force?: boolean }): Promise<CFO[]> {
        if (!force && !cfoListStateInvalid.value) {
            return new Promise((resolve) => {
                resolve(allCFOs.value);
            });
        }

        return new Promise((resolve, reject) => {
            services.getAllCFOs().then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve CFO list' });
                    return;
                }

                if (cfoListStateInvalid.value) {
                    updateCFOListInvalidState(false);
                }

                const cfos = CFO.ofMulti(data.result);

                if (force && data.result && isEquals(allCFOs.value, cfos)) {
                    reject({ message: 'CFO list is up to date', isUpToDate: true });
                    return;
                }

                loadCFOList(cfos);

                resolve(cfos);
            }).catch(error => {
                if (force) {
                    logger.error('failed to force load CFO list', error);
                } else {
                    logger.error('failed to load CFO list', error);
                }

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve CFO list' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function saveCFO({ cfo, beforeResolve }: { cfo: CFO, beforeResolve?: BeforeResolveFunction }): Promise<CFO> {
        return new Promise((resolve, reject) => {
            let promise: ApiResponsePromise<CFOInfoResponse>;

            if (!cfo.id) {
                promise = services.addCFO(cfo.toCreateRequest());
            } else {
                promise = services.modifyCFO(cfo.toModifyRequest());
            }

            promise.then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (!cfo.id) {
                        reject({ message: 'Unable to add CFO' });
                    } else {
                        reject({ message: 'Unable to save CFO' });
                    }
                    return;
                }

                const newCFO = CFO.of(data.result);

                if (beforeResolve) {
                    beforeResolve(() => {
                        if (!cfo.id) {
                            addCFOToList(newCFO);
                        } else {
                            updateCFOInList(newCFO);
                        }
                    });
                } else {
                    if (!cfo.id) {
                        addCFOToList(newCFO);
                    } else {
                        updateCFOInList(newCFO);
                    }
                }

                resolve(newCFO);
            }).catch(error => {
                logger.error('failed to save CFO', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (!cfo.id) {
                        reject({ message: 'Unable to add CFO' });
                    } else {
                        reject({ message: 'Unable to save CFO' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function hideCFO({ cfo, hidden }: { cfo: CFO, hidden: boolean }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.hideCFO({
                id: cfo.id,
                hidden: hidden
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (hidden) {
                        reject({ message: 'Unable to hide this CFO' });
                    } else {
                        reject({ message: 'Unable to unhide this CFO' });
                    }
                    return;
                }

                updateCFOVisibilityInList({ cfo, hidden });

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to change CFO visibility', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (hidden) {
                        reject({ message: 'Unable to hide this CFO' });
                    } else {
                        reject({ message: 'Unable to unhide this CFO' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function changeCFODisplayOrder({ cfoId, from, to }: { cfoId: string, from: number, to: number }): Promise<void> {
        return new Promise((resolve, reject) => {
            const currentCFO = allCFOsMap.value[cfoId];

            if (!currentCFO || !allCFOs.value[to]) {
                reject({ message: 'Unable to move CFO' });
                return;
            }

            if (!cfoListStateInvalid.value) {
                updateCFOListInvalidState(true);
            }

            updateCFODisplayOrderInList({ from, to });

            resolve();
        });
    }

    function updateCFODisplayOrders(): Promise<boolean> {
        const newDisplayOrders: CFONewDisplayOrderRequest[] = [];

        for (const [cfo, index] of itemAndIndex(allCFOs.value)) {
            newDisplayOrders.push({
                id: cfo.id,
                displayOrder: index + 1
            });
        }

        return new Promise((resolve, reject) => {
            services.moveCFO({
                newDisplayOrders: newDisplayOrders
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to move CFO' });
                    return;
                }

                if (cfoListStateInvalid.value) {
                    updateCFOListInvalidState(false);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to save CFOs display order', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to move CFO' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function deleteCFO({ cfo, beforeResolve }: { cfo: CFO, beforeResolve?: BeforeResolveFunction }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.deleteCFO({
                id: cfo.id
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to delete this CFO' });
                    return;
                }

                if (beforeResolve) {
                    beforeResolve(() => {
                        removeCFOFromList(cfo);
                    });
                } else {
                    removeCFOFromList(cfo);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to delete CFO', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete this CFO' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return {
        allCFOs,
        allCFOsMap,
        cfoListStateInvalid,
        allVisibleCFOs,
        allAvailableCFOsCount,
        updateCFOListInvalidState,
        resetCFOs,
        loadAllCFOs,
        saveCFO,
        hideCFO,
        changeCFODisplayOrder,
        updateCFODisplayOrders,
        deleteCFO
    }
});
