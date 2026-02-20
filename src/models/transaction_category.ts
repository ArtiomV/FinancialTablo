import type { ColorValue } from '@/core/color.ts';
import { CategoryType } from '@/core/category.ts';
import { DEFAULT_CATEGORY_ICON_ID } from '@/consts/icon.ts';
import { DEFAULT_CATEGORY_COLOR } from '@/consts/color.ts';

export class TransactionCategory implements TransactionCategoryInfoResponse {
    public id: string;
    public name: string;
    public type: CategoryType;
    public icon: string;
    public color: ColorValue;
    public comment: string;
    public displayOrder: number;
    public visible: boolean;
    public activityType: number;
    public costType: number;

    private constructor(id: string, name: string, type: CategoryType, icon: string, color: ColorValue, comment: string, displayOrder: number, visible: boolean, activityType: number = 0, costType: number = 0) {
        this.id = id;
        this.name = name;
        this.type = type;
        this.icon = icon;
        this.color = color;
        this.comment = comment;
        this.displayOrder = displayOrder;
        this.visible = visible;
        this.activityType = activityType;
        this.costType = costType;
    }

    public get hidden(): boolean {
        return !this.visible;
    }

    public equals(other: TransactionCategory): boolean {
        return this.id === other.id &&
            this.name === other.name &&
            this.type === other.type &&
            this.icon === other.icon &&
            this.color === other.color &&
            this.comment === other.comment &&
            this.displayOrder === other.displayOrder &&
            this.visible === other.visible &&
            this.activityType === other.activityType &&
            this.costType === other.costType;
    }

    public fillFrom(other: TransactionCategory): void {
        this.id = other.id;
        this.name = other.name;
        this.type = other.type;
        this.icon = other.icon;
        this.color = other.color;
        this.comment = other.comment;
        this.visible = other.visible;
        this.activityType = other.activityType;
        this.costType = other.costType;
    }

    public clone(): TransactionCategory {
        return new TransactionCategory(
            this.id,
            this.name,
            this.type,
            this.icon,
            this.color,
            this.comment,
            this.displayOrder,
            this.visible,
            this.activityType,
            this.costType
        );
    }

    public toCreateRequest(clientSessionId: string): TransactionCategoryCreateRequest {
        return {
            name: this.name,
            type: this.type,
            icon: this.icon,
            color: this.color,
            comment: this.comment,
            activityType: this.activityType,
            costType: this.costType,
            clientSessionId: clientSessionId
        };
    }

    public toModifyRequest(): TransactionCategoryModifyRequest {
        return {
            id: this.id,
            name: this.name,
            icon: this.icon,
            color: this.color,
            comment: this.comment,
            hidden: !this.visible,
            activityType: this.activityType,
            costType: this.costType
        };
    }

    public static of(categoryResponse: TransactionCategoryInfoResponse): TransactionCategory {
        return new TransactionCategory(
            categoryResponse.id,
            categoryResponse.name,
            categoryResponse.type,
            categoryResponse.icon,
            categoryResponse.color,
            categoryResponse.comment,
            categoryResponse.displayOrder,
            !categoryResponse.hidden,
            categoryResponse.activityType || 0,
            categoryResponse.costType || 0
        );
    }

    public static ofMulti(categoryResponses: TransactionCategoryInfoResponse[]): TransactionCategory[] {
        const categories: TransactionCategory[] = [];

        for (const categoryResponse of categoryResponses) {
            categories.push(TransactionCategory.of(categoryResponse));
        }

        return categories;
    }

    public static ofMap(categoriesByType: Record<number, TransactionCategoryInfoResponse[]>): Record<number, TransactionCategory[]> {
        const ret: Record<number, TransactionCategory[]> = {};

        for (const [categoryType, categories] of Object.entries(categoriesByType)) {
            ret[parseInt(categoryType)] = TransactionCategory.ofMulti(categories);
        }

        return ret;
    }

    public static findNameById(categories: TransactionCategory[], id: string): string | null {
        for (const category of categories) {
            if (category.id === id) {
                return category.name;
            }
        }

        return null;
    }

    public static createNewCategory(type?: CategoryType): TransactionCategory {
        return new TransactionCategory('', '', type || CategoryType.Income, DEFAULT_CATEGORY_ICON_ID, DEFAULT_CATEGORY_COLOR, '', 0, true, 1, 0);
    }
}

export interface TransactionCategoryCreateRequest {
    readonly name: string;
    readonly type: number;
    readonly icon: string;
    readonly color: string;
    readonly comment: string;
    readonly activityType: number;
    readonly costType: number;
    readonly clientSessionId: string;
}

export interface TransactionCategoryCreateBatchRequest {
    readonly categories: TransactionCategoryCreateRequest[];
}

export interface TransactionCategoryModifyRequest {
    readonly id: string;
    readonly name: string;
    readonly icon: string;
    readonly color: string;
    readonly comment: string;
    readonly hidden: boolean;
    readonly activityType: number;
    readonly costType: number;
}

export interface TransactionCategoryHideRequest {
    readonly id: string;
    readonly hidden: boolean;
}

export interface TransactionCategoryMoveRequest {
    readonly newDisplayOrders: TransactionCategoryNewDisplayOrderRequest[];
}

export interface TransactionCategoryNewDisplayOrderRequest {
    readonly id: string;
    readonly displayOrder: number;
}

export interface TransactionCategoryDeleteRequest {
    readonly id: string;
}

export interface TransactionCategoryInfoResponse {
    readonly id: string;
    readonly name: string;
    readonly type: number;
    readonly icon: string;
    readonly color: string;
    readonly comment: string;
    readonly displayOrder: number;
    readonly hidden: boolean;
    readonly activityType?: number;
    readonly costType?: number;
}
