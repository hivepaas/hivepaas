import { useCallback, useEffect, useState } from "react";

import type { WebSocketHandlers, WebSocketReadyState, WebSocketSubscription } from "@infrastructure/websocket";

import { LogsViewer } from "./logs-viewer.com";
import { parseLogsViewerFrames } from "./logs-viewer.utils";
import { useBufferedLogFrames } from "./use-buffered-log-frames";

export interface StreamingLogsViewerProps {
    subscribe: (handlers: WebSocketHandlers, signal?: AbortSignal) => Promise<WebSocketSubscription>;
    isNotStarted?: boolean;
    isInProgress?: boolean;
    downloadFileName?: string;
    defaultShowDebugLogs?: boolean;
    fontSize?: number;
    themeId?: string;
    height?: number | string;
    isFullView?: boolean;
    isFullHeight?: boolean;
    onStreamClosedWhileInProgress?: () => void;
}

export function StreamingLogsViewer({
    subscribe,
    isNotStarted = false,
    isInProgress = false,
    downloadFileName = "logs.txt",
    defaultShowDebugLogs = false,
    fontSize,
    themeId,
    height,
    isFullView = false,
    isFullHeight = false,
    onStreamClosedWhileInProgress,
}: StreamingLogsViewerProps) {
    const { frames: logs, appendFrames, reset } = useBufferedLogFrames();
    const [webSocketReadyState, setWebSocketReadyState] = useState<WebSocketReadyState>(WebSocket.CLOSED);
    const [refreshVersion, setRefreshVersion] = useState(0);
    const [isRefreshPending, setIsRefreshPending] = useState(false);
    const [suppressAutoReconnect, setSuppressAutoReconnect] = useState(false);

    const canLoadLogs = !isNotStarted;
    const shouldConnect = canLoadLogs && !suppressAutoReconnect;
    const isConnectionActive = webSocketReadyState === WebSocket.CONNECTING || webSocketReadyState === WebSocket.OPEN;
    const isStreaming = webSocketReadyState === WebSocket.OPEN;
    const showRefresh = canLoadLogs && webSocketReadyState === WebSocket.CLOSED;

    useEffect(() => {
        reset();
        setSuppressAutoReconnect(false);
        setIsRefreshPending(false);
    }, [reset, subscribe]);

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
        reset();

        void subscribe(
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
                        console.error("Failed to parse streaming log frame", error);
                    }
                },
                onMessageError: error => {
                    console.error("Failed to read streaming log frame", error);
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

                    if (!isInProgress || didRefetchAfterClose) {
                        return;
                    }

                    didRefetchAfterClose = true;
                    setSuppressAutoReconnect(true);
                    onStreamClosedWhileInProgress?.();
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
                    console.error("Failed to connect streaming logs", error);
                    setWebSocketReadyState(WebSocket.CLOSED);
                    setIsRefreshPending(false);
                }
            });

        return () => {
            isDisposed = true;
            abortController.abort();
            subscription?.close();
        };
    }, [appendFrames, isInProgress, onStreamClosedWhileInProgress, refreshVersion, reset, shouldConnect, subscribe]);

    return (
        <LogsViewer
            frames={logs}
            height={height}
            isFullView={isFullView}
            isFullHeight={isFullHeight}
            isStreaming={isStreaming}
            onRefresh={!isConnectionActive && showRefresh ? handleRefresh : undefined}
            isRefreshPending={isRefreshPending}
            hasLineNumbers={false}
            fontSize={fontSize}
            themeId={themeId}
            downloadFileName={downloadFileName}
            defaultShowDebugLogs={defaultShowDebugLogs}
        />
    );
}
