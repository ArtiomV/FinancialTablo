export class Asset implements AssetInfoResponse {
    public id: string;
    public name: string;
    public cfoId: string;
    public locationId: string;
    public assetType: number;
    public purchaseDate: number;
    public purchaseCost: number;
    public usefulLifeMonths: number;
    public salvageValue: number;
    public status: number;
    public commissionDate: number;
    public decommissionDate: number;
    public installedCapacityWatts: number;
    public comment: string;
    public displayOrder: number;
    public hidden: boolean;

    private constructor(id: string, name: string, cfoId: string, locationId: string, assetType: number, purchaseDate: number, purchaseCost: number, usefulLifeMonths: number, salvageValue: number, status: number, commissionDate: number, decommissionDate: number, installedCapacityWatts: number, comment: string, displayOrder: number, hidden: boolean) {
        this.id = id;
        this.name = name;
        this.cfoId = cfoId;
        this.locationId = locationId;
        this.assetType = assetType;
        this.purchaseDate = purchaseDate;
        this.purchaseCost = purchaseCost;
        this.usefulLifeMonths = usefulLifeMonths;
        this.salvageValue = salvageValue;
        this.status = status;
        this.commissionDate = commissionDate;
        this.decommissionDate = decommissionDate;
        this.installedCapacityWatts = installedCapacityWatts;
        this.comment = comment;
        this.displayOrder = displayOrder;
        this.hidden = hidden;
    }

    public toCreateRequest(clientSessionId?: string): AssetCreateRequest {
        return {
            name: this.name,
            cfoId: this.cfoId,
            locationId: this.locationId,
            assetType: this.assetType,
            purchaseDate: this.purchaseDate,
            purchaseCost: this.purchaseCost,
            usefulLifeMonths: this.usefulLifeMonths,
            salvageValue: this.salvageValue,
            status: this.status,
            commissionDate: this.commissionDate,
            decommissionDate: this.decommissionDate,
            installedCapacityWatts: this.installedCapacityWatts,
            comment: this.comment,
            clientSessionId: clientSessionId
        };
    }

    public toModifyRequest(): AssetModifyRequest {
        return {
            id: this.id,
            name: this.name,
            cfoId: this.cfoId,
            locationId: this.locationId,
            assetType: this.assetType,
            purchaseDate: this.purchaseDate,
            purchaseCost: this.purchaseCost,
            usefulLifeMonths: this.usefulLifeMonths,
            salvageValue: this.salvageValue,
            status: this.status,
            commissionDate: this.commissionDate,
            decommissionDate: this.decommissionDate,
            installedCapacityWatts: this.installedCapacityWatts,
            comment: this.comment,
            hidden: this.hidden
        };
    }

    public clone(): Asset {
        return new Asset(this.id, this.name, this.cfoId, this.locationId, this.assetType, this.purchaseDate, this.purchaseCost, this.usefulLifeMonths, this.salvageValue, this.status, this.commissionDate, this.decommissionDate, this.installedCapacityWatts, this.comment, this.displayOrder, this.hidden);
    }

    public static of(response: AssetInfoResponse): Asset {
        return new Asset(response.id, response.name, response.cfoId, response.locationId, response.assetType, response.purchaseDate, response.purchaseCost, response.usefulLifeMonths, response.salvageValue, response.status, response.commissionDate, response.decommissionDate, response.installedCapacityWatts, response.comment, response.displayOrder, response.hidden);
    }

    public static ofMulti(responses: AssetInfoResponse[]): Asset[] {
        const assets: Asset[] = [];

        for (const response of responses) {
            assets.push(Asset.of(response));
        }

        return assets;
    }

    public static createNew(): Asset {
        return new Asset('', '', '0', '0', 1, 0, 0, 0, 0, 1, 0, 0, 0, '', 0, false);
    }
}

export interface AssetCreateRequest {
    readonly name: string;
    readonly cfoId: string;
    readonly locationId: string;
    readonly assetType: number;
    readonly purchaseDate: number;
    readonly purchaseCost: number;
    readonly usefulLifeMonths: number;
    readonly salvageValue: number;
    readonly status: number;
    readonly commissionDate: number;
    readonly decommissionDate: number;
    readonly installedCapacityWatts: number;
    readonly comment: string;
    readonly clientSessionId?: string;
}

export interface AssetModifyRequest {
    readonly id: string;
    readonly name: string;
    readonly cfoId: string;
    readonly locationId: string;
    readonly assetType: number;
    readonly purchaseDate: number;
    readonly purchaseCost: number;
    readonly usefulLifeMonths: number;
    readonly salvageValue: number;
    readonly status: number;
    readonly commissionDate: number;
    readonly decommissionDate: number;
    readonly installedCapacityWatts: number;
    readonly comment: string;
    readonly hidden: boolean;
}

export interface AssetHideRequest {
    readonly id: string;
    readonly hidden: boolean;
}

export interface AssetMoveRequest {
    readonly newDisplayOrders: AssetNewDisplayOrderRequest[];
}

export interface AssetNewDisplayOrderRequest {
    readonly id: string;
    readonly displayOrder: number;
}

export interface AssetDeleteRequest {
    readonly id: string;
}

export interface AssetInfoResponse {
    readonly id: string;
    readonly name: string;
    readonly cfoId: string;
    readonly locationId: string;
    readonly assetType: number;
    readonly purchaseDate: number;
    readonly purchaseCost: number;
    readonly usefulLifeMonths: number;
    readonly salvageValue: number;
    readonly status: number;
    readonly commissionDate: number;
    readonly decommissionDate: number;
    readonly installedCapacityWatts: number;
    readonly comment: string;
    readonly displayOrder: number;
    readonly hidden: boolean;
}
