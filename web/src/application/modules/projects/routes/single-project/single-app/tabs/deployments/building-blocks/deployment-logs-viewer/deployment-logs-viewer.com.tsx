import { useCallback } from "react";

import type { WebSocketHandlers } from "@infrastructure/websocket";
import { useAppDeploymentLogsWsApi } from "~/projects/api";
import { EAppDeploymentStatus as AppDeploymentStatus } from "~/projects/module-shared/enums";
import type { EAppDeploymentStatus } from "~/projects/module-shared/enums";

import { StreamingLogsViewer } from "@application/shared/components";

import type { OpenApiConstant } from "@infrastructure/api";

export function DeploymentLogsViewer({
    projectID,
    env,
    appID,
    deploymentID,
    status,
    fontSize,
    themeId,
    height,
    isFullView,
    isFullHeight,
    onStreamClosedWhileInProgress,
}: DeploymentLogsViewerProps) {
    const { streams } = useAppDeploymentLogsWsApi();

    const subscribe = useCallback(
        (handlers: WebSocketHandlers, signal?: AbortSignal) =>
            streams.subscribe({ projectID, env, appID, deploymentID }, handlers, signal),
        [appID, deploymentID, env, projectID, streams],
    );

    return (
        <StreamingLogsViewer
            subscribe={subscribe}
            isNotStarted={status === AppDeploymentStatus.NotStarted}
            isInProgress={status === AppDeploymentStatus.InProgress}
            downloadFileName={`deployment-${deploymentID}-logs.txt`}
            defaultShowDebugLogs
            fontSize={fontSize}
            themeId={themeId}
            height={height}
            isFullView={isFullView}
            isFullHeight={isFullHeight}
            onStreamClosedWhileInProgress={onStreamClosedWhileInProgress}
        />
    );
}

interface DeploymentLogsScope {
    projectID: string;
    env: string;
    appID: string;
    deploymentID: string;
    status: OpenApiConstant<EAppDeploymentStatus>;
    fontSize?: number;
    themeId?: string;
    height?: number | string;
    isFullView?: boolean;
    isFullHeight?: boolean;
    onStreamClosedWhileInProgress: () => void;
}

type DeploymentLogsViewerProps = DeploymentLogsScope;
