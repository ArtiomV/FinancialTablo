export class InvestorPayment implements InvestorPaymentInfoResponse {
    public id: string;
    public dealId: string;
    public paymentDate: number;
    public amount: number;
    public paymentType: number;
    public transactionId: string;
    public comment: string;

    private constructor(id: string, dealId: string, paymentDate: number, amount: number, paymentType: number, transactionId: string, comment: string) {
        this.id = id;
        this.dealId = dealId;
        this.paymentDate = paymentDate;
        this.amount = amount;
        this.paymentType = paymentType;
        this.transactionId = transactionId;
        this.comment = comment;
    }

    public toCreateRequest(): InvestorPaymentCreateRequest {
        return {
            dealId: this.dealId,
            paymentDate: this.paymentDate,
            amount: this.amount,
            paymentType: this.paymentType,
            transactionId: this.transactionId,
            comment: this.comment
        };
    }

    public toModifyRequest(): InvestorPaymentModifyRequest {
        return {
            id: this.id,
            dealId: this.dealId,
            paymentDate: this.paymentDate,
            amount: this.amount,
            paymentType: this.paymentType,
            transactionId: this.transactionId,
            comment: this.comment
        };
    }

    public clone(): InvestorPayment {
        return new InvestorPayment(this.id, this.dealId, this.paymentDate, this.amount, this.paymentType, this.transactionId, this.comment);
    }

    public static of(response: InvestorPaymentInfoResponse): InvestorPayment {
        return new InvestorPayment(response.id, response.dealId, response.paymentDate, response.amount, response.paymentType, response.transactionId, response.comment);
    }

    public static ofMulti(responses: InvestorPaymentInfoResponse[]): InvestorPayment[] {
        const payments: InvestorPayment[] = [];

        for (const response of responses) {
            payments.push(InvestorPayment.of(response));
        }

        return payments;
    }

    public static createNew(dealId: string): InvestorPayment {
        return new InvestorPayment('', dealId, 0, 0, 1, '0', '');
    }
}

export interface InvestorPaymentCreateRequest {
    readonly dealId: string;
    readonly paymentDate: number;
    readonly amount: number;
    readonly paymentType: number;
    readonly transactionId: string;
    readonly comment: string;
}

export interface InvestorPaymentModifyRequest {
    readonly id: string;
    readonly dealId: string;
    readonly paymentDate: number;
    readonly amount: number;
    readonly paymentType: number;
    readonly transactionId: string;
    readonly comment: string;
}

export interface InvestorPaymentDeleteRequest {
    readonly id: string;
}

export interface InvestorPaymentInfoResponse {
    readonly id: string;
    readonly dealId: string;
    readonly paymentDate: number;
    readonly amount: number;
    readonly paymentType: number;
    readonly transactionId: string;
    readonly comment: string;
}
