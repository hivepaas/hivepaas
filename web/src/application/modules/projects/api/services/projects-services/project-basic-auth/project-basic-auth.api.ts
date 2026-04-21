import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { ProjectBasicAuthApiValidator } from "./project-basic-auth.api.validator";
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

export class ProjectBasicAuthApi extends BaseApi {
    public constructor(private readonly validator: ProjectBasicAuthApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectBasicAuth_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBasicAuth_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/basic-auth`, {
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
        request: ProjectBasicAuth_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBasicAuth_FindOneById_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/basic-auth/${id}`, {
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
        request: ProjectBasicAuth_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBasicAuth_CreateOne_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/basic-auth`, payload, {
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
        request: ProjectBasicAuth_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBasicAuth_UpdateOne_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/basic-auth/${id}`, payload, {
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
        request: ProjectBasicAuth_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectBasicAuth_UpdateStatus_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/basic-auth/${id}/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: ProjectBasicAuth_DeleteOne_Req): Promise<Result<ProjectBasicAuth_DeleteOne_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/basic-auth/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
