import {
    BaseWebSocketApi,
    type WebSocketHandlers,
    type WebSocketSubscription,
    toWebSocketError,
} from "@infrastructure/websocket";
import { Err, Ok, type Result } from "oxide.ts";

import { session } from "@infrastructure/api";

import { buildAppLogsQueryParams } from "./app-logs.api";
import type { AppLogs_GetLogs_Req } from "./app-logs.api.contracts";

export type AppLogsWs_StreamLogs_Req = AppLogs_GetLogs_Req;

export type AppLogsWs_StreamLogs_Res = WebSocketSubscription;

export type AppLogsWsHandlers = WebSocketHandlers;

export class AppLogsWsApi extends BaseWebSocketApi {
    streamLogs(
        request: AppLogsWs_StreamLogs_Req,
        handlers: AppLogsWsHandlers,
        signal?: AbortSignal,
    ): Result<AppLogsWs_StreamLogs_Res, Error> {
        const { projectID, appID } = request.data;
        const accessToken = session.getToken();

        if (!accessToken) {
            return Err(new Error("Access token not found."));
        }

        try {
            const url = this.client.buildUrl(
                `projects/${encodeURIComponent(projectID)}/apps/${encodeURIComponent(appID)}/logs`,
                buildAppLogsQueryParams({ ...request.data, follow: true }),
            );

            return Ok(
                this.client.connect(url, handlers, {
                    signal,
                    closeOnError: true,
                    protocols: ["access_token", accessToken],
                }),
            );
        } catch (error) {
            return Err(toWebSocketError(error, "Failed to stream app logs."));
        }
    }
}
