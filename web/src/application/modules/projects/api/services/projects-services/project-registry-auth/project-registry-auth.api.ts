import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectRegistryAuth_CreateOne_Req,
    ProjectRegistryAuth_CreateOne_Res,
    ProjectRegistryAuth_DeleteOne_Req,
    ProjectRegistryAuth_DeleteOne_Res,
    ProjectRegistryAuth_FindManyPaginated_Req,
    ProjectRegistryAuth_FindManyPaginated_Res,
    ProjectRegistryAuth_FindOneById_Req,
    ProjectRegistryAuth_FindOneById_Res,
    ProjectRegistryAuth_UpdateMeta_Req,
    ProjectRegistryAuth_UpdateMeta_Res,
    ProjectRegistryAuth_UpdateOne_Req,
    ProjectRegistryAuth_UpdateOne_Res,
} from "./project-registry-auth.api.contracts";
import type { ProjectRegistryAuthApiValidator } from "./project-registry-auth.api.validator";

function getProjectRegistryAuthBasePath(projectID: string, env?: string): string {
    if (env) {
        return `/projects/${projectID}/${encodeURIComponent(env)}/registry-auth`;
    }

    return `/projects/${projectID}/registry-auth`;
}

export class ProjectRegistryAuthApi extends BaseApi {
    public constructor(private readonly validator: ProjectRegistryAuthApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectRegistryAuth_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectRegistryAuth_FindManyPaginated_Res, Error>> {
        const { projectID, env, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(getProjectRegistryAuthBasePath(projectID, env), {
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
        request: ProjectRegistryAuth_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectRegistryAuth_FindOneById_Res, Error>> {
        const { projectID, env, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`${getProjectRegistryAuthBasePath(projectID, env)}/${id}`, {
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
        request: ProjectRegistryAuth_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectRegistryAuth_CreateOne_Res, Error>> {
        const { projectID, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/registry-auth`, json, {
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
        request: ProjectRegistryAuth_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectRegistryAuth_UpdateOne_Res, Error>> {
        const { projectID, id, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/registry-auth/${id}`, json, {
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
        request: ProjectRegistryAuth_UpdateMeta_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectRegistryAuth_UpdateMeta_Res, Error>> {
        const { projectID, id, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/registry-auth/${id}/status`, json, {
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
        request: ProjectRegistryAuth_DeleteOne_Req,
    ): Promise<Result<ProjectRegistryAuth_DeleteOne_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/registry-auth/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
