import { useCallback, useEffect, useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { listBox } from "@lib/styles";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { AppDeploymentsCommands, AppDeploymentsQueries } from "~/projects/data";
import { EAppDeploymentStatus } from "~/projects/module-shared/enums";

import { useLogViewerControls } from "@application/shared/components";

import type { OpenApiConstant } from "@infrastructure/api";

import {
    DeploymentLogsViewer,
    DeploymentSummaryCard,
    DeploymentSummaryCardSkeleton,
    useDeploymentCurrentTime,
} from "../../building-blocks";
import { showDeploymentCancelToast } from "../../utils";

const DEPLOYMENT_DETAILS_REFETCH_INTERVAL_MS = 5_000;

function shouldPollDeploymentDetails(
    status: OpenApiConstant<EAppDeploymentStatus> | undefined,
    shouldPollAfterStreamClose: boolean,
): boolean {
    return status === EAppDeploymentStatus.NotStarted || (shouldPollAfterStreamClose && isDeploymentInProgress(status));
}

function isDeploymentInProgress(status?: OpenApiConstant<EAppDeploymentStatus>): boolean {
    return status === EAppDeploymentStatus.InProgress;
}

function isDeploymentTerminal(status?: OpenApiConstant<EAppDeploymentStatus>): boolean {
    return (
        status === EAppDeploymentStatus.Canceled ||
        status === EAppDeploymentStatus.Done ||
        status === EAppDeploymentStatus.Failed
    );
}

export function AppDeploymentDetailsRoute() {
    const {
        id: projectId,
        env,
        appId,
        deploymentId,
    } = useParams<{
        id: string;
        env: string;
        appId: string;
        deploymentId: string;
    }>();

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");
    invariant(deploymentId, "deploymentId must be defined");

    const [shouldPollAfterStreamClose, setShouldPollAfterStreamClose] = useState(false);
    const { isFullscreen, toggleFullscreen, fontSize, cycleFontSize } = useLogViewerControls();

    const {
        data: deploymentResponse,
        isFetching,
        refetch: refetchDeployment,
    } = AppDeploymentsQueries.useFindOneById(
        {
            projectID: projectId,
            env,
            appID: appId,
            deploymentID: deploymentId,
        },
        {
            refetchInterval: query =>
                shouldPollDeploymentDetails(query.state.data?.data.status, shouldPollAfterStreamClose)
                    ? DEPLOYMENT_DETAILS_REFETCH_INTERVAL_MS
                    : false,
        },
    );
    const deployment = deploymentResponse?.data;

    const hasActiveDeployment = useMemo(() => isDeploymentInProgress(deployment?.status), [deployment?.status]);
    const now = useDeploymentCurrentTime(hasActiveDeployment);
    const handleStreamClosedWhileInProgress = useCallback(() => {
        setShouldPollAfterStreamClose(true);
        void refetchDeployment();
    }, [refetchDeployment]);

    useEffect(() => {
        setShouldPollAfterStreamClose(false);
    }, [deploymentId]);

    useEffect(() => {
        if (shouldPollAfterStreamClose && isDeploymentTerminal(deployment?.status)) {
            setShouldPollAfterStreamClose(false);
        }
    }, [deployment?.status, shouldPollAfterStreamClose]);

    const { mutate: cancelDeployment, isPending: isCancelling } = AppDeploymentsCommands.useCancel({
        onSuccess: response => {
            showDeploymentCancelToast(response);
        },
    });

    return (
        <section
            className={cn(
                listBox,
                isFullscreen &&
                    "!max-w-none !w-auto fixed inset-4 z-50 min-h-0 rounded-lg border bg-background p-4 shadow-2xl flex flex-col",
            )}
        >
            <div className={cn("flex flex-col gap-5", isFullscreen && "flex-1 min-h-0 flex flex-col")}>
                {isFetching && !deployment ? (
                    <DeploymentSummaryCardSkeleton variant="details" />
                ) : deployment ? (
                    <DeploymentSummaryCard
                        deployment={deployment}
                        now={now}
                        variant="details"
                        isCancelling={isCancelling}
                        isFullscreen={isFullscreen}
                        fontSize={fontSize}
                        onToggleFullscreen={toggleFullscreen}
                        onCycleFontSize={cycleFontSize}
                        onCancel={id => {
                            cancelDeployment({
                                projectID: projectId,
                                env,
                                appID: appId,
                                deploymentID: id,
                            });
                        }}
                    >
                        <DeploymentLogsViewer
                            projectID={projectId}
                            env={env}
                            appID={appId}
                            deploymentID={deploymentId}
                            status={deployment.status}
                            fontSize={fontSize}
                            height={isFullscreen ? "100%" : undefined}
                            onStreamClosedWhileInProgress={handleStreamClosedWhileInProgress}
                        />
                    </DeploymentSummaryCard>
                ) : (
                    <div className="rounded-lg border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
                        Deployment not found.
                    </div>
                )}
            </div>
        </section>
    );
}
