<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="title-and-toolbar d-flex align-center">
                        <span>{{ tt('CFOs') }}</span>
                        <v-btn class="ms-3" color="default" variant="outlined"
                               :disabled="loading || updating || hasEditingCFO" @click="add">{{ tt('Add') }}</v-btn>
                        <v-btn class="ms-3" color="primary" variant="tonal"
                               :disabled="loading || updating || hasEditingCFO" @click="saveSortResult"
                               v-if="displayOrderModified">{{ tt('Save Display Order') }}</v-btn>
                        <v-btn density="compact" color="default" variant="text" size="24"
                               class="ms-2" :icon="true" :disabled="loading || updating || hasEditingCFO"
                               :loading="loading" @click="reload">
                            <template #loader>
                                <v-progress-circular indeterminate size="20"/>
                            </template>
                            <v-icon :icon="mdiRefresh" size="24" />
                            <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                        </v-btn>
                        <v-spacer/>
                        <v-btn density="comfortable" color="default" variant="text" class="ms-2"
                               :disabled="loading || updating || hasEditingCFO" :icon="true">
                            <v-icon :icon="mdiDotsVertical" />
                            <v-menu activator="parent">
                                <v-list>
                                    <v-list-item :prepend-icon="mdiEyeOutline"
                                                 :title="tt('Show Hidden CFOs')"
                                                 v-if="!showHidden" @click="showHidden = true"></v-list-item>
                                    <v-list-item :prepend-icon="mdiEyeOffOutline"
                                                 :title="tt('Hide Hidden CFOs')"
                                                 v-if="showHidden" @click="showHidden = false"></v-list-item>
                                </v-list>
                            </v-menu>
                        </v-btn>
                    </div>
                </template>

                <v-table class="cfos-table table-striped" :hover="!loading">
                    <thead>
                    <tr>
                        <th>
                            <div class="d-flex align-center">
                                <span>{{ tt('CFO Name') }}</span>
                                <span class="ms-4">{{ tt('Comment') }}</span>
                                <v-spacer/>
                                <span>{{ tt('Operation') }}</span>
                            </div>
                        </th>
                    </tr>
                    </thead>

                    <tbody v-if="loading && noAvailableCFO && !newCFO">
                    <tr :key="itemIdx" v-for="itemIdx in [ 1, 2, 3, 4, 5 ]">
                        <td class="px-0">
                            <v-skeleton-loader type="text" :loading="true"></v-skeleton-loader>
                        </td>
                    </tr>
                    </tbody>

                    <tbody v-if="!loading && noAvailableCFO && !newCFO">
                    <tr>
                        <td>{{ tt('No available CFO') }}</td>
                    </tr>
                    </tbody>

                    <draggable-list tag="tbody"
                                    item-key="id"
                                    handle=".drag-handle"
                                    ghost-class="dragging-item"
                                    :class="{ 'has-bottom-border': newCFO }"
                                    :disabled="noAvailableCFO"
                                    v-model="cfos"
                                    @change="onMove">
                        <template #item="{ element }">
                            <tr class="cfos-table-row text-sm" v-if="showHidden || !element.hidden">
                                <td>
                                    <div class="d-flex align-center">
                                        <!-- Display mode -->
                                        <div class="d-flex align-center" v-if="editingCFO.id !== element.id">
                                            <span class="cfo-name" :class="{ 'text-medium-emphasis': element.hidden }">{{ element.name }}</span>
                                            <span class="cfo-comment text-medium-emphasis ms-4" v-if="element.comment">{{ element.comment }}</span>
                                        </div>

                                        <!-- Edit mode -->
                                        <div class="d-flex align-center w-100 me-2" v-else-if="editingCFO.id === element.id">
                                            <v-text-field class="me-2" type="text"
                                                          density="compact" variant="underlined"
                                                          :disabled="loading || updating"
                                                          :placeholder="tt('CFO Name')"
                                                          v-model="editingCFO.name"
                                                          @keyup.enter="save(editingCFO)">
                                            </v-text-field>
                                            <v-text-field class="me-2" type="text"
                                                          density="compact" variant="underlined"
                                                          :disabled="loading || updating"
                                                          :placeholder="tt('Comment')"
                                                          v-model="editingCFO.comment"
                                                          @keyup.enter="save(editingCFO)">
                                            </v-text-field>
                                        </div>

                                        <v-spacer/>

                                        <!-- Action buttons - display mode -->
                                        <v-btn class="px-2 ms-2" color="default"
                                               density="comfortable" variant="text"
                                               :class="{ 'd-none': loading, 'hover-display': !loading }"
                                               :prepend-icon="element.hidden ? mdiEyeOutline : mdiEyeOffOutline"
                                               :loading="cfoHiding[element.id]"
                                               :disabled="loading || updating"
                                               v-if="editingCFO.id !== element.id"
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
                                               :loading="cfoUpdating[element.id]"
                                               :disabled="loading || updating"
                                               v-if="editingCFO.id !== element.id"
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
                                               :loading="cfoRemoving[element.id]"
                                               :disabled="loading || updating"
                                               v-if="editingCFO.id !== element.id"
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
                                               :loading="cfoUpdating[element.id]"
                                               :disabled="loading || updating || !isCFOModified(element)"
                                               v-if="editingCFO.id === element.id" @click="save(editingCFO)">
                                            <template #loader>
                                                <v-progress-circular indeterminate size="20" width="2"/>
                                            </template>
                                            {{ tt('Save') }}
                                        </v-btn>
                                        <v-btn class="px-2" color="default"
                                               density="comfortable" variant="text"
                                               :prepend-icon="mdiClose"
                                               :disabled="loading || updating"
                                               v-if="editingCFO.id === element.id" @click="cancelSave(editingCFO)">
                                            {{ tt('Cancel') }}
                                        </v-btn>

                                        <span class="ms-2">
                                            <v-icon :class="!loading && !updating && !hasEditingCFO && availableCFOCount > 1 ? 'drag-handle' : 'disabled'"
                                                    :icon="mdiDrag"/>
                                            <v-tooltip activator="parent" v-if="!loading && !updating && !hasEditingCFO && availableCFOCount > 1">{{ tt('Drag to Reorder') }}</v-tooltip>
                                        </span>
                                    </div>
                                </td>
                            </tr>
                        </template>
                    </draggable-list>

                    <tbody v-if="newCFO">
                    <tr class="text-sm" :class="{ 'even-row': (availableCFOCount & 1) === 1}">
                        <td>
                            <div class="d-flex align-center">
                                <v-text-field class="me-2" type="text" color="primary"
                                              density="compact" variant="underlined"
                                              :disabled="loading || updating"
                                              :placeholder="tt('CFO Name')"
                                              v-model="newCFO.name"
                                              @keyup.enter="save(newCFO)">
                                </v-text-field>
                                <v-text-field class="me-2" type="text" color="primary"
                                              density="compact" variant="underlined"
                                              :disabled="loading || updating"
                                              :placeholder="tt('Comment')"
                                              v-model="newCFO.comment"
                                              @keyup.enter="save(newCFO)">
                                </v-text-field>

                                <v-spacer/>

                                <v-btn class="px-2" density="comfortable" variant="text"
                                       :prepend-icon="mdiCheck"
                                       :loading="cfoUpdating['']"
                                       :disabled="loading || updating || !isCFOModified(newCFO)"
                                       @click="save(newCFO)">
                                    <template #loader>
                                        <v-progress-circular indeterminate size="20" width="2"/>
                                    </template>
                                    {{ tt('Save') }}
                                </v-btn>
                                <v-btn class="px-2" color="default"
                                       density="comfortable" variant="text"
                                       :prepend-icon="mdiClose"
                                       :disabled="loading || updating"
                                       @click="cancelSave(newCFO)">
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

