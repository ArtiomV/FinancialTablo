<template>
    <v-dialog width="800" :persistent="isCategoryModified" v-model="showState">
        <v-card class="pa-sm-1 pa-md-2">
            <template #title>
                <div class="d-flex align-center">
                    <h4 class="text-h4">{{ tt(title) }}</h4>
                    <v-progress-circular indeterminate size="22" class="ms-2" v-if="loading"></v-progress-circular>
                </div>
            </template>
            <v-card-text class="d-flex flex-column flex-md-row flex-grow-1 overflow-y-auto">
                <v-form class="w-100 mt-2">
                    <v-row>
                        <v-col cols="12" md="12">
                            <v-text-field
                                type="text"
                                persistent-placeholder
                                :disabled="loading || submitting"
                                :label="tt('Category Name')"
                                :placeholder="tt('Category Name')"
                                v-model="category.name"
                            />
                        </v-col>
                        <v-col cols="12" md="12">
                            <v-textarea
                                type="text"
                                persistent-placeholder
                                rows="3"
                                :disabled="loading || submitting"
                                :label="tt('Description')"
                                :placeholder="tt('Your category description (optional)')"
                                v-model="category.comment"
                            />
                        </v-col>
                        <v-col cols="12" md="6">
                            <v-select
                                persistent-placeholder
                                :disabled="loading || submitting"
                                :label="tt('Activity Type')"
                                :items="activityTypeOptions"
                                item-title="label"
                                item-value="value"
                                v-model="category.activityType"
                            />
                        </v-col>
                        <v-col cols="12" md="6" v-if="category.type === CategoryType.Expense">
                            <v-select
                                persistent-placeholder
                                :disabled="loading || submitting"
                                :label="tt('Cost Type')"
                                :items="costTypeOptions"
                                item-title="label"
                                item-value="value"
                                v-model="category.costType"
                            />
                        </v-col>
                        <v-col class="py-0" cols="12" md="12" v-if="editCategoryId">
                            <v-switch :disabled="loading || submitting"
                                      :label="tt('Visible')" v-model="category.visible"/>
                        </v-col>
                    </v-row>
                </v-form>
            </v-card-text>
            <v-card-text>
                <div class="w-100 d-flex justify-center flex-wrap mt-sm-1 mt-md-2 gap-4">
                    <v-tooltip :disabled="!inputIsEmpty" :text="inputEmptyProblemMessage ? tt(inputEmptyProblemMessage) : ''">
                        <template v-slot:activator="{ props }">
                            <div v-bind="props" class="d-inline-block">
                                <v-btn :disabled="inputIsEmpty || loading || submitting" @click="save">
                                    {{ tt(saveButtonTitle) }}
                                    <v-progress-circular indeterminate size="22" class="ms-2" v-if="submitting"></v-progress-circular>
                                </v-btn>
                            </div>
                        </template>
                    </v-tooltip>
                    <v-btn color="secondary" variant="tonal"
                           :disabled="loading || submitting" @click="cancel">{{ tt('Cancel') }}</v-btn>
                </div>
            </v-card-text>
        </v-card>
    </v-dialog>

    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import SnackBar from '@/components/desktop/SnackBar.vue';

import { ref, computed, useTemplateRef, type ComputedRef } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { useCategoryEditPageBase } from '@/views/base/categories/CategoryEditPageBase.ts';

import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';

import type { ColorValue } from '@/core/color.ts';
import { CategoryType } from '@/core/category.ts';
// @ts-ignore
import { ALL_CATEGORY_ICONS } from '@/consts/icon.ts';
// @ts-ignore
import { ALL_CATEGORY_COLORS } from '@/consts/color.ts';
import { TransactionCategory } from '@/models/transaction_category.ts';

import { generateRandomUUID } from '@/lib/misc.ts';

interface TransactionCategoryEditResponse {
    message: string;
}

type SnackBarType = InstanceType<typeof SnackBar>;

const { tt } = useI18n();
const {
    editCategoryId,
    clientSessionId,
    loading,
    submitting,
    category,
    title,
    saveButtonTitle,
    inputEmptyProblemMessage,
    inputIsEmpty
} = useCategoryEditPageBase();

const transactionCategoriesStore = useTransactionCategoriesStore();

const snackbar = useTemplateRef<SnackBarType>('snackbar');

const showState = ref<boolean>(false);

let resolveFunc: ((value: TransactionCategoryEditResponse) => void) | null = null;
let rejectFunc: ((reason?: unknown) => void) | null = null;

const isCategoryModified = computed<boolean>(() => {
    if (!editCategoryId.value) { // Add
        return !category.value.equals(TransactionCategory.createNewCategory(category.value.type));
    } else { // Edit
        return true;
    }
});

const activityTypeOptions: ComputedRef<{ label: string; value: number }[]> = computed(() => [
    { label: tt('Operating'), value: 1 },
    { label: tt('Investing'), value: 2 },
    { label: tt('Financing'), value: 3 }
]);

const costTypeOptions: ComputedRef<{ label: string; value: number }[]> = computed(() => [
    { label: tt('None'), value: 0 },
    { label: tt('Cost of Goods Sold'), value: 1 },
    { label: tt('Operational'), value: 2 },
    { label: tt('Financial'), value: 3 }
]);

function open(options: { id?: string; type?: CategoryType; currentCategory?: TransactionCategory, color?: ColorValue, icon?: string }): Promise<TransactionCategoryEditResponse> {
    showState.value = true;
    loading.value = true;
    submitting.value = false;

    const newTransactionCategory = TransactionCategory.createNewCategory();
    category.value.fillFrom(newTransactionCategory);

    if (options.id) {
        if (options.currentCategory) {
            category.value.fillFrom(options.currentCategory);
        }

        editCategoryId.value = options.id;
        transactionCategoriesStore.getCategory({
            categoryId: editCategoryId.value
        }).then(response => {
            category.value.fillFrom(response);
            loading.value = false;
        }).catch(error => {
            loading.value = false;
            showState.value = false;

            if (!error.processed) {
                if (rejectFunc) {
                    rejectFunc(error);
                }
            }
        });
    } else if (options.type) {
        editCategoryId.value = null;

        const categoryType = options.type;

        if (categoryType !== CategoryType.Income &&
            categoryType !== CategoryType.Expense &&
            categoryType !== CategoryType.Transfer) {
            loading.value = false;
            showState.value = false;

            return Promise.reject('Parameter Invalid');
        }

        category.value.type = categoryType;

        if (options.color) {
            category.value.color = options.color;
        }

        if (options.icon) {
            category.value.icon = options.icon;
        }

        clientSessionId.value = generateRandomUUID();
        loading.value = false;
    }

    return new Promise((resolve, reject) => {
        resolveFunc = resolve;
        rejectFunc = reject;
    });
}

function save(): void {
    const problemMessage = inputEmptyProblemMessage.value;

    if (problemMessage) {
        snackbar.value?.showMessage(problemMessage);
        return;
    }

    submitting.value = true;

    transactionCategoriesStore.saveCategory({
        category: category.value,
        isEdit: !!editCategoryId.value,
        clientSessionId: clientSessionId.value
    }).then(() => {
        submitting.value = false;

        let message = 'You have saved this category';

        if (!editCategoryId.value) {
            message = 'You have added a new category';
        }

        resolveFunc?.({ message });
        showState.value = false;
    }).catch(error => {
        submitting.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function cancel(): void {
    rejectFunc?.();
    showState.value = false;
}

defineExpose({
    open
});
</script>
