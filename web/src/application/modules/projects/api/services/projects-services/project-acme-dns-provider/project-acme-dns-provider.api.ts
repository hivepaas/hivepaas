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

export class ProjectAcmeDnsProviderApi extends BaseApi {
    public constructor(private readonly validator: ProjectAcmeDnsProviderApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectAcmeDnsProvider_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectAcmeDnsProvider_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting, kind } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        const params = {
            ...query.build(),
            ...(kind ? { kind } : {}),
        };

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/acme-dns-providers`, {
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
                map(this.validator.findOneById),
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
                map(this.validator.createOne),
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
                map(this.validator.updateOne),
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
                map(this.validator.updateStatus),
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
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
