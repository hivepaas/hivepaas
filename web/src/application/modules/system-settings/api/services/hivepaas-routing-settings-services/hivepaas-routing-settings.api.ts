import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    HivePaaSRoutingSettings_FindOne_Req,
    HivePaaSRoutingSettings_FindOne_Res,
    HivePaaSRoutingSettings_UpdateOne_Req,
    HivePaaSRoutingSettings_UpdateOne_Res,
} from "./hivepaas-routing-settings.api.contracts";
import type { HivePaaSRoutingSettingsApiValidator } from "./hivepaas-routing-settings.api.validator";

export class HivePaaSRoutingSettingsApi extends BaseApi {
    public constructor(private readonly validator: HivePaaSRoutingSettingsApiValidator) {
        super();
    }

    async findOne(
        _request: HivePaaSRoutingSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSRoutingSettings_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/hivepaas/routing-settings", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: HivePaaSRoutingSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSRoutingSettings_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put("/system/hivepaas/routing-settings", payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}

export { HivePaaSRoutingSettingsApi as HivePaaSHttpSettingsApi };
