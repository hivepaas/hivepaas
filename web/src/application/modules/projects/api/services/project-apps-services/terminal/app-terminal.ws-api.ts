import {
    BaseWebSocketApi,
    type WebSocketHandlers,
    type WebSocketSubscription,
    toWebSocketError,
} from "@infrastructure/websocket";
import { Err, Ok, type Result } from "oxide.ts";
import type { AppTerminalResizeMessage, AppTerminalWs_Open_Req } from "~/projects/api/services";

import { session } from "@infrastructure/api";

export type AppTerminalWs_Open_Res = WebSocketSubscription;

export type AppTerminalWsHandlers = WebSocketHandlers;

export function buildAppTerminalResizeMessage(width: number, height: number): AppTerminalResizeMessage {
    return {
        type: "resize",
        width,
        height,
    };
}

export class AppTerminalWsApi extends BaseWebSocketApi {
    open(
        request: AppTerminalWs_Open_Req,
        handlers: AppTerminalWsHandlers,
        signal?: AbortSignal,
    ): Result<AppTerminalWs_Open_Res, Error> {
        const accessToken = session.getToken();

        if (!accessToken) {
            return Err(new Error("Access token not found."));
        }

        try {
            const { projectID, appID, shell, width, height } = request.data;
            const url = this.client.buildUrl(
                `projects/${encodeURIComponent(projectID)}/apps/${encodeURIComponent(appID)}/terminal`,
                {
                    shell,
                    w: width,
                    h: height,
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
            return Err(toWebSocketError(error, "Failed to open app terminal."));
        }
    }
}
