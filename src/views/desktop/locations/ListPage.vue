<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="title-and-toolbar d-flex align-center">
                        <span>{{ tt('Locations') }}</span>
                        <v-btn class="ms-3" color="default" variant="outlined"
                               :disabled="loading || updating || hasEditingLocation" @click="add">{{ tt('Add') }}</v-btn>
                        <v-btn class="ms-3" color="primary" variant="tonal"
                               :disabled="loading || updating || hasEditingLocation" @click="saveSortResult"
                               v-if="displayOrderModified">{{ tt('Save Display Order') }}</v-btn>
                        <v-btn density="compact" color="default" variant="text" size="24"
                               class="ms-2" :icon="true" :disabled="loading || updating || hasEditingLocation"
                               :loading="loading" @click="reload">
                            <template #loader>
                                <v-progress-circular indeterminate size="20"/>
                            </template>
                            <v-icon :icon="mdiRefresh" size="24" />
                            <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                        </v-btn>
                        <v-spacer/>
                        <v-btn density="comfortable" color="default" variant="text" class="ms-2"
                               :disabled="loading || updating || hasEditingLocation" :icon="true">
                            <v-icon :icon="mdiDotsVertical" />
                            <v-menu activator="parent">
                                <v-list>
                                    <v-list-item :prepend-icon="mdiEyeOutline"
                                                 :title="tt('Show Hidden Locations')"
                                                 v-if="!showHidden" @click="showHidden = true"></v-list-item>
                                    <v-list-item :prepend-icon="mdiEyeOffOutline"
                                                 :title="tt('Hide Hidden Locations')"
                                                 v-if="showHidden" @click="showHidden = false"></v-list-item>
                                </v-list>
                            </v-menu>
                        </v-btn>
                    </div>
                </template>

                <v-table class="locations-table table-striped" :hover="!loading">
                    <thead>
                    <tr>
                        <th>
                            <div class="d-flex align-center">
                                <span>{{ tt('Location Name') }}</span>
                                <span class="ms-4">{{ tt('Address') }}</span>
                                <span class="ms-4">{{ tt('Type') }}</span>
                                <span class="ms-4">{{ tt('CFO') }}</span>
                                <v-spacer/>
                                <span>{{ tt('Operation') }}</span>
                            </div>
                        </th>
                    </tr>
                    </thead>

                    <tbody v-if="loading && noAvailableLocation && !newLocation">
                    <tr :key="itemIdx" v-for="itemIdx in [ 1, 2, 3, 4, 5 ]">
                        <td class="px-0">
                            <v-skeleton-loader type="text" :loading="true"></v-skeleton-loader>
                        </td>
                    </tr>
                    </tbody>

                    <tbody v-if="!loading && noAvailableLocation && !newLocation">
                    <tr>
                        <td>{{ tt('No available location') }}</td>
                    </tr>
                    </tbody>

                    <draggable-list tag="tbody"
                                    item-key="id"
                                    handle=".drag-handle"
                                    ghost-class="dragging-item"
                                    :class="{ 'has-bottom-border': newLocation }"
                                    :disabled="noAvailableLocation"
                                    v-model="locations"
                                    @change="onMove">
                        <template #item="{ element }">
                            <tr class="locations-table-row text-sm" v-if="showHidden || !element.hidden">
                                <td>
                                    <div class="d-flex align-center">
                                        <!-- Display mode -->
                                        <div class="d-flex align-center" v-if="editingLocation.id !== element.id">
                                            <span class="location-name" :class="{ 'text-medium-emphasis': element.hidden }">{{ element.name }}</span>
                                            <span class="location-detail text-medium-emphasis ms-4" v-if="element.address">{{ element.address }}</span>
                                            <span class="location-detail text-medium-emphasis ms-4">{{ getLocationTypeName(element.locationType) }}</span>
                                            <span class="location-detail text-medium-emphasis ms-4" v-if="getCFOName(element.cfoId)">{{ getCFOName(element.cfoId) }}</span>
                                        </div>

                                        <!-- Edit mode -->
                                        <div class="d-flex align-center w-100 me-2 flex-wrap" v-else-if="editingLocation.id === element.id">
                                            <v-text-field class="me-2 location-edit-field" type="text"
                                                          density="compact" variant="underlined"
                                                          :disabled="loading || updating"
                                                          :placeholder="tt('Location Name')"
                                                          v-model="editingLocation.name"
                                                          @keyup.enter="save(editingLocation)">
                                            </v-text-field>
                                            <v-text-field class="me-2 location-edit-field" type="text"
                                                          density="compact" variant="underlined"
                                                          :disabled="loading || updating"
                                                          :placeholder="tt('Address')"
                                                          v-model="editingLocation.address"
                                                          @keyup.enter="save(editingLocation)">
                                            </v-text-field>
                                            <v-select class="me-2 location-edit-field-sm"
                                                      density="compact" variant="underlined"
                                                      :disabled="loading || updating"
                                                      :items="locationTypes"
                                                      item-title="name"
                                                      item-value="type"
                                                      v-model="editingLocation.locationType">
                                            </v-select>
                                            <v-select class="me-2 location-edit-field-sm"
                                                      density="compact" variant="underlined"
                                                      :disabled="loading || updating"
                                                      :items="cfoOptions"
                                                      item-title="name"
                                                      item-value="id"
                                                      v-model="editingLocation.cfoId">
                                            </v-select>
                                            <v-text-field class="me-2 location-edit-field" type="text"
                                                          density="compact" variant="underlined"
                                                          :disabled="loading || updating"
                                                          :placeholder="tt('Comment')"
                                                          v-model="editingLocation.comment"
                                                          @keyup.enter="save(editingLocation)">
                                            </v-text-field>
                                        </div>

                                        <v-spacer/>

                                        <!-- Action buttons - display mode -->
                                        <v-btn class="px-2 ms-2" color="default"
                                               density="comfortable" variant="text"
                                               :class="{ 'd-none': loading, 'hover-display': !loading }"
                                               :prepend-icon="element.hidden ? mdiEyeOutline : mdiEyeOffOutline"
                                               :loading="locationHiding[element.id]"
                                               :disabled="loading || updating"
                                               v-if="editingLocation.id !== element.id"
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
                                               :loading="locationUpdating[element.id]"
                                               :disabled="loading || updating"
                                               v-if="editingLocation.id !== element.id"
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
                                               :loading="locationRemoving[element.id]"
                                               :disabled="loading || updating"
                                               v-if="editingLocation.id !== element.id"
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
                                               :loading="locationUpdating[element.id]"
                                               :disabled="loading || updating || !isLocationModified(element)"
                                               v-if="editingLocation.id === element.id" @click="save(editingLocation)">
                                            <template #loader>
                                                <v-progress-circular indeterminate size="20" width="2"/>
                                            </template>
                                            {{ tt('Save') }}
                                        </v-btn>
                                        <v-btn class="px-2" color="default"
                                               density="comfortable" variant="text"
                                               :prepend-icon="mdiClose"
                                               :disabled="loading || updating"
                                               v-if="editingLocation.id === element.id" @click="cancelSave(editingLocation)">
                                            {{ tt('Cancel') }}
                                        </v-btn>

                                        <span class="ms-2">
                                            <v-icon :class="!loading && !updating && !hasEditingLocation && availableLocationCount > 1 ? 'drag-handle' : 'disabled'"
                                                    :icon="mdiDrag"/>
                                            <v-tooltip activator="parent" v-if="!loading && !updating && !hasEditingLocation && availableLocationCount > 1">{{ tt('Drag to Reorder') }}</v-tooltip>
                                        </span>
                                    </div>
                                </td>
                            </tr>
                        </template>
                    </draggable-list>

                    <tbody v-if="newLocation">
                    <tr class="text-sm" :class="{ 'even-row': (availableLocationCount & 1) === 1}">
                        <td>
                            <div class="d-flex align-center flex-wrap">
                                <v-text-field class="me-2 location-edit-field" type="text" color="primary"
                                              density="compact" variant="underlined"
                                              :disabled="loading || updating"
                                              :placeholder="tt('Location Name')"
                                              v-model="newLocation.name"
                                              @keyup.enter="save(newLocation)">
                                </v-text-field>
                                <v-text-field class="me-2 location-edit-field" type="text" color="primary"
                                              density="compact" variant="underlined"
                                              :disabled="loading || updating"
                                              :placeholder="tt('Address')"
                                              v-model="newLocation.address"
                                              @keyup.enter="save(newLocation)">
                                </v-text-field>
                                <v-select class="me-2 location-edit-field-sm"
                                          density="compact" variant="underlined"
                                          :disabled="loading || updating"
                                          :items="locationTypes"
                                          item-title="name"
                                          item-value="type"
                                          v-model="newLocation.locationType">
                                </v-select>
                                <v-select class="me-2 location-edit-field-sm"
                                          density="compact" variant="underlined"
                                          :disabled="loading || updating"
                                          :items="cfoOptions"
                                          item-title="name"
                                          item-value="id"
                                          v-model="newLocation.cfoId">
                                </v-select>
                                <v-text-field class="me-2 location-edit-field" type="text" color="primary"
                                              density="compact" variant="underlined"
                                              :disabled="loading || updating"
                                              :placeholder="tt('Comment')"
                                              v-model="newLocation.comment"
                                              @keyup.enter="save(newLocation)">
                                </v-text-field>

                                <v-spacer/>

                                <v-btn class="px-2" density="comfortable" variant="text"
                                       :prepend-icon="mdiCheck"
                                       :loading="locationUpdating['']"
                                       :disabled="loading || updating || !isLocationModified(newLocation)"
                                       @click="save(newLocation)">
                                    <template #loader>
                                        <v-progress-circular indeterminate size="20" width="2"/>
                                    </template>
                                    {{ tt('Save') }}
                                </v-btn>
                                <v-btn class="px-2" color="default"
                                       density="comfortable" variant="text"
                                       :prepend-icon="mdiClose"
                                       :disabled="loading || updating"
                                       @click="cancelSave(newLocation)">
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

