import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    SystemBackup_Execute_Req,
    SystemBackup_Execute_Res,
    SystemBackup_FindOne_Req,
    SystemBackup_FindOne_Res,
    SystemBackup_UpdateOne_Req,
    SystemBackup_UpdateOne_Res,
} from "./system-backup.api.contracts";
import type { SystemBackupApiValidator } from "./system-backup.api.validator";

export class SystemBackupApi extends BaseApi {
    public constructor(private readonly validator: SystemBackupApiValidator) {
        super();
    }

    async findOne(
        _request: SystemBackup_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemBackup_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/settings/backup", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: SystemBackup_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemBackup_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put("/system/settings/backup", payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async execute(
        _request: SystemBackup_Execute_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemBackup_Execute_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/system/settings/backup/exec", {}, { signal })).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
