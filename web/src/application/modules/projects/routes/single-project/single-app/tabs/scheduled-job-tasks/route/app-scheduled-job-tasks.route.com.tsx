import { useMemo } from "react";

import { cn } from "@/lib/utils";
import { listBox } from "@lib/styles";
import { useParams } from "react-router";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppScheduledJobsCommands, AppScheduledJobsQueries } from "~/projects/data";
import { EAppScheduledJobTaskStatus } from "~/projects/module-shared/enums";

import { AppLink, TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { TablePagination } from "@/components/ui";

import {
    ScheduledJobTaskSummaryCard,
    ScheduledJobTaskSummaryCardSkeleton,
    useScheduledJobTaskCurrentTime,
} from "../building-blocks";

export function AppScheduledJobTasksRoute() {
    const {
        id: projectId,
        appId,
        scheduledJobId,
    } = useParams<{
        id: string;
        appId: string;
        scheduledJobId: string;
    }>();
    const { pagination, setPagination, sorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");
    invariant(scheduledJobId, "scheduledJobId must be defined");

    const { data: { data: tasks, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        AppScheduledJobsQueries.useFindTasksManyPaginated({
            projectID: projectId,
            appID: appId,
            scheduledJobID: scheduledJobId,
            pagination,
            sorting,
            search,
        });

    const hasActiveTask = useMemo(
        () => tasks.some(task => task.status === EAppScheduledJobTaskStatus.InProgress),
        [tasks],
    );
    const now = useScheduledJobTaskCurrentTime(hasActiveTask);
    const pageCount = Math.max(1, Math.ceil(meta.page.total / pagination.size));

    const { mutate: cancelTask, isPending: isCancelling } = AppScheduledJobsCommands.useCancelTask({
        onSuccess: () => {
            toast.success("Task cancel requested");
        },
    });

    return (
        <section className={cn(listBox)}>
            <div className="flex flex-col gap-5">
                <TableActions
                    search={{ value: search, onChange: setSearch }}
                    renderActions={
                        <AppLink.Basic
                            to={ROUTE.projects.single.apps.single.configuration.scheduledJobs.$route(projectId, appId)}
                            className="text-sm font-medium text-primary underline-offset-4 hover:underline"
                        >
                            Go to Scheduled Jobs
                        </AppLink.Basic>
                    }
                />

                <div className="flex flex-col gap-4">
                    {isFetching && tasks.length === 0 ? (
                        <>
                            <ScheduledJobTaskSummaryCardSkeleton />
                            <ScheduledJobTaskSummaryCardSkeleton />
                            <ScheduledJobTaskSummaryCardSkeleton />
                        </>
                    ) : tasks.length === 0 ? (
                        <div className="rounded-lg border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
                            No tasks found.
                        </div>
                    ) : (
                        tasks.map(task => (
                            <ScheduledJobTaskSummaryCard
                                key={task.id}
                                task={task}
                                now={now}
                                isCancelling={isCancelling}
                                onClick={() => {
                                    navigate.modules(
                                        ROUTE.projects.single.apps.single.scheduledJobTasks.details.$route(
                                            projectId,
                                            appId,
                                            scheduledJobId,
                                            task.id,
                                        ),
                                    );
                                }}
                                onCancel={taskID => {
                                    cancelTask({
                                        projectID: projectId,
                                        appID: appId,
                                        scheduledJobID: scheduledJobId,
                                        taskID,
                                    });
                                }}
                            />
                        ))
                    )}
                </div>

                {meta.page.total > 0 && (
                    <TablePagination
                        pageIndex={pagination.page - 1}
                        pageSize={pagination.size}
                        pageCount={pageCount}
                        totalCount={meta.page.total}
                        showPageSizeSelector={false}
                        onPageChange={pageIndex => {
                            setPagination(prev => ({ ...prev, page: pageIndex + 1 }));
                        }}
                    />
                )}
            </div>
        </section>
    );
}
