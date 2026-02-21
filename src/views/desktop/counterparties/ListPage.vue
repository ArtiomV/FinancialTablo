<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="title-and-toolbar d-flex align-center">
                        <span>{{ tt('Counterparties') }}</span>
                        <v-btn class="ms-3" color="default" variant="outlined"
                               :disabled="loading || updating || hasEditingCounterparty" @click="add">{{ tt('Add') }}</v-btn>
                        <v-btn class="ms-3" color="primary" variant="tonal"
                               :disabled="loading || updating || hasEditingCounterparty" @click="saveSortResult"
                               v-if="displayOrderModified">{{ tt('Save Display Order') }}</v-btn>
                        <v-btn density="compact" color="default" variant="text" size="24"
                               class="ms-2" :icon="true" :disabled="loading || updating || hasEditingCounterparty"
                               :loading="loading" @click="reload">
                            <template #loader>
                                <v-progress-circular indeterminate size="20"/>
                            </template>
                            <v-icon :icon="mdiRefresh" size="24" />
                            <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                        </v-btn>
                        <v-spacer/>
                        <v-btn density="comfortable" color="default" variant="text" class="ms-2"
                               :disabled="loading || updating || hasEditingCounterparty" :icon="true">
                            <v-icon :icon="mdiDotsVertical" />
                            <v-menu activator="parent">
                                <v-list>
                                    <v-list-item :prepend-icon="mdiEyeOutline"
                                                 :title="tt('Show Hidden Counterparties')"
                                                 v-if="!showHidden" @click="showHidden = true"></v-list-item>
                                    <v-list-item :prepend-icon="mdiEyeOffOutline"
                                                 :title="tt('Hide Hidden Counterparties')"
                                                 v-if="showHidden" @click="showHidden = false"></v-list-item>
                                </v-list>
                            </v-menu>
                        </v-btn>
                    </div>
                </template>

                <v-table class="counterparties-table table-striped" :hover="!loading">
                    <thead>
                    <tr>
                        <th>
                            <div class="d-flex align-center">
                                <span>{{ tt('Counterparty Name') }}</span>
                                <span class="ms-4">{{ tt('Comment') }}</span>
                                <v-spacer/>
                                <span>{{ tt('Operation') }}</span>
                            </div>
                        </th>
                    </tr>
                    </thead>

                    <tbody v-if="loading && noAvailableCounterparty && !newCounterparty">
                    <tr :key="itemIdx" v-for="itemIdx in [ 1, 2, 3, 4, 5 ]">
                        <td class="px-0">
                            <v-skeleton-loader type="text" :loading="true"></v-skeleton-loader>
                        </td>
                    </tr>
                    </tbody>

                    <tbody v-if="!loading && noAvailableCounterparty && !newCounterparty">
                    <tr>
                        <td>{{ tt('No available counterparty') }}</td>
                    </tr>
                    </tbody>

                    <draggable-list tag="tbody"
                                    item-key="id"
                                    handle=".drag-handle"
                                    ghost-class="dragging-item"
                                    :class="{ 'has-bottom-border': newCounterparty }"
                                    :disabled="noAvailableCounterparty"
                                    v-model="counterparties"
                                    @change="onMove">
                        <template #item="{ element }">
                            <tr class="counterparties-table-row text-sm" v-if="(showHidden || !element.hidden) && displayedCounterpartyIds.has(element.id)">
                                <td>
                                    <div class="d-flex align-center">
                                        <!-- Display mode -->
                                        <div class="d-flex align-center" v-if="editingCounterparty.id !== element.id">
                                            <span class="counterparty-name" :class="{ 'text-medium-emphasis': element.hidden }">{{ element.name }}</span>
                                            <span class="counterparty-comment text-medium-emphasis ms-4" v-if="element.comment">{{ element.comment }}</span>
                                        </div>

                                        <!-- Edit mode -->
                                        <div class="d-flex align-center w-100 me-2" v-else-if="editingCounterparty.id === element.id">
                                            <v-text-field class="me-2" type="text"
                                                          density="compact" variant="underlined"
                                                          :disabled="loading || updating"
                                                          :placeholder="tt('Counterparty Name')"
                                                          v-model="editingCounterparty.name"
                                                          @keyup.enter="save(editingCounterparty)">
                                            </v-text-field>
                                            <v-select class="counterparty-type-select me-2"
                                                      density="compact" variant="underlined"
                                                      :disabled="loading || updating"
                                                      :items="counterpartyTypeOptions"
                                                      item-title="name"
                                                      item-value="type"
                                                      v-model="editingCounterparty.type">
                                            </v-select>
                                            <v-text-field class="me-2" type="text"
                                                          density="compact" variant="underlined"
                                                          :disabled="loading || updating"
                                                          :placeholder="tt('Comment')"
                                                          v-model="editingCounterparty.comment"
                                                          @keyup.enter="save(editingCounterparty)">
                                            </v-text-field>
                                        </div>

                                        <v-spacer/>

                                        <!-- Action buttons - display mode -->
                                        <v-btn class="px-2 ms-2" color="default"
                                               density="comfortable" variant="text"
                                               :class="{ 'd-none': loading, 'hover-display': !loading }"
                                               :prepend-icon="element.hidden ? mdiEyeOutline : mdiEyeOffOutline"
                                               :loading="counterpartyHiding[element.id]"
                                               :disabled="loading || updating"
                                               v-if="editingCounterparty.id !== element.id"
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
                                               :loading="counterpartyUpdating[element.id]"
                                               :disabled="loading || updating"
                                               v-if="editingCounterparty.id !== element.id"
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
                                               :loading="counterpartyRemoving[element.id]"
                                               :disabled="loading || updating"
                                               v-if="editingCounterparty.id !== element.id"
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
                                               :loading="counterpartyUpdating[element.id]"
                                               :disabled="loading || updating || !isCounterpartyModified(element)"
                                               v-if="editingCounterparty.id === element.id" @click="save(editingCounterparty)">
                                            <template #loader>
                                                <v-progress-circular indeterminate size="20" width="2"/>
                                            </template>
                                            {{ tt('Save') }}
                                        </v-btn>
                                        <v-btn class="px-2" color="default"
                                               density="comfortable" variant="text"
                                               :prepend-icon="mdiClose"
                                               :disabled="loading || updating"
                                               v-if="editingCounterparty.id === element.id" @click="cancelSave(editingCounterparty)">
                                            {{ tt('Cancel') }}
                                        </v-btn>

                                        <span class="ms-2">
                                            <v-icon :class="!loading && !updating && !hasEditingCounterparty && availableCounterpartyCount > 1 ? 'drag-handle' : 'disabled'"
                                                    :icon="mdiDrag"/>
                                            <v-tooltip activator="parent" v-if="!loading && !updating && !hasEditingCounterparty && availableCounterpartyCount > 1">{{ tt('Drag to Reorder') }}</v-tooltip>
                                        </span>
                                    </div>
                                </td>
                            </tr>
                        </template>
                    </draggable-list>

                    <tbody v-if="newCounterparty" ref="newCounterpartyRow">
                    <tr class="text-sm" :class="{ 'even-row': (availableCounterpartyCount & 1) === 1}">
                        <td>
                            <div class="d-flex align-center">
                                <v-text-field class="me-2" type="text" color="primary"
                                              density="compact" variant="underlined"
                                              :disabled="loading || updating"
                                              :placeholder="tt('Counterparty Name')"
                                              v-model="newCounterparty.name"
                                              @keyup.enter="save(newCounterparty)">
                                </v-text-field>
                                <v-select class="counterparty-type-select me-2"
                                          density="compact" variant="underlined"
                                          :disabled="loading || updating"
                                          :items="counterpartyTypeOptions"
                                          item-title="name"
                                          item-value="type"
                                          v-model="newCounterparty.type">
                                </v-select>
                                <v-text-field class="me-2" type="text" color="primary"
                                              density="compact" variant="underlined"
                                              :disabled="loading || updating"
                                              :placeholder="tt('Comment')"
                                              v-model="newCounterparty.comment"
                                              @keyup.enter="save(newCounterparty)">
                                </v-text-field>

                                <v-spacer/>

                                <v-btn class="px-2" density="comfortable" variant="text"
                                       :prepend-icon="mdiCheck"
                                       :loading="counterpartyUpdating['']"
                                       :disabled="loading || updating || !isCounterpartyModified(newCounterparty)"
                                       @click="save(newCounterparty)">
                                    <template #loader>
                                        <v-progress-circular indeterminate size="20" width="2"/>
                                    </template>
                                    {{ tt('Save') }}
                                </v-btn>
                                <v-btn class="px-2" color="default"
                                       density="comfortable" variant="text"
                                       :prepend-icon="mdiClose"
                                       :disabled="loading || updating"
                                       @click="cancelSave(newCounterparty)">
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
                <div class="d-flex justify-center my-4" v-if="hasMoreCounterparties">
                    <v-btn variant="tonal" color="primary" @click="displayLimit += 50">
                        {{ tt('Load More') }}
                    </v-btn>
                </div>
            </v-card>
        </v-col>
    </v-row>

    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';

