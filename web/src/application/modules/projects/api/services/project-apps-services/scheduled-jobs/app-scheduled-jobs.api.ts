import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    AppScheduledJobsApiValidator,
    AppScheduledJobs_CreateOne_Req,
    AppScheduledJobs_CreateOne_Res,
    AppScheduledJobs_DeleteOne_Req,
    AppScheduledJobs_DeleteOne_Res,
    AppScheduledJobs_FindManyPaginated_Req,
    AppScheduledJobs_FindManyPaginated_Res,
    AppScheduledJobs_FindOneById_Req,
    AppScheduledJobs_FindOneById_Res,
    AppScheduledJobs_RunNow_Req,
    AppScheduledJobs_RunNow_Res,
    AppScheduledJobs_UpdateOne_Req,
    AppScheduledJobs_UpdateOne_Res,
    AppScheduledJobs_UpdateStatus_Req,
    AppScheduledJobs_UpdateStatus_Res,
    AppScheduledJobs_Upsert_Payload,
} from "~/projects/api/services/project-apps-services";

import { BaseApi, parseApiError } from "@infrastructure/api";

function toObjectIdPayload(ref?: { id: string } | null) {
    return {
        id: ref?.id ?? "",
    };
}

function toOptionalObjectIdPayload(ref?: { id: string } | null) {
    if (!ref?.id) {
        return undefined;
    }

    return {
        id: ref.id,
    };
}

function toUpsertPayload(
    payload: AppScheduledJobs_Upsert_Payload | (AppScheduledJobs_Upsert_Payload & { updateVer: number }),
) {
    const { success, failure, ...notification } = payload.notification;
    const successPayload = toOptionalObjectIdPayload(success);
    const failurePayload = toOptionalObjectIdPayload(failure);

    return {
        ...payload,
        app: toObjectIdPayload(payload.app),
        notification: {
            ...notification,
            ...(successPayload ? { success: successPayload } : {}),
            ...(failurePayload ? { failure: failurePayload } : {}),
        },
    };
}

export class AppScheduledJobsApi extends BaseApi {
    public constructor(private readonly validator: AppScheduledJobsApiValidator) {
        super();
    }

    async findManyPaginated(
        request: AppScheduledJobs_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppScheduledJobs_FindManyPaginated_Res, Error>> {
        const { projectID, appID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/sched-jobs`, {
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
        request: AppScheduledJobs_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppScheduledJobs_FindOneById_Res, Error>> {
        const { projectID, appID, scheduledJobID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/sched-jobs/${scheduledJobID}`, {
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
        request: AppScheduledJobs_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppScheduledJobs_CreateOne_Res, Error>> {
        const { projectID, appID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/apps/${appID}/sched-jobs`, toUpsertPayload(payload), {
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
        request: AppScheduledJobs_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppScheduledJobs_UpdateOne_Res, Error>> {
        const { projectID, appID, scheduledJobID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(
                    `/projects/${projectID}/apps/${appID}/sched-jobs/${scheduledJobID}`,
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
        request: AppScheduledJobs_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppScheduledJobs_UpdateStatus_Res, Error>> {
        const { projectID, appID, scheduledJobID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(
                    `/projects/${projectID}/apps/${appID}/sched-jobs/${scheduledJobID}/status`,
                    payload,
                    {
                        signal,
                    },
                ),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: AppScheduledJobs_DeleteOne_Req): Promise<Result<AppScheduledJobs_DeleteOne_Res, Error>> {
        const { projectID, appID, scheduledJobID } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/apps/${appID}/sched-jobs/${scheduledJobID}`)).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async runNow(
        request: AppScheduledJobs_RunNow_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppScheduledJobs_RunNow_Res, Error>> {
        const { projectID, appID, scheduledJobID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(
                    `/projects/${projectID}/apps/${appID}/sched-jobs/${scheduledJobID}/exec`,
                    {},
                    { signal },
                ),
            ).pipe(
                map(this.validator.runNow),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
