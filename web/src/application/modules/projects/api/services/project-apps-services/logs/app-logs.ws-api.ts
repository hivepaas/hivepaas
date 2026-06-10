import {
    BaseWebSocketApi,
    type WebSocketHandlers,
    type WebSocketSubscription,
    toWebSocketError,
} from "@infrastructure/websocket";
import { Err, Ok, type Result, match } from "oxide.ts";

import type { AppLogsApi } from "./app-logs.api";
import { buildAppLogsQueryParams } from "./app-logs.api";
import type { AppLogs_GetLogs_Req } from "./app-logs.api.contracts";

export type AppLogsWs_StreamLogs_Req = AppLogs_GetLogs_Req;

export type AppLogsWs_StreamLogs_Res = WebSocketSubscription;

export type AppLogsWsHandlers = WebSocketHandlers;

export class AppLogsWsApi extends BaseWebSocketApi {
    public constructor(private readonly appLogsApi: AppLogsApi) {
        super();
    }

    async streamLogs(
        request: AppLogsWs_StreamLogs_Req,
        handlers: AppLogsWsHandlers,
        signal?: AbortSignal,
    ): Promise<Result<AppLogsWs_StreamLogs_Res, Error>> {
        const { projectID, appID } = request.data;
        const logsTokenResult = await this.appLogsApi.getToken({ data: { projectID, appID } }, signal);

        return match(logsTokenResult, {
            Ok: logsTokenResponse => {
                try {
                    const url = this.client.buildUrl(
                        `projects/${encodeURIComponent(projectID)}/apps/${encodeURIComponent(appID)}/logs`,
                        {
                            ...buildAppLogsQueryParams({ ...request.data, follow: true }),
                            token: logsTokenResponse.data.token,
                        },
                    );

                    return Ok(this.client.connect(url, handlers, { signal, closeOnError: true }));
                } catch (error) {
                    return Err(toWebSocketError(error, "Failed to stream app logs."));
                }
            },
            Err: error => Err(error),
        });
    }
}
