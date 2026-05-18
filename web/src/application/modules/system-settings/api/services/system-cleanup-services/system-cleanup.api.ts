import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    SystemCleanup_Execute_Req,
    SystemCleanup_Execute_Res,
    SystemCleanup_FindOne_Req,
    SystemCleanup_FindOne_Res,
    SystemCleanup_UpdateOne_Req,
    SystemCleanup_UpdateOne_Res,
} from "./system-cleanup.api.contracts";
import type { SystemCleanupApiValidator } from "./system-cleanup.api.validator";

export class SystemCleanupApi extends BaseApi {
    public constructor(private readonly validator: SystemCleanupApiValidator) {
        super();
    }

    async findOne(
        _request: SystemCleanup_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemCleanup_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/settings/cleanup", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: SystemCleanup_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemCleanup_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put("/system/settings/cleanup", payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async execute(
        _request: SystemCleanup_Execute_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemCleanup_Execute_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/system/settings/cleanup/exec", {}, { signal })).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
