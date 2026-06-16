import { useMemo } from "react";

import { cn } from "@/lib/utils";
import { listBox } from "@lib/styles";
import { toast } from "sonner";
import { SystemTasksCommands, SystemTasksQueries } from "~/system-status/data";
import { SystemTaskStatus } from "~/system-status/domain";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { TablePagination } from "@/components/ui";

import { SystemTaskSummaryCard, SystemTaskSummaryCardSkeleton, useSystemTaskCurrentTime } from "../building-blocks";

export function SystemTasksRoute() {
    const { pagination, setPagination, sorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();

    const { data: { data: tasks, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        SystemTasksQueries.useFindManyPaginated({
            pagination,
            sorting,
            search,
        });

    const hasActiveTask = useMemo(() => tasks.some(task => task.status === SystemTaskStatus.InProgress), [tasks]);
    const now = useSystemTaskCurrentTime(hasActiveTask);
    const pageCount = Math.max(1, Math.ceil(meta.page.total / pagination.size));

    const { mutate: cancelTask, isPending: isCancelling } = SystemTasksCommands.useCancel({
        onSuccess: () => {
            toast.success("Task cancel requested");
        },
    });

    return (
        <section className={cn(listBox)}>
            <div className="flex flex-col gap-5">
                <TableActions search={{ value: search, onChange: setSearch }} />

                <div className="flex flex-col gap-4">
                    {isFetching && tasks.length === 0 ? (
                        <>
                            <SystemTaskSummaryCardSkeleton />
                            <SystemTaskSummaryCardSkeleton />
                            <SystemTaskSummaryCardSkeleton />
                        </>
                    ) : tasks.length === 0 ? (
                        <div className="rounded-lg border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
                            No tasks found.
                        </div>
                    ) : (
                        tasks.map(task => (
                            <SystemTaskSummaryCard
                                key={task.id}
                                task={task}
                                now={now}
                                isCancelling={isCancelling}
                                onClick={() => {
                                    navigate.modules(ROUTE.systemStatus.tasks.details.$route(task.id));
                                }}
                                onCancel={taskID => {
                                    cancelTask({ taskID });
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