import { ref, computed, useTemplateRef, nextTick } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

import { useCounterpartiesStore } from '@/stores/counterparty.ts';

import { Counterparty, CounterpartyType } from '@/models/counterparty.ts';

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

const counterpartiesStore = useCounterpartiesStore();

const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');
const snackbar = useTemplateRef<SnackBarType>('snackbar');
const newCounterpartyRow = useTemplateRef<HTMLElement>('newCounterpartyRow');

const loading = ref<boolean>(true);
const updating = ref<boolean>(false);
const showHidden = ref<boolean>(false);
const displayOrderModified = ref<boolean>(false);
const newCounterparty = ref<Counterparty | null>(null);
const editingCounterparty = ref<Counterparty>(Counterparty.createNew());
const counterpartyUpdating = ref<Record<string, boolean>>({});
const counterpartyHiding = ref<Record<string, boolean>>({});
const counterpartyRemoving = ref<Record<string, boolean>>({});

const counterpartyTypeOptions = computed(() => [
    { name: tt('Person'), type: CounterpartyType.Person },
    { name: tt('Company'), type: CounterpartyType.Company }
]);

const counterparties = computed<Counterparty[]>(() => counterpartiesStore.allCounterparties);
const displayLimit = ref<number>(50);

const displayedCounterpartyIds = computed<Set<string>>(() => {
    const ids = new Set<string>();
    let count = 0;
    for (const cp of counterparties.value) {
        if (showHidden.value || !cp.hidden) {
            count++;
            if (count <= displayLimit.value) {
                ids.add(cp.id);
            }
        }
    }
    return ids;
});

