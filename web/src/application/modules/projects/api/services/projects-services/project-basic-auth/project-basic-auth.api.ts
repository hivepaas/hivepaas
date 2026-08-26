import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectBasicAuth_CreateOne_Req,
    ProjectBasicAuth_CreateOne_Res,
    ProjectBasicAuth_DeleteOne_Req,
    ProjectBasicAuth_DeleteOne_Res,
    ProjectBasicAuth_FindManyPaginated_Req,
    ProjectBasicAuth_FindManyPaginated_Res,
    ProjectBasicAuth_FindOneById_Req,
    ProjectBasicAuth_FindOneById_Res,
    ProjectBasicAuth_UpdateOne_Req,
    ProjectBasicAuth_UpdateOne_Res,
    ProjectBasicAuth_UpdateStatus_Req,
    ProjectBasicAuth_UpdateStatus_Res,
} from "./project-basic-auth.api.contracts";
import type { ProjectBasicAuthApiValidator } from "./project-basic-auth.api.validator";

function getProjectBasicAuthBasePath(projectID: string, env?: string): string {
    if (env) {
        return `/projects/${projectID}/${encodeURIComponent(env)}/basic-auth`;
    }

    return `/projects/${projectID}/basic-auth`;
}

export class ProjectBasicAuthApi extends BaseApi {
    public constructor(private readonly validator: ProjectBasicAuthApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectBasicAuth_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBasicAuth_FindManyPaginated_Res, Error>> {
        const { projectID, env, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(getProjectBasicAuthBasePath(projectID, env), {
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
        request: ProjectBasicAuth_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBasicAuth_FindOneById_Res, Error>> {
        const { projectID, env, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`${getProjectBasicAuthBasePath(projectID, env)}/${id}`, {
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
        request: ProjectBasicAuth_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBasicAuth_CreateOne_Res, Error>> {
        const { projectID, env, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.post(getProjectBasicAuthBasePath(projectID, env), json, {
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
        request: ProjectBasicAuth_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBasicAuth_UpdateOne_Res, Error>> {
        const { projectID, env, id, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.put(`${getProjectBasicAuthBasePath(projectID, env)}/${id}`, json, {
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
        request: ProjectBasicAuth_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBasicAuth_UpdateStatus_Res, Error>> {
        const { projectID, env, id, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.put(`${getProjectBasicAuthBasePath(projectID, env)}/${id}/status`, json, {
                    signal,
                }),
            ).pipe(
                map(response => this.validator.updateStatus(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: ProjectBasicAuth_DeleteOne_Req): Promise<Result<ProjectBasicAuth_DeleteOne_Res, Error>> {
        const { projectID, env, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`${getProjectBasicAuthBasePath(projectID, env)}/${id}`)).pipe(
                map(response => this.validator.deleteOne(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
