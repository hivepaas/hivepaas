import {
    BaseWebSocketApi,
    type WebSocketHandlers,
    type WebSocketSubscription,
    toWebSocketError,
} from "@infrastructure/websocket";
import { Err, Ok, type Result, match } from "oxide.ts";

import type { AppDeploymentsApi } from "./app-deployments.api";
import type { AppDeployments_GetLogsToken_Req } from "./app-deployments.api.contracts";

const DEFAULT_LOG_TAIL = 100;

export type AppDeploymentLogsWs_StreamLogs_Req = AppDeployments_GetLogsToken_Req;

export type AppDeploymentLogsWs_StreamLogs_Res = WebSocketSubscription;

export type AppDeploymentLogsWsHandlers = WebSocketHandlers;

export class AppDeploymentLogsWsApi extends BaseWebSocketApi {
    public constructor(private readonly appDeploymentsApi: AppDeploymentsApi) {
        super();
    }

    async streamLogs(
        request: AppDeploymentLogsWs_StreamLogs_Req,
        handlers: AppDeploymentLogsWsHandlers,
        signal?: AbortSignal,
    ): Promise<Result<AppDeploymentLogsWs_StreamLogs_Res, Error>> {
        const logsTokenResult = await this.appDeploymentsApi.getLogsToken(request, signal);

        return match(logsTokenResult, {
            Ok: logsTokenResponse => {
                try {
                    const { projectID, appID, deploymentID } = request.data;
                    const url = this.client.buildUrl(
                        `projects/${encodeURIComponent(projectID)}/apps/${encodeURIComponent(
                            appID,
                        )}/deployments/${encodeURIComponent(deploymentID)}/logs`,
                        {
                            follow: true,
                            tail: DEFAULT_LOG_TAIL,
                            token: logsTokenResponse.data.token,
                        },
                    );

                    return Ok(this.client.connect(url, handlers, { signal, closeOnError: true }));
                } catch (error) {
                    return Err(toWebSocketError(error, "Failed to stream deployment logs."));
                }
            },
            Err: error => Err(error),
        });
    }
}
