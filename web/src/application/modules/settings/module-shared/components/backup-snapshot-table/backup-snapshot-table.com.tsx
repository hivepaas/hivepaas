import { useCallback, useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { ListFilter } from "lucide-react";
import { useSearchParams } from "react-router";
import { BackupSnapshotQueries } from "~/settings/data/queries";
import type { BackupSnapshot, BackupSnapshotScope } from "~/settings/domain";

import { AppLink, TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useTableState } from "@application/shared/hooks/table";

import { Button, DataTable } from "@/components/ui";
import { Badge } from "@/components/ui/badge";

import { BackupSnapshotDeleteDialog } from "./backup-snapshot-delete-dialog.com";
import { BackupSnapshotDetails } from "./backup-snapshot-details.com";
import { BackupSnapshotFilterBar } from "./backup-snapshot-filter-bar.com";
import { snapshotRestoreRoute } from "./backup-snapshot-restore.helpers";
import { BackupSnapshotTableDefs } from "./backup-snapshot-table.defs";
import { type BackupSnapshotFilterValues, countActiveFilters, isTagFilter } from "./backup-snapshot-table.helpers";

/** The run that took a snapshot, where the view can link it. */
function runRoute(scope: BackupSnapshotScope, snapshot: BackupSnapshot): string | undefined {
    if (!snapshot.runId) {
        return undefined;
    }
    if (scope.type === "settings") {
        return ROUTE.operations.tasks.details.$route(snapshot.runId);
    }
    if (!snapshot.app || snapshot.app.deleted || !snapshot.app.env) {
        return undefined;
    }
    return ROUTE.projects.single.apps.single.tasks.details.$route(
        scope.projectId,
        snapshot.app.env,
        snapshot.app.id,
        snapshot.runId,
    );
}

/**
 * The snapshots a scope sees, with their filters. A link in may set the filters
 * through the URL: `repo`, `app` and `tag` (several). `fixedTags` narrow the list
 * for good: they are not shown as filters, and Clear Filter keeps them.
 */
export function BackupSnapshotTable({ scope, fixedTags = [] }: Props) {
    const [searchParams] = useSearchParams();
    const { pagination, setPagination, search, setSearch } = useTableState();
    const [filters, setFilters] = useState<BackupSnapshotFilterValues>(() => ({
        repo: searchParams.get("repo") ?? undefined,
        app: searchParams.get("app") ?? undefined,
        tags: searchParams.getAll("tag").filter(tag => isTagFilter(tag) && !fixedTags.includes(tag)),
    }));
    // Open from the start when a link in set a filter, so it shows.
    const [isFilterOpen, setIsFilterOpen] = useState(() => countActiveFilters(filters) > 0);
    const [details, setDetails] = useState<BackupSnapshot | null>(null);
    const [deleting, setDeleting] = useState<BackupSnapshot | null>(null);
    const { navigate } = useAppNavigate();

    const openRestore = useCallback(
        (snapshot: BackupSnapshot) => {
            navigate.modules(snapshotRestoreRoute(scope, snapshot.id));
        },
        [navigate, scope],
    );

    const { data = { ...DEFAULT_PAGINATED_DATA, repos: [] }, isFetching } = BackupSnapshotQueries.useFindManyPaginated({
        scope,
        pagination,
        search,
        repo: filters.repo ? [filters.repo] : undefined,
        app: filters.app ? [filters.app] : undefined,
        tag: fixedTags.length + filters.tags.length > 0 ? [...fixedTags, ...filters.tags] : undefined,
        fromDate: filters.fromDate,
        toDate: filters.toDate,
    });

    const activeFilterCount = countActiveFilters(filters);

    function handleFiltersChange(next: BackupSnapshotFilterValues) {
        setFilters(next);
        setPagination(prev => ({ ...prev, page: 1 }));
    }

    const columns = useMemo(
        () =>
            BackupSnapshotTableDefs.columns(scope, {
                onViewDetails: setDetails,
                onRestore: openRestore,
                onDelete: setDeleting,
            }),
        [scope, openRestore],
    );

    const detailsRunRoute = details ? runRoute(scope, details) : undefined;

    return (
        <div className="flex flex-col gap-4">
            <p className="text-sm text-muted-foreground">
                What each repository held when it was last synced. A repository&apos;s Sync, on its page, brings in what
                changed outside HivePaaS.
            </p>

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
                                activeFilterCount > 0 && "border-primary/50 text-foreground bg-primary/5",
                            )}
                            aria-label="Toggle filter view"
                        >
                            <ListFilter className="size-4 text-muted-foreground" />
                            <span>Filter</span>
                            {activeFilterCount > 0 && (
                                <Badge
                                    variant="secondary"
                                    className="size-4.5 p-0 flex items-center justify-center rounded-full text-[10px] font-semibold bg-primary text-primary-foreground"
                                >
                                    {activeFilterCount}
                                </Badge>
                            )}
                        </Button>

                        {activeFilterCount > 0 && (
                            <button
                                type="button"
                                onClick={() => {
                                    handleFiltersChange({ tags: [] });
                                }}
                                className="text-xs text-muted-foreground hover:text-foreground font-medium underline underline-offset-4 transition-colors cursor-pointer py-1 px-1.5"
                            >
                                Clear Filter
                            </button>
                        )}
                    </div>
                }
            />

            {isFilterOpen && (
                <BackupSnapshotFilterBar
                    scope={scope}
                    repos={data.repos}
                    filters={filters}
                    onChange={handleFiltersChange}
                />
            )}

            <DataTable
                columns={columns}
                data={data.data}
                pageSize={pagination.size}
                manualPagination
                totalCount={data.meta.page.total}
                enablePagination
                isLoading={isFetching}
                onPaginationChange={setPagination}
            />

            <BackupSnapshotDetails
                scope={scope}
                snapshot={details}
                runLink={
                    detailsRunRoute && details ? (
                        <AppLink.Modules
                            to={detailsRunRoute}
                            className="font-mono text-link hover:underline"
                        >
                            {details.runId}
                        </AppLink.Modules>
                    ) : undefined
                }
                onRestore={snapshot => {
                    setDetails(null);
                    openRestore(snapshot);
                }}
                onOpenChange={open => {
                    if (!open) {
                        setDetails(null);
                    }
                }}
            />
            <BackupSnapshotDeleteDialog
                scope={scope}
                snapshot={deleting}
                onOpenChange={open => {
                    if (!open) {
                        setDeleting(null);
                    }
                }}
            />
        </div>
    );
}

interface Props {
    scope: BackupSnapshotScope;
    /** Tags every snapshot listed carries, whatever the filters. */
    fixedTags?: string[];
}
