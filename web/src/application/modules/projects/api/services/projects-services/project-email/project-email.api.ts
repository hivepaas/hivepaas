import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectEmail_CreateOne_Req,
    ProjectEmail_CreateOne_Res,
    ProjectEmail_DeleteOne_Req,
    ProjectEmail_DeleteOne_Res,
    ProjectEmail_FindManyPaginated_Req,
    ProjectEmail_FindManyPaginated_Res,
    ProjectEmail_FindOneById_Req,
    ProjectEmail_FindOneById_Res,
    ProjectEmail_UpdateOne_Req,
    ProjectEmail_UpdateOne_Res,
    ProjectEmail_UpdateStatus_Req,
    ProjectEmail_UpdateStatus_Res,
} from "./project-email.api.contracts";
import type { ProjectEmailApiValidator } from "./project-email.api.validator";

export class ProjectEmailApi extends BaseApi {
    public constructor(private readonly validator: ProjectEmailApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectEmail_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectEmail_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/emails`, {
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
        request: ProjectEmail_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectEmail_FindOneById_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/emails/${id}`, {
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
        request: ProjectEmail_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectEmail_CreateOne_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/emails`, payload, {
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
        request: ProjectEmail_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectEmail_UpdateOne_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/emails/${id}`, payload, {
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
        request: ProjectEmail_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectEmail_UpdateStatus_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/emails/${id}/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: ProjectEmail_DeleteOne_Req): Promise<Result<ProjectEmail_DeleteOne_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/emails/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
