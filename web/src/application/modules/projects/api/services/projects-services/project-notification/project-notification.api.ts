import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectNotification_CreateOne_Req,
    ProjectNotification_CreateOne_Res,
    ProjectNotification_DeleteOne_Req,
    ProjectNotification_DeleteOne_Res,
    ProjectNotification_FindManyPaginated_Req,
    ProjectNotification_FindManyPaginated_Res,
    ProjectNotification_FindOneById_Req,
    ProjectNotification_FindOneById_Res,
    ProjectNotification_UpdateOne_Req,
    ProjectNotification_UpdateOne_Res,
    ProjectNotification_UpdateStatus_Req,
    ProjectNotification_UpdateStatus_Res,
} from "./project-notification.api.contracts";
import type { ProjectNotificationApiValidator } from "./project-notification.api.validator";

export class ProjectNotificationApi extends BaseApi {
    public constructor(private readonly validator: ProjectNotificationApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectNotification_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectNotification_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/notifications`, {
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
        request: ProjectNotification_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectNotification_FindOneById_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/notifications/${id}`, {
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
        request: ProjectNotification_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectNotification_CreateOne_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(`/projects/${projectID}/notifications`, payload, { signal })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: ProjectNotification_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectNotification_UpdateOne_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/projects/${projectID}/notifications/${id}`, payload, { signal })).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateStatus(
        request: ProjectNotification_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectNotification_UpdateStatus_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`/projects/${projectID}/notifications/${id}/status`, payload, { signal })).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(
        request: ProjectNotification_DeleteOne_Req,
    ): Promise<Result<ProjectNotification_DeleteOne_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/notifications/${id}`)).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
