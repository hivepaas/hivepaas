import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    HivePaaSRequestInfo_FindOne_Req,
    HivePaaSRequestInfo_FindOne_Res,
} from "./hivepaas-request-info.api.contracts";
import type { HivePaaSRequestInfoApiValidator } from "./hivepaas-request-info.api.validator";

export class HivePaaSRequestInfoApi extends BaseApi {
    public constructor(private readonly validator: HivePaaSRequestInfoApiValidator) {
        super();
    }

    async findOne(
        _request: HivePaaSRequestInfo_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSRequestInfo_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/hivepaas/request-info", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
