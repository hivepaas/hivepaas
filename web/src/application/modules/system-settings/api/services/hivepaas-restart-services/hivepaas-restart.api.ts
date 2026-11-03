import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { HivePaaSRestart_Execute_Req, HivePaaSRestart_Execute_Res } from "./hivepaas-restart.api.contracts";
import type { HivePaaSRestartApiValidator } from "./hivepaas-restart.api.validator";

export class HivePaaSRestartApi extends BaseApi {
    public constructor(private readonly validator: HivePaaSRestartApiValidator) {
        super();
    }

    async execute(
        request: HivePaaSRestart_Execute_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSRestart_Execute_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.post("/system/hivepaas/restart", payload, { signal })).pipe(
                map(this.validator.execute),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
