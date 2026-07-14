import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    HivePaaSServiceSettings_FindOne_Req,
    HivePaaSServiceSettings_FindOne_Res,
    HivePaaSServiceSettings_UpdateOne_Req,
    HivePaaSServiceSettings_UpdateOne_Res,
} from "./hivepaas-service-settings.api.contracts";
import type { HivePaaSServiceSettingsApiValidator } from "./hivepaas-service-settings.api.validator";

export class HivePaaSServiceSettingsApi extends BaseApi {
    public constructor(private readonly validator: HivePaaSServiceSettingsApiValidator) {
        super();
    }

    async findOne(
        _request: HivePaaSServiceSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSServiceSettings_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/hivepaas/service-settings", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: HivePaaSServiceSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSServiceSettings_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put("/system/hivepaas/service-settings", payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
