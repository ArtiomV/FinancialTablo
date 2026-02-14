<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="title-and-toolbar d-flex align-center">
                        <span>{{ tt('Assets') }}</span>
                        <v-btn class="ms-3" color="default" variant="outlined"
                               :disabled="loading || updating || hasEditingAsset" @click="add">{{ tt('Add') }}</v-btn>
                        <v-btn class="ms-3" color="primary" variant="tonal"
                               :disabled="loading || updating || hasEditingAsset" @click="saveSortResult"
                               v-if="displayOrderModified">{{ tt('Save Display Order') }}</v-btn>
                        <v-btn density="compact" color="default" variant="text" size="24"
                               class="ms-2" :icon="true" :disabled="loading || updating || hasEditingAsset"
                               :loading="loading" @click="reload">
                            <template #loader>
                                <v-progress-circular indeterminate size="20"/>
                            </template>
                            <v-icon :icon="mdiRefresh" size="24" />
                            <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                        </v-btn>
                        <v-spacer/>
                        <v-btn density="comfortable" color="default" variant="text" class="ms-2"
                               :disabled="loading || updating || hasEditingAsset" :icon="true">
                            <v-icon :icon="mdiDotsVertical" />
                            <v-menu activator="parent">
                                <v-list>
                                    <v-list-item :prepend-icon="mdiEyeOutline"
                                                 :title="tt('Show Hidden Assets')"
                                                 v-if="!showHidden" @click="showHidden = true"></v-list-item>
                                    <v-list-item :prepend-icon="mdiEyeOffOutline"
                                                 :title="tt('Hide Hidden Assets')"
                                                 v-if="showHidden" @click="showHidden = false"></v-list-item>
                                </v-list>
                            </v-menu>
                        </v-btn>
                    </div>
                </template>

                <v-table class="assets-table table-striped" :hover="!loading">
                    <thead>
                    <tr>
                        <th>
                            <div class="d-flex align-center">
                                <span>{{ tt('Asset Name') }}</span>
                                <span class="ms-4">{{ tt('Type') }}</span>
                                <span class="ms-4">{{ tt('Status') }}</span>
                                <span class="ms-4">{{ tt('Location') }}</span>
                                <v-spacer/>
                                <span>{{ tt('Operation') }}</span>
                            </div>
                        </th>
                    </tr>
                    </thead>

                    <tbody v-if="loading && noAvailableAsset && !newAsset">
                    <tr :key="itemIdx" v-for="itemIdx in [ 1, 2, 3, 4, 5 ]">
                        <td class="px-0">
                            <v-skeleton-loader type="text" :loading="true"></v-skeleton-loader>
                        </td>
                    </tr>
                    </tbody>

                    <tbody v-if="!loading && noAvailableAsset && !newAsset">
                    <tr>
                        <td>{{ tt('No available asset') }}</td>
                    </tr>
                    </tbody>

                    <draggable-list tag="tbody"
                                    item-key="id"
                                    handle=".drag-handle"
                                    ghost-class="dragging-item"
                                    :class="{ 'has-bottom-border': newAsset }"
                                    :disabled="noAvailableAsset"
                                    v-model="assets"
                                    @change="onMove">
                        <template #item="{ element }">
                            <tr class="assets-table-row text-sm" v-if="showHidden || !element.hidden">
                                <td>
                                    <div class="d-flex align-center">
                                        <!-- Display mode -->
                                        <div class="d-flex align-center" v-if="editingAsset.id !== element.id">
                                            <span class="asset-name" :class="{ 'text-medium-emphasis': element.hidden }">{{ element.name }}</span>
                                            <span class="asset-detail text-medium-emphasis ms-4">{{ getAssetTypeName(element.assetType) }}</span>
                                            <span class="asset-detail text-medium-emphasis ms-4">{{ getAssetStatusName(element.status) }}</span>
                                            <span class="asset-detail text-medium-emphasis ms-4" v-if="getLocationName(element.locationId)">{{ getLocationName(element.locationId) }}</span>
                                        </div>

                                        <!-- Edit mode -->
                                        <div class="d-flex align-center w-100 me-2 flex-wrap" v-else-if="editingAsset.id === element.id">
                                            <v-text-field class="me-2 asset-edit-field" type="text"
                                                          density="compact" variant="underlined"
                                                          :disabled="loading || updating"
                                                          :placeholder="tt('Asset Name')"
                                                          v-model="editingAsset.name"
                                                          @keyup.enter="save(editingAsset)">
                                            </v-text-field>
                                            <v-select class="me-2 asset-edit-field-sm"
                                                      density="compact" variant="underlined"
                                                      :disabled="loading || updating"
                                                      :items="assetTypes"
                                                      item-title="name"
                                                      item-value="type"
                                                      v-model="editingAsset.assetType">
                                            </v-select>
                                            <v-select class="me-2 asset-edit-field-sm"
                                                      density="compact" variant="underlined"
                                                      :disabled="loading || updating"
                                                      :items="assetStatuses"
                                                      item-title="name"
                                                      item-value="status"
                                                      v-model="editingAsset.status">
                                            </v-select>
                                            <v-select class="me-2 asset-edit-field-sm"
                                                      density="compact" variant="underlined"
                                                      :disabled="loading || updating"
                                                      :items="locationOptions"
                                                      item-title="name"
                                                      item-value="id"
                                                      v-model="editingAsset.locationId">
                                            </v-select>
                                            <v-text-field class="me-2 asset-edit-field" type="text"
                                                          density="compact" variant="underlined"
                                                          :disabled="loading || updating"
                                                          :placeholder="tt('Comment')"
                                                          v-model="editingAsset.comment"
                                                          @keyup.enter="save(editingAsset)">
                                            </v-text-field>
                                        </div>

                                        <v-spacer/>

                                        <!-- Action buttons - display mode -->
                                        <v-btn class="px-2 ms-2" color="default"
                                               density="comfortable" variant="text"
                                               :class="{ 'd-none': loading, 'hover-display': !loading }"
                                               :prepend-icon="element.hidden ? mdiEyeOutline : mdiEyeOffOutline"
                                               :loading="assetHiding[element.id]"
                                               :disabled="loading || updating"
                                               v-if="editingAsset.id !== element.id"
                                               @click="hide(element, !element.hidden)">
                                            <template #loader>
                                                <v-progress-circular indeterminate size="20" width="2"/>
                                            </template>
                                            {{ element.hidden ? tt('Show') : tt('Hide') }}
                                        </v-btn>
                                        <v-btn class="px-2" color="default"
                                               density="comfortable" variant="text"
                                               :class="{ 'd-none': loading, 'hover-display': !loading }"
                                               :prepend-icon="mdiPencilOutline"
                                               :loading="assetUpdating[element.id]"
                                               :disabled="loading || updating"
                                               v-if="editingAsset.id !== element.id"
                                               @click="edit(element)">
                                            <template #loader>
                                                <v-progress-circular indeterminate size="20" width="2"/>
                                            </template>
                                            {{ tt('Edit') }}
                                        </v-btn>
                                        <v-btn class="px-2" color="default"
                                               density="comfortable" variant="text"
                                               :class="{ 'd-none': loading, 'hover-display': !loading }"
                                               :prepend-icon="mdiDeleteOutline"
                                               :loading="assetRemoving[element.id]"
                                               :disabled="loading || updating"
                                               v-if="editingAsset.id !== element.id"
                                               @click="remove(element)">
                                            <template #loader>
                                                <v-progress-circular indeterminate size="20" width="2"/>
                                            </template>
                                            {{ tt('Delete') }}
                                        </v-btn>

                                        <!-- Action buttons - edit mode -->
                                        <v-btn class="px-2"
                                               density="comfortable" variant="text"
                                               :prepend-icon="mdiCheck"
                                               :loading="assetUpdating[element.id]"
                                               :disabled="loading || updating || !isAssetModified(element)"
                                               v-if="editingAsset.id === element.id" @click="save(editingAsset)">
                                            <template #loader>
                                                <v-progress-circular indeterminate size="20" width="2"/>
                                            </template>
                                            {{ tt('Save') }}
                                        </v-btn>
                                        <v-btn class="px-2" color="default"
                                               density="comfortable" variant="text"
                                               :prepend-icon="mdiClose"
                                               :disabled="loading || updating"
                                               v-if="editingAsset.id === element.id" @click="cancelSave(editingAsset)">
                                            {{ tt('Cancel') }}
                                        </v-btn>

                                        <span class="ms-2">
                                            <v-icon :class="!loading && !updating && !hasEditingAsset && availableAssetCount > 1 ? 'drag-handle' : 'disabled'"
                                                    :icon="mdiDrag"/>
                                            <v-tooltip activator="parent" v-if="!loading && !updating && !hasEditingAsset && availableAssetCount > 1">{{ tt('Drag to Reorder') }}</v-tooltip>
                                        </span>
                                    </div>
                                </td>
                            </tr>
                        </template>
                    </draggable-list>

                    <tbody v-if="newAsset">
                    <tr class="text-sm" :class="{ 'even-row': (availableAssetCount & 1) === 1}">
                        <td>
                            <div class="d-flex align-center flex-wrap">
                                <v-text-field class="me-2 asset-edit-field" type="text" color="primary"
                                              density="compact" variant="underlined"
                                              :disabled="loading || updating"
                                              :placeholder="tt('Asset Name')"
                                              v-model="newAsset.name"
                                              @keyup.enter="save(newAsset)">
                                </v-text-field>
                                <v-select class="me-2 asset-edit-field-sm"
                                          density="compact" variant="underlined"
                                          :disabled="loading || updating"
                                          :items="assetTypes"
                                          item-title="name"
                                          item-value="type"
                                          v-model="newAsset.assetType">
                                </v-select>
                                <v-select class="me-2 asset-edit-field-sm"
                                          density="compact" variant="underlined"
                                          :disabled="loading || updating"
                                          :items="assetStatuses"
                                          item-title="name"
                                          item-value="status"
                                          v-model="newAsset.status">
                                </v-select>
                                <v-select class="me-2 asset-edit-field-sm"
                                          density="compact" variant="underlined"
                                          :disabled="loading || updating"
                                          :items="locationOptions"
                                          item-title="name"
                                          item-value="id"
                                          v-model="newAsset.locationId">
                                </v-select>
                                <v-text-field class="me-2 asset-edit-field" type="text" color="primary"
                                              density="compact" variant="underlined"
                                              :disabled="loading || updating"
                                              :placeholder="tt('Comment')"
                                              v-model="newAsset.comment"
                                              @keyup.enter="save(newAsset)">
                                </v-text-field>

                                <v-spacer/>

                                <v-btn class="px-2" density="comfortable" variant="text"
                                       :prepend-icon="mdiCheck"
                                       :loading="assetUpdating['']"
                                       :disabled="loading || updating || !isAssetModified(newAsset)"
                                       @click="save(newAsset)">
                                    <template #loader>
                                        <v-progress-circular indeterminate size="20" width="2"/>
                                    </template>
                                    {{ tt('Save') }}
                                </v-btn>
                                <v-btn class="px-2" color="default"
                                       density="comfortable" variant="text"
                                       :prepend-icon="mdiClose"
                                       :disabled="loading || updating"
                                       @click="cancelSave(newAsset)">
                                    {{ tt('Cancel') }}
                                </v-btn>
                                <span class="ms-2">
                                    <v-icon class="disabled" :icon="mdiDrag"/>
                                </span>
                            </div>
                        </td>
                    </tr>
                    </tbody>
                </v-table>
            </v-card>
        </v-col>
    </v-row>

    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';

