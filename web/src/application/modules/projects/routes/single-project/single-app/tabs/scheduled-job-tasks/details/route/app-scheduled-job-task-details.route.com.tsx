import { useCallback, useMemo } from "react";

import { cn } from "@/lib/utils";
import { listBox } from "@lib/styles";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppScheduledJobsCommands, AppScheduledJobsQueries } from "~/projects/data";
import { EAppScheduledJobTaskStatus } from "~/projects/module-shared/enums";

import {
    ScheduledJobTaskLogsViewer,
    ScheduledJobTaskSummaryCard,
    ScheduledJobTaskSummaryCardSkeleton,
    useScheduledJobTaskCurrentTime,
} from "../../building-blocks";

const TASK_DETAILS_REFETCH_INTERVAL_MS = 5_000;

function shouldPollTaskDetails(status?: EAppScheduledJobTaskStatus): boolean {
    return status === EAppScheduledJobTaskStatus.NotStarted;
}

export function AppScheduledJobTaskDetailsRoute() {
    const {
        id: projectId,
        appId,
        scheduledJobId,
        taskId,
    } = useParams<{
        id: string;
        appId: string;
        scheduledJobId: string;
        taskId: string;
    }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");
    invariant(scheduledJobId, "scheduledJobId must be defined");
    invariant(taskId, "taskId must be defined");

    const {
        data: taskResponse,
        isFetching,
        refetch: refetchTask,
    } = AppScheduledJobsQueries.useFindTaskById(
        {
            projectID: projectId,
            appID: appId,
            scheduledJobID: scheduledJobId,
            taskID: taskId,
        },
        {
            refetchInterval: query =>
                shouldPollTaskDetails(query.state.data?.data.status) ? TASK_DETAILS_REFETCH_INTERVAL_MS : false,
        },
    );
    const task = taskResponse?.data;

    const hasActiveTask = useMemo(() => task?.status === EAppScheduledJobTaskStatus.InProgress, [task?.status]);
    const now = useScheduledJobTaskCurrentTime(hasActiveTask);
    const handleStreamClosedWhileInProgress = useCallback(() => {
        void refetchTask();
    }, [refetchTask]);

    const { mutate: cancelTask, isPending: isCancelling } = AppScheduledJobsCommands.useCancelTask({
        onSuccess: () => {
            toast.success("Task cancel requested");
        },
    });

    return (
        <section className={cn(listBox, "p-0")}>
            <div className="flex flex-col gap-5">
                {isFetching && !task ? (
                    <ScheduledJobTaskSummaryCardSkeleton variant="details" />
                ) : task ? (
                    <ScheduledJobTaskSummaryCard
                        task={task}
                        now={now}
                        variant="details"
                        isCancelling={isCancelling}
                        onCancel={id => {
                            cancelTask({ taskID: id });
                        }}
                    >
                        <ScheduledJobTaskLogsViewer
                            projectID={projectId}
                            appID={appId}
                            scheduledJobID={scheduledJobId}
                            taskID={taskId}
                            status={task.status}
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
