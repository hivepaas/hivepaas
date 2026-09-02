import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectBackupRepo_DeleteOne_Req,
    ProjectBackupRepo_DeleteOne_Res,
    ProjectBackupRepo_FindManyPaginated_Req,
    ProjectBackupRepo_FindManyPaginated_Res,
    ProjectBackupRepo_FindOneById_Req,
    ProjectBackupRepo_FindOneById_Res,
    ProjectBackupRepo_UpdateStatus_Req,
    ProjectBackupRepo_UpdateStatus_Res,
} from "./project-backup-repo.api.contracts";
import type { ProjectBackupRepoApiValidator } from "./project-backup-repo.api.validator";

function getProjectBackupRepoBasePath(projectID: string, env?: string): string {
    if (env && env !== "all") {
        return `/projects/${projectID}/${encodeURIComponent(env)}/backup-repos`;
    }

    return `/projects/${projectID}/backup-repos`;
}

export class ProjectBackupRepoApi extends BaseApi {
    public constructor(private readonly validator: ProjectBackupRepoApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectBackupRepo_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBackupRepo_FindManyPaginated_Res, Error>> {
        const { projectID, env, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        const params = query.build();

        return lastValueFrom(
            from(
                this.client.v1.get(getProjectBackupRepoBasePath(projectID, env), {
                    params,
                    signal,
                }),
            ).pipe(
                map(response => this.validator.findManyPaginated(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findOneById(
        request: ProjectBackupRepo_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBackupRepo_FindOneById_Res, Error>> {
        const { projectID, env, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`${getProjectBackupRepoBasePath(projectID, env)}/${id}`, {
                    signal,
                }),
            ).pipe(
                map(response => this.validator.findOneById(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateStatus(
        request: ProjectBackupRepo_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBackupRepo_UpdateStatus_Res, Error>> {
        const { projectID, env, id, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.put(`${getProjectBackupRepoBasePath(projectID, env)}/${id}/status`, json, {
                    signal,
                }),
            ).pipe(
                map(response => this.validator.updateStatus(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(
        request: ProjectBackupRepo_DeleteOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBackupRepo_DeleteOne_Res, Error>> {
        const { projectID, env, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`${getProjectBackupRepoBasePath(projectID, env)}/${id}`, { signal })).pipe(
                map(response => this.validator.deleteOne(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
