import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    HivePaaSLoggingSettings_FindOne_Req,
    HivePaaSLoggingSettings_FindOne_Res,
    HivePaaSLoggingSettings_UpdateOne_Req,
    HivePaaSLoggingSettings_UpdateOne_Res,
} from "./hivepaas-logging-settings.api.contracts";
import type { HivePaaSLoggingSettingsApiValidator } from "./hivepaas-logging-settings.api.validator";

export class HivePaaSLoggingSettingsApi extends BaseApi {
    public constructor(private readonly validator: HivePaaSLoggingSettingsApiValidator) {
        super();
    }

    async findOne(
        _request: HivePaaSLoggingSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSLoggingSettings_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/hivepaas/logging-settings", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: HivePaaSLoggingSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSLoggingSettings_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        // Saving also deploys, updates or removes the logging stack, so this can
        // take several seconds.
        return lastValueFrom(
            from(this.client.v1.put("/system/hivepaas/logging-settings", { data: payload }, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
