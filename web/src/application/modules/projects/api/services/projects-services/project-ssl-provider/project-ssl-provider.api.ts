import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectSslProvider_CreateOne_Req,
    ProjectSslProvider_CreateOne_Res,
    ProjectSslProvider_DeleteOne_Req,
    ProjectSslProvider_DeleteOne_Res,
    ProjectSslProvider_FindManyPaginated_Req,
    ProjectSslProvider_FindManyPaginated_Res,
    ProjectSslProvider_FindOneById_Req,
    ProjectSslProvider_FindOneById_Res,
    ProjectSslProvider_UpdateOne_Req,
    ProjectSslProvider_UpdateOne_Res,
    ProjectSslProvider_UpdateStatus_Req,
    ProjectSslProvider_UpdateStatus_Res,
} from "./project-ssl-provider.api.contracts";
import type { ProjectSslProviderApiValidator } from "./project-ssl-provider.api.validator";

function getProjectSslProviderBasePath(projectID: string, env?: string): string {
    if (env) {
        return `/projects/${projectID}/${encodeURIComponent(env)}/ssl-providers`;
    }

    return `/projects/${projectID}/ssl-providers`;
}

export class ProjectSslProviderApi extends BaseApi {
    public constructor(private readonly validator: ProjectSslProviderApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectSslProvider_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslProvider_FindManyPaginated_Res, Error>> {
        const { projectID, env, search, pagination, sorting, kind } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        const params = {
            ...query.build(),
            ...(kind ? { kind } : {}),
        };

        return lastValueFrom(
            from(
                this.client.v1.get(getProjectSslProviderBasePath(projectID, env), {
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
        request: ProjectSslProvider_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslProvider_FindOneById_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/ssl-providers/${id}`, {
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
        request: ProjectSslProvider_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslProvider_CreateOne_Res, Error>> {
        const { projectID, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/ssl-providers`, json, {
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
        request: ProjectSslProvider_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslProvider_UpdateOne_Res, Error>> {
        const { projectID, id, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/ssl-providers/${id}`, json, {
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
        request: ProjectSslProvider_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslProvider_UpdateStatus_Res, Error>> {
        const { projectID, id, payload } = request.data;
        const json = { ...payload, inheritable: true };

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/ssl-providers/${id}/status`, json, {
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
        request: ProjectSslProvider_DeleteOne_Req,
    ): Promise<Result<ProjectSslProvider_DeleteOne_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/ssl-providers/${id}`)).pipe(
                map(response => this.validator.deleteOne(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