import { ref, computed, useTemplateRef } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

import { useAssetsStore } from '@/stores/asset.ts';
import { useLocationsStore } from '@/stores/location.ts';
import { useCFOsStore } from '@/stores/cfo.ts';

import { Asset } from '@/models/asset.ts';

import {
    mdiRefresh,
    mdiPencilOutline,
    mdiCheck,
    mdiClose,
    mdiEyeOffOutline,
    mdiEyeOutline,
    mdiDeleteOutline,
    mdiDrag,
    mdiDotsVertical
} from '@mdi/js';

type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;
type SnackBarType = InstanceType<typeof SnackBar>;

const { tt } = useI18n();

const assetsStore = useAssetsStore();
const locationsStore = useLocationsStore();
const cfosStore = useCFOsStore();

const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');
const snackbar = useTemplateRef<SnackBarType>('snackbar');

const loading = ref<boolean>(true);
const updating = ref<boolean>(false);
const showHidden = ref<boolean>(false);
const displayOrderModified = ref<boolean>(false);
const newAsset = ref<Asset | null>(null);
const editingAsset = ref<Asset>(Asset.createNew());
const assetUpdating = ref<Record<string, boolean>>({});
const assetHiding = ref<Record<string, boolean>>({});
const assetRemoving = ref<Record<string, boolean>>({});

