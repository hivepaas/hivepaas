import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    AppHealthChecksApiValidator,
    AppHealthChecks_CreateOne_Req,
    AppHealthChecks_CreateOne_Res,
    AppHealthChecks_DeleteOne_Req,
    AppHealthChecks_DeleteOne_Res,
    AppHealthChecks_FindManyPaginated_Req,
    AppHealthChecks_FindManyPaginated_Res,
    AppHealthChecks_FindOneById_Req,
    AppHealthChecks_FindOneById_Res,
    AppHealthChecks_UpdateOne_Req,
    AppHealthChecks_UpdateOne_Res,
    AppHealthChecks_UpdateStatus_Req,
    AppHealthChecks_UpdateStatus_Res,
    AppHealthChecks_Upsert_Payload,
} from "~/projects/api/services/project-apps-services";

import { BaseApi, parseApiError } from "@infrastructure/api";

function toObjectIdPayload(ref?: { id: string }) {
    return {
        id: ref?.id ?? "",
    };
}

function toUpsertPayload(
    payload: AppHealthChecks_Upsert_Payload | (AppHealthChecks_Upsert_Payload & { updateVer: number }),
) {
    return {
        ...payload,
        notification: {
            ...payload.notification,
            success: toObjectIdPayload(payload.notification.success),
            failure: toObjectIdPayload(payload.notification.failure),
        },
    };
}

export class AppHealthChecksApi extends BaseApi {
    public constructor(private readonly validator: AppHealthChecksApiValidator) {
        super();
    }

    async findManyPaginated(
        request: AppHealthChecks_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppHealthChecks_FindManyPaginated_Res, Error>> {
        const { projectID, appID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/healthchecks`, {
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
        request: AppHealthChecks_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppHealthChecks_FindOneById_Res, Error>> {
        const { projectID, appID, healthCheckID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/healthchecks/${healthCheckID}`, {
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
        request: AppHealthChecks_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppHealthChecks_CreateOne_Res, Error>> {
        const { projectID, appID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/apps/${appID}/healthchecks`, toUpsertPayload(payload), {
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
        request: AppHealthChecks_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppHealthChecks_UpdateOne_Res, Error>> {
        const { projectID, appID, healthCheckID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(
                    `/projects/${projectID}/apps/${appID}/healthchecks/${healthCheckID}`,
                    toUpsertPayload(payload),
                    { signal },
                ),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateStatus(
        request: AppHealthChecks_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppHealthChecks_UpdateStatus_Res, Error>> {
        const { projectID, appID, healthCheckID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(
                    `/projects/${projectID}/apps/${appID}/healthchecks/${healthCheckID}/status`,
                    payload,
                    { signal },
                ),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: AppHealthChecks_DeleteOne_Req): Promise<Result<AppHealthChecks_DeleteOne_Res, Error>> {
        const { projectID, appID, healthCheckID } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/apps/${appID}/healthchecks/${healthCheckID}`)).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
