import { useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { ListFilter, Plus } from "lucide-react";
import { EnvScheduledJobsQueries } from "~/projects/data/queries";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";
import { PermissionTooltipAction } from "@application/shared/permissions";

import { Button, DataTable } from "@/components/ui";
import { Badge } from "@/components/ui/badge";

import { type EnvScheduledJobFilterValues, EnvScheduledJobsFilterBar } from "./env-scheduled-jobs-filter-bar.com";
import { EnvScheduledJobsTableDefs } from "./env-scheduled-jobs-table.defs";

/** How many filters are set, for the Filter button's badge. */
function countFilters(filters: EnvScheduledJobFilterValues): number {
    return [filters.appId, filters.jobType, filters.status].filter(Boolean).length;
}

/** An env's scheduled jobs: its own, and those of the apps in it. */
export function EnvScheduledJobsTable({ projectId, env }: Props) {
    const { pagination, setPagination, sorting, setSorting, search, setSearch } = useTableState();
    const { navigate } = useAppNavigate();
    const [isFilterOpen, setIsFilterOpen] = useState(false);
    const [filters, setFilters] = useState<EnvScheduledJobFilterValues>({});

    const activeFilterCount = countFilters(filters);
    const hasActiveFilters = activeFilterCount > 0;

    function handleFiltersChange(nextFilters: EnvScheduledJobFilterValues) {
        setFilters(nextFilters);
        setPagination(prev => ({ ...prev, page: 1 }));
    }

    const { data: { data: jobs, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        EnvScheduledJobsQueries.useFindManyPaginated({
            projectID: projectId,
            env,
            pagination,
            sorting,
            search,
            ...(filters.jobType ? { jobTypes: [filters.jobType] } : {}),
            ...(filters.appId ? { appId: filters.appId } : {}),
            ...(filters.status ? { statuses: [filters.status] } : {}),
        });
    const columns = useMemo(() => EnvScheduledJobsTableDefs.columns(projectId, env), [projectId, env]);

    return (
        <div className="flex flex-col gap-4">
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
                                onClick={() => {
                                    handleFiltersChange({});
                                }}
                                className="text-xs text-muted-foreground hover:text-foreground font-medium underline underline-offset-4 transition-colors cursor-pointer py-1 px-1.5"
                            >
                                Clear Filter
                            </button>
                        )}
                    </div>
                }
                renderActions={
                    <PermissionTooltipAction
                        id={MODULE_IDS.Project}
                        action="write"
                    >
                        {({ isDenied }) => (
                            <Button
                                disabled={isDenied}
                                onClick={() => {
                                    navigate.modules(
                                        ROUTE.projects.single.providerConfiguration.scheduledJobs.createSequence.$route(
                                            projectId,
                                        ),
                                    );
                                }}
                            >
                                <Plus className="size-4" /> New Job Sequence
                            </Button>
                        )}
                    </PermissionTooltipAction>
                }
            />

            {isFilterOpen && (
                <EnvScheduledJobsFilterBar
                    projectId={projectId}
                    env={env}
                    filters={filters}
                    onChange={handleFiltersChange}
                />
            )}

            <DataTable
                columns={columns}
                data={jobs}
                pageSize={pagination.size}
                manualPagination
                totalCount={meta.page.total}
                manualSorting
                enableSorting
                enablePagination
                isLoading={isFetching}
                onPaginationChange={setPagination}
                onSortingChange={setSorting}
            />
        </div>
    );
}

interface Props {
    projectId: string;
    env: string;
}
