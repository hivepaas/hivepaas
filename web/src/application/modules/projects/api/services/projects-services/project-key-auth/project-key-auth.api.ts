import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectKeyAuth_CreateOne_Req,
    ProjectKeyAuth_CreateOne_Res,
    ProjectKeyAuth_DeleteOne_Req,
    ProjectKeyAuth_DeleteOne_Res,
    ProjectKeyAuth_FindManyPaginated_Req,
    ProjectKeyAuth_FindManyPaginated_Res,
    ProjectKeyAuth_FindOneById_Req,
    ProjectKeyAuth_FindOneById_Res,
    ProjectKeyAuth_UpdateOne_Req,
    ProjectKeyAuth_UpdateOne_Res,
    ProjectKeyAuth_UpdateStatus_Req,
    ProjectKeyAuth_UpdateStatus_Res,
} from "./project-key-auth.api.contracts";
import type { ProjectKeyAuthApiValidator } from "./project-key-auth.api.validator";

function getProjectKeyAuthBasePath(projectID: string, env?: string): string {
    if (env) {
        return `/projects/${projectID}/${encodeURIComponent(env)}/key-auth`;
    }

    return `/projects/${projectID}/key-auth`;
}

export class ProjectKeyAuthApi extends BaseApi {
    public constructor(private readonly validator: ProjectKeyAuthApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectKeyAuth_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectKeyAuth_FindManyPaginated_Res, Error>> {
        const { projectID, env, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(getProjectKeyAuthBasePath(projectID, env), {
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
        request: ProjectKeyAuth_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectKeyAuth_FindOneById_Res, Error>> {
        const { projectID, env, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`${getProjectKeyAuthBasePath(projectID, env)}/${id}`, {
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
        request: ProjectKeyAuth_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectKeyAuth_CreateOne_Res, Error>> {
        const { projectID, env, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(getProjectKeyAuthBasePath(projectID, env), payload, {
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
        request: ProjectKeyAuth_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectKeyAuth_UpdateOne_Res, Error>> {
        const { projectID, env, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`${getProjectKeyAuthBasePath(projectID, env)}/${id}`, payload, {
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
        request: ProjectKeyAuth_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectKeyAuth_UpdateStatus_Res, Error>> {
        const { projectID, env, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`${getProjectKeyAuthBasePath(projectID, env)}/${id}/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(response => this.validator.updateStatus(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: ProjectKeyAuth_DeleteOne_Req): Promise<Result<ProjectKeyAuth_DeleteOne_Res, Error>> {
        const { projectID, env, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`${getProjectKeyAuthBasePath(projectID, env)}/${id}`)).pipe(
                map(response => this.validator.deleteOne(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
