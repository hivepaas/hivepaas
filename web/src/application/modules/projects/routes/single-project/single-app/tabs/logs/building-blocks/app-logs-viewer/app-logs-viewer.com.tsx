import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import type { WebSocketReadyState, WebSocketSubscription } from "@infrastructure/websocket";
import { useAppLogsWsApi } from "~/projects/api";
import type { AppLogFrame, AppLogs_GetLogs_Req } from "~/projects/api/services";
import { AppLogsQueries } from "~/projects/data";

import { LogsViewer, type LogsViewerFrame, parseLogsViewerFrames } from "@application/shared/components";

import { AppLogsToolbarFilters, AppLogsToolbarStart } from "../app-logs-toolbar";

const DEFAULT_LOG_LINES = 100;
const APP_LOG_VIEWER_HEIGHT = "clamp(700px, calc(100vh - 300px), 2000px)";

export function AppLogsViewer({
    tabID,
    projectID,
    appID,
    tabLabel,
    taskId,
    isActive,
    shouldAutoStream,
    onReadyStateChange,
}: AppLogsViewerProps) {
    const [logs, setLogs] = useState<LogsViewerFrame[]>([]);
    const [webSocketReadyState, setWebSocketReadyState] = useState<WebSocketReadyState>(WebSocket.CLOSED);
    const [lines, setLines] = useState<number | undefined>(DEFAULT_LOG_LINES);
    const [since, setSince] = useState<Date | undefined>();
    const [duration, setDuration] = useState<string | undefined>();
    const subscriptionRef = useRef<WebSocketSubscription | null>(null);
    const abortControllerRef = useRef<AbortController | null>(null);
    const hasAutoStartedRef = useRef(false);
    const { streams } = useAppLogsWsApi();
    const isConnectionActive = webSocketReadyState === WebSocket.CONNECTING || webSocketReadyState === WebSocket.OPEN;
    const isStreaming = webSocketReadyState === WebSocket.OPEN;
    const hasTimeFilter = since !== undefined || duration !== undefined;

    const setReadyState = useCallback(
        (readyState: WebSocketReadyState) => {
            setWebSocketReadyState(readyState);
            onReadyStateChange(tabID, readyState);
        },
        [onReadyStateChange, tabID],
    );

    const request = useMemo<AppLogs_GetLogs_Req["data"]>(
        () => ({
            projectID,
            appID,
            taskId,
            tail: hasTimeFilter ? undefined : lines,
            since,
            duration,
        }),
        [appID, duration, hasTimeFilter, lines, projectID, since, taskId],
    );

    const { refetch: refreshLogs, isFetching: isRefreshPending } = AppLogsQueries.useGetLogs(
        {
            ...request,
            follow: false,
        },
        {
            enabled: false,
        },
    );

    const closeStream = useCallback(() => {
        abortControllerRef.current?.abort();
        abortControllerRef.current = null;
        subscriptionRef.current?.close();
        subscriptionRef.current = null;
        setReadyState(WebSocket.CLOSED);
    }, [setReadyState]);

    const handleStream = useCallback(() => {
        if (isConnectionActive) {
            return;
        }

        closeStream();
        setLogs([]);
        setReadyState(WebSocket.CONNECTING);

        const abortController = new AbortController();
        abortControllerRef.current = abortController;

        void streams
            .subscribe(
                {
                    ...request,
                    follow: true,
                },
                {
                    onMessage: message => {
                        try {
                            const frames = parseLogsViewerFrames(message);

                            if (frames.length === 0) {
                                return;
                            }

                            setLogs(current => [...current, ...frames]);
                        } catch (error) {
                            console.error("Failed to parse app log frame", error);
                        }
                    },
                    onMessageError: error => {
                        console.error("Failed to read app log frame", error);
                    },
                    onError: () => {
                        setReadyState(WebSocket.CLOSING);
                    },
                    onClose: () => {
                        subscriptionRef.current = null;
                        abortControllerRef.current = null;
                        setReadyState(WebSocket.CLOSED);
                    },
                    onReadyStateChange: readyState => {
                        setReadyState(readyState);
                    },
                },
                abortController.signal,
            )
            .then(subscription => {
                if (abortController.signal.aborted) {
                    subscription.close();
                    return;
                }

                subscriptionRef.current = subscription;
                setReadyState(subscription.getReadyState());
            })
            .catch((error: unknown) => {
                if (!abortController.signal.aborted) {
                    console.error("Failed to connect app logs", error);
                    setReadyState(WebSocket.CLOSED);
                }
            });
    }, [closeStream, isConnectionActive, request, setReadyState, streams]);

    const handleRefresh = useCallback(async () => {
        if (isConnectionActive) {
            return;
        }

        const result = await refreshLogs();

        if (result.data) {
            setLogs(toLogsViewerFrames(result.data.data));
        }
    }, [isConnectionActive, refreshLogs]);

    useEffect(() => {
        if (!shouldAutoStream || !isActive || hasAutoStartedRef.current) {
            return;
        }

        hasAutoStartedRef.current = true;
        handleStream();
    }, [handleStream, isActive, shouldAutoStream]);

    useEffect(() => {
        return () => {
            hasAutoStartedRef.current = false;
            closeStream();
        };
    }, [closeStream]);

    return (
        <LogsViewer
            frames={logs}
            isStreaming={isStreaming}
            isRefreshPending={isRefreshPending}
            hasLineNumbers={false}
            height={APP_LOG_VIEWER_HEIGHT}
            fontSize="0.875rem"
            downloadFileName={taskId ? `app-logs-${tabLabel}.txt` : "app-logs-aggregation.txt"}
            toolbarStart={
                <AppLogsToolbarStart
                    isConnectionActive={isConnectionActive}
                    isStreaming={isStreaming}
                    isRefreshPending={isRefreshPending}
                    onStream={handleStream}
                    onRefresh={() => {
                        void handleRefresh();
                    }}
                    onStop={closeStream}
                />
            }
            toolbarFilters={
                <AppLogsToolbarFilters
                    lines={lines}
                    since={since}
                    duration={duration}
                    isLinesHidden={hasTimeFilter}
                    onLinesChange={setLines}
                    onSinceChange={setSince}
                    onDurationChange={setDuration}
                />
            }
        />
    );
}

function toLogsViewerFrames(frames: AppLogFrame[]): LogsViewerFrame[] {
    return frames.map(frame => ({
        type: frame.type as LogsViewerFrame["type"],
        data: frame.data,
        ts: frame.ts,
    }));
}

interface AppLogsViewerProps {
    tabID: string;
    projectID: string;
    appID: string;
    tabLabel: string;
    taskId?: string;
    isActive: boolean;
    shouldAutoStream: boolean;
    onReadyStateChange: (tabID: string, readyState: WebSocketReadyState) => void;
}
