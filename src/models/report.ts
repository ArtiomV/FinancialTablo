export interface CashFlowActivityLine {
    readonly categoryId: string;
    readonly categoryName: string;
    readonly income: number;
    readonly expense: number;
    readonly net: number;
}

export interface CashFlowActivity {
    readonly activityType: number;
    readonly activityName: string;
    readonly lines: CashFlowActivityLine[];
    readonly totalIncome: number;
    readonly totalExpense: number;
    readonly totalNet: number;
}

export interface CashFlowResponse {
    readonly activities: CashFlowActivity[];
    readonly totalNet: number;
}

export interface PnLResponse {
    readonly revenue: number;
    readonly costOfGoods: number;
    readonly grossProfit: number;
    readonly operatingExpense: number;
    readonly depreciation: number;
    readonly operatingProfit: number;
    readonly financialExpense: number;
    readonly taxExpense: number;
    readonly netProfit: number;
}

export interface BalanceLine {
    readonly label: string;
    readonly amount: number;
}

export interface BalanceResponse {
    readonly assetLines: BalanceLine[];
    readonly totalAssets: number;
    readonly liabilityLines: BalanceLine[];
    readonly totalLiability: number;
    readonly equity: number;
}

export interface PaymentCalendarItem {
    readonly date: number;
    readonly type: string;
    readonly amount: number;
    readonly description: string;
    readonly currency: string;
}

export interface PaymentCalendarResponse {
    readonly items: PaymentCalendarItem[];
}