const hasMoreCounterparties = computed<boolean>(() => {
    let visibleCount = 0;
    for (const cp of counterparties.value) {
        if (showHidden.value || !cp.hidden) {
            visibleCount++;
        }
    }
    return visibleCount > displayLimit.value;
});

const noAvailableCounterparty = computed<boolean>(() => {
    if (!counterparties.value || counterparties.value.length < 1) {
        return true;
    }

    if (showHidden.value) {
        return false;
    }

    for (const counterparty of counterparties.value) {
        if (!counterparty.hidden) {
            return false;
        }
    }

    return true;
});

const availableCounterpartyCount = computed<number>(() => {
    if (!counterparties.value) {
        return 0;
    }

    if (showHidden.value) {
        return counterparties.value.length;
    }

    let count = 0;

    for (const counterparty of counterparties.value) {
        if (!counterparty.hidden) {
            count++;
        }
    }

    return count;
});

const hasEditingCounterparty = computed<boolean>(() => {
    return !!(newCounterparty.value || (editingCounterparty.value.id && editingCounterparty.value.id !== ''));
});

function isCounterpartyModified(counterparty: Counterparty): boolean {
    if (counterparty.id) {
        const original = counterpartiesStore.allCounterpartiesMap[counterparty.id];

        if (!original) {
            return false;
        }

        return editingCounterparty.value.name !== original.name
            || editingCounterparty.value.type !== original.type
            || editingCounterparty.value.comment !== original.comment;
    } else {
        return counterparty.name !== '';
    }
}

