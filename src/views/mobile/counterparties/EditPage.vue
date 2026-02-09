<template>
    <f7-page @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :class="{ 'disabled': loading }" :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title :title="tt(title)"></f7-nav-title>
            <f7-nav-right :class="{ 'disabled': loading }">
                <f7-link icon-f7="checkmark_alt" :class="{ 'disabled': inputIsEmpty || submitting }" @click="save"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-list strong inset dividers class="margin-top skeleton-text" v-if="loading">
            <f7-list-input label="Counterparty Name" placeholder="Your counterparty name"></f7-list-input>
            <f7-list-item class="list-item-with-header-and-title" header="Counterparty Type" title="Person"></f7-list-item>
            <f7-list-input label="Description" type="textarea" placeholder="Your counterparty description (optional)"></f7-list-input>
        </f7-list>

        <f7-list form strong inset dividers class="margin-top" v-else-if="!loading">
            <f7-list-input
                type="text"
                clear-button
                :label="tt('Counterparty Name')"
                :placeholder="tt('Your counterparty name')"
                v-model:value="counterparty.name"
            ></f7-list-input>

            <f7-list-item
                link="#" no-chevron
                class="list-item-with-header-and-title"
                :header="tt('Counterparty Type')"
                :title="getCounterpartyTypeName(counterparty.type)"
                @click="showTypeSheet = true"
            >
                <list-item-selection-sheet value-type="item"
                                           key-field="type" value-field="type" title-field="displayName"
                                           :items="allCounterpartyTypes"
                                           v-model:show="showTypeSheet"
                                           v-model="counterparty.type">
                </list-item-selection-sheet>
            </f7-list-item>

            <f7-list-input
                type="textarea"
                style="height: auto"
                :label="tt('Description')"
                :placeholder="tt('Your counterparty description (optional)')"
                v-textarea-auto-size
                v-model:value="counterparty.comment"
            ></f7-list-input>
        </f7-list>

        <f7-list strong inset dividers class="margin-top" v-if="!loading && editCounterpartyId">
            <f7-list-button :title="tt('Delete')" color="red" @click="remove(false)"></f7-list-button>
        </f7-list>

        <f7-actions close-by-outside-click close-on-escape :opened="showDeleteActionSheet" @actions:closed="showDeleteActionSheet = false">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Are you sure you want to delete this counterparty?') }}</f7-actions-label>
                <f7-actions-button color="red" @click="remove(true)">{{ tt('Delete') }}</f7-actions-button>
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
import { useI18nUIComponents, showLoading, hideLoading } from '@/lib/ui/mobile.ts';

import { useCounterpartiesStore } from '@/stores/counterparty.ts';

import { Counterparty, CounterpartyType } from '@/models/counterparty.ts';

interface CounterpartyTypeOption {
    readonly type: CounterpartyType;
    readonly displayName: string;
}

const props = defineProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>();

const { tt } = useI18n();
const { showAlert, showToast, routeBackOnError } = useI18nUIComponents();

const counterpartiesStore = useCounterpartiesStore();

const editCounterpartyId = ref<string>('');
const loading = ref<boolean>(true);
const submitting = ref<boolean>(false);
const loadingError = ref<unknown | null>(null);
const counterparty = ref<Counterparty>(Counterparty.createNew());
const showTypeSheet = ref<boolean>(false);
const showDeleteActionSheet = ref<boolean>(false);

const title = computed<string>(() => {
    if (editCounterpartyId.value) {
        return 'Edit Counterparty';
    } else {
        return 'Add Counterparty';
    }
});

const inputIsEmpty = computed<boolean>(() => {
    return !counterparty.value.name;
});

const allCounterpartyTypes = computed<CounterpartyTypeOption[]>(() => {
    return [
        { type: CounterpartyType.Person, displayName: tt('Person') },
        { type: CounterpartyType.Company, displayName: tt('Company') }
    ];
});

function getCounterpartyTypeName(type: CounterpartyType): string {
    for (const option of allCounterpartyTypes.value) {
        if (option.type === type) {
            return option.displayName;
        }
    }
    return '';
}

function init(): void {
    const query = props.f7route.query;

    if (query['id']) {
        loading.value = true;
        editCounterpartyId.value = query['id'];

        counterpartiesStore.loadAllCounterparties({
            force: false
        }).then(() => {
            const existingCounterparty = counterpartiesStore.allCounterpartiesMap[editCounterpartyId.value];

            if (existingCounterparty) {
                counterparty.value = existingCounterparty.clone();
            } else {
                showToast('This counterparty does not exist');
                loadingError.value = 'This counterparty does not exist';
            }

            loading.value = false;
        }).catch(error => {
            if (error.processed) {
                loading.value = false;
            } else {
                loadingError.value = error;
                showToast(error.message || error);
            }
        });
    } else {
        loading.value = false;
    }
}

function save(): void {
    const router = props.f7router;

    if (!counterparty.value.name) {
        showAlert('Counterparty name cannot be empty');
        return;
    }

    submitting.value = true;
    showLoading(() => submitting.value);

    counterpartiesStore.saveCounterparty({
        counterparty: counterparty.value
    }).then(() => {
        submitting.value = false;
        hideLoading();

        if (!editCounterpartyId.value) {
            showToast('You have added a new counterparty');
        } else {
            showToast('You have saved this counterparty');
        }

        router.back();
    }).catch(error => {
        submitting.value = false;
        hideLoading();

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function remove(confirm: boolean): void {
    if (!confirm) {
        showDeleteActionSheet.value = true;
        return;
    }

    showDeleteActionSheet.value = false;
    showLoading();

    counterpartiesStore.deleteCounterparty({
        counterparty: counterparty.value
    }).then(() => {
        hideLoading();
        showToast('You have deleted this counterparty');
        props.f7router.back();
    }).catch(error => {
        hideLoading();

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function onPageAfterIn(): void {
    routeBackOnError(props.f7router, loadingError);
}

init();
</script>
