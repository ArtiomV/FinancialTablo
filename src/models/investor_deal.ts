export class InvestorDeal implements InvestorDealInfoResponse {
    public id: string;
    public investorName: string;
    public cfoId: string;
    public investmentDate: number;
    public investmentAmount: number;
    public currency: string;
    public dealType: number;
    public annualRate: number;
    public profitSharePct: number;
    public fixedPayment: number;
    public repaymentStartDate: number;
    public repaymentEndDate: number;
    public totalToRepay: number;
    public comment: string;

    private constructor(id: string, investorName: string, cfoId: string, investmentDate: number, investmentAmount: number, currency: string, dealType: number, annualRate: number, profitSharePct: number, fixedPayment: number, repaymentStartDate: number, repaymentEndDate: number, totalToRepay: number, comment: string) {
        this.id = id;
        this.investorName = investorName;
        this.cfoId = cfoId;
        this.investmentDate = investmentDate;
        this.investmentAmount = investmentAmount;
        this.currency = currency;
        this.dealType = dealType;
        this.annualRate = annualRate;
        this.profitSharePct = profitSharePct;
        this.fixedPayment = fixedPayment;
        this.repaymentStartDate = repaymentStartDate;
        this.repaymentEndDate = repaymentEndDate;
        this.totalToRepay = totalToRepay;
        this.comment = comment;
    }

    public toCreateRequest(): InvestorDealCreateRequest {
        return {
            investorName: this.investorName,
            cfoId: this.cfoId,
            investmentDate: this.investmentDate,
            investmentAmount: this.investmentAmount,
            currency: this.currency,
            dealType: this.dealType,
            annualRate: this.annualRate,
            profitSharePct: this.profitSharePct,
            fixedPayment: this.fixedPayment,
            repaymentStartDate: this.repaymentStartDate,
            repaymentEndDate: this.repaymentEndDate,
            totalToRepay: this.totalToRepay,
            comment: this.comment
        };
    }

    public toModifyRequest(): InvestorDealModifyRequest {
        return {
            id: this.id,
            investorName: this.investorName,
            cfoId: this.cfoId,
            investmentDate: this.investmentDate,
            investmentAmount: this.investmentAmount,
            currency: this.currency,
            dealType: this.dealType,
            annualRate: this.annualRate,
            profitSharePct: this.profitSharePct,
            fixedPayment: this.fixedPayment,
            repaymentStartDate: this.repaymentStartDate,
            repaymentEndDate: this.repaymentEndDate,
            totalToRepay: this.totalToRepay,
            comment: this.comment
        };
    }

    public clone(): InvestorDeal {
        return new InvestorDeal(this.id, this.investorName, this.cfoId, this.investmentDate, this.investmentAmount, this.currency, this.dealType, this.annualRate, this.profitSharePct, this.fixedPayment, this.repaymentStartDate, this.repaymentEndDate, this.totalToRepay, this.comment);
    }

    public static of(response: InvestorDealInfoResponse): InvestorDeal {
        return new InvestorDeal(response.id, response.investorName, response.cfoId, response.investmentDate, response.investmentAmount, response.currency, response.dealType, response.annualRate, response.profitSharePct, response.fixedPayment, response.repaymentStartDate, response.repaymentEndDate, response.totalToRepay, response.comment);
    }

    public static ofMulti(responses: InvestorDealInfoResponse[]): InvestorDeal[] {
        const deals: InvestorDeal[] = [];

        for (const response of responses) {
            deals.push(InvestorDeal.of(response));
        }

        return deals;
    }

    public static createNew(): InvestorDeal {
        return new InvestorDeal('', '', '0', 0, 0, 'RUB', 1, 0, 0, 0, 0, 0, 0, '');
    }
}

export interface InvestorDealCreateRequest {
    readonly investorName: string;
    readonly cfoId: string;
    readonly investmentDate: number;
    readonly investmentAmount: number;
    readonly currency: string;
    readonly dealType: number;
    readonly annualRate: number;
    readonly profitSharePct: number;
    readonly fixedPayment: number;
    readonly repaymentStartDate: number;
    readonly repaymentEndDate: number;
    readonly totalToRepay: number;
    readonly comment: string;
}

export interface InvestorDealModifyRequest {
    readonly id: string;
    readonly investorName: string;
    readonly cfoId: string;
    readonly investmentDate: number;
    readonly investmentAmount: number;
    readonly currency: string;
    readonly dealType: number;
    readonly annualRate: number;
    readonly profitSharePct: number;
    readonly fixedPayment: number;
    readonly repaymentStartDate: number;
    readonly repaymentEndDate: number;
    readonly totalToRepay: number;
    readonly comment: string;
}

export interface InvestorDealDeleteRequest {
    readonly id: string;
}

export interface InvestorDealInfoResponse {
    readonly id: string;
    readonly investorName: string;
    readonly cfoId: string;
    readonly investmentDate: number;
    readonly investmentAmount: number;
    readonly currency: string;
    readonly dealType: number;
    readonly annualRate: number;
    readonly profitSharePct: number;
    readonly fixedPayment: number;
    readonly repaymentStartDate: number;
    readonly repaymentEndDate: number;
    readonly totalToRepay: number;
    readonly comment: string;
}