import { useLocationsStore } from '@/stores/location.ts';
import { useCFOsStore } from '@/stores/cfo.ts';

import { Location } from '@/models/location.ts';

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

const locationsStore = useLocationsStore();
const cfosStore = useCFOsStore();

const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');
const snackbar = useTemplateRef<SnackBarType>('snackbar');

const loading = ref<boolean>(true);
const updating = ref<boolean>(false);
const showHidden = ref<boolean>(false);
const displayOrderModified = ref<boolean>(false);
const newLocation = ref<Location | null>(null);
const editingLocation = ref<Location>(Location.createNew());
const locationUpdating = ref<Record<string, boolean>>({});
const locationHiding = ref<Record<string, boolean>>({});
const locationRemoving = ref<Record<string, boolean>>({});

const locations = computed<Location[]>(() => locationsStore.allLocations);

const locationTypes = computed(() => [
    { type: 1, name: tt('Office') },
    { type: 2, name: tt('Warehouse') },
    { type: 3, name: tt('Store') },
    { type: 4, name: tt('Production') },
    { type: 5, name: tt('Other') }
]);

const cfoOptions = computed(() => {
    const options = [{ id: '0', name: tt('No CFO') }];
    for (const cfo of cfosStore.allCFOs) {
        if (!cfo.hidden) {
            options.push({ id: cfo.id, name: cfo.name });
        }
    }
    return options;
});