const assets = computed<Asset[]>(() => assetsStore.allAssets);

const assetTypes = computed(() => [
    { type: 1, name: tt('Equipment') },
    { type: 2, name: tt('Furniture') },
    { type: 3, name: tt('Vehicle') },
    { type: 4, name: tt('Electronics') },
    { type: 5, name: tt('Real Estate') },
    { type: 6, name: tt('Other') }
]);

const assetStatuses = computed(() => [
    { status: 1, name: tt('Active') },
    { status: 2, name: tt('Decommissioned') },
    { status: 3, name: tt('Sold') }
]);

const locationOptions = computed(() => {
    const options = [{ id: '0', name: tt('No Location') }];
    for (const location of locationsStore.allLocations) {
        if (!location.hidden) {
            options.push({ id: location.id, name: location.name });
        }
    }
    return options;
});

const noAvailableAsset = computed<boolean>(() => {
    if (!assets.value || assets.value.length < 1) {
        return true;
    }

    if (showHidden.value) {
        return false;
    }

    for (const asset of assets.value) {
        if (!asset.hidden) {
            return false;
        }
    }

    return true;
});

const availableAssetCount = computed<number>(() => {
    if (!assets.value) {
        return 0;
    }

    if (showHidden.value) {
        return assets.value.length;
    }

    let count = 0;

    for (const asset of assets.value) {
        if (!asset.hidden) {
            count++;
        }
    }

    return count;
});

