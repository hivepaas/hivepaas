import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    SystemBackupRepoCleanup_Execute_Req,
    SystemBackupRepoCleanup_Execute_Res,
    SystemBackupRepoCleanup_FindOne_Req,
    SystemBackupRepoCleanup_FindOne_Res,
    SystemBackupRepoCleanup_UpdateOne_Req,
    SystemBackupRepoCleanup_UpdateOne_Res,
} from "./system-backup-repo-cleanup.api.contracts";
import type { SystemBackupRepoCleanupApiValidator } from "./system-backup-repo-cleanup.api.validator";

export class SystemBackupRepoCleanupApi extends BaseApi {
    public constructor(private readonly validator: SystemBackupRepoCleanupApiValidator) {
        super();
    }

    async findOne(
        _request: SystemBackupRepoCleanup_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemBackupRepoCleanup_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/settings/backup-repo-cleanup", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: SystemBackupRepoCleanup_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemBackupRepoCleanup_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put("/system/settings/backup-repo-cleanup", payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async execute(
        request: SystemBackupRepoCleanup_Execute_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemBackupRepoCleanup_Execute_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.post("/system/settings/backup-repo-cleanup/exec", request.data, { signal })).pipe(
                map(this.validator.execute),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
