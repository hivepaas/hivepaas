import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    HivePaaSHttpSettings_FindOne_Req,
    HivePaaSHttpSettings_FindOne_Res,
    HivePaaSHttpSettings_UpdateOne_Req,
    HivePaaSHttpSettings_UpdateOne_Res,
} from "./hivepaas-http-settings.api.contracts";
import type { HivePaaSHttpSettingsApiValidator } from "./hivepaas-http-settings.api.validator";

export class HivePaaSHttpSettingsApi extends BaseApi {
    public constructor(private readonly validator: HivePaaSHttpSettingsApiValidator) {
        super();
    }

    async findOne(
        _request: HivePaaSHttpSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSHttpSettings_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/hivepaas/http-settings", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: HivePaaSHttpSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSHttpSettings_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put("/system/hivepaas/http-settings", payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
