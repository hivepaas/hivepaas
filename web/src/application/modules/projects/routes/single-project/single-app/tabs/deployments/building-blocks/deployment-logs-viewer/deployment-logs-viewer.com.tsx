import { useCallback, useEffect, useMemo, useState } from "react";

import type { WebSocketReadyState, WebSocketSubscription } from "@infrastructure/websocket";
import { useAppDeploymentLogsWsApi } from "~/projects/api";
import type { EAppDeploymentStatus } from "~/projects/module-shared/enums";
import { EAppDeploymentStatus as AppDeploymentStatus } from "~/projects/module-shared/enums";

import { LogsViewer, parseLogsViewerFrames, useBufferedLogFrames } from "@application/shared/components";

import type { OpenApiConstant } from "@infrastructure/api";

const DEPLOYMENT_LOG_VIEWER_HEIGHT = "clamp(700px, calc(100vh - 340px), 2000px)";

export function DeploymentLogsViewer({
    projectID,
    env,
    appID,
    deploymentID,
    status,
    onStreamClosedWhileInProgress,
}: DeploymentLogsViewerProps) {
    const { frames: logs, appendFrames, reset } = useBufferedLogFrames();
    const [webSocketReadyState, setWebSocketReadyState] = useState<WebSocketReadyState>(WebSocket.CLOSED);
    const [refreshVersion, setRefreshVersion] = useState(0);
    const [isRefreshPending, setIsRefreshPending] = useState(false);
    const [suppressAutoReconnect, setSuppressAutoReconnect] = useState(false);
    const { streams } = useAppDeploymentLogsWsApi();
    const canLoadLogs = status !== AppDeploymentStatus.NotStarted;
    const shouldConnect = canLoadLogs && !suppressAutoReconnect;
    const isConnectionActive = webSocketReadyState === WebSocket.CONNECTING || webSocketReadyState === WebSocket.OPEN;
    const isStreaming = webSocketReadyState === WebSocket.OPEN;
    const showRefresh = canLoadLogs && webSocketReadyState === WebSocket.CLOSED;

    const streamRequest = useMemo(
        () => ({
            projectID,
            env,
            appID,
            deploymentID,
        }),
        [appID, deploymentID, env, projectID],
    );

    useEffect(() => {
        reset();
        setSuppressAutoReconnect(false);
        setIsRefreshPending(false);
    }, [deploymentID, reset]);

    const handleRefresh = useCallback(() => {
        if (!showRefresh || isRefreshPending) {
            return;
        }

        setIsRefreshPending(true);
        setSuppressAutoReconnect(false);
        setRefreshVersion(current => current + 1);
    }, [isRefreshPending, showRefresh]);

    useEffect(() => {
        if (!shouldConnect) {
            setWebSocketReadyState(WebSocket.CLOSED);
            setIsRefreshPending(false);
            return;
        }

        let isDisposed = false;
        let didRefetchAfterClose = false;
        let subscription: WebSocketSubscription | null = null;
        const abortController = new AbortController();

        setWebSocketReadyState(WebSocket.CONNECTING);

        if (refreshVersion > 0) {
            reset();
        }

        void streams
            .subscribe(
                streamRequest,
                {
                    onMessage: message => {
                        if (isDisposed) {
                            return;
                        }

                        try {
                            const frames = parseLogsViewerFrames(message);

                            if (frames.length === 0) {
                                return;
                            }

                            appendFrames(frames);
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
                    onClose: () => {
                        if (isDisposed) {
                            return;
                        }

                        setIsRefreshPending(false);
                        setWebSocketReadyState(WebSocket.CLOSED);

                        if (status !== AppDeploymentStatus.InProgress || didRefetchAfterClose) {
                            return;
                        }

                        didRefetchAfterClose = true;
                        setSuppressAutoReconnect(true);
                        onStreamClosedWhileInProgress();
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
                setIsRefreshPending(false);
            })
            .catch((error: unknown) => {
                if (!isDisposed) {
                    console.error("Failed to connect deployment logs", error);
                    setWebSocketReadyState(WebSocket.CLOSED);
                    setIsRefreshPending(false);
                }
            });

        return () => {
            isDisposed = true;
            abortController.abort();
            subscription?.close();
        };
    }, [appendFrames, onStreamClosedWhileInProgress, refreshVersion, reset, shouldConnect, status, streamRequest, streams]);

    return (
        <LogsViewer
            frames={logs}
            height={DEPLOYMENT_LOG_VIEWER_HEIGHT}
            isStreaming={isStreaming}
            onRefresh={!isConnectionActive && showRefresh ? handleRefresh : undefined}
            isRefreshPending={isRefreshPending}
            hasLineNumbers={false}
            fontSize="0.875rem"
            downloadFileName="deployment-logs.txt"
            defaultShowDebugLogs
        />
    );
}

interface DeploymentLogsScope {
    projectID: string;
    env: string;
    appID: string;
    deploymentID: string;
    status: OpenApiConstant<EAppDeploymentStatus>;
    onStreamClosedWhileInProgress: () => void;
}

type DeploymentLogsViewerProps = DeploymentLogsScope;
