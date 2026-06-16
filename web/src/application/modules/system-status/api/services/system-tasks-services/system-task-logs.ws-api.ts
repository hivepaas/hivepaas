import {
    BaseWebSocketApi,
    type WebSocketHandlers,
    type WebSocketSubscription,
    toWebSocketError,
} from "@infrastructure/websocket";
import { Err, Ok, type Result } from "oxide.ts";

import { session } from "@infrastructure/api";

import { buildSystemTaskLogsQueryParams } from "./system-tasks.api";
import type { SystemTasks_GetLogs_Req } from "./system-tasks.api.contracts";

export type SystemTaskLogsWs_StreamLogs_Req = SystemTasks_GetLogs_Req;

export type SystemTaskLogsWs_StreamLogs_Res = WebSocketSubscription;

export type SystemTaskLogsWsHandlers = WebSocketHandlers;

export class SystemTaskLogsWsApi extends BaseWebSocketApi {
    streamLogs(
        request: SystemTaskLogsWs_StreamLogs_Req,
        handlers: SystemTaskLogsWsHandlers,
        signal?: AbortSignal,
    ): Result<SystemTaskLogsWs_StreamLogs_Res, Error> {
        const { taskID } = request.data;
        const accessToken = session.getToken();

        if (!accessToken) {
            return Err(new Error("Access token not found."));
        }

        try {
            const url = this.client.buildUrl(
                `system/tasks/${encodeURIComponent(taskID)}/logs`,
                buildSystemTaskLogsQueryParams({
                    ...request.data,
                    follow: true,
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
            return Err(toWebSocketError(error, "Failed to stream system task logs."));
        }
    }
}
