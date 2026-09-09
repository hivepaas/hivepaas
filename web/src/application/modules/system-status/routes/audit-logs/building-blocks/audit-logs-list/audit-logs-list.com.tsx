import { useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { ListFilter } from "lucide-react";
import { AuditLogsQueries } from "~/system-status/data";
import type { AuditLog, AuditLogScope } from "~/system-status/domain";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";

import { TablePagination } from "@/components/ui";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

import { AuditLogSummaryCard, AuditLogSummaryCardSkeleton } from "../audit-log-summary-card/audit-log-summary-card.com";
import { type AuditLogFilterValues, AuditLogsFilterBar } from "../audit-logs-filter-bar/audit-logs-filter-bar.com";

export interface AuditLogsListProps {
    scope?: AuditLogScope;
    onSelectLog?: (log: AuditLog) => void;
    className?: string;
}

export function AuditLogsList({ scope = { type: "global" }, onSelectLog, className }: AuditLogsListProps) {
    // Default to 20 items per page as requested
    const { pagination, setPagination, sorting, search, setSearch } = useTableState({
        pagination: { page: 1, size: 20 },
    });

    // Filter panel toggle state
    const [isFilterOpen, setIsFilterOpen] = useState(false);

    // Active filter values
    const [filters, setFilters] = useState<AuditLogFilterValues>({});

    const activeFilterCount = useMemo(() => {
        let count = 0;
        if (filters.type) count++;
        if (filters.source) count++;
        if (filters.result) count++;
        if (filters.actorId) count++;
        if (filters.fromDate) count++;
        if (filters.toDate) count++;
        if (filters.projectId) count++;
        if (filters.appId) count++;
        return count;
    }, [filters]);

    const hasActiveFilters = activeFilterCount > 0;

    function handleFiltersChange(nextFilters: AuditLogFilterValues) {
        setFilters(nextFilters);
        setPagination(prev => ({ ...prev, page: 1 }));
    }

    function handleClearFilters() {
        setFilters({});
        setPagination(prev => ({ ...prev, page: 1 }));
    }

    // Single expanded card state - only at most 1 item expanded at a time
    const [expandedLogId, setExpandedLogId] = useState<string | null>(null);

    // Cache of fetched detailed logs - avoids re-fetching already retrieved logs
    const [detailsCache, setDetailsCache] = useState<Record<string, AuditLog>>({});

    function handleDetailLoaded(detail: AuditLog) {
        setDetailsCache(prev => {
            if (prev[detail.id]) {
                return prev;
            }
            return { ...prev, [detail.id]: detail };
        });
    }

    function handleToggleExpand(log: AuditLog) {
        setExpandedLogId(prev => (prev === log.id ? null : log.id));
        if (onSelectLog) {
            onSelectLog(log);
        }
    }

    const { data: { data: logs, meta } = DEFAULT_PAGINATED_DATA, isFetching } = AuditLogsQueries.useFindManyPaginated({
        scope,
        pagination,
        sorting,
        search,
        type: filters.type ? [filters.type] : undefined,
        source: filters.source ? [filters.source] : undefined,
        result: filters.result ? [filters.result] : undefined,
        actorID: filters.actorId ? [filters.actorId] : undefined,
        fromDate: filters.fromDate,
        toDate: filters.toDate,
        projectID: filters.projectId,
        appID: filters.appId,
    });

    const pageCount = Math.max(1, Math.ceil(meta.page.total / pagination.size));

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
                <AuditLogsFilterBar
                    scope={scope}
                    filters={filters}
                    onChange={handleFiltersChange}
                />
            )}

            <div className="flex flex-col gap-4">
                {isFetching && logs.length === 0 ? (
                    <>
                        <AuditLogSummaryCardSkeleton />
                        <AuditLogSummaryCardSkeleton />
                        <AuditLogSummaryCardSkeleton />
                        <AuditLogSummaryCardSkeleton />
                    </>
                ) : logs.length === 0 ? (
                    <div className="rounded-lg border border-border p-8 text-center text-sm text-muted-foreground">
                        No audit logs found.
                    </div>
                ) : (
                    logs.map(log => (
                        <AuditLogSummaryCard
                            key={log.id}
                            log={log}
                            scope={scope}
                            isExpanded={expandedLogId === log.id}
                            cachedDetail={detailsCache[log.id]}
                            onDetailLoaded={handleDetailLoaded}
                            onToggleExpand={() => {
                                handleToggleExpand(log);
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
