import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    HivePaaSSecuritySettings_FindOne_Req,
    HivePaaSSecuritySettings_FindOne_Res,
    HivePaaSSecuritySettings_UpdateOne_Req,
    HivePaaSSecuritySettings_UpdateOne_Res,
} from "./hivepaas-security-settings.api.contracts";
import type { HivePaaSSecuritySettingsApiValidator } from "./hivepaas-security-settings.api.validator";

export class HivePaaSSecuritySettingsApi extends BaseApi {
    public constructor(private readonly validator: HivePaaSSecuritySettingsApiValidator) {
        super();
    }

    async findOne(
        _request: HivePaaSSecuritySettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSSecuritySettings_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/hivepaas/security-settings", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: HivePaaSSecuritySettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSSecuritySettings_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put("/system/hivepaas/security-settings", payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