const noAvailableLocation = computed<boolean>(() => {
    if (!locations.value || locations.value.length < 1) {
        return true;
    }

    if (showHidden.value) {
        return false;
    }

    for (const location of locations.value) {
        if (!location.hidden) {
            return false;
        }
    }

    return true;
});

const availableLocationCount = computed<number>(() => {
    if (!locations.value) {
        return 0;
    }

    if (showHidden.value) {
        return locations.value.length;
    }

    let count = 0;

    for (const location of locations.value) {
        if (!location.hidden) {
            count++;
        }
    }

    return count;
});

const hasEditingLocation = computed<boolean>(() => {
    return !!(newLocation.value || (editingLocation.value.id && editingLocation.value.id !== ''));
});

function getLocationTypeName(type: number): string {
    switch (type) {
        case 1: return tt('Office');
        case 2: return tt('Warehouse');
        case 3: return tt('Store');
        case 4: return tt('Production');
        case 5: return tt('Other');
        default: return tt('Other');
    }
}

function getCFOName(cfoId: string): string {
    if (!cfoId || cfoId === '0') {
        return '';
    }
    const cfo = cfosStore.allCFOsMap[cfoId];
    return cfo ? cfo.name : '';
}

function isLocationModified(location: Location): boolean {
    if (location.id) {
        const original = locationsStore.allLocationsMap[location.id];

        if (!original) {
            return false;
        }

        return editingLocation.value.name !== original.name
            || editingLocation.value.address !== original.address
            || editingLocation.value.locationType !== original.locationType
            || editingLocation.value.cfoId !== original.cfoId
            || editingLocation.value.comment !== original.comment;
    } else {
        return location.name !== '';
    }
}

