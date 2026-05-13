import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectImService_CreateOne_Req,
    ProjectImService_CreateOne_Res,
    ProjectImService_DeleteOne_Req,
    ProjectImService_DeleteOne_Res,
    ProjectImService_FindManyPaginated_Req,
    ProjectImService_FindManyPaginated_Res,
    ProjectImService_FindOneById_Req,
    ProjectImService_FindOneById_Res,
    ProjectImService_UpdateOne_Req,
    ProjectImService_UpdateOne_Res,
    ProjectImService_UpdateStatus_Req,
    ProjectImService_UpdateStatus_Res,
} from "./project-im-service.api.contracts";
import type { ProjectImServiceApiValidator } from "./project-im-service.api.validator";

export class ProjectImServiceApi extends BaseApi {
    public constructor(private readonly validator: ProjectImServiceApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectImService_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectImService_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/im-services`, {
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
        request: ProjectImService_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectImService_FindOneById_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/im-services/${id}`, {
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
        request: ProjectImService_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectImService_CreateOne_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/im-services`, payload, {
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
        request: ProjectImService_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectImService_UpdateOne_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/im-services/${id}`, payload, {
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
        request: ProjectImService_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectImService_UpdateStatus_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/im-services/${id}/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: ProjectImService_DeleteOne_Req): Promise<Result<ProjectImService_DeleteOne_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/im-services/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
