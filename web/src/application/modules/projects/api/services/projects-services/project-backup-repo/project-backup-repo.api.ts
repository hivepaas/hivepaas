import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectBackupRepo_Cleanup_Req,
    ProjectBackupRepo_Cleanup_Res,
    ProjectBackupRepo_CreateOne_Req,
    ProjectBackupRepo_CreateOne_Res,
    ProjectBackupRepo_DeleteOne_Req,
    ProjectBackupRepo_DeleteOne_Res,
    ProjectBackupRepo_FindManyPaginated_Req,
    ProjectBackupRepo_FindManyPaginated_Res,
    ProjectBackupRepo_FindOneById_Req,
    ProjectBackupRepo_FindOneById_Res,
    ProjectBackupRepo_Sync_Req,
    ProjectBackupRepo_Sync_Res,
    ProjectBackupRepo_UpdateOne_Req,
    ProjectBackupRepo_UpdateOne_Res,
    ProjectBackupRepo_UpdatePassword_Req,
    ProjectBackupRepo_UpdatePassword_Res,
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

    async createOne(
        request: ProjectBackupRepo_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBackupRepo_CreateOne_Res, Error>> {
        const { projectID, env, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.post(getProjectBackupRepoBasePath(projectID, env), json, {
                    signal,
                }),
            ).pipe(
                map(response => this.validator.createOne(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: ProjectBackupRepo_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBackupRepo_UpdateOne_Res, Error>> {
        const { projectID, env, id, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.put(`${getProjectBackupRepoBasePath(projectID, env)}/${id}`, json, {
                    signal,
                }),
            ).pipe(
                map(response => this.validator.updateOne(response)),
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

    async updatePassword(
        request: ProjectBackupRepo_UpdatePassword_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBackupRepo_UpdatePassword_Res, Error>> {
        const { projectID, env, id, payload } = request.data;
        const json = {
            updateVer: payload.updateVer,
            currentPassword: payload.currentPassword,
            newPassword: payload.newPassword,
            inheritable: payload.inheritable,
            default: payload.default,
        };

        return lastValueFrom(
            from(
                this.client.v1.put(`${getProjectBackupRepoBasePath(projectID, env)}/${id}/password`, json, {
                    signal,
                }),
            ).pipe(
                map(response => this.validator.updatePassword(response)),
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

    async cleanup(
        request: ProjectBackupRepo_Cleanup_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBackupRepo_Cleanup_Res, Error>> {
        const { projectID, env, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`${getProjectBackupRepoBasePath(projectID, env)}/${id}/cleanup`, {}, { signal }),
            ).pipe(
                map(response => this.validator.cleanup(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async sync(
        request: ProjectBackupRepo_Sync_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBackupRepo_Sync_Res, Error>> {
        const { projectID, env, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`${getProjectBackupRepoBasePath(projectID, env)}/${id}/sync`, {}, { signal }),
            ).pipe(
                map(response => this.validator.sync(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
