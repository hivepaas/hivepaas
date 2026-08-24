import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectAcmeDnsProvider_CreateOne_Req,
    ProjectAcmeDnsProvider_CreateOne_Res,
    ProjectAcmeDnsProvider_DeleteOne_Req,
    ProjectAcmeDnsProvider_DeleteOne_Res,
    ProjectAcmeDnsProvider_FindManyPaginated_Req,
    ProjectAcmeDnsProvider_FindManyPaginated_Res,
    ProjectAcmeDnsProvider_FindOneById_Req,
    ProjectAcmeDnsProvider_FindOneById_Res,
    ProjectAcmeDnsProvider_UpdateOne_Req,
    ProjectAcmeDnsProvider_UpdateOne_Res,
    ProjectAcmeDnsProvider_UpdateStatus_Req,
    ProjectAcmeDnsProvider_UpdateStatus_Res,
} from "./project-acme-dns-provider.api.contracts";
import type { ProjectAcmeDnsProviderApiValidator } from "./project-acme-dns-provider.api.validator";

function getProjectAcmeDnsProviderBasePath(projectID: string, env?: string): string {
    if (env) {
        return `/projects/${projectID}/${encodeURIComponent(env)}/acme-dns-providers`;
    }

    return `/projects/${projectID}/acme-dns-providers`;
}

export class ProjectAcmeDnsProviderApi extends BaseApi {
    public constructor(private readonly validator: ProjectAcmeDnsProviderApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectAcmeDnsProvider_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAcmeDnsProvider_FindManyPaginated_Res, Error>> {
        const { projectID, env, search, pagination, sorting, kind } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        const params = {
            ...query.build(),
            ...(kind ? { kind } : {}),
        };

        return lastValueFrom(
            from(
                this.client.v1.get(getProjectAcmeDnsProviderBasePath(projectID, env), {
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
        request: ProjectAcmeDnsProvider_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAcmeDnsProvider_FindOneById_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/acme-dns-providers/${id}`, {
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
        request: ProjectAcmeDnsProvider_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAcmeDnsProvider_CreateOne_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/acme-dns-providers`, payload, {
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
        request: ProjectAcmeDnsProvider_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAcmeDnsProvider_UpdateOne_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/acme-dns-providers/${id}`, payload, {
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
        request: ProjectAcmeDnsProvider_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAcmeDnsProvider_UpdateStatus_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/acme-dns-providers/${id}/status`, payload, {
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
        request: ProjectAcmeDnsProvider_DeleteOne_Req,
    ): Promise<Result<ProjectAcmeDnsProvider_DeleteOne_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/acme-dns-providers/${id}`)).pipe(
                map(response => this.validator.deleteOne(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
