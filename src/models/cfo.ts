export class CFO implements CFOInfoResponse {
    public id: string;
    public name: string;
    public color: string;
    public comment: string;
    public displayOrder: number;
    public hidden: boolean;

    private constructor(id: string, name: string, color: string, comment: string, displayOrder: number, hidden: boolean) {
        this.id = id;
        this.name = name;
        this.color = color;
        this.comment = comment;
        this.displayOrder = displayOrder;
        this.hidden = hidden;
    }

    public toCreateRequest(clientSessionId?: string): CFOCreateRequest {
        return {
            name: this.name,
            color: this.color,
            comment: this.comment,
            clientSessionId: clientSessionId
        };
    }

    public toModifyRequest(): CFOModifyRequest {
        return {
            id: this.id,
            name: this.name,
            color: this.color,
            comment: this.comment,
            hidden: this.hidden
        };
    }

    public clone(): CFO {
        return new CFO(this.id, this.name, this.color, this.comment, this.displayOrder, this.hidden);
    }

    public static of(response: CFOInfoResponse): CFO {
        return new CFO(response.id, response.name, response.color, response.comment, response.displayOrder, response.hidden);
    }

    public static ofMulti(responses: CFOInfoResponse[]): CFO[] {
        const cfos: CFO[] = [];

        for (const response of responses) {
            cfos.push(CFO.of(response));
        }

        return cfos;
    }

    public static createNew(): CFO {
        return new CFO('', '', '000000', '', 0, false);
    }
}

export interface CFOCreateRequest {
    readonly name: string;
    readonly color: string;
    readonly comment: string;
    readonly clientSessionId?: string;
}

export interface CFOModifyRequest {
    readonly id: string;
    readonly name: string;
    readonly color: string;
    readonly comment: string;
    readonly hidden: boolean;
}

export interface CFOHideRequest {
    readonly id: string;
    readonly hidden: boolean;
}

export interface CFOMoveRequest {
    readonly newDisplayOrders: CFONewDisplayOrderRequest[];
}

export interface CFONewDisplayOrderRequest {
    readonly id: string;
    readonly displayOrder: number;
}

export interface CFODeleteRequest {
    readonly id: string;
}

export interface CFOInfoResponse {
    readonly id: string;
    readonly name: string;
    readonly color: string;
    readonly comment: string;
    readonly displayOrder: number;
    readonly hidden: boolean;
}
