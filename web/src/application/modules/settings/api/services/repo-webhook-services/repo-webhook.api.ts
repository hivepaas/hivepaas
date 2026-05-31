import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    RepoWebhook_CreateOne_Req,
    RepoWebhook_CreateOne_Res,
    RepoWebhook_DeleteOne_Req,
    RepoWebhook_DeleteOne_Res,
    RepoWebhook_FindManyPaginated_Req,
    RepoWebhook_FindManyPaginated_Res,
    RepoWebhook_FindOneById_Req,
    RepoWebhook_FindOneById_Res,
    RepoWebhook_UpdateOne_Req,
    RepoWebhook_UpdateOne_Res,
    RepoWebhook_UpdateStatus_Req,
    RepoWebhook_UpdateStatus_Res,
} from "./repo-webhook.api.contracts";
import type { RepoWebhookApiValidator } from "./repo-webhook.api.validator";

export class RepoWebhookApi extends BaseApi {
    public constructor(private readonly validator: RepoWebhookApiValidator) {
        super();
    }

    async findManyPaginated(
        request: RepoWebhook_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<RepoWebhook_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/settings/repo-webhooks", {
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
        request: RepoWebhook_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<RepoWebhook_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/settings/repo-webhooks/${id}`, {
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
        request: RepoWebhook_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<RepoWebhook_CreateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/repo-webhooks", payload, {
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
        request: RepoWebhook_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<RepoWebhook_UpdateOne_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/repo-webhooks/${id}`, payload, {
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
        request: RepoWebhook_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<RepoWebhook_UpdateStatus_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/repo-webhooks/${id}/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: RepoWebhook_DeleteOne_Req): Promise<Result<RepoWebhook_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/settings/repo-webhooks/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
