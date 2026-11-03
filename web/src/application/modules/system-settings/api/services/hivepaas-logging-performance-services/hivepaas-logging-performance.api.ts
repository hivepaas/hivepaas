import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    HivePaaSLoggingPerformance_FindOne_Req,
    HivePaaSLoggingPerformance_FindOne_Res,
    HivePaaSLoggingPerformance_UpdateOne_Req,
    HivePaaSLoggingPerformance_UpdateOne_Res,
} from "./hivepaas-logging-performance.api.contracts";
import type { HivePaaSLoggingPerformanceApiValidator } from "./hivepaas-logging-performance.api.validator";

export class HivePaaSLoggingPerformanceApi extends BaseApi {
    public constructor(private readonly validator: HivePaaSLoggingPerformanceApiValidator) {
        super();
    }

    async findOne(
        _request: HivePaaSLoggingPerformance_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSLoggingPerformance_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/settings/logging/performance", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: HivePaaSLoggingPerformance_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSLoggingPerformance_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        // Saved with the logging settings and applied by each node's agent on its own, within 30 seconds:
        // nothing is deployed here.
        return lastValueFrom(
            from(this.client.v1.put("/system/settings/logging/performance", payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
