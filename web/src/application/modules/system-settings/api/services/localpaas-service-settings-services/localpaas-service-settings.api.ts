import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    LocalPaaSServiceSettings_FindOne_Req,
    LocalPaaSServiceSettings_FindOne_Res,
    LocalPaaSServiceSettings_UpdateOne_Req,
    LocalPaaSServiceSettings_UpdateOne_Res,
} from "./localpaas-service-settings.api.contracts";
import type { LocalPaaSServiceSettingsApiValidator } from "./localpaas-service-settings.api.validator";

export class LocalPaaSServiceSettingsApi extends BaseApi {
    public constructor(private readonly validator: LocalPaaSServiceSettingsApiValidator) {
        super();
    }

    async findOne(
        _request: LocalPaaSServiceSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<LocalPaaSServiceSettings_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/hivepaas/service-settings", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: LocalPaaSServiceSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<LocalPaaSServiceSettings_UpdateOne_Res, Error>> {
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
