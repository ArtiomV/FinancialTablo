export class TaxRecord implements TaxRecordInfoResponse {
    public id: string;
    public cfoId: string;
    public taxType: number;
    public periodYear: number;
    public periodQuarter: number;
    public taxableIncome: number;
    public taxAmount: number;
    public paidAmount: number;
    public dueDate: number;
    public status: number;
    public comment: string;

    private constructor(id: string, cfoId: string, taxType: number, periodYear: number, periodQuarter: number, taxableIncome: number, taxAmount: number, paidAmount: number, dueDate: number, status: number, comment: string) {
        this.id = id;
        this.cfoId = cfoId;
        this.taxType = taxType;
        this.periodYear = periodYear;
        this.periodQuarter = periodQuarter;
        this.taxableIncome = taxableIncome;
        this.taxAmount = taxAmount;
        this.paidAmount = paidAmount;
        this.dueDate = dueDate;
        this.status = status;
        this.comment = comment;
    }

    public toCreateRequest(): TaxRecordCreateRequest {
        return {
            cfoId: this.cfoId,
            taxType: this.taxType,
            periodYear: this.periodYear,
            periodQuarter: this.periodQuarter,
            taxableIncome: this.taxableIncome,
            taxAmount: this.taxAmount,
            paidAmount: this.paidAmount,
            dueDate: this.dueDate,
            status: this.status,
            comment: this.comment
        };
    }

    public toModifyRequest(): TaxRecordModifyRequest {
        return {
            id: this.id,
            cfoId: this.cfoId,
            taxType: this.taxType,
            periodYear: this.periodYear,
            periodQuarter: this.periodQuarter,
            taxableIncome: this.taxableIncome,
            taxAmount: this.taxAmount,
            paidAmount: this.paidAmount,
            dueDate: this.dueDate,
            status: this.status,
            comment: this.comment
        };
    }

    public clone(): TaxRecord {
        return new TaxRecord(this.id, this.cfoId, this.taxType, this.periodYear, this.periodQuarter, this.taxableIncome, this.taxAmount, this.paidAmount, this.dueDate, this.status, this.comment);
    }

    public static of(response: TaxRecordInfoResponse): TaxRecord {
        return new TaxRecord(response.id, response.cfoId, response.taxType, response.periodYear, response.periodQuarter, response.taxableIncome, response.taxAmount, response.paidAmount, response.dueDate, response.status, response.comment);
    }

    public static ofMulti(responses: TaxRecordInfoResponse[]): TaxRecord[] {
        const records: TaxRecord[] = [];
        for (const response of responses) {
            records.push(TaxRecord.of(response));
        }
        return records;
    }

    public static createNew(): TaxRecord {
        const now = new Date();
        return new TaxRecord('', '0', 1, now.getFullYear(), Math.ceil((now.getMonth() + 1) / 3), 0, 0, 0, 0, 1, '');
    }
}

export interface TaxRecordCreateRequest {
    readonly cfoId: string;
    readonly taxType: number;
    readonly periodYear: number;
    readonly periodQuarter: number;
    readonly taxableIncome: number;
    readonly taxAmount: number;
    readonly paidAmount: number;
    readonly dueDate: number;
    readonly status: number;
    readonly comment: string;
}

export interface TaxRecordModifyRequest {
    readonly id: string;
    readonly cfoId: string;
    readonly taxType: number;
    readonly periodYear: number;
    readonly periodQuarter: number;
    readonly taxableIncome: number;
    readonly taxAmount: number;
    readonly paidAmount: number;
    readonly dueDate: number;
    readonly status: number;
    readonly comment: string;
}

export interface TaxRecordDeleteRequest {
    readonly id: string;
}

export interface TaxRecordInfoResponse {
    readonly id: string;
    readonly cfoId: string;
    readonly taxType: number;
    readonly periodYear: number;
    readonly periodQuarter: number;
    readonly taxableIncome: number;
    readonly taxAmount: number;
    readonly paidAmount: number;
    readonly dueDate: number;
    readonly status: number;
    readonly comment: string;
}
