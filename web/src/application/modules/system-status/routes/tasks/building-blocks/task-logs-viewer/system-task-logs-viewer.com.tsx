import { useCallback } from "react";

import type { WebSocketHandlers } from "@infrastructure/websocket";
import { useSystemTaskLogsWsApi } from "~/system-status/api";
import { SystemTaskStatus, type SystemTaskStatus as SystemTaskStatusValue } from "~/system-status/domain";

import { StreamingLogsViewer } from "@application/shared/components";

export function SystemTaskLogsViewer({
    taskID,
    status,
    fontSize,
    themeId,
    height,
    isFullView,
    isFullHeight,
    onStreamClosedWhileInProgress,
}: SystemTaskLogsViewerProps) {
    const { streams } = useSystemTaskLogsWsApi();

    const subscribe = useCallback(
        (handlers: WebSocketHandlers, signal?: AbortSignal) => streams.subscribe({ taskID }, handlers, signal),
        [streams, taskID],
    );

    return (
        <StreamingLogsViewer
            subscribe={subscribe}
            isNotStarted={status === SystemTaskStatus.NotStarted}
            isInProgress={status === SystemTaskStatus.InProgress}
            downloadFileName={`system-task-${taskID}-logs.txt`}
            fontSize={fontSize}
            themeId={themeId}
            height={height}
            isFullView={isFullView}
            isFullHeight={isFullHeight}
            onStreamClosedWhileInProgress={onStreamClosedWhileInProgress}
        />
    );
}

interface SystemTaskLogsViewerProps {
    taskID: string;
    status: SystemTaskStatusValue;
    fontSize?: number;
    themeId?: string;
    height?: number | string;
    isFullView?: boolean;
    isFullHeight?: boolean;
    onStreamClosedWhileInProgress: () => void;
}
