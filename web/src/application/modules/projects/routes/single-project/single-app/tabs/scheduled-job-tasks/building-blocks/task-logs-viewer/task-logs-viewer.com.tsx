import { useCallback, useEffect, useMemo, useState } from "react";

import type { WebSocketReadyState, WebSocketSubscription } from "@infrastructure/websocket";
import { useAppScheduledJobTaskLogsWsApi } from "~/projects/api";
import type { EAppScheduledJobTaskStatus } from "~/projects/module-shared/enums";
import { EAppScheduledJobTaskStatus as TaskStatus } from "~/projects/module-shared/enums";

import { LogsViewer, type LogsViewerFrame, parseLogsViewerFrames } from "@application/shared/components";

const TASK_LOG_VIEWER_HEIGHT = "clamp(700px, calc(100vh - 340px), 2000px)";

export function ScheduledJobTaskLogsViewer({
    projectID,
    appID,
    scheduledJobID,
    taskID,
    status,
    onStreamClosedWhileInProgress,
}: ScheduledJobTaskLogsViewerProps) {
    const [logs, setLogs] = useState<LogsViewerFrame[]>([]);
    const [webSocketReadyState, setWebSocketReadyState] = useState<WebSocketReadyState>(WebSocket.CLOSED);
    const [refreshVersion, setRefreshVersion] = useState(0);
    const [isRefreshPending, setIsRefreshPending] = useState(false);
    const [suppressAutoReconnect, setSuppressAutoReconnect] = useState(false);
    const { streams } = useAppScheduledJobTaskLogsWsApi();
    const canLoadLogs = status !== TaskStatus.NotStarted;
    const shouldConnect = canLoadLogs && !suppressAutoReconnect;
    const isConnectionActive = webSocketReadyState === WebSocket.CONNECTING || webSocketReadyState === WebSocket.OPEN;
    const isStreaming = webSocketReadyState === WebSocket.OPEN;
    const showRefresh = canLoadLogs && webSocketReadyState === WebSocket.CLOSED;

    const streamRequest = useMemo(
        () => ({
            projectID,
            appID,
            scheduledJobID,
            taskID,
        }),
        [appID, projectID, scheduledJobID, taskID],
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
                            console.error("Failed to parse scheduled job task log frame", error);
                        }
                    },
                    onMessageError: error => {
                        console.error("Failed to read scheduled job task log frame", error);
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

                        if (status !== TaskStatus.InProgress || didRefetchAfterClose) {
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
                    console.error("Failed to connect scheduled job task logs", error);
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
            downloadFileName={`scheduled-job-task-${taskID}-logs.txt`}
        />
    );
}

interface ScheduledJobTaskLogsScope {
    projectID: string;
    appID: string;
    scheduledJobID: string;
    taskID: string;
    status: EAppScheduledJobTaskStatus;
    onStreamClosedWhileInProgress: () => void;
}

type ScheduledJobTaskLogsViewerProps = ScheduledJobTaskLogsScope;
