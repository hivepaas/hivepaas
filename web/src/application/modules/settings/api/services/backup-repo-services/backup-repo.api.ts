import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    BackupRepo_DeleteOne_Req,
    BackupRepo_DeleteOne_Res,
    BackupRepo_FindManyPaginated_Req,
    BackupRepo_FindManyPaginated_Res,
    BackupRepo_FindOneById_Req,
    BackupRepo_FindOneById_Res,
    BackupRepo_UpdateStatus_Req,
    BackupRepo_UpdateStatus_Res,
} from "./backup-repo.api.contracts";
import type { BackupRepoApiValidator } from "./backup-repo.api.validator";

export class BackupRepoApi extends BaseApi {
    public constructor(private readonly validator: BackupRepoApiValidator) {
        super();
    }

    async findManyPaginated(
        request: BackupRepo_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<BackupRepo_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        const params = query.build();

        return lastValueFrom(
            from(
                this.client.v1.get("/settings/backup-repos", {
                    params,
                    signal,
                }),
            ).pipe(
                map(this.validator.findManyPaginated),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findOneById(
        request: BackupRepo_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<BackupRepo_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/settings/backup-repos/${id}`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateStatus(
        request: BackupRepo_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<BackupRepo_UpdateStatus_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/backup-repos/${id}/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(
        request: BackupRepo_DeleteOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<BackupRepo_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/settings/backup-repos/${id}`, { signal })).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
