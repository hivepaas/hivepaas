import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectCloudStorage_CreateOne_Req,
    ProjectCloudStorage_CreateOne_Res,
    ProjectCloudStorage_DeleteOne_Req,
    ProjectCloudStorage_DeleteOne_Res,
    ProjectCloudStorage_FindManyPaginated_Req,
    ProjectCloudStorage_FindManyPaginated_Res,
    ProjectCloudStorage_FindOneById_Req,
    ProjectCloudStorage_FindOneById_Res,
    ProjectCloudStorage_UpdateMeta_Req,
    ProjectCloudStorage_UpdateMeta_Res,
    ProjectCloudStorage_UpdateOne_Req,
    ProjectCloudStorage_UpdateOne_Res,
} from "./project-cloud-storage.api.contracts";
import type { ProjectCloudStorageApiValidator } from "./project-cloud-storage.api.validator";

function getProjectCloudStorageBasePath(projectID: string, env?: string): string {
    if (env) {
        return `/projects/${projectID}/${encodeURIComponent(env)}/cloud-storages`;
    }

    return `/projects/${projectID}/cloud-storages`;
}

export class ProjectCloudStorageApi extends BaseApi {
    public constructor(private readonly validator: ProjectCloudStorageApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectCloudStorage_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCloudStorage_FindManyPaginated_Res, Error>> {
        const { projectID, env, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(getProjectCloudStorageBasePath(projectID, env), {
                    params: query.build(),
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
        request: ProjectCloudStorage_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCloudStorage_FindOneById_Res, Error>> {
        const { projectID, env, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`${getProjectCloudStorageBasePath(projectID, env)}/${id}`, {
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
        request: ProjectCloudStorage_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCloudStorage_CreateOne_Res, Error>> {
        const { projectID, env, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(this.client.v1.post(getProjectCloudStorageBasePath(projectID, env), json, { signal })).pipe(
                map(response => this.validator.createOne(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: ProjectCloudStorage_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCloudStorage_UpdateOne_Res, Error>> {
        const { projectID, env, id, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.put(`${getProjectCloudStorageBasePath(projectID, env)}/${id}`, json, {
                    signal,
                }),
            ).pipe(
                map(response => this.validator.updateOne(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateMeta(
        request: ProjectCloudStorage_UpdateMeta_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectCloudStorage_UpdateMeta_Res, Error>> {
        const { projectID, env, id, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.put(`${getProjectCloudStorageBasePath(projectID, env)}/${id}/status`, json, {
                    signal,
                }),
            ).pipe(
                map(response => this.validator.updateMeta(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(
        request: ProjectCloudStorage_DeleteOne_Req,
    ): Promise<Result<ProjectCloudStorage_DeleteOne_Res, Error>> {
        const { projectID, env, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`${getProjectCloudStorageBasePath(projectID, env)}/${id}`)).pipe(
                map(response => this.validator.deleteOne(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
