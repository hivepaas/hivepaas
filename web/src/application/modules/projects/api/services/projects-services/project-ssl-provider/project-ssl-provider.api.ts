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

export class ProjectSslProviderApi extends BaseApi {
    public constructor(private readonly validator: ProjectSslProviderApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectSslProvider_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslProvider_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/ssl-providers`, {
                    params: query.build(),
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
                map(this.validator.findOneById),
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

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/ssl-providers`, payload, {
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
        request: ProjectSslProvider_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslProvider_UpdateOne_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/ssl-providers/${id}`, payload, {
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
        request: ProjectSslProvider_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSslProvider_UpdateStatus_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/ssl-providers/${id}/status`, payload, {
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
        request: ProjectSslProvider_DeleteOne_Req,
    ): Promise<Result<ProjectSslProvider_DeleteOne_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/ssl-providers/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
