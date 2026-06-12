import {
    BaseWebSocketApi,
    type WebSocketHandlers,
    type WebSocketSubscription,
    toWebSocketError,
} from "@infrastructure/websocket";
import { Err, Ok, type Result } from "oxide.ts";

import { session } from "@infrastructure/api";

const DEFAULT_LOG_TAIL = 100;

export type AppDeploymentLogsWs_StreamLogs_Req = {
    data: {
        projectID: string;
        appID: string;
        deploymentID: string;
    };
};

export type AppDeploymentLogsWs_StreamLogs_Res = WebSocketSubscription;

export type AppDeploymentLogsWsHandlers = WebSocketHandlers;

export class AppDeploymentLogsWsApi extends BaseWebSocketApi {
    streamLogs(
        request: AppDeploymentLogsWs_StreamLogs_Req,
        handlers: AppDeploymentLogsWsHandlers,
        signal?: AbortSignal,
    ): Result<AppDeploymentLogsWs_StreamLogs_Res, Error> {
        const accessToken = session.getToken();

        if (!accessToken) {
            return Err(new Error("Access token not found."));
        }

        try {
            const { projectID, appID, deploymentID } = request.data;
            const url = this.client.buildUrl(
                `projects/${encodeURIComponent(projectID)}/apps/${encodeURIComponent(
                    appID,
                )}/deployments/${encodeURIComponent(deploymentID)}/logs`,
                {
                    follow: true,
                    tail: DEFAULT_LOG_TAIL,
                },
            );

            return Ok(
                this.client.connect(url, handlers, {
                    signal,
                    closeOnError: true,
                    protocols: ["access_token", accessToken],
                }),
            );
        } catch (error) {
            return Err(toWebSocketError(error, "Failed to stream deployment logs."));
        }
    }
}
