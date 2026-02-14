<template>
    <v-menu v-model="isOpen" :close-on-content-click="false" location="bottom end">
        <template #activator="{ props: filterProps }">
            <v-btn variant="outlined" size="small" :prepend-icon="mdiFilterVariant"
                   v-bind="filterProps">
                {{ tt('Filters') }}
                <v-badge v-if="activeFilterCount > 0" :content="activeFilterCount"
                         color="primary" floating />
            </v-btn>
        </template>
        <v-card width="380" class="pa-4">
            <div class="text-subtitle-2 mb-2">{{ tt('Filters') }}</div>
            <div class="d-flex ga-2 mb-2">
                <v-text-field density="compact" hide-details type="number"
                              :label="tt('Min Amount')" v-model.number="localAmountMin" />
                <v-text-field density="compact" hide-details type="number"
                              :label="tt('Max Amount')" v-model.number="localAmountMax" />
            </div>
            <v-autocomplete density="compact" hide-details class="mb-2"
                            item-title="name" item-value="id" clearable
                            :label="tt('Transaction Categories')"
                            :items="categoryList"
                            :model-value="localCategoryId"
                            @update:model-value="localCategoryId = $event || ''" />
            <v-autocomplete density="compact" hide-details class="mb-2"
                            item-title="name" item-value="id" clearable
                            :label="tt('Account')"
                            :items="accountList"
                            :model-value="localAccountId"
                            @update:model-value="localAccountId = $event || ''" />
            <v-autocomplete density="compact" hide-details class="mb-2"
                            item-title="name" item-value="id" clearable
                            :label="tt('Counterparty')"
                            :items="counterpartyList"
                            :model-value="localCounterpartyId"
                            @update:model-value="localCounterpartyId = $event || ''" />
            <v-text-field density="compact" hide-details class="mb-2"
                          :prepend-inner-icon="mdiMagnify"
                          :placeholder="tt('Search transaction description')"
                          v-model="localKeyword"
                          @keyup.enter="applyFilters" />
            <div class="d-flex justify-end ga-2 mt-3">
                <v-btn variant="text" size="small" @click="clearAll">{{ tt('Clear All Filters') }}</v-btn>
                <v-btn color="primary" variant="tonal" size="small" @click="applyFilters">{{ tt('Apply') }}</v-btn>
            </div>
        </v-card>
    </v-menu>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from '@/locales/helpers.ts';
import { mdiFilterVariant, mdiMagnify } from '@mdi/js';

interface FilterItem {
    id: string;
    name: string;
}

interface FilterState {
    keyword: string;
    accountId: string;
    counterpartyId: string;
    categoryId: string;
    amountMin: number | null;
    amountMax: number | null;
}

interface TransactionFilterPanelProps {
    categoryList: FilterItem[];
    accountList: FilterItem[];
    counterpartyList: FilterItem[];
    currentFilters: FilterState;
}

const props = defineProps<TransactionFilterPanelProps>();

const emit = defineEmits<{
    apply: [filters: FilterState];
    clear: [];
}>();

const { tt } = useI18n();

const isOpen = ref<boolean>(false);
const localKeyword = ref<string>(props.currentFilters.keyword);
const localAccountId = ref<string>(props.currentFilters.accountId);
const localCounterpartyId = ref<string>(props.currentFilters.counterpartyId);
const localCategoryId = ref<string>(props.currentFilters.categoryId);
const localAmountMin = ref<number | null>(props.currentFilters.amountMin);
const localAmountMax = ref<number | null>(props.currentFilters.amountMax);

const activeFilterCount = computed<number>(() => {
    let count = 0;
    if (localKeyword.value) count++;
    if (localAccountId.value) count++;
    if (localCounterpartyId.value) count++;
    if (localCategoryId.value) count++;
    if (localAmountMin.value !== null || localAmountMax.value !== null) count++;
    return count;
});

function applyFilters(): void {
    isOpen.value = false;
    emit('apply', {
        keyword: localKeyword.value,
        accountId: localAccountId.value,
        counterpartyId: localCounterpartyId.value,
        categoryId: localCategoryId.value,
        amountMin: localAmountMin.value,
        amountMax: localAmountMax.value,
    });
}

function clearAll(): void {
    localKeyword.value = '';
    localAccountId.value = '';
    localCounterpartyId.value = '';
    localCategoryId.value = '';
    localAmountMin.value = null;
    localAmountMax.value = null;
    isOpen.value = false;
    emit('clear');
}
</script>
