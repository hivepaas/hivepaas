import { useCallback, useEffect, useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { listBox } from "@lib/styles";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppScheduledJobsCommands, AppScheduledJobsQueries } from "~/projects/data";
import { EAppScheduledJobTaskStatus } from "~/projects/module-shared/enums";

import { useLogViewerControls } from "@application/shared/components";

import type { OpenApiConstant } from "@infrastructure/api";

import {
    ScheduledJobTaskLogsViewer,
    ScheduledJobTaskSummaryCard,
    ScheduledJobTaskSummaryCardSkeleton,
    useScheduledJobTaskCurrentTime,
} from "../../building-blocks";

const TASK_DETAILS_REFETCH_INTERVAL_MS = 5_000;

function shouldPollTaskDetails(
    status: OpenApiConstant<EAppScheduledJobTaskStatus> | undefined,
    shouldPollAfterStreamClose: boolean,
): boolean {
    return status === EAppScheduledJobTaskStatus.NotStarted || (shouldPollAfterStreamClose && isTaskInProgress(status));
}

function isTaskInProgress(status?: OpenApiConstant<EAppScheduledJobTaskStatus>): boolean {
    return status === EAppScheduledJobTaskStatus.InProgress;
}

function isTaskTerminal(status?: OpenApiConstant<EAppScheduledJobTaskStatus>): boolean {
    return (
        status === EAppScheduledJobTaskStatus.Canceled ||
        status === EAppScheduledJobTaskStatus.Done ||
        status === EAppScheduledJobTaskStatus.Failed
    );
}

export function AppScheduledJobTaskDetailsRoute() {
    const {
        id: projectId,
        env,
        appId,
        scheduledJobId,
        taskId,
    } = useParams<{
        id: string;
        env: string;
        appId: string;
        scheduledJobId: string;
        taskId: string;
    }>();

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");
    invariant(scheduledJobId, "scheduledJobId must be defined");
    invariant(taskId, "taskId must be defined");

    const [shouldPollAfterStreamClose, setShouldPollAfterStreamClose] = useState(false);
    const { isFullscreen, toggleFullscreen, isFullView, toggleFullView, fontSize, cycleFontSize } =
        useLogViewerControls();

    const {
        data: taskResponse,
        isFetching,
        refetch: refetchTask,
    } = AppScheduledJobsQueries.useFindTaskById(
        {
            projectID: projectId,
            env,
            appID: appId,
            scheduledJobID: scheduledJobId,
            taskID: taskId,
        },
        {
            refetchInterval: query =>
                shouldPollTaskDetails(query.state.data?.data.status, shouldPollAfterStreamClose)
                    ? TASK_DETAILS_REFETCH_INTERVAL_MS
                    : false,
        },
    );
    const task = taskResponse?.data;

    const hasActiveTask = useMemo(() => isTaskInProgress(task?.status), [task?.status]);
    const now = useScheduledJobTaskCurrentTime(hasActiveTask);
    const handleStreamClosedWhileInProgress = useCallback(() => {
        setShouldPollAfterStreamClose(true);
        void refetchTask();
    }, [refetchTask]);

    useEffect(() => {
        setShouldPollAfterStreamClose(false);
    }, [taskId]);

    useEffect(() => {
        if (shouldPollAfterStreamClose && isTaskTerminal(task?.status)) {
            setShouldPollAfterStreamClose(false);
        }
    }, [shouldPollAfterStreamClose, task?.status]);

    const { mutate: cancelTask, isPending: isCancelling } = AppScheduledJobsCommands.useCancelTask({
        onSuccess: () => {
            toast.success("Task cancel requested");
        },
    });

    return (
        <section
            className={cn(
                listBox,
                isFullView && "max-w-none",
                isFullscreen &&
                    "!max-w-none !w-auto fixed inset-4 z-50 min-h-0 rounded-lg border bg-background p-4 shadow-2xl flex flex-col",
            )}
        >
            <div className={cn("flex flex-col gap-5", isFullscreen && "flex-1 min-h-0")}>
                {isFetching && !task ? (
                    <ScheduledJobTaskSummaryCardSkeleton variant="details" />
                ) : task ? (
                    <ScheduledJobTaskSummaryCard
                        task={task}
                        now={now}
                        variant="details"
                        isCancelling={isCancelling}
                        isFullscreen={isFullscreen}
                        isFullView={isFullView}
                        fontSize={fontSize}
                        onToggleFullscreen={toggleFullscreen}
                        onToggleFullView={toggleFullView}
                        onCycleFontSize={cycleFontSize}
                        onCancel={id => {
                            cancelTask({
                                projectID: projectId,
                                env,
                                appID: appId,
                                scheduledJobID: scheduledJobId,
                                taskID: id,
                            });
                        }}
                    >
                        <ScheduledJobTaskLogsViewer
                            projectID={projectId}
                            env={env}
                            appID={appId}
                            scheduledJobID={scheduledJobId}
                            taskID={taskId}
                            status={task.status}
                            fontSize={fontSize}
                            height={isFullscreen ? "100%" : undefined}
                            isFullView={isFullView}
                            onStreamClosedWhileInProgress={handleStreamClosedWhileInProgress}
                        />
                    </ScheduledJobTaskSummaryCard>
                ) : (
                    <div className="rounded-lg border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
                        Task not found.
                    </div>
                )}
            </div>
        </section>
    );
}
