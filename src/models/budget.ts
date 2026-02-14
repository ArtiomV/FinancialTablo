export interface BudgetInfoResponse {
    readonly id: string;
    readonly cfoId: string;
    readonly categoryId: string;
    readonly year: number;
    readonly month: number;
    readonly plannedAmount: number;
    readonly comment: string;
}

export interface BudgetSaveRequest {
    readonly year: number;
    readonly month: number;
    readonly cfoId: string;
    readonly budgets: BudgetItemRequest[];
}

export interface BudgetItemRequest {
    readonly categoryId: string;
    readonly plannedAmount: number;
    readonly comment: string;
}

export interface PlanFactLineResponse {
    readonly categoryId: string;
    readonly categoryName: string;
    readonly categoryType: number;
    readonly plannedAmount: number;
    readonly factAmount: number;
    readonly deviation: number;
    readonly deviationPct: number | null;
}

export interface PlanFactResponse {
    readonly year: number;
    readonly month: number;
    readonly cfoId: string;
    readonly lines: PlanFactLineResponse[];
}
