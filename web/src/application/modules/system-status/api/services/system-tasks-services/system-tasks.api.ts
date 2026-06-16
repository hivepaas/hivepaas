import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    SystemTasks_Cancel_Req,
    SystemTasks_Cancel_Res,
    SystemTasks_FindManyPaginated_Req,
    SystemTasks_FindManyPaginated_Res,
    SystemTasks_FindOneById_Req,
    SystemTasks_FindOneById_Res,
    SystemTasks_GetLogs_Req,
} from "./system-tasks.api.contracts";
import type { SystemTasksApiValidator } from "./system-tasks.api.validator";

export type SystemTaskLogsQueryParams = Record<string, string | number | boolean>;

export function buildSystemTaskLogsQueryParams(request: SystemTasks_GetLogs_Req["data"]): SystemTaskLogsQueryParams {
    const { follow, tail, since, duration, timestamps } = request;

    return {
        ...(follow === undefined ? {} : { follow }),
        ...(tail === undefined ? {} : { tail }),
        ...(since ? { since: since.toISOString() } : {}),
        ...(duration ? { duration } : {}),
        ...(timestamps === undefined ? {} : { timestamps }),
    };
}

export class SystemTasksApi extends BaseApi {
    public constructor(private readonly validator: SystemTasksApiValidator) {
        super();
    }

    async findManyPaginated(
        request: SystemTasks_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemTasks_FindManyPaginated_Res, Error>> {
        const { pagination, sorting, search, jobName, targetId, status } = request.data;
        const query = this.queryBuilder.getInstance();

        query
            .pagination(pagination)
            .sorting(sorting)
            .search(search)
            .filterBy({
                jobName: jobName ? [jobName] : undefined,
                targetId,
                status,
            });

        return lastValueFrom(
            from(
                this.client.v1.get("/system/tasks", {
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
        request: SystemTasks_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemTasks_FindOneById_Res, Error>> {
        const { taskID } = request.data;

        return lastValueFrom(
            from(this.client.v1.get(`/system/tasks/${taskID}`, { signal })).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async cancel(request: SystemTasks_Cancel_Req): Promise<Result<SystemTasks_Cancel_Res, Error>> {
        const { taskID } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(`/system/tasks/${taskID}/cancel`, {})).pipe(
                map(this.validator.cancel),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
