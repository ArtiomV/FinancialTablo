<template>
    <f7-page :ptr="!sortable" @ptr:refresh="reload" @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :class="{ 'disabled': loading }" :back-link="tt('Back')" v-if="!sortable"></f7-nav-left>
            <f7-nav-left v-else-if="sortable">
                <f7-link icon-f7="xmark" :class="{ 'disabled': displayOrderSaving }" @click="cancelSort"></f7-link>
            </f7-nav-left>
            <f7-nav-title :title="tt('Counterparties')"></f7-nav-title>
            <f7-nav-right :class="{ 'navbar-compact-icons': true, 'disabled': loading }">
                <f7-link icon-f7="ellipsis" :class="{ 'disabled': sortable }" @click="showMoreActionSheet = true"></f7-link>
                <f7-link icon-f7="plus" href="/counterparty/add" v-if="!sortable"></f7-link>
                <f7-link icon-f7="checkmark_alt" :class="{ 'disabled': displayOrderSaving || !displayOrderModified }" @click="saveSortResult" v-else-if="sortable"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-list strong inset dividers class="counterparty-item-list margin-top skeleton-text" v-if="loading">
            <f7-list-item :key="itemIdx" v-for="itemIdx in [ 1, 2, 3 ]">
                <template #media>
                    <f7-icon f7="person_fill"></f7-icon>
                </template>
                <template #title>
                    <div class="display-flex">
                        <div class="counterparty-list-item-content list-item-valign-middle padding-inline-start-half">Counterparty Name</div>
                    </div>
                </template>
            </f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="counterparty-item-list margin-top" v-if="!loading && noAvailableCounterparty">
            <f7-list-item :title="tt('No available counterparty')"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers sortable class="counterparty-item-list margin-top"
                 :sortable-enabled="sortable" @sortable:sort="onSort"
                 v-if="!loading && allCounterparties.length > 0">
            <f7-list-item swipeout
                          :class="{ 'actual-first-child': counterparty.id === firstShowingId, 'actual-last-child': counterparty.id === lastShowingId }"
                          :id="getCounterpartyDomId(counterparty)"
                          :key="counterparty.id"
                          v-for="counterparty in allCounterparties"
                          v-show="showHidden || !counterparty.hidden"
                          @taphold="setSortable()">
                <template #media>
                    <f7-icon :f7="counterparty.type === CounterpartyType.Company ? 'building_2_fill' : 'person_fill'">
                        <f7-badge color="gray" class="right-bottom-icon" v-if="counterparty.hidden">
                            <f7-icon f7="eye_slash_fill"></f7-icon>
                        </f7-badge>
                    </f7-icon>
                </template>
                <template #title>
                    <div class="display-flex">
                        <div class="counterparty-list-item-content list-item-valign-middle padding-inline-start-half">
                            {{ counterparty.name }}
                        </div>
                    </div>
                </template>
                <f7-swipeout-actions :left="textDirection === TextDirection.LTR"
                                     :right="textDirection === TextDirection.RTL"
                                     v-if="sortable">
                    <f7-swipeout-button :color="counterparty.hidden ? 'blue' : 'gray'" class="padding-horizontal"
                                        overswipe close @click="hide(counterparty, !counterparty.hidden)">
                        <f7-icon :f7="counterparty.hidden ? 'eye' : 'eye_slash'"></f7-icon>
                    </f7-swipeout-button>
                </f7-swipeout-actions>
                <f7-swipeout-actions :left="textDirection === TextDirection.RTL"
                                     :right="textDirection === TextDirection.LTR"
                                     v-if="!sortable">
                    <f7-swipeout-button color="orange" close :text="tt('Edit')" @click="edit(counterparty)"></f7-swipeout-button>
                    <f7-swipeout-button color="red" class="padding-horizontal" @click="remove(counterparty, false)">
                        <f7-icon f7="trash"></f7-icon>
                    </f7-swipeout-button>
                </f7-swipeout-actions>
            </f7-list-item>
        </f7-list>

        <f7-actions close-by-outside-click close-on-escape :opened="showMoreActionSheet" @actions:closed="showMoreActionSheet = false">
            <f7-actions-group>
                <f7-actions-button @click="setSortable()">{{ tt('Sort') }}</f7-actions-button>
                <f7-actions-button v-if="!showHidden" @click="showHidden = true">{{ tt('Show Hidden Counterparties') }}</f7-actions-button>
                <f7-actions-button v-if="showHidden" @click="showHidden = false">{{ tt('Hide Hidden Counterparties') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>

        <f7-actions close-by-outside-click close-on-escape :opened="showDeleteActionSheet" @actions:closed="showDeleteActionSheet = false">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Are you sure you want to delete this counterparty?') }}</f7-actions-label>
                <f7-actions-button color="red" @click="remove(counterpartyToDelete, true)">{{ tt('Delete') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>
    </f7-page>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import type { Router } from 'framework7/types';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents, showLoading, hideLoading, onSwipeoutDeleted } from '@/lib/ui/mobile.ts';

import { useCounterpartiesStore } from '@/stores/counterparty.ts';

import { TextDirection } from '@/core/text.ts';
import { Counterparty, CounterpartyType } from '@/models/counterparty.ts';

const props = defineProps<{
    f7router: Router.Router;
}>();

const { tt, getCurrentLanguageTextDirection } = useI18n();
const { showAlert, showToast, routeBackOnError } = useI18nUIComponents();

const counterpartiesStore = useCounterpartiesStore();

const loading = ref<boolean>(true);
const loadingError = ref<unknown | null>(null);
const sortable = ref<boolean>(false);
const showHidden = ref<boolean>(false);
const showMoreActionSheet = ref<boolean>(false);
const showDeleteActionSheet = ref<boolean>(false);
const counterpartyToDelete = ref<Counterparty | null>(null);
const displayOrderModified = ref<boolean>(false);
const displayOrderSaving = ref<boolean>(false);

const textDirection = computed<TextDirection>(() => getCurrentLanguageTextDirection());

const allCounterparties = computed<Counterparty[]>(() => counterpartiesStore.allCounterparties);

const noAvailableCounterparty = computed<boolean>(() => {
    if (showHidden.value) {
        return counterpartiesStore.allAvailableCounterpartiesCount < 1;
    } else {
        return counterpartiesStore.allVisibleCounterparties.length < 1;
    }
});

const firstShowingId = computed<string | null>(() => {
    for (const counterparty of allCounterparties.value) {
        if (showHidden.value || !counterparty.hidden) {
            return counterparty.id;
        }
    }
    return null;
});

const lastShowingId = computed<string | null>(() => {
    for (let i = allCounterparties.value.length - 1; i >= 0; i--) {
        const counterparty = allCounterparties.value[i]!;
        if (showHidden.value || !counterparty.hidden) {
            return counterparty.id;
        }
    }
    return null;
});

function getCounterpartyDomId(counterparty: Counterparty): string {
    return 'counterparty_' + counterparty.id;
}

function parseCounterpartyIdFromDomId(domId: string): string | null {
    if (!domId || domId.indexOf('counterparty_') !== 0) {
        return null;
    }

    return domId.substring(13); // counterparty_
}

function init(): void {
    loading.value = true;

    counterpartiesStore.loadAllCounterparties({
        force: false
    }).then(() => {
        loading.value = false;
    }).catch(error => {
        if (error.processed) {
            loading.value = false;
        } else {
            loadingError.value = error;
            showToast(error.message || error);
        }
    });
}

function reload(done?: () => void): void {
    if (sortable.value) {
        done?.();
        return;
    }

    const force = !!done;

    counterpartiesStore.loadAllCounterparties({
        force: force
    }).then(() => {
        done?.();

        if (force) {
            showToast('Counterparty list has been updated');
        }
    }).catch(error => {
        done?.();

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function edit(counterparty: Counterparty): void {
    props.f7router.navigate('/counterparty/edit?id=' + counterparty.id);
}

function hide(counterparty: Counterparty, hidden: boolean): void {
    showLoading();

    counterpartiesStore.hideCounterparty({
        counterparty: counterparty,
        hidden: hidden
    }).then(() => {
        hideLoading();
    }).catch(error => {
        hideLoading();

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function remove(counterparty: Counterparty | null, confirm: boolean): void {
    if (!counterparty) {
        showAlert('An error occurred');
        return;
    }

    if (!confirm) {
        counterpartyToDelete.value = counterparty;
        showDeleteActionSheet.value = true;
        return;
    }

    showDeleteActionSheet.value = false;
    counterpartyToDelete.value = null;
    showLoading();

    counterpartiesStore.deleteCounterparty({
        counterparty: counterparty,
        beforeResolve: (done) => {
            onSwipeoutDeleted(getCounterpartyDomId(counterparty), done);
        }
    }).then(() => {
        hideLoading();
    }).catch(error => {
        hideLoading();

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function setSortable(): void {
    if (sortable.value) {
        return;
    }

    showHidden.value = true;
    sortable.value = true;
    displayOrderModified.value = false;
}

function saveSortResult(): void {
    if (!displayOrderModified.value) {
        showHidden.value = false;
        sortable.value = false;
        return;
    }

    displayOrderSaving.value = true;
    showLoading();

    counterpartiesStore.updateCounterpartyDisplayOrders().then(() => {
        displayOrderSaving.value = false;
        hideLoading();

        showHidden.value = false;
        sortable.value = false;
        displayOrderModified.value = false;
    }).catch(error => {
        displayOrderSaving.value = false;
        hideLoading();

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function cancelSort(): void {
    if (!displayOrderModified.value) {
        showHidden.value = false;
        sortable.value = false;
        return;
    }

    displayOrderSaving.value = true;
    showLoading();

    counterpartiesStore.loadAllCounterparties({
        force: false
    }).then(() => {
        displayOrderSaving.value = false;
        hideLoading();

        showHidden.value = false;
        sortable.value = false;
        displayOrderModified.value = false;
    }).catch(error => {
        displayOrderSaving.value = false;
        hideLoading();

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function onSort(event: { el: { id: string }, from: number, to: number }): void {
    if (!event || !event.el || !event.el.id) {
        showToast('Unable to move counterparty');
        return;
    }

    const id = parseCounterpartyIdFromDomId(event.el.id);

    if (!id) {
        showToast('Unable to move counterparty');
        return;
    }

    counterpartiesStore.changeCounterpartyDisplayOrder({
        counterpartyId: id,
        from: event.from,
        to: event.to
    }).then(() => {
        displayOrderModified.value = true;
    }).catch(error => {
        showToast(error.message || error);
    });
}

function onPageAfterIn(): void {
    if (counterpartiesStore.counterpartyListStateInvalid && !loading.value) {
        reload();
    }

    routeBackOnError(props.f7router, loadingError);
}

init();
</script>

<style>
.counterparty-item-list.list .item-media + .item-inner {
    margin-inline-start: 5px;
}

.counterparty-list-item-content {
    overflow: hidden;
    text-overflow: ellipsis;
}
</style>