const hasEditingAsset = computed<boolean>(() => {
    return !!(newAsset.value || (editingAsset.value.id && editingAsset.value.id !== ''));
});

function getAssetTypeName(type: number): string {
    switch (type) {
        case 1: return tt('Equipment');
        case 2: return tt('Furniture');
        case 3: return tt('Vehicle');
        case 4: return tt('Electronics');
        case 5: return tt('Real Estate');
        case 6: return tt('Other');
        default: return tt('Other');
    }
}

function getAssetStatusName(status: number): string {
    switch (status) {
        case 1: return tt('Active');
        case 2: return tt('Decommissioned');
        case 3: return tt('Sold');
        default: return tt('Active');
    }
}

function getLocationName(locationId: string): string {
    if (!locationId || locationId === '0') {
        return '';
    }
    const location = locationsStore.allLocationsMap[locationId];
    return location ? location.name : '';
}

function isAssetModified(asset: Asset): boolean {
    if (asset.id) {
        const original = assetsStore.allAssetsMap[asset.id];

        if (!original) {
            return false;
        }

        return editingAsset.value.name !== original.name
            || editingAsset.value.assetType !== original.assetType
            || editingAsset.value.status !== original.status
            || editingAsset.value.locationId !== original.locationId
            || editingAsset.value.comment !== original.comment;
    } else {
        return asset.name !== '';
    }
}

