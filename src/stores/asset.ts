import { ref, computed } from 'vue';
import { defineStore } from 'pinia';

import { type BeforeResolveFunction, itemAndIndex } from '@/core/base.ts';

import {
    type AssetInfoResponse,
    type AssetNewDisplayOrderRequest,
    Asset
} from '@/models/asset.ts';

import { isEquals } from '@/lib/common.ts';

import logger from '@/lib/logger.ts';
import services, { type ApiResponsePromise } from '@/lib/services.ts';

export const useAssetsStore = defineStore('assets', () => {
    const allAssets = ref<Asset[]>([]);
    const allAssetsMap = ref<Record<string, Asset>>({});
    const assetListStateInvalid = ref<boolean>(true);

    const allVisibleAssets = computed<Asset[]>(() => {
        const visibleAssets: Asset[] = [];

        for (const asset of allAssets.value) {
            if (!asset.hidden) {
                visibleAssets.push(asset);
            }
        }

        return visibleAssets;
    });

    const allAvailableAssetsCount = computed<number>(() => allAssets.value.length);

    function loadAssetList(assets: Asset[]): void {
        allAssets.value = assets;
        allAssetsMap.value = {};

        for (const asset of assets) {
            allAssetsMap.value[asset.id] = asset;
        }
    }

    function addAssetToList(asset: Asset): void {
        allAssets.value.push(asset);
        allAssetsMap.value[asset.id] = asset;
    }

    function updateAssetInList(currentAsset: Asset): void {
        for (const [asset, index] of itemAndIndex(allAssets.value)) {
            if (asset.id === currentAsset.id) {
                allAssets.value.splice(index, 1, currentAsset);
                break;
            }
        }

        allAssetsMap.value[currentAsset.id] = currentAsset;
    }

    function updateAssetDisplayOrderInList({ from, to }: { from: number, to: number }): void {
        allAssets.value.splice(to, 0, allAssets.value.splice(from, 1)[0] as Asset);
    }

    function updateAssetVisibilityInList({ asset, hidden }: { asset: Asset, hidden: boolean }): void {
        if (allAssetsMap.value[asset.id]) {
            allAssetsMap.value[asset.id]!.hidden = hidden;
        }
    }

    function removeAssetFromList(currentAsset: Asset): void {
        for (const [asset, index] of itemAndIndex(allAssets.value)) {
            if (asset.id === currentAsset.id) {
                allAssets.value.splice(index, 1);
                break;
            }
        }

        if (allAssetsMap.value[currentAsset.id]) {
            delete allAssetsMap.value[currentAsset.id];
        }
    }

    function updateAssetListInvalidState(invalidState: boolean): void {
        assetListStateInvalid.value = invalidState;
    }

    function resetAssets(): void {
        allAssets.value = [];
        allAssetsMap.value = {};
        assetListStateInvalid.value = true;
    }

    function loadAllAssets({ force }: { force?: boolean }): Promise<Asset[]> {
        if (!force && !assetListStateInvalid.value) {
            return new Promise((resolve) => {
                resolve(allAssets.value);
            });
        }

        return new Promise((resolve, reject) => {
            services.getAllAssets().then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to retrieve asset list' });
                    return;
                }

                if (assetListStateInvalid.value) {
                    updateAssetListInvalidState(false);
                }

                const assets = Asset.ofMulti(data.result);

                if (force && data.result && isEquals(allAssets.value, assets)) {
                    reject({ message: 'Asset list is up to date', isUpToDate: true });
                    return;
                }

                loadAssetList(assets);

                resolve(assets);
            }).catch(error => {
                if (force) {
                    logger.error('failed to force load asset list', error);
                } else {
                    logger.error('failed to load asset list', error);
                }

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to retrieve asset list' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function saveAsset({ asset, beforeResolve }: { asset: Asset, beforeResolve?: BeforeResolveFunction }): Promise<Asset> {
        return new Promise((resolve, reject) => {
            let promise: ApiResponsePromise<AssetInfoResponse>;

            if (!asset.id) {
                promise = services.addAsset(asset.toCreateRequest());
            } else {
                promise = services.modifyAsset(asset.toModifyRequest());
            }

            promise.then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (!asset.id) {
                        reject({ message: 'Unable to add asset' });
                    } else {
                        reject({ message: 'Unable to save asset' });
                    }
                    return;
                }

                const newAsset = Asset.of(data.result);

                if (beforeResolve) {
                    beforeResolve(() => {
                        if (!asset.id) {
                            addAssetToList(newAsset);
                        } else {
                            updateAssetInList(newAsset);
                        }
                    });
                } else {
                    if (!asset.id) {
                        addAssetToList(newAsset);
                    } else {
                        updateAssetInList(newAsset);
                    }
                }

                resolve(newAsset);
            }).catch(error => {
                logger.error('failed to save asset', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (!asset.id) {
                        reject({ message: 'Unable to add asset' });
                    } else {
                        reject({ message: 'Unable to save asset' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function hideAsset({ asset, hidden }: { asset: Asset, hidden: boolean }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.hideAsset({
                id: asset.id,
                hidden: hidden
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    if (hidden) {
                        reject({ message: 'Unable to hide this asset' });
                    } else {
                        reject({ message: 'Unable to unhide this asset' });
                    }
                    return;
                }

                updateAssetVisibilityInList({ asset, hidden });

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to change asset visibility', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    if (hidden) {
                        reject({ message: 'Unable to hide this asset' });
                    } else {
                        reject({ message: 'Unable to unhide this asset' });
                    }
                } else {
                    reject(error);
                }
            });
        });
    }

    function changeAssetDisplayOrder({ assetId, from, to }: { assetId: string, from: number, to: number }): Promise<void> {
        return new Promise((resolve, reject) => {
            const currentAsset = allAssetsMap.value[assetId];

            if (!currentAsset || !allAssets.value[to]) {
                reject({ message: 'Unable to move asset' });
                return;
            }

            if (!assetListStateInvalid.value) {
                updateAssetListInvalidState(true);
            }

            updateAssetDisplayOrderInList({ from, to });

            resolve();
        });
    }

    function updateAssetDisplayOrders(): Promise<boolean> {
        const newDisplayOrders: AssetNewDisplayOrderRequest[] = [];

        for (const [asset, index] of itemAndIndex(allAssets.value)) {
            newDisplayOrders.push({
                id: asset.id,
                displayOrder: index + 1
            });
        }

        return new Promise((resolve, reject) => {
            services.moveAsset({
                newDisplayOrders: newDisplayOrders
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to move asset' });
                    return;
                }

                if (assetListStateInvalid.value) {
                    updateAssetListInvalidState(false);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to save assets display order', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to move asset' });
                } else {
                    reject(error);
                }
            });
        });
    }

    function deleteAsset({ asset, beforeResolve }: { asset: Asset, beforeResolve?: BeforeResolveFunction }): Promise<boolean> {
        return new Promise((resolve, reject) => {
            services.deleteAsset({
                id: asset.id
            }).then(response => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    reject({ message: 'Unable to delete this asset' });
                    return;
                }

                if (beforeResolve) {
                    beforeResolve(() => {
                        removeAssetFromList(asset);
                    });
                } else {
                    removeAssetFromList(asset);
                }

                resolve(data.result);
            }).catch(error => {
                logger.error('failed to delete asset', error);

                if (error.response && error.response.data && error.response.data.errorMessage) {
                    reject({ error: error.response.data });
                } else if (!error.processed) {
                    reject({ message: 'Unable to delete this asset' });
                } else {
                    reject(error);
                }
            });
        });
    }

    return {
        allAssets,
        allAssetsMap,
        assetListStateInvalid,
        allVisibleAssets,
        allAvailableAssetsCount,
        updateAssetListInvalidState,
        resetAssets,
        loadAllAssets,
        saveAsset,
        hideAsset,
        changeAssetDisplayOrder,
        updateAssetDisplayOrders,
        deleteAsset
    }
});
