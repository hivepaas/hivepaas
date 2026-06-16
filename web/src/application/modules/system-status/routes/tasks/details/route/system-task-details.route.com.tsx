import { useCallback, useEffect, useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { listBox } from "@lib/styles";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { SystemTasksCommands, SystemTasksQueries } from "~/system-status/data";
import { SystemTaskStatus } from "~/system-status/domain";

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
        <section className={cn(listBox, "p-0")}>
            <div className="flex flex-col gap-5">
                {isFetching && !task ? (
                    <SystemTaskSummaryCardSkeleton variant="details" />
                ) : task ? (
                    <SystemTaskSummaryCard
                        task={task}
                        now={now}
                        variant="details"
                        isCancelling={isCancelling}
                        onCancel={id => {
                            cancelTask({ taskID: id });
                        }}
                    >
                        <SystemTaskLogsViewer
                            taskID={taskId}
                            status={task.status}
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
