import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectRepoWebhook_CreateOne_Req,
    ProjectRepoWebhook_CreateOne_Res,
    ProjectRepoWebhook_DeleteOne_Req,
    ProjectRepoWebhook_DeleteOne_Res,
    ProjectRepoWebhook_FindManyPaginated_Req,
    ProjectRepoWebhook_FindManyPaginated_Res,
    ProjectRepoWebhook_FindOneById_Req,
    ProjectRepoWebhook_FindOneById_Res,
    ProjectRepoWebhook_UpdateOne_Req,
    ProjectRepoWebhook_UpdateOne_Res,
    ProjectRepoWebhook_UpdateStatus_Req,
    ProjectRepoWebhook_UpdateStatus_Res,
} from "./project-repo-webhook.api.contracts";
import type { ProjectRepoWebhookApiValidator } from "./project-repo-webhook.api.validator";

export class ProjectRepoWebhookApi extends BaseApi {
    public constructor(private readonly validator: ProjectRepoWebhookApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectRepoWebhook_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectRepoWebhook_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/repo-webhooks`, {
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
        request: ProjectRepoWebhook_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectRepoWebhook_FindOneById_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/repo-webhooks/${id}`, {
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
        request: ProjectRepoWebhook_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectRepoWebhook_CreateOne_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/repo-webhooks`, payload, {
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
        request: ProjectRepoWebhook_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectRepoWebhook_UpdateOne_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/repo-webhooks/${id}`, payload, {
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
        request: ProjectRepoWebhook_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectRepoWebhook_UpdateStatus_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/repo-webhooks/${id}/status`, payload, {
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
        request: ProjectRepoWebhook_DeleteOne_Req,
    ): Promise<Result<ProjectRepoWebhook_DeleteOne_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/repo-webhooks/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
