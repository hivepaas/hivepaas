import { useEffect, useState } from "react";

import type { WebSocketReadyState, WebSocketSubscription } from "@infrastructure/websocket";
import { useAppDeploymentLogsWsApi } from "~/projects/api";

import { LogsViewer, type LogsViewerFrame, parseLogsViewerFrames } from "@application/shared/components";

export function DeploymentLogsViewer({ projectID, appID, deploymentID }: DeploymentLogsViewerProps) {
    const [logs, setLogs] = useState<LogsViewerFrame[]>([]);
    const [webSocketReadyState, setWebSocketReadyState] = useState<WebSocketReadyState>(WebSocket.CLOSED);
    const { streams } = useAppDeploymentLogsWsApi();
    const isStreaming = webSocketReadyState === WebSocket.OPEN;

    useEffect(() => {
        let isDisposed = false;
        let subscription: WebSocketSubscription | null = null;
        const abortController = new AbortController();

        void streams
            .subscribe(
                { projectID, appID, deploymentID },
                {
                    onMessage: message => {
                        try {
                            const frames = parseLogsViewerFrames(message);

                            if (frames.length === 0) {
                                return;
                            }

                            setLogs(current => [...current, ...frames]);
                        } catch (error) {
                            console.error("Failed to parse deployment log frame", error);
                        }
                    },
                    onMessageError: error => {
                        console.error("Failed to read deployment log frame", error);
                    },
                    onError: () => {
                        if (isDisposed) {
                            return;
                        }

                        setWebSocketReadyState(WebSocket.CLOSING);
                    },
                    onReadyStateChange: readyState => {
                        if (isDisposed) {
                            return;
                        }

                        setWebSocketReadyState(readyState);
                    },
                },
                abortController.signal,
            )
            .then(currentSubscription => {
                if (isDisposed) {
                    currentSubscription.close();
                    return;
                }

                subscription = currentSubscription;
                setWebSocketReadyState(currentSubscription.getReadyState());
            })
            .catch((error: unknown) => {
                if (!isDisposed) {
                    console.error("Failed to connect deployment logs", error);
                    setWebSocketReadyState(WebSocket.CLOSED);
                }
            });

        return () => {
            isDisposed = true;
            abortController.abort();
            subscription?.close();
        };
    }, [appID, deploymentID, projectID, streams]);

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

type DeploymentLogsViewerProps = DeploymentLogsScope;