import { useCFOsStore } from '@/stores/cfo.ts';

import { CFO } from '@/models/cfo.ts';

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

const cfosStore = useCFOsStore();

const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');
const snackbar = useTemplateRef<SnackBarType>('snackbar');

const loading = ref<boolean>(true);
const updating = ref<boolean>(false);
const showHidden = ref<boolean>(false);
const displayOrderModified = ref<boolean>(false);
const newCFO = ref<CFO | null>(null);
const editingCFO = ref<CFO>(CFO.createNew());
const cfoUpdating = ref<Record<string, boolean>>({});
const cfoHiding = ref<Record<string, boolean>>({});
const cfoRemoving = ref<Record<string, boolean>>({});

const cfos = computed<CFO[]>(() => cfosStore.allCFOs);

const noAvailableCFO = computed<boolean>(() => {
    if (!cfos.value || cfos.value.length < 1) {
        return true;
    }

    if (showHidden.value) {
        return false;
    }

    for (const cfo of cfos.value) {
        if (!cfo.hidden) {
            return false;
        }
    }

    return true;
});

const availableCFOCount = computed<number>(() => {
    if (!cfos.value) {
        return 0;
    }

    if (showHidden.value) {
        return cfos.value.length;
    }

    let count = 0;

    for (const cfo of cfos.value) {
        if (!cfo.hidden) {
            count++;
        }
    }

    return count;
});

