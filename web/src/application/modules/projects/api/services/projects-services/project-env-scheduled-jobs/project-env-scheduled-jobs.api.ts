import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import { toUpsertPayload } from "~/projects/api/services/project-apps-services/scheduled-jobs/app-scheduled-jobs.api";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    EnvScheduledJobs_CreateOne_Req,
    EnvScheduledJobs_CreateOne_Res,
    EnvScheduledJobs_DeleteOne_Req,
    EnvScheduledJobs_DeleteOne_Res,
    EnvScheduledJobs_FindManyPaginated_Req,
    EnvScheduledJobs_FindManyPaginated_Res,
    EnvScheduledJobs_FindOneById_Req,
    EnvScheduledJobs_FindOneById_Res,
    EnvScheduledJobs_RunNow_Req,
    EnvScheduledJobs_RunNow_Res,
    EnvScheduledJobs_UpdateOne_Req,
    EnvScheduledJobs_UpdateOne_Res,
    EnvScheduledJobs_UpdateStatus_Req,
    EnvScheduledJobs_UpdateStatus_Res,
} from "./project-env-scheduled-jobs.api.contracts";
import type { EnvScheduledJobsApiValidator } from "./project-env-scheduled-jobs.api.validator";

function basePath(projectID: string, env: string): string {
    return `/projects/${projectID}/${encodeURIComponent(env)}/sched-jobs`;
}

/** An env's scheduled jobs: its own (job sequences), and, in the list, those of its apps. */
export class EnvScheduledJobsApi extends BaseApi {
    public constructor(private readonly validator: EnvScheduledJobsApiValidator) {
        super();
    }

    async findManyPaginated(
        request: EnvScheduledJobs_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<EnvScheduledJobs_FindManyPaginated_Res, Error>> {
        const { projectID, env, search, pagination, sorting, jobTypes, appId } = request.data;
        const query = this.queryBuilder.getInstance();
        query
            .pagination(pagination)
            .sorting(sorting)
            .search(search)
            .filterBy({ jobType: jobTypes, appId: [appId] });

        return lastValueFrom(
            from(this.client.v1.get(basePath(projectID, env), { params: query.build(), signal })).pipe(
                map(this.validator.findManyPaginated),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findOneById(
        request: EnvScheduledJobs_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<EnvScheduledJobs_FindOneById_Res, Error>> {
        const { projectID, env, scheduledJobID } = request.data;

        return lastValueFrom(
            from(this.client.v1.get(`${basePath(projectID, env)}/${scheduledJobID}`, { signal })).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async createOne(
        request: EnvScheduledJobs_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<EnvScheduledJobs_CreateOne_Res, Error>> {
        const { projectID, env, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(basePath(projectID, env), toUpsertPayload(payload), { signal })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: EnvScheduledJobs_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<EnvScheduledJobs_UpdateOne_Res, Error>> {
        const { projectID, env, scheduledJobID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`${basePath(projectID, env)}/${scheduledJobID}`, toUpsertPayload(payload), {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateStatus(
        request: EnvScheduledJobs_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<EnvScheduledJobs_UpdateStatus_Res, Error>> {
        const { projectID, env, scheduledJobID, payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put(`${basePath(projectID, env)}/${scheduledJobID}/status`, payload, { signal })).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: EnvScheduledJobs_DeleteOne_Req): Promise<Result<EnvScheduledJobs_DeleteOne_Res, Error>> {
        const { projectID, env, scheduledJobID } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`${basePath(projectID, env)}/${scheduledJobID}`)).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async runNow(
        request: EnvScheduledJobs_RunNow_Req,
        signal?: AbortSignal,
    ): Promise<Result<EnvScheduledJobs_RunNow_Res, Error>> {
        const { projectID, env, scheduledJobID } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(`${basePath(projectID, env)}/${scheduledJobID}/exec`, {}, { signal })).pipe(
                map(this.validator.runNow),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
