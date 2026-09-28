import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type { BackupSnapshotScope } from "~/settings/domain";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    BackupSnapshot_DeleteOne_Req,
    BackupSnapshot_DeleteOne_Res,
    BackupSnapshot_FindManyPaginated_Req,
    BackupSnapshot_FindManyPaginated_Res,
    BackupSnapshot_FindOneById_Req,
    BackupSnapshot_FindOneById_Res,
} from "./backup-snapshot.api.contracts";
import type { BackupSnapshotApiValidator } from "./backup-snapshot.api.validator";

/** Where a scope's snapshots are served. */
export function getBackupSnapshotBasePath(scope: BackupSnapshotScope): string {
    switch (scope.type) {
        case "settings":
            return "/settings/backup-snapshots";
        case "project":
            if (scope.env && scope.env !== "all") {
                return `/projects/${scope.projectId}/${encodeURIComponent(scope.env)}/backup-snapshots`;
            }
            return `/projects/${scope.projectId}/backup-snapshots`;
        case "app":
            return `/projects/${scope.projectId}/${encodeURIComponent(scope.env)}/apps/${scope.appId}/backup-snapshots`;
    }
}

export class BackupSnapshotApi extends BaseApi {
    public constructor(private readonly validator: BackupSnapshotApiValidator) {
        super();
    }

    async findManyPaginated(
        request: BackupSnapshot_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<BackupSnapshot_FindManyPaginated_Res, Error>> {
        const { scope, pagination, search, repo, app, tag, fromDate, toDate } = request.data;
        const query = this.queryBuilder.getInstance();
        query
            .pagination(pagination)
            .search(search)
            .filterBy({
                repo,
                app,
                tag,
                fromDate: fromDate ? [fromDate] : undefined,
                toDate: toDate ? [toDate] : undefined,
            });

        return lastValueFrom(
            from(this.client.v1.get(getBackupSnapshotBasePath(scope), { params: query.build(), signal })).pipe(
                map(this.validator.findManyPaginated),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findOneById(
        request: BackupSnapshot_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<BackupSnapshot_FindOneById_Res, Error>> {
        const { scope, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.get(`${getBackupSnapshotBasePath(scope)}/${id}`, { signal })).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(
        request: BackupSnapshot_DeleteOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<BackupSnapshot_DeleteOne_Res, Error>> {
        const { scope, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`${getBackupSnapshotBasePath(scope)}/${id}`, { signal })).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