function reload(): void {
    if (hasEditingLocation.value) {
        return;
    }

    loading.value = true;

    locationsStore.loadAllLocations({
        force: true
    }).then(() => {
        loading.value = false;
        displayOrderModified.value = false;

        snackbar.value?.showMessage('Location list has been updated');
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
    newLocation.value = Location.createNew();
}

function edit(location: Location): void {
    editingLocation.value = location.clone();
}

function save(location: Location): void {
    updating.value = true;
    locationUpdating.value[location.id || ''] = true;

    locationsStore.saveLocation({
        location: location
    }).then(() => {
        updating.value = false;
        locationUpdating.value[location.id || ''] = false;

        if (location.id) {
            editingLocation.value = Location.createNew();
        } else {
            newLocation.value = null;
        }
    }).catch(error => {
        updating.value = false;
        locationUpdating.value[location.id || ''] = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function cancelSave(location: Location): void {
    if (location.id) {
        editingLocation.value = Location.createNew();
    } else {
        newLocation.value = null;
    }
}

function saveSortResult(): void {
    if (!displayOrderModified.value) {
        return;
    }

    loading.value = true;

    locationsStore.updateLocationDisplayOrders().then(() => {
        loading.value = false;
        displayOrderModified.value = false;
    }).catch(error => {
        loading.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function hide(location: Location, hidden: boolean): void {
    updating.value = true;
    locationHiding.value[location.id] = true;

    locationsStore.hideLocation({
        location: location,
        hidden: hidden
    }).then(() => {
        updating.value = false;
        locationHiding.value[location.id] = false;
    }).catch(error => {
        updating.value = false;
        locationHiding.value[location.id] = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function remove(location: Location): void {
    confirmDialog.value?.open('Are you sure you want to delete this location?').then(() => {
        updating.value = true;
        locationRemoving.value[location.id] = true;

        locationsStore.deleteLocation({
            location: location
        }).then(() => {
            updating.value = false;
            locationRemoving.value[location.id] = false;
        }).catch(error => {
            updating.value = false;
            locationRemoving.value[location.id] = false;

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
        snackbar.value?.showMessage('Unable to move location');
        return;
    }

    locationsStore.changeLocationDisplayOrder({
        locationId: moveEvent.element.id,
        from: moveEvent.oldIndex,
        to: moveEvent.newIndex
    }).then(() => {
        displayOrderModified.value = true;
    }).catch(error => {
        snackbar.value?.showError(error);
    });
}

// Load CFOs first, then locations
cfosStore.loadAllCFOs({ force: false }).then(() => {
    return locationsStore.loadAllLocations({ force: false });
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
.locations-table tr.locations-table-row .hover-display {
    display: none;
}

.locations-table tr.locations-table-row:hover .hover-display {
    display: inline-grid;
}

.locations-table tr:not(:last-child) > td > div {
    padding-bottom: 1px;
}

.locations-table .has-bottom-border tr:last-child > td > div {
    padding-bottom: 1px;
}

.locations-table .v-text-field .v-field__input {
    font-size: 0.875rem;
    padding-top: 0;
    color: rgba(var(--v-theme-on-surface));
}

.locations-table .location-name {
    font-size: 0.875rem;
}

.locations-table .location-detail {
    font-size: 0.8125rem;
}

.locations-table tr .v-text-field .v-field__input {
    padding-bottom: 1px;
}

.locations-table tr .v-input--density-compact .v-field__input {
    padding-bottom: 1px;
}

.location-edit-field {
    min-width: 120px;
    max-width: 200px;
}

.location-edit-field-sm {
    min-width: 100px;
    max-width: 150px;
}
</style>
