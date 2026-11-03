import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { HivePaaSProject_FindOne_Req, HivePaaSProject_FindOne_Res } from "./hivepaas-project.api.contracts";
import type { HivePaaSProjectApiValidator } from "./hivepaas-project.api.validator";

export class HivePaaSProjectApi extends BaseApi {
    public constructor(private readonly validator: HivePaaSProjectApiValidator) {
        super();
    }

    async findOne(
        _request: HivePaaSProject_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<HivePaaSProject_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/hivepaas/project", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
