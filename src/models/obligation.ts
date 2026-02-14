export class Obligation implements ObligationInfoResponse {
    public id: string;
    public obligationType: number;
    public counterpartyId: string;
    public cfoId: string;
    public amount: number;
    public currency: string;
    public dueDate: number;
    public status: number;
    public paidAmount: number;
    public comment: string;

    private constructor(id: string, obligationType: number, counterpartyId: string, cfoId: string, amount: number, currency: string, dueDate: number, status: number, paidAmount: number, comment: string) {
        this.id = id;
        this.obligationType = obligationType;
        this.counterpartyId = counterpartyId;
        this.cfoId = cfoId;
        this.amount = amount;
        this.currency = currency;
        this.dueDate = dueDate;
        this.status = status;
        this.paidAmount = paidAmount;
        this.comment = comment;
    }

    public toCreateRequest(): ObligationCreateRequest {
        return {
            obligationType: this.obligationType,
            counterpartyId: this.counterpartyId,
            cfoId: this.cfoId,
            amount: this.amount,
            currency: this.currency,
            dueDate: this.dueDate,
            status: this.status,
            paidAmount: this.paidAmount,
            comment: this.comment
        };
    }

    public toModifyRequest(): ObligationModifyRequest {
        return {
            id: this.id,
            obligationType: this.obligationType,
            counterpartyId: this.counterpartyId,
            cfoId: this.cfoId,
            amount: this.amount,
            currency: this.currency,
            dueDate: this.dueDate,
            status: this.status,
            paidAmount: this.paidAmount,
            comment: this.comment
        };
    }

    public clone(): Obligation {
        return new Obligation(this.id, this.obligationType, this.counterpartyId, this.cfoId, this.amount, this.currency, this.dueDate, this.status, this.paidAmount, this.comment);
    }

    public static of(response: ObligationInfoResponse): Obligation {
        return new Obligation(response.id, response.obligationType, response.counterpartyId, response.cfoId, response.amount, response.currency, response.dueDate, response.status, response.paidAmount, response.comment);
    }

    public static ofMulti(responses: ObligationInfoResponse[]): Obligation[] {
        const obligations: Obligation[] = [];
        for (const response of responses) {
            obligations.push(Obligation.of(response));
        }
        return obligations;
    }

    public static createNew(obligationType: number = 1): Obligation {
        return new Obligation('', obligationType, '0', '0', 0, 'RUB', 0, 1, 0, '');
    }
}

export interface ObligationCreateRequest {
    readonly obligationType: number;
    readonly counterpartyId: string;
    readonly cfoId: string;
    readonly amount: number;
    readonly currency: string;
    readonly dueDate: number;
    readonly status: number;
    readonly paidAmount: number;
    readonly comment: string;
}

export interface ObligationModifyRequest {
    readonly id: string;
    readonly obligationType: number;
    readonly counterpartyId: string;
    readonly cfoId: string;
    readonly amount: number;
    readonly currency: string;
    readonly dueDate: number;
    readonly status: number;
    readonly paidAmount: number;
    readonly comment: string;
}

export interface ObligationDeleteRequest {
    readonly id: string;
}

export interface ObligationInfoResponse {
    readonly id: string;
    readonly obligationType: number;
    readonly counterpartyId: string;
    readonly cfoId: string;
    readonly amount: number;
    readonly currency: string;
    readonly dueDate: number;
    readonly status: number;
    readonly paidAmount: number;
    readonly comment: string;
}
