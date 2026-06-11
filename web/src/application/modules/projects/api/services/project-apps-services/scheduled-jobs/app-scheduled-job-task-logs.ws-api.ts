import {
    BaseWebSocketApi,
    type WebSocketHandlers,
    type WebSocketSubscription,
    toWebSocketError,
} from "@infrastructure/websocket";
import { Err, Ok, type Result } from "oxide.ts";

import { session } from "@infrastructure/api";

import { buildAppScheduledJobTaskLogsQueryParams } from "./app-scheduled-jobs.api";
import type { AppScheduledJobTasks_GetLogs_Req } from "./app-scheduled-jobs.api.contracts";

const DEFAULT_LOG_TAIL = 100;

export type AppScheduledJobTaskLogsWs_StreamLogs_Req = AppScheduledJobTasks_GetLogs_Req;

export type AppScheduledJobTaskLogsWs_StreamLogs_Res = WebSocketSubscription;

export type AppScheduledJobTaskLogsWsHandlers = WebSocketHandlers;

export class AppScheduledJobTaskLogsWsApi extends BaseWebSocketApi {
    streamLogs(
        request: AppScheduledJobTaskLogsWs_StreamLogs_Req,
        handlers: AppScheduledJobTaskLogsWsHandlers,
        signal?: AbortSignal,
    ): Result<AppScheduledJobTaskLogsWs_StreamLogs_Res, Error> {
        const { projectID, appID, scheduledJobID, taskID } = request.data;
        const accessToken = session.getToken();

        if (!accessToken) {
            return Err(new Error("Access token not found."));
        }

        try {
            const url = this.client.buildUrl(
                `projects/${encodeURIComponent(projectID)}/apps/${encodeURIComponent(
                    appID,
                )}/sched-jobs/${encodeURIComponent(scheduledJobID)}/tasks/${encodeURIComponent(taskID)}/logs`,
                buildAppScheduledJobTaskLogsQueryParams({
                    ...request.data,
                    follow: true,
                    tail: request.data.tail ?? DEFAULT_LOG_TAIL,
                }),
            );

            return Ok(
                this.client.connect(url, handlers, {
                    signal,
                    closeOnError: true,
                    protocols: ["access_token", accessToken],
                }),
            );
        } catch (error) {
            return Err(toWebSocketError(error, "Failed to stream scheduled job task logs."));
        }
    }
}
