export class Location implements LocationInfoResponse {
    public id: string;
    public name: string;
    public cfoId: string;
    public address: string;
    public locationType: number;
    public monthlyRent: number;
    public monthlyElectricity: number;
    public monthlyInternet: number;
    public comment: string;
    public displayOrder: number;
    public hidden: boolean;

    private constructor(id: string, name: string, cfoId: string, address: string, locationType: number, monthlyRent: number, monthlyElectricity: number, monthlyInternet: number, comment: string, displayOrder: number, hidden: boolean) {
        this.id = id;
        this.name = name;
        this.cfoId = cfoId;
        this.address = address;
        this.locationType = locationType;
        this.monthlyRent = monthlyRent;
        this.monthlyElectricity = monthlyElectricity;
        this.monthlyInternet = monthlyInternet;
        this.comment = comment;
        this.displayOrder = displayOrder;
        this.hidden = hidden;
    }

    public toCreateRequest(clientSessionId?: string): LocationCreateRequest {
        return {
            name: this.name,
            cfoId: this.cfoId,
            address: this.address,
            locationType: this.locationType,
            monthlyRent: this.monthlyRent,
            monthlyElectricity: this.monthlyElectricity,
            monthlyInternet: this.monthlyInternet,
            comment: this.comment,
            clientSessionId: clientSessionId
        };
    }

    public toModifyRequest(): LocationModifyRequest {
        return {
            id: this.id,
            name: this.name,
            cfoId: this.cfoId,
            address: this.address,
            locationType: this.locationType,
            monthlyRent: this.monthlyRent,
            monthlyElectricity: this.monthlyElectricity,
            monthlyInternet: this.monthlyInternet,
            comment: this.comment,
            hidden: this.hidden
        };
    }

    public clone(): Location {
        return new Location(this.id, this.name, this.cfoId, this.address, this.locationType, this.monthlyRent, this.monthlyElectricity, this.monthlyInternet, this.comment, this.displayOrder, this.hidden);
    }

    public static of(response: LocationInfoResponse): Location {
        return new Location(response.id, response.name, response.cfoId, response.address, response.locationType, response.monthlyRent, response.monthlyElectricity, response.monthlyInternet, response.comment, response.displayOrder, response.hidden);
    }

    public static ofMulti(responses: LocationInfoResponse[]): Location[] {
        const locations: Location[] = [];

        for (const response of responses) {
            locations.push(Location.of(response));
        }

        return locations;
    }

    public static createNew(): Location {
        return new Location('', '', '0', '', 1, 0, 0, 0, '', 0, false);
    }
}

export interface LocationCreateRequest {
    readonly name: string;
    readonly cfoId: string;
    readonly address: string;
    readonly locationType: number;
    readonly monthlyRent: number;
    readonly monthlyElectricity: number;
    readonly monthlyInternet: number;
    readonly comment: string;
    readonly clientSessionId?: string;
}

export interface LocationModifyRequest {
    readonly id: string;
    readonly name: string;
    readonly cfoId: string;
    readonly address: string;
    readonly locationType: number;
    readonly monthlyRent: number;
    readonly monthlyElectricity: number;
    readonly monthlyInternet: number;
    readonly comment: string;
    readonly hidden: boolean;
}

export interface LocationHideRequest {
    readonly id: string;
    readonly hidden: boolean;
}

export interface LocationMoveRequest {
    readonly newDisplayOrders: LocationNewDisplayOrderRequest[];
}

export interface LocationNewDisplayOrderRequest {
    readonly id: string;
    readonly displayOrder: number;
}

export interface LocationDeleteRequest {
    readonly id: string;
}

export interface LocationInfoResponse {
    readonly id: string;
    readonly name: string;
    readonly cfoId: string;
    readonly address: string;
    readonly locationType: number;
    readonly monthlyRent: number;
    readonly monthlyElectricity: number;
    readonly monthlyInternet: number;
    readonly comment: string;
    readonly displayOrder: number;
    readonly hidden: boolean;
}
