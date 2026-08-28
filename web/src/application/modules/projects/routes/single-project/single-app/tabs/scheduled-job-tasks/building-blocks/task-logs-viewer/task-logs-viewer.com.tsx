import { useCallback } from "react";

import type { WebSocketHandlers } from "@infrastructure/websocket";
import { useAppScheduledJobTaskLogsWsApi } from "~/projects/api";
import { EAppScheduledJobTaskStatus as TaskStatus } from "~/projects/module-shared/enums";
import type { EAppScheduledJobTaskStatus } from "~/projects/module-shared/enums";

import { StreamingLogsViewer } from "@application/shared/components";

import type { OpenApiConstant } from "@infrastructure/api";

export function ScheduledJobTaskLogsViewer({
    projectID,
    env,
    appID,
    scheduledJobID,
    taskID,
    status,
    fontSize,
    height,
    isFullView,
    onStreamClosedWhileInProgress,
}: ScheduledJobTaskLogsViewerProps) {
    const { streams } = useAppScheduledJobTaskLogsWsApi();

    const subscribe = useCallback(
        (handlers: WebSocketHandlers, signal?: AbortSignal) =>
            streams.subscribe({ projectID, env, appID, scheduledJobID, taskID }, handlers, signal),
        [appID, env, projectID, scheduledJobID, streams, taskID],
    );

    return (
        <StreamingLogsViewer
            subscribe={subscribe}
            isNotStarted={status === TaskStatus.NotStarted}
            isInProgress={status === TaskStatus.InProgress}
            downloadFileName={`scheduled-job-task-${taskID}-logs.txt`}
            fontSize={fontSize}
            height={height}
            isFullView={isFullView}
            onStreamClosedWhileInProgress={onStreamClosedWhileInProgress}
        />
    );
}

interface ScheduledJobTaskLogsScope {
    projectID: string;
    env: string;
    appID: string;
    scheduledJobID: string;
    taskID: string;
    status: OpenApiConstant<EAppScheduledJobTaskStatus>;
    fontSize?: number;
    height?: number | string;
    isFullView?: boolean;
    onStreamClosedWhileInProgress: () => void;
}

type ScheduledJobTaskLogsViewerProps = ScheduledJobTaskLogsScope;
