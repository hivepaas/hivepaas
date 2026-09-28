import { useMemo, useState } from "react";

import { format } from "date-fns";
import { X } from "lucide-react";
import { useSearchParams } from "react-router";
import { type SearchableFilterItem, SearchableFilterSelect } from "~/operations/routes/tasks";
import { BackupSnapshotQueries } from "~/settings/data/queries";
import type { BackupSnapshot, BackupSnapshotScope } from "~/settings/domain";

import { AppLink, TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { AppsPublicQueries } from "@application/shared/data-public/queries";
import { useTableState } from "@application/shared/hooks/table";

import { Button, DataTable, Input } from "@/components/ui";
import { DateTimePicker } from "@/components/ui/date-time-picker";

import { BackupSnapshotDeleteDialog } from "./backup-snapshot-delete-dialog.com";
import { BackupSnapshotDetails } from "./backup-snapshot-details.com";
import { BackupSnapshotTableDefs } from "./backup-snapshot-table.defs";
import { isTagFilter } from "./backup-snapshot-table.helpers";

const ALL = "all";
const LIST_ALL_PAGE = { page: 1, size: 1000 };

interface Filters {
    repo?: string;
    app?: string;
    tags: string[];
    fromDate?: string;
    toDate?: string;
}

function FilterField({ label, children }: { label: string; children: React.ReactNode }) {
    return (
        <div className="flex flex-col gap-1.5 min-w-0">
            <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">{label}</span>
            {children}
        </div>
    );
}

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
 * through the URL: `repo`, `app` and `tag` (several).
 */
export function BackupSnapshotTable({ scope }: Props) {
    const [searchParams] = useSearchParams();
    const { pagination, setPagination, search, setSearch } = useTableState();
    const [filters, setFilters] = useState<Filters>(() => ({
        repo: searchParams.get("repo") ?? undefined,
        app: searchParams.get("app") ?? undefined,
        tags: searchParams.getAll("tag").filter(isTagFilter),
    }));
    const [tagInput, setTagInput] = useState("");
    const [details, setDetails] = useState<BackupSnapshot | null>(null);
    const [deleting, setDeleting] = useState<BackupSnapshot | null>(null);

    const { data = { ...DEFAULT_PAGINATED_DATA, repos: [] }, isFetching } = BackupSnapshotQueries.useFindManyPaginated({
        scope,
        pagination,
        search,
        repo: filters.repo ? [filters.repo] : undefined,
        app: filters.app ? [filters.app] : undefined,
        tag: filters.tags.length > 0 ? filters.tags : undefined,
        fromDate: filters.fromDate,
        toDate: filters.toDate,
    });

    const projectId = scope.type === "settings" ? "" : scope.projectId;
    const { data: appsData } = AppsPublicQueries.useFindMany(
        { projectID: projectId, pagination: LIST_ALL_PAGE },
        { enabled: scope.type === "project" },
    );

    const repoItems: SearchableFilterItem[] = useMemo(
        () => [
            { value: ALL, label: "All Repositories" },
            ...data.repos.map(repo => ({ value: repo.id, label: repo.name, searchKey: repo.name })),
        ],
        [data.repos],
    );
    const appItems: SearchableFilterItem[] = useMemo(
        () => [
            { value: ALL, label: "All Apps" },
            ...(appsData?.data ?? []).map(app => ({
                value: app.id,
                label: app.name,
                searchKey: app.name,
                avatar: { name: app.name },
            })),
        ],
        [appsData?.data],
    );

    const columns = useMemo(
        () =>
            BackupSnapshotTableDefs.columns(scope, {
                onViewDetails: setDetails,
                onDelete: setDeleting,
            }),
        [scope],
    );

    function addTag() {
        const tag = tagInput.trim();
        if (!isTagFilter(tag) || filters.tags.includes(tag)) {
            return;
        }
        setFilters({ ...filters, tags: [...filters.tags, tag] });
        setTagInput("");
    }

    const detailsRunRoute = details ? runRoute(scope, details) : undefined;

    return (
        <div className="flex flex-col gap-4">
            <p className="text-sm text-muted-foreground">
                What each repository held when it was last synced. A repository&apos;s Sync, on its page, brings in what
                changed outside HivePaaS.
            </p>

            <div className="rounded-lg border border-border/80 bg-card/60 p-3.5 sm:p-4 shadow-2xs flex flex-col gap-3.5">
                <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
                    <FilterField label="Repository">
                        <SearchableFilterSelect
                            value={filters.repo ?? ALL}
                            onValueChange={value => {
                                setFilters({ ...filters, repo: value === ALL ? undefined : value });
                            }}
                            placeholder="All Repositories"
                            searchPlaceholder="Search repositories..."
                            emptyText="No repositories found."
                            items={repoItems}
                        />
                    </FilterField>
                    {scope.type === "project" && (
                        <FilterField label="App">
                            <SearchableFilterSelect
                                value={filters.app ?? ALL}
                                onValueChange={value => {
                                    setFilters({ ...filters, app: value === ALL ? undefined : value });
                                }}
                                placeholder="All Apps"
                                searchPlaceholder="Search apps..."
                                emptyText="No apps found."
                                items={appItems}
                            />
                        </FilterField>
                    )}
                    <FilterField label="Tag">
                        <Input
                            value={tagInput}
                            onChange={event => {
                                setTagInput(event.target.value);
                            }}
                            onKeyDown={event => {
                                if (event.key === "Enter") {
                                    event.preventDefault();
                                    addTag();
                                }
                            }}
                            placeholder="key:value, then Enter"
                            className="h-9 text-xs sm:text-sm font-mono"
                        />
                    </FilterField>
                    <FilterField label="From Date">
                        <DateTimePicker
                            value={filters.fromDate ? new Date(`${filters.fromDate}T00:00:00`) : undefined}
                            onChange={date => {
                                setFilters({ ...filters, fromDate: date ? format(date, "yyyy-MM-dd") : undefined });
                            }}
                            placeholder="From date"
                            granularity="day"
                            showClearButton
                            className="h-9 text-xs sm:text-sm"
                        />
                    </FilterField>
                    <FilterField label="To Date">
                        <DateTimePicker
                            value={filters.toDate ? new Date(`${filters.toDate}T00:00:00`) : undefined}
                            onChange={date => {
                                setFilters({ ...filters, toDate: date ? format(date, "yyyy-MM-dd") : undefined });
                            }}
                            placeholder="To date"
                            granularity="day"
                            showClearButton
                            className="h-9 text-xs sm:text-sm"
                        />
                    </FilterField>
                </div>
                {filters.tags.length > 0 && (
                    <div className="flex flex-wrap gap-2">
                        {filters.tags.map(tag => (
                            <Button
                                key={tag}
                                type="button"
                                variant="outline"
                                size="sm"
                                className="h-7 gap-1 font-mono text-xs"
                                onClick={() => {
                                    setFilters({ ...filters, tags: filters.tags.filter(t => t !== tag) });
                                }}
                            >
                                {tag}
                                <X className="size-3" />
                            </Button>
                        ))}
                    </div>
                )}
            </div>

            <TableActions search={{ value: search, onChange: setSearch }} />
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
}
