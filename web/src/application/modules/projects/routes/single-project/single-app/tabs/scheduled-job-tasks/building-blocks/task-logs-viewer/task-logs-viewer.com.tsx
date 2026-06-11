import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { cn } from "@/lib/utils";
import type { WebSocketReadyState, WebSocketSubscription } from "@infrastructure/websocket";
import { LoaderCircle } from "lucide-react";
import { useAppScheduledJobTaskLogsWsApi } from "~/projects/api";
import { AppScheduledJobsQueries } from "~/projects/data";
import type { AppScheduledJobTaskLogFrame } from "~/projects/domain";
import type { EAppScheduledJobTaskStatus } from "~/projects/module-shared/enums";
import { EAppScheduledJobTaskStatus as TaskStatus } from "~/projects/module-shared/enums";

import { Button } from "@/components/ui";
import { LogsViewer, type LogsViewerFrame, parseLogsViewerFrames } from "@application/shared/components";

const TASK_LOG_VIEWER_HEIGHT = "clamp(700px, calc(100vh - 300px), 2000px)";

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
    const subscriptionRef = useRef<WebSocketSubscription | null>(null);
    const abortControllerRef = useRef<AbortController | null>(null);
    const didRefetchAfterCloseRef = useRef(false);
    const hasAutoStartedRef = useRef(false);
    const { streams } = useAppScheduledJobTaskLogsWsApi();
    const canStream = status !== TaskStatus.NotStarted;
    const isConnectionActive = webSocketReadyState === WebSocket.CONNECTING || webSocketReadyState === WebSocket.OPEN;
    const isStreaming = webSocketReadyState === WebSocket.OPEN;

    const streamRequest = useMemo(
        () => ({
            projectID,
            appID,
            scheduledJobID,
            taskID,
        }),
        [appID, projectID, scheduledJobID, taskID],
    );

    const { refetch: refreshLogs, isFetching: isRefreshPending } = AppScheduledJobsQueries.useGetTaskLogs(
        {
            ...streamRequest,
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
        if (!canStream || isConnectionActive) {
            return;
        }

        closeStream();
        didRefetchAfterCloseRef.current = false;
        setLogs([]);
        setWebSocketReadyState(WebSocket.CONNECTING);

        const abortController = new AbortController();
        abortControllerRef.current = abortController;

        void streams
            .subscribe(
                streamRequest,
                {
                    onMessage: message => {
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
                        setWebSocketReadyState(WebSocket.CLOSING);
                    },
                    onClose: () => {
                        subscriptionRef.current = null;
                        abortControllerRef.current = null;
                        setWebSocketReadyState(WebSocket.CLOSED);

                        if (status !== TaskStatus.InProgress || didRefetchAfterCloseRef.current) {
                            return;
                        }

                        didRefetchAfterCloseRef.current = true;
                        onStreamClosedWhileInProgress();
                    },
                    onReadyStateChange: readyState => {
                        setWebSocketReadyState(readyState);
                    },
                },
                abortController.signal,
            )
            .then(currentSubscription => {
                if (abortController.signal.aborted) {
                    currentSubscription.close();
                    return;
                }

                subscriptionRef.current = currentSubscription;
                setWebSocketReadyState(currentSubscription.getReadyState());
            })
            .catch((error: unknown) => {
                if (!abortController.signal.aborted) {
                    console.error("Failed to connect scheduled job task logs", error);
                    setWebSocketReadyState(WebSocket.CLOSED);
                }
            });
    }, [canStream, closeStream, isConnectionActive, onStreamClosedWhileInProgress, status, streamRequest, streams]);

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
        hasAutoStartedRef.current = false;
        closeStream();
        setLogs([]);
    }, [closeStream, taskID]);

    useEffect(() => {
        if (!canStream) {
            hasAutoStartedRef.current = false;
            closeStream();
            return;
        }

        if (hasAutoStartedRef.current) {
            return;
        }

        hasAutoStartedRef.current = true;
        handleStream();
    }, [canStream, closeStream, handleStream]);

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
            height={TASK_LOG_VIEWER_HEIGHT}
            fontSize="0.875rem"
            downloadFileName={`scheduled-job-task-${taskID}-logs.txt`}
            toolbarStart={
                <ScheduledJobTaskLogsToolbarStart
                    canStream={canStream}
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
        />
    );
}

function ScheduledJobTaskLogsToolbarStart({
    canStream,
    isConnectionActive,
    isStreaming,
    isRefreshPending,
    onStream,
    onRefresh,
    onStop,
}: ScheduledJobTaskLogsToolbarStartProps) {
    if (isConnectionActive) {
        return (
            <div className="flex min-w-0 items-center gap-3">
                <span className="flex items-center gap-2 text-sm text-rose-500">
                    <LoaderCircle className={cn("size-4", isStreaming && "animate-spin")} />
                    streaming
                </span>
                <Button
                    type="button"
                    variant="link"
                    className="h-auto p-0 text-sm text-primary"
                    onClick={onStop}
                >
                    Stop
                </Button>
            </div>
        );
    }

    return (
        <div className="flex min-w-0 items-center gap-3">
            {canStream ? (
                <Button
                    type="button"
                    variant="link"
                    className="h-auto p-0 text-sm text-primary"
                    onClick={onStream}
                >
                    Stream
                </Button>
            ) : null}
            <Button
                type="button"
                variant="link"
                className="h-auto p-0 text-sm text-primary"
                isLoading={isRefreshPending}
                onClick={onRefresh}
            >
                Refresh
            </Button>
        </div>
    );
}

function toLogsViewerFrames(frames: AppScheduledJobTaskLogFrame[]): LogsViewerFrame[] {
    return frames.map(frame => ({
        type: frame.type as LogsViewerFrame["type"],
        data: frame.data,
        ts: frame.ts,
    }));
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

interface ScheduledJobTaskLogsToolbarStartProps {
    canStream: boolean;
    isConnectionActive: boolean;
    isStreaming: boolean;
    isRefreshPending: boolean;
    onStream: () => void;
    onRefresh: () => void;
    onStop: () => void;
}
