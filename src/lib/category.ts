import { reversed, keys, values } from '@/core/base.ts';
import { type LocalizedPresetCategory, CategoryType } from '@/core/category.ts';
import { TransactionType } from '@/core/transaction.ts';
import {
    type TransactionCategoryCreateRequest,
    TransactionCategory
} from '@/models/transaction_category.ts';

export function transactionTypeToCategoryType(transactionType: TransactionType): CategoryType | null {
    if (transactionType === TransactionType.Income) {
        return CategoryType.Income;
    } else if (transactionType === TransactionType.Expense) {
        return CategoryType.Expense;
    } else if (transactionType === TransactionType.Transfer) {
        return CategoryType.Transfer;
    } else {
        return null;
    }
}

export function categoryTypeToTransactionType(categoryType: CategoryType): TransactionType | null {
    if (categoryType === CategoryType.Income) {
        return TransactionType.Income;
    } else if (categoryType === CategoryType.Expense) {
        return TransactionType.Expense;
    } else if (categoryType === CategoryType.Transfer) {
        return TransactionType.Transfer;
    } else {
        return null;
    }
}

export function localizedPresetCategoriesToTransactionCategoryCreateRequests(presetCategories: LocalizedPresetCategory[]): TransactionCategoryCreateRequest[] {
    const categories: TransactionCategoryCreateRequest[] = [];

    for (const presetCategory of presetCategories) {
        const category: TransactionCategoryCreateRequest = {
            name: presetCategory.name,
            type: presetCategory.type,
            icon: presetCategory.icon,
            color: presetCategory.color,
            comment: '',
            clientSessionId: ''
        };

        categories.push(category);
    }

    return categories;
}

export function getCategoryNameById(categoryId: string | null | undefined, allCategories?: TransactionCategory[]): string {
    if (!allCategories || !categoryId) {
        return '';
    }

    for (const category of allCategories) {
        if (category.id === categoryId) {
            return category.name;
        }
    }

    return '';
}

export function filterTransactionCategories(allTransactionCategories: Record<number, TransactionCategory[]>, allowCategoryTypes?: Record<number, boolean>, allowCategoryName?: string, showHidden?: boolean): Record<string, TransactionCategory[]> {
    const ret: Record<string, TransactionCategory[]> = {};
    const hasAllowCategoryTypes = allowCategoryTypes
        && (allowCategoryTypes[CategoryType.Income]
            || allowCategoryTypes[CategoryType.Expense]
            || allowCategoryTypes[CategoryType.Transfer]);

    const allCategoryTypes = [ CategoryType.Income, CategoryType.Expense, CategoryType.Transfer ];
    const lowercaseFilterContent = allowCategoryName ? allowCategoryName.toLowerCase() : '';

    for (const categoryType of allCategoryTypes) {
        const allCategories = allTransactionCategories[categoryType];

        if (!allCategories || allCategories.length < 1) {
            continue;
        }

        if (hasAllowCategoryTypes && !allowCategoryTypes[categoryType]) {
            continue;
        }

        const allFilteredCategories: TransactionCategory[] = [];

        for (const category of allCategories) {
            if (!showHidden && category.hidden) {
                continue;
            }

            if (lowercaseFilterContent && !category.name.toLowerCase().includes(lowercaseFilterContent)) {
                continue;
            }

            allFilteredCategories.push(category);
        }

        ret[`${categoryType}`] = allFilteredCategories;
    }

    return ret;
}

export function allVisibleTransactionCategoriesByType(allTransactionCategories: Record<number, TransactionCategory[]>, categoryType: number): TransactionCategory[] {
    const allCategories = allTransactionCategories[categoryType];
    const visibleCategories: TransactionCategory[] = [];

    if (!allCategories) {
        return visibleCategories;
    }

    for (const category of allCategories) {
        if (category.hidden) {
            continue;
        }

        visibleCategories.push(category);
    }

    return visibleCategories;
}

export function getFinalCategoryIdsByFilteredCategoryIds(allTransactionCategoriesMap: Record<number, TransactionCategory>, filteredCategoryIds: Record<string, boolean>): string {
    let finalCategoryIds = '';

    if (!allTransactionCategoriesMap) {
        return finalCategoryIds;
    }

    for (const category of values(allTransactionCategoriesMap)) {
        if (filteredCategoryIds && filteredCategoryIds[category.id]) {
            continue;
        }

        if (finalCategoryIds.length > 0) {
            finalCategoryIds += ',';
        }

        finalCategoryIds += category.id;
    }

    return finalCategoryIds;
}

export function isCategoryIdAvailable(categories: TransactionCategory[], categoryId: string): boolean {
    if (!categories || !categories.length) {
        return false;
    }

    for (const category of categories) {
        if (category.hidden) {
            continue;
        }

        if (category.id === categoryId) {
            return true;
        }
    }

    return false;
}

export function getFirstVisibleCategoryId(categories?: TransactionCategory[]): string {
    if (!categories || !categories.length) {
        return '';
    }

    for (const category of categories) {
        if (category.hidden) {
            continue;
        }

        return category.id;
    }

    return '';
}

export function isNoAvailableCategory(categories: TransactionCategory[], showHidden: boolean): boolean {
    for (const category of categories) {
        if (showHidden || !category.hidden) {
            return false;
        }
    }

    return true;
}

export function getAvailableCategoryCount(categories: TransactionCategory[], showHidden: boolean): number {
    let count = 0;

    for (const category of categories) {
        if (showHidden || !category.hidden) {
            count++;
        }
    }

    return count;
}

export function getFirstShowingId(categories: TransactionCategory[], showHidden: boolean): string | null {
    for (const category of categories) {
        if (showHidden || !category.hidden) {
            return category.id;
        }
    }

    return null;
}

export function getLastShowingId(categories: TransactionCategory[], showHidden: boolean): string | null {
    for (const category of reversed(categories)) {
        if (showHidden || !category.hidden) {
            return category.id;
        }
    }

    return null;
}

export function getCategoryMapByName(allCategories: TransactionCategory[]): Record<string, TransactionCategory> {
    const ret: Record<string, TransactionCategory> = {};

    if (!allCategories) {
        return ret;
    }

    for (const category of allCategories) {
        ret[category.name] = category;
    }

    return ret;
}

export function selectAll(filterCategoryIds: Record<string, boolean>, allTransactionCategoriesMap: Record<string, TransactionCategory>): void {
    for (const categoryId of keys(filterCategoryIds)) {
        const category = allTransactionCategoriesMap[categoryId];

        if (category) {
            filterCategoryIds[category.id] = false;
        }
    }
}

export function selectNone(filterCategoryIds: Record<string, boolean>, allTransactionCategoriesMap: Record<string, TransactionCategory>): void {
    for (const categoryId of keys(filterCategoryIds)) {
        const category = allTransactionCategoriesMap[categoryId];

        if (category) {
            filterCategoryIds[category.id] = true;
        }
    }
}

export function selectInvert(filterCategoryIds: Record<string, boolean>, allTransactionCategoriesMap: Record<string, TransactionCategory>): void {
    for (const categoryId of keys(filterCategoryIds)) {
        const category = allTransactionCategoriesMap[categoryId];

        if (category) {
            filterCategoryIds[category.id] = !filterCategoryIds[category.id];
        }
    }
}
