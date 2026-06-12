import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    AppLogsApiValidator,
    AppLogs_GetInfo_Req,
    AppLogs_GetInfo_Res,
    AppLogs_GetLogs_Req,
    AppLogs_GetLogs_Res,
} from "~/projects/api/services";

import { BaseApi, parseApiError } from "@infrastructure/api";

export type AppLogsQueryParams = Record<string, string | number | boolean>;

export function buildAppLogsQueryParams(request: AppLogs_GetLogs_Req["data"]): AppLogsQueryParams {
    const { taskId, follow, tail, since, duration, timestamps } = request;

    return {
        ...(taskId ? { taskId } : {}),
        ...(follow === undefined ? {} : { follow }),
        ...(tail === undefined ? {} : { tail }),
        ...(since ? { since: since.toISOString() } : {}),
        ...(duration ? { duration } : {}),
        ...(timestamps === undefined ? {} : { timestamps }),
    };
}

export class AppLogsApi extends BaseApi {
    public constructor(private readonly validator: AppLogsApiValidator) {
        super();
    }

    async getInfo(request: AppLogs_GetInfo_Req, signal?: AbortSignal): Promise<Result<AppLogs_GetInfo_Res, Error>> {
        const { projectID, appID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/logs/info`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.getInfo),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async getLogs(request: AppLogs_GetLogs_Req, signal?: AbortSignal): Promise<Result<AppLogs_GetLogs_Res, Error>> {
        const { projectID, appID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/logs`, {
                    params: buildAppLogsQueryParams({ ...request.data, follow: false }),
                    signal,
                }),
            ).pipe(
                map(this.validator.getLogs),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
