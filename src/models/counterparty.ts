export enum CounterpartyType {
    Person = 1,
    Company = 2
}

export class Counterparty implements CounterpartyInfoResponse {
    public id: string;
    public name: string;
    public type: CounterpartyType;
    public comment: string;
    public icon: string;
    public color: string;
    public displayOrder: number;
    public hidden: boolean;

    private constructor(id: string, name: string, type: CounterpartyType, comment: string, icon: string, color: string, displayOrder: number, hidden: boolean) {
        this.id = id;
        this.name = name;
        this.type = type;
        this.comment = comment;
        this.icon = icon;
        this.color = color;
        this.displayOrder = displayOrder;
        this.hidden = hidden;
    }

    public toCreateRequest(clientSessionId?: string): CounterpartyCreateRequest {
        return {
            name: this.name,
            type: this.type,
            icon: this.icon,
            color: this.color,
            comment: this.comment,
            clientSessionId: clientSessionId
        };
    }

    public toModifyRequest(): CounterpartyModifyRequest {
        return {
            id: this.id,
            name: this.name,
            type: this.type,
            icon: this.icon,
            color: this.color,
            comment: this.comment,
            hidden: this.hidden
        };
    }

    public clone(): Counterparty {
        return new Counterparty(this.id, this.name, this.type, this.comment, this.icon, this.color, this.displayOrder, this.hidden);
    }

    public static of(response: CounterpartyInfoResponse): Counterparty {
        return new Counterparty(response.id, response.name, response.type, response.comment, response.icon, response.color, response.displayOrder, response.hidden);
    }

    public static ofMulti(responses: CounterpartyInfoResponse[]): Counterparty[] {
        const counterparties: Counterparty[] = [];

        for (const response of responses) {
            counterparties.push(Counterparty.of(response));
        }

        return counterparties;
    }

    public static createNew(): Counterparty {
        return new Counterparty('', '', CounterpartyType.Person, '', '', '000000', 0, false);
    }
}

export interface CounterpartyCreateRequest {
    readonly name: string;
    readonly type: CounterpartyType;
    readonly icon: string;
    readonly color: string;
    readonly comment: string;
    readonly clientSessionId?: string;
}

export interface CounterpartyModifyRequest {
    readonly id: string;
    readonly name: string;
    readonly type: CounterpartyType;
    readonly icon: string;
    readonly color: string;
    readonly comment: string;
    readonly hidden: boolean;
}

export interface CounterpartyHideRequest {
    readonly id: string;
    readonly hidden: boolean;
}

export interface CounterpartyMoveRequest {
    readonly newDisplayOrders: CounterpartyNewDisplayOrderRequest[];
}

export interface CounterpartyNewDisplayOrderRequest {
    readonly id: string;
    readonly displayOrder: number;
}

export interface CounterpartyDeleteRequest {
    readonly id: string;
}

export interface CounterpartyInfoResponse {
    readonly id: string;
    readonly name: string;
    readonly type: CounterpartyType;
    readonly comment: string;
    readonly icon: string;
    readonly color: string;
    readonly displayOrder: number;
    readonly hidden: boolean;
}
