import { useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { ListFilter } from "lucide-react";
import { toast } from "sonner";
import { SystemTasksCommands, SystemTasksQueries } from "~/operations/data";
import { type SystemTask, type SystemTaskScope, SystemTaskStatus } from "~/operations/domain";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { Button, TablePagination } from "@/components/ui";
import { Badge } from "@/components/ui/badge";

import {
    type SystemTaskFilterValues,
    SystemTasksFilterBar,
} from "../system-tasks-filter-bar/system-tasks-filter-bar.com";
import {
    SystemTaskSummaryCard,
    SystemTaskSummaryCardSkeleton,
} from "../task-summary-card/system-task-summary-card.com";
import { useSystemTaskCurrentTime } from "../task-summary-card/system-task-summary-card.hooks";

export interface SystemTasksListProps {
    scope?: SystemTaskScope;
    onSelectTask?: (task: SystemTask) => void;
    className?: string;
}

export function SystemTasksList({ scope = { type: "global" }, onSelectTask, className }: SystemTasksListProps) {
    const { pagination, setPagination, sorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();

    // Filter panel toggle state
    const [isFilterOpen, setIsFilterOpen] = useState(false);

    // Active filter values
    const [filters, setFilters] = useState<SystemTaskFilterValues>({});

    const activeFilterCount = useMemo(() => {
        let count = 0;
        if (filters.type) count++;
        if (filters.status) count++;
        if (filters.fromDate) count++;
        if (filters.toDate) count++;
        if (filters.projectId) count++;
        if (filters.appId) count++;
        if (filters.scopeOnly) count++;
        return count;
    }, [filters]);

    const hasActiveFilters = activeFilterCount > 0;

    function handleFiltersChange(nextFilters: SystemTaskFilterValues) {
        setFilters(nextFilters);
        setPagination(prev => ({ ...prev, page: 1 }));
    }

    function handleClearFilters() {
        setFilters({});
        setPagination(prev => ({ ...prev, page: 1 }));
    }

    const { data: { data: tasks, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        SystemTasksQueries.useFindManyPaginated({
            scope,
            pagination,
            sorting,
            search,
            type: filters.type ? [filters.type] : undefined,
            status: filters.status ? [filters.status] : undefined,
            fromDate: filters.fromDate,
            toDate: filters.toDate,
            projectID: filters.projectId,
            appID: filters.appId,
            scopeOnly: filters.scopeOnly,
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
        <div className={cn("flex flex-col gap-4", className)}>
            <TableActions
                search={{ value: search, onChange: setSearch }}
                renderAfterSearch={
                    <div className="flex items-center gap-2">
                        <Button
                            variant={isFilterOpen ? "secondary" : "outline"}
                            size="sm"
                            onClick={() => {
                                setIsFilterOpen(prev => !prev);
                            }}
                            className={cn(
                                "h-9 gap-1.5 px-3 font-medium transition-colors",
                                hasActiveFilters && "border-primary/50 text-foreground bg-primary/5",
                            )}
                            aria-label="Toggle filter view"
                        >
                            <ListFilter className="size-4 text-muted-foreground" />
                            <span>Filter</span>
                            {hasActiveFilters && (
                                <Badge
                                    variant="secondary"
                                    className="size-4.5 p-0 flex items-center justify-center rounded-full text-[10px] font-semibold bg-primary text-primary-foreground"
                                >
                                    {activeFilterCount}
                                </Badge>
                            )}
                        </Button>

                        {hasActiveFilters && (
                            <button
                                type="button"
                                onClick={handleClearFilters}
                                className="text-xs text-muted-foreground hover:text-foreground font-medium underline underline-offset-4 transition-colors cursor-pointer py-1 px-1.5"
                            >
                                Clear Filter
                            </button>
                        )}
                    </div>
                }
            />

            {isFilterOpen && (
                <SystemTasksFilterBar
                    scope={scope}
                    filters={filters}
                    onChange={handleFiltersChange}
                />
            )}

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
                                if (onSelectTask) {
                                    onSelectTask(task);
                                } else {
                                    navigate.modules(ROUTE.operations.tasks.details.$route(task.id));
                                }
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
                    pageSizeOptions={[10, 20, 50, 100]}
                    onPageChange={pageIndex => {
                        setPagination(prev => ({ ...prev, page: pageIndex + 1 }));
                    }}
                    onPageSizeChange={pageSize => {
                        setPagination(prev => ({ ...prev, size: pageSize, page: 1 }));
                    }}
                />
            )}
        </div>
    );
}