const hasEditingCFO = computed<boolean>(() => {
    return !!(newCFO.value || (editingCFO.value.id && editingCFO.value.id !== ''));
});

function isCFOModified(cfo: CFO): boolean {
    if (cfo.id) {
        const original = cfosStore.allCFOsMap[cfo.id];

        if (!original) {
            return false;
        }

        return editingCFO.value.name !== original.name
            || editingCFO.value.comment !== original.comment;
    } else {
        return cfo.name !== '';
    }
}

function reload(): void {
    if (hasEditingCFO.value) {
        return;
    }

    loading.value = true;

    cfosStore.loadAllCFOs({
        force: true
    }).then(() => {
        loading.value = false;
        displayOrderModified.value = false;

        snackbar.value?.showMessage('CFO list has been updated');
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
    newCFO.value = CFO.createNew();
}

function edit(cfo: CFO): void {
    editingCFO.value = cfo.clone();
}

function save(cfo: CFO): void {
    updating.value = true;
    cfoUpdating.value[cfo.id || ''] = true;

    cfosStore.saveCFO({
        cfo: cfo
    }).then(() => {
        updating.value = false;
        cfoUpdating.value[cfo.id || ''] = false;

        if (cfo.id) {
            editingCFO.value = CFO.createNew();
        } else {
            newCFO.value = null;
        }
    }).catch(error => {
        updating.value = false;
        cfoUpdating.value[cfo.id || ''] = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function cancelSave(cfo: CFO): void {
    if (cfo.id) {
        editingCFO.value = CFO.createNew();
    } else {
        newCFO.value = null;
    }
}

function saveSortResult(): void {
    if (!displayOrderModified.value) {
        return;
    }

    loading.value = true;

    cfosStore.updateCFODisplayOrders().then(() => {
        loading.value = false;
        displayOrderModified.value = false;
    }).catch(error => {
        loading.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function hide(cfo: CFO, hidden: boolean): void {
    updating.value = true;
    cfoHiding.value[cfo.id] = true;

    cfosStore.hideCFO({
        cfo: cfo,
        hidden: hidden
    }).then(() => {
        updating.value = false;
        cfoHiding.value[cfo.id] = false;
    }).catch(error => {
        updating.value = false;
        cfoHiding.value[cfo.id] = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function remove(cfo: CFO): void {
    confirmDialog.value?.open('Are you sure you want to delete this CFO?').then(() => {
        updating.value = true;
        cfoRemoving.value[cfo.id] = true;

        cfosStore.deleteCFO({
            cfo: cfo
        }).then(() => {
            updating.value = false;
            cfoRemoving.value[cfo.id] = false;
        }).catch(error => {
            updating.value = false;
            cfoRemoving.value[cfo.id] = false;

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
        snackbar.value?.showMessage('Unable to move CFO');
        return;
    }

    cfosStore.changeCFODisplayOrder({
        cfoId: moveEvent.element.id,
        from: moveEvent.oldIndex,
        to: moveEvent.newIndex
    }).then(() => {
        displayOrderModified.value = true;
    }).catch(error => {
        snackbar.value?.showError(error);
    });
}

cfosStore.loadAllCFOs({
    force: false
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
.cfos-table tr.cfos-table-row .hover-display {
    display: none;
}

.cfos-table tr.cfos-table-row:hover .hover-display {
    display: inline-grid;
}

.cfos-table tr:not(:last-child) > td > div {
    padding-bottom: 1px;
}

.cfos-table .has-bottom-border tr:last-child > td > div {
    padding-bottom: 1px;
}

.cfos-table .v-text-field .v-field__input {
    font-size: 0.875rem;
    padding-top: 0;
    color: rgba(var(--v-theme-on-surface));
}

.cfos-table .cfo-name {
    font-size: 0.875rem;
}

.cfos-table .cfo-comment {
    font-size: 0.8125rem;
}

.cfos-table tr .v-text-field .v-field__input {
    padding-bottom: 1px;
}

.cfos-table tr .v-input--density-compact .v-field__input {
    padding-bottom: 1px;
}
</style>
