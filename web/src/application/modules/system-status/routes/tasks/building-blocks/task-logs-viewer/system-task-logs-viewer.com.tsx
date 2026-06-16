import { useCallback, useEffect, useMemo, useState } from "react";

import type { WebSocketReadyState, WebSocketSubscription } from "@infrastructure/websocket";
import { useSystemTaskLogsWsApi } from "~/system-status/api";
import { SystemTaskStatus, type SystemTaskStatus as SystemTaskStatusValue } from "~/system-status/domain";

import { LogsViewer, type LogsViewerFrame, parseLogsViewerFrames } from "@application/shared/components";

const TASK_LOG_VIEWER_HEIGHT = "clamp(700px, calc(100vh - 340px), 2000px)";

export function SystemTaskLogsViewer({ taskID, status, onStreamClosedWhileInProgress }: SystemTaskLogsViewerProps) {
    const [logs, setLogs] = useState<LogsViewerFrame[]>([]);
    const [webSocketReadyState, setWebSocketReadyState] = useState<WebSocketReadyState>(WebSocket.CLOSED);
    const [refreshVersion, setRefreshVersion] = useState(0);
    const [isRefreshPending, setIsRefreshPending] = useState(false);
    const [suppressAutoReconnect, setSuppressAutoReconnect] = useState(false);
    const { streams } = useSystemTaskLogsWsApi();
    const canLoadLogs = status !== SystemTaskStatus.NotStarted;
    const shouldConnect = canLoadLogs && !suppressAutoReconnect;
    const isConnectionActive = webSocketReadyState === WebSocket.CONNECTING || webSocketReadyState === WebSocket.OPEN;
    const isStreaming = webSocketReadyState === WebSocket.OPEN;
    const showRefresh = canLoadLogs && webSocketReadyState === WebSocket.CLOSED;

    const streamRequest = useMemo(
        () => ({
            taskID,
        }),
        [taskID],
    );

    useEffect(() => {
        setLogs([]);
        setSuppressAutoReconnect(false);
        setIsRefreshPending(false);
    }, [taskID]);

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
            setLogs([]);
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

                            setLogs(current => [...current, ...frames]);
                        } catch (error) {
                            console.error("Failed to parse system task log frame", error);
                        }
                    },
                    onMessageError: error => {
                        console.error("Failed to read system task log frame", error);
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

                        if (status !== SystemTaskStatus.InProgress || didRefetchAfterClose) {
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
                    console.error("Failed to connect system task logs", error);
                    setWebSocketReadyState(WebSocket.CLOSED);
                    setIsRefreshPending(false);
                }
            });

        return () => {
            isDisposed = true;
            abortController.abort();
            subscription?.close();
        };
    }, [onStreamClosedWhileInProgress, refreshVersion, shouldConnect, status, streamRequest, streams]);

    return (
        <LogsViewer
            frames={logs}
            isStreaming={isStreaming}
            onRefresh={!isConnectionActive && showRefresh ? handleRefresh : undefined}
            isRefreshPending={isRefreshPending}
            hasLineNumbers={false}
            height={TASK_LOG_VIEWER_HEIGHT}
            fontSize="0.875rem"
            downloadFileName={`system-task-${taskID}-logs.txt`}
        />
    );
}

interface SystemTaskLogsViewerProps {
    taskID: string;
    status: SystemTaskStatusValue;
    onStreamClosedWhileInProgress: () => void;
}
