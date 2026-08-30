import { useCallback, useEffect, useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { listBox } from "@lib/styles";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { SystemTasksCommands, SystemTasksQueries } from "~/system-status/data";
import { SystemTaskStatus } from "~/system-status/domain";

import { useLogViewerControls } from "@application/shared/components";

import {
    SystemTaskLogsViewer,
    SystemTaskSummaryCard,
    SystemTaskSummaryCardSkeleton,
    useSystemTaskCurrentTime,
} from "../../building-blocks";

const TASK_DETAILS_REFETCH_INTERVAL_MS = 5_000;

function shouldPollTaskDetails(status: SystemTaskStatus | undefined, shouldPollAfterStreamClose: boolean): boolean {
    return status === SystemTaskStatus.NotStarted || (shouldPollAfterStreamClose && isTaskInProgress(status));
}

function isTaskInProgress(status?: SystemTaskStatus): boolean {
    return status === SystemTaskStatus.InProgress;
}

function isTaskTerminal(status?: SystemTaskStatus): boolean {
    return (
        status === SystemTaskStatus.Canceled || status === SystemTaskStatus.Done || status === SystemTaskStatus.Failed
    );
}

export function SystemTaskDetailsRoute() {
    const { taskId } = useParams<{ taskId: string }>();

    invariant(taskId, "taskId must be defined");

    const [shouldPollAfterStreamClose, setShouldPollAfterStreamClose] = useState(false);
    const {
        isFullscreen,
        toggleFullscreen,
        isFullView,
        toggleFullView,
        isFullHeight,
        toggleFullHeight,
        fontSize,
        cycleFontSize,
        themeId,
        changeTheme,
    } = useLogViewerControls();

    const {
        data: taskResponse,
        isFetching,
        refetch: refetchTask,
    } = SystemTasksQueries.useFindOneById(
        {
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
    const now = useSystemTaskCurrentTime(hasActiveTask);
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

    const { mutate: cancelTask, isPending: isCancelling } = SystemTasksCommands.useCancel({
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
                    "!max-w-none !w-auto fixed inset-1.5 md:inset-4 z-50 min-h-0 rounded-lg border bg-background p-2.5 md:p-4 shadow-2xl flex flex-col",
            )}
        >
            <div className={cn("flex flex-col gap-2 sm:gap-5", isFullscreen && "flex-1 min-h-0")}>
                {isFetching && !task ? (
                    <SystemTaskSummaryCardSkeleton variant="details" />
                ) : task ? (
                    <SystemTaskSummaryCard
                        task={task}
                        now={now}
                        variant="details"
                        isCancelling={isCancelling}
                        isFullscreen={isFullscreen}
                        isFullView={isFullView}
                        isFullHeight={isFullHeight}
                        fontSize={fontSize}
                        themeId={themeId}
                        onToggleFullscreen={toggleFullscreen}
                        onToggleFullView={toggleFullView}
                        onToggleFullHeight={toggleFullHeight}
                        onCycleFontSize={cycleFontSize}
                        onSelectTheme={changeTheme}
                        onCancel={id => {
                            cancelTask({ taskID: id });
                        }}
                    >
                        <SystemTaskLogsViewer
                            taskID={taskId}
                            status={task.status}
                            fontSize={fontSize}
                            themeId={themeId}
                            height={isFullscreen ? "100%" : undefined}
                            isFullView={isFullView}
                            isFullHeight={isFullHeight}
                            onStreamClosedWhileInProgress={handleStreamClosedWhileInProgress}
                        />
                    </SystemTaskSummaryCard>
                ) : (
                    <div className="rounded-lg border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
                        Task not found.
                    </div>
                )}
            </div>
        </section>
    );
}
