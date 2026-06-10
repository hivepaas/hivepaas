import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import type { WebSocketReadyState, WebSocketSubscription } from "@infrastructure/websocket";
import { useAppLogsWsApi } from "~/projects/api";
import type { AppLogFrame, AppLogs_GetLogs_Req } from "~/projects/api/services";
import { AppLogsQueries } from "~/projects/data";

import { LogsViewer, type LogsViewerFrame, parseLogsViewerFrames } from "@application/shared/components";

import { AppLogsToolbarFilters, AppLogsToolbarStart } from "../app-logs-toolbar";

const DEFAULT_LOG_LINES = 100;

export function AppLogsViewer({ projectID, appID, tabLabel, taskId, isActive }: AppLogsViewerProps) {
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
        setWebSocketReadyState(WebSocket.CLOSED);
    }, []);

    const handleStream = useCallback(() => {
        if (isConnectionActive) {
            return;
        }

        closeStream();
        setLogs([]);
        setWebSocketReadyState(WebSocket.CONNECTING);

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
                        setWebSocketReadyState(WebSocket.CLOSING);
                    },
                    onClose: () => {
                        subscriptionRef.current = null;
                        abortControllerRef.current = null;
                        setWebSocketReadyState(WebSocket.CLOSED);
                    },
                    onReadyStateChange: readyState => {
                        setWebSocketReadyState(readyState);
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
                setWebSocketReadyState(subscription.getReadyState());
            })
            .catch((error: unknown) => {
                if (!abortController.signal.aborted) {
                    console.error("Failed to connect app logs", error);
                    setWebSocketReadyState(WebSocket.CLOSED);
                }
            });
    }, [closeStream, isConnectionActive, request, streams]);

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
        if (!isActive || hasAutoStartedRef.current) {
            return;
        }

        hasAutoStartedRef.current = true;
        handleStream();
    }, [handleStream, isActive]);

    useEffect(() => {
        if (!isActive && isConnectionActive) {
            closeStream();
        }
    }, [closeStream, isActive, isConnectionActive]);

    useEffect(() => {
        return () => {
            closeStream();
        };
    }, [closeStream]);

    return (
        <LogsViewer
            frames={logs}
            isStreaming={isStreaming}
            isRefreshPending={isRefreshPending}
            hasLineNumbers={false}
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
                    isLinesDisabled={hasTimeFilter}
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
    projectID: string;
    appID: string;
    tabLabel: string;
    taskId?: string;
    isActive: boolean;
}