function reload(): void {
    if (hasEditingAsset.value) {
        return;
    }

    loading.value = true;

    assetsStore.loadAllAssets({
        force: true
    }).then(() => {
        loading.value = false;
        displayOrderModified.value = false;

        snackbar.value?.showMessage('Asset list has been updated');
    }).catch(error => {
        loading.value = false;

        if (error && error.isUpToDate) {
            displayOrderModified.value = false;
        }

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function add(): void {
    newAsset.value = Asset.createNew();
}

function edit(asset: Asset): void {
    editingAsset.value = asset.clone();
}

function save(asset: Asset): void {
    updating.value = true;
    assetUpdating.value[asset.id || ''] = true;

    assetsStore.saveAsset({
        asset: asset
    }).then(() => {
        updating.value = false;
        assetUpdating.value[asset.id || ''] = false;

        if (asset.id) {
            editingAsset.value = Asset.createNew();
        } else {
            newAsset.value = null;
        }
    }).catch(error => {
        updating.value = false;
        assetUpdating.value[asset.id || ''] = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function cancelSave(asset: Asset): void {
    if (asset.id) {
        editingAsset.value = Asset.createNew();
    } else {
        newAsset.value = null;
    }
}

function saveSortResult(): void {
    if (!displayOrderModified.value) {
        return;
    }

    loading.value = true;

    assetsStore.updateAssetDisplayOrders().then(() => {
        loading.value = false;
        displayOrderModified.value = false;
    }).catch(error => {
        loading.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function hide(asset: Asset, hidden: boolean): void {
    updating.value = true;
    assetHiding.value[asset.id] = true;

    assetsStore.hideAsset({
        asset: asset,
        hidden: hidden
    }).then(() => {
        updating.value = false;
        assetHiding.value[asset.id] = false;
    }).catch(error => {
        updating.value = false;
        assetHiding.value[asset.id] = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function remove(asset: Asset): void {
    confirmDialog.value?.open('Are you sure you want to delete this asset?').then(() => {
        updating.value = true;
        assetRemoving.value[asset.id] = true;

        assetsStore.deleteAsset({
            asset: asset
        }).then(() => {
            updating.value = false;
            assetRemoving.value[asset.id] = false;
        }).catch(error => {
            updating.value = false;
            assetRemoving.value[asset.id] = false;

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    });
}

function onMove(event: { moved: { element: { id: string }; oldIndex: number; newIndex: number } }): void {
    if (!event || !event.moved) {
        return;
    }

    const moveEvent = event.moved;

    if (!moveEvent.element || !moveEvent.element.id) {
        snackbar.value?.showMessage('Unable to move asset');
        return;
    }

    assetsStore.changeAssetDisplayOrder({
        assetId: moveEvent.element.id,
        from: moveEvent.oldIndex,
        to: moveEvent.newIndex
    }).then(() => {
        displayOrderModified.value = true;
    }).catch(error => {
        snackbar.value?.showError(error);
    });
}

// Load dependencies first, then assets
Promise.all([
    cfosStore.loadAllCFOs({ force: false }),
    locationsStore.loadAllLocations({ force: false })
]).then(() => {
    return assetsStore.loadAllAssets({ force: false });
}).then(() => {
    loading.value = false;
}).catch(error => {
    loading.value = false;

    if (!error.processed) {
        snackbar.value?.showError(error);
    }
});
</script>

<style>
.assets-table tr.assets-table-row .hover-display {
    display: none;
}

.assets-table tr.assets-table-row:hover .hover-display {
    display: inline-grid;
}

.assets-table tr:not(:last-child) > td > div {
    padding-bottom: 1px;
}

.assets-table .has-bottom-border tr:last-child > td > div {
    padding-bottom: 1px;
}

.assets-table .v-text-field .v-field__input {
    font-size: 0.875rem;
    padding-top: 0;
    color: rgba(var(--v-theme-on-surface));
}

.assets-table .asset-name {
    font-size: 0.875rem;
}

.assets-table .asset-detail {
    font-size: 0.8125rem;
}

.assets-table tr .v-text-field .v-field__input {
    padding-bottom: 1px;
}

.assets-table tr .v-input--density-compact .v-field__input {
    padding-bottom: 1px;
}

.asset-edit-field {
    min-width: 120px;
    max-width: 200px;
}

.asset-edit-field-sm {
    min-width: 100px;
    max-width: 150px;
}
</style>
