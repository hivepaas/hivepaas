import { useEffect, useMemo, useRef, useState } from "react";

import { AppDeploymentsQueries } from "~/projects/data";

import { EnvConfig } from "@config";

import { LogsViewer, type LogsViewerFrame, parseLogsViewerFrames } from "@application/shared/components";

const DEFAULT_LOG_TAIL = 100;

type WebSocketReadyState = WebSocket["readyState"];

function buildDeploymentLogsWsUrl({ projectID, appID, deploymentID, token }: DeploymentLogsWsUrlParams): string {
    const baseUrl = EnvConfig.API_URL.endsWith("/") ? EnvConfig.API_URL : `${EnvConfig.API_URL}/`;
    const url = new URL(
        `projects/${encodeURIComponent(projectID)}/apps/${encodeURIComponent(appID)}/deployments/${encodeURIComponent(deploymentID)}/logs`,
        baseUrl,
    );

    url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
    appendDeploymentLogsQuery(url, token);

    return url.toString();
}

function appendDeploymentLogsQuery(url: URL, token: string) {
    url.searchParams.set("follow", "true");
    url.searchParams.set("tail", String(DEFAULT_LOG_TAIL));
    url.searchParams.set("token", token);
}

async function readWebSocketMessageData(data: MessageEvent["data"]): Promise<string> {
    if (typeof data === "string") {
        return data;
    }

    if (data instanceof Blob) {
        return data.text();
    }

    if (data instanceof ArrayBuffer) {
        return new TextDecoder().decode(data);
    }

    return "";
}

export function DeploymentLogsViewer({ projectID, appID, deploymentID }: DeploymentLogsViewerProps) {
    const [logs, setLogs] = useState<LogsViewerFrame[]>([]);
    const [webSocketReadyState, setWebSocketReadyState] = useState<WebSocketReadyState>(WebSocket.CLOSED);
    const webSocketRef = useRef<WebSocket | null>(null);
    const logsTokenRequest = useMemo(() => ({ projectID, appID, deploymentID }), [projectID, appID, deploymentID]);
    const { refetch: fetchLogsToken } = AppDeploymentsQueries.useGetLogsToken(logsTokenRequest, {
        enabled: false,
        staleTime: 0,
        gcTime: 0,
    });
    const isStreaming = webSocketReadyState === WebSocket.OPEN;

    useEffect(() => {
        let isDisposed = false;
        let ws: WebSocket | null = null;

        function updateReadyState(currentWebSocket: WebSocket) {
            if (isDisposed || webSocketRef.current !== currentWebSocket) {
                return;
            }

            setWebSocketReadyState(currentWebSocket.readyState);
        }

        async function connect() {
            try {
                const logsTokenResponse = await fetchLogsToken({ throwOnError: true });
                const token = logsTokenResponse.data?.data.token;

                if (isDisposed || !token) {
                    return;
                }

                const wsUrl = buildDeploymentLogsWsUrl({ projectID, appID, deploymentID, token });
                const currentWebSocket = new WebSocket(wsUrl);
                ws = currentWebSocket;
                webSocketRef.current = currentWebSocket;
                setWebSocketReadyState(currentWebSocket.readyState);

                currentWebSocket.addEventListener("open", () => {
                    updateReadyState(currentWebSocket);
                });

                currentWebSocket.addEventListener("message", event => {
                    void readWebSocketMessageData(event.data)
                        .then(parseLogsViewerFrames)
                        .then(frames => {
                            if (frames.length === 0) {
                                return;
                            }

                            setLogs(current => [...current, ...frames]);
                        })
                        .catch((error: unknown) => {
                            console.error("Failed to parse deployment log frame", error);
                        });
                });

                currentWebSocket.addEventListener("close", () => {
                    updateReadyState(currentWebSocket);
                    if (webSocketRef.current === currentWebSocket) {
                        webSocketRef.current = null;
                    }
                });

                currentWebSocket.addEventListener("error", () => {
                    if (isDisposed || webSocketRef.current !== currentWebSocket) {
                        return;
                    }

                    setWebSocketReadyState(WebSocket.CLOSING);
                    if (
                        currentWebSocket.readyState !== WebSocket.CLOSING &&
                        currentWebSocket.readyState !== WebSocket.CLOSED
                    ) {
                        currentWebSocket.close();
                    }
                });
            } catch (error) {
                if (!isDisposed) {
                    console.error("Failed to connect deployment logs", error);
                    setWebSocketReadyState(WebSocket.CLOSED);
                }
            }
        }

        void connect();

        return () => {
            isDisposed = true;
            setWebSocketReadyState(WebSocket.CLOSED);
            ws?.close();
            if (webSocketRef.current === ws) {
                webSocketRef.current = null;
            }
        };
    }, [appID, deploymentID, fetchLogsToken, projectID]);

    return (
        <LogsViewer
            frames={logs}
            isStreaming={isStreaming}
            downloadFileName="deployment-logs.txt"
        />
    );
}

interface DeploymentLogsScope {
    projectID: string;
    appID: string;
    deploymentID: string;
}

type DeploymentLogsWsUrlParams = DeploymentLogsScope & {
    token: string;
};

type DeploymentLogsViewerProps = DeploymentLogsScope;