function reload(): void {
    if (hasEditingCounterparty.value) {
        return;
    }

    loading.value = true;

    counterpartiesStore.loadAllCounterparties({
        force: true
    }).then(() => {
        loading.value = false;
        displayOrderModified.value = false;

        snackbar.value?.showMessage('Counterparty list has been updated');
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
    newCounterparty.value = Counterparty.createNew();
    nextTick(() => {
        newCounterpartyRow.value?.scrollIntoView({ behavior: 'smooth', block: 'center' });
    });
}

function edit(counterparty: Counterparty): void {
    editingCounterparty.value = counterparty.clone();
}

function save(counterparty: Counterparty): void {
    updating.value = true;
    counterpartyUpdating.value[counterparty.id || ''] = true;

    counterpartiesStore.saveCounterparty({
        counterparty: counterparty
    }).then(() => {
        updating.value = false;
        counterpartyUpdating.value[counterparty.id || ''] = false;

        if (counterparty.id) {
            editingCounterparty.value = Counterparty.createNew();
        } else {
            newCounterparty.value = null;
        }
    }).catch(error => {
        updating.value = false;
        counterpartyUpdating.value[counterparty.id || ''] = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function cancelSave(counterparty: Counterparty): void {
    if (counterparty.id) {
        editingCounterparty.value = Counterparty.createNew();
    } else {
        newCounterparty.value = null;
    }
}

function saveSortResult(): void {
    if (!displayOrderModified.value) {
        return;
    }

    loading.value = true;

    counterpartiesStore.updateCounterpartyDisplayOrders().then(() => {
        loading.value = false;
        displayOrderModified.value = false;
    }).catch(error => {
        loading.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function hide(counterparty: Counterparty, hidden: boolean): void {
    updating.value = true;
    counterpartyHiding.value[counterparty.id] = true;

    counterpartiesStore.hideCounterparty({
        counterparty: counterparty,
        hidden: hidden
    }).then(() => {
        updating.value = false;
        counterpartyHiding.value[counterparty.id] = false;
    }).catch(error => {
        updating.value = false;
        counterpartyHiding.value[counterparty.id] = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function remove(counterparty: Counterparty): void {
    confirmDialog.value?.open('Are you sure you want to delete this counterparty?').then(() => {
        updating.value = true;
        counterpartyRemoving.value[counterparty.id] = true;

        counterpartiesStore.deleteCounterparty({
            counterparty: counterparty
        }).then(() => {
            updating.value = false;
            counterpartyRemoving.value[counterparty.id] = false;
        }).catch(error => {
            updating.value = false;
            counterpartyRemoving.value[counterparty.id] = false;

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
        snackbar.value?.showMessage('Unable to move counterparty');
        return;
    }

    counterpartiesStore.changeCounterpartyDisplayOrder({
        counterpartyId: moveEvent.element.id,
        from: moveEvent.oldIndex,
        to: moveEvent.newIndex
    }).then(() => {
        displayOrderModified.value = true;
    }).catch(error => {
        snackbar.value?.showError(error);
    });
}

counterpartiesStore.loadAllCounterparties({
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
.counterparties-table tr.counterparties-table-row .hover-display {
    display: none;
}

.counterparties-table tr.counterparties-table-row:hover .hover-display {
    display: inline-grid;
}

.counterparties-table tr:not(:last-child) > td > div {
    padding-bottom: 1px;
}

.counterparties-table .has-bottom-border tr:last-child > td > div {
    padding-bottom: 1px;
}

.counterparties-table tr.counterparties-table-row .right-bottom-icon .v-badge__badge {
    padding-bottom: 1px;
}

.counterparties-table .v-text-field .v-input__prepend {
    margin-inline-end: 0;
    color: rgba(var(--v-theme-on-surface));
}

.counterparties-table .v-text-field .v-input__prepend .v-badge > .v-badge__wrapper > .v-icon {
    opacity: var(--v-medium-emphasis-opacity);
}

.counterparties-table .v-text-field.v-input--plain-underlined .v-input__prepend {
    padding-top: 10px;
}

.counterparties-table .v-text-field .v-field__input {
    font-size: 0.875rem;
    padding-top: 0;
    color: rgba(var(--v-theme-on-surface));
}

.counterparties-table .counterparty-name {
    font-size: 0.875rem;
}

.counterparties-table .counterparty-comment {
    font-size: 0.8125rem;
}

.counterparties-table .counterparty-type-select {
    max-width: 140px;
}

.counterparties-table tr .v-text-field .v-field__input {
    padding-bottom: 1px;
}

.counterparties-table tr .v-input--density-compact .v-field__input {
    padding-bottom: 1px;
}
</style>
