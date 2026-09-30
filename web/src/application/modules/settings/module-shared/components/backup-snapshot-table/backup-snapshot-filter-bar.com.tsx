import { useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { format } from "date-fns";
import { X } from "lucide-react";
import { type SearchableFilterItem, SearchableFilterSelect } from "~/operations/routes/tasks";
import type { BackupSnapshotRepoRef, BackupSnapshotScope } from "~/settings/domain";

import { AppsPublicQueries } from "@application/shared/data-public/queries";

import { Input } from "@/components/ui";
import { Badge } from "@/components/ui/badge";
import { DateTimePicker } from "@/components/ui/date-time-picker";

import { type BackupSnapshotFilterValues, isTagFilter, tagClassName } from "./backup-snapshot-table.helpers";

const ALL = "all";
const LIST_ALL_PAGE = { page: 1, size: 1000 };

function FilterField({ label, children }: { label: string; children: React.ReactNode }) {
    return (
        <div className="flex flex-col gap-1.5 min-w-0">
            <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">{label}</span>
            {children}
        </div>
    );
}

/** The snapshots' filters, laid out as Operations › Tasks' filter bar. */
export function BackupSnapshotFilterBar({ scope, repos, filters, onChange }: Props) {
    const [tagInput, setTagInput] = useState("");

    // An env's view lists only that env's apps; the project's view lists them all, each with its env.
    const projectId = scope.type === "project" ? scope.projectId : "";
    const env = scope.type === "project" ? scope.env : undefined;
    const { data: appsData } = AppsPublicQueries.useFindMany(
        { projectID: projectId, env, pagination: LIST_ALL_PAGE },
        { enabled: scope.type === "project" },
    );

    const repoItems: SearchableFilterItem[] = useMemo(
        () => [
            { value: ALL, label: "All Repositories" },
            ...repos.map(repo => ({ value: repo.id, label: repo.name, searchKey: repo.name })),
        ],
        [repos],
    );
    const appItems: SearchableFilterItem[] = useMemo(
        () => [
            { value: ALL, label: "All Apps" },
            ...(appsData?.data ?? []).map(app => ({
                value: app.id,
                label: app.name,
                searchKey: `${app.name} ${app.env ?? ""}`,
                badge: env ? undefined : app.env,
                avatar: { name: app.name },
            })),
        ],
        [appsData?.data, env],
    );

    function addTag() {
        const tag = tagInput.trim();
        if (!isTagFilter(tag)) {
            return;
        }
        if (!filters.tags.includes(tag)) {
            onChange({ ...filters, tags: [...filters.tags, tag] });
        }
        setTagInput("");
    }

    return (
        <div className="rounded-lg border border-border/80 bg-card/60 p-3.5 sm:p-4 shadow-2xs backdrop-blur-xs flex flex-col gap-3.5 transition-all duration-200">
            {/* One column per field on a wide screen, so the row fills the panel: a
                project's snapshots have the App field too. */}
            <div
                className={cn(
                    "grid grid-cols-1 sm:grid-cols-2 gap-3",
                    scope.type === "project" ? "md:grid-cols-3 lg:grid-cols-5" : "lg:grid-cols-4",
                )}
            >
                <FilterField label="Repository">
                    <SearchableFilterSelect
                        value={filters.repo ?? ALL}
                        onValueChange={value => {
                            onChange({ ...filters, repo: value === ALL ? undefined : value });
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
                                onChange({ ...filters, app: value === ALL ? undefined : value });
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
                            onChange({ ...filters, fromDate: date ? format(date, "yyyy-MM-dd") : undefined });
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
                            onChange({ ...filters, toDate: date ? format(date, "yyyy-MM-dd") : undefined });
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
                        <Badge
                            key={tag}
                            asChild
                            variant="outline"
                            className={cn("h-7 gap-1 px-2 font-mono text-xs cursor-pointer", tagClassName(tag))}
                        >
                            <button
                                type="button"
                                aria-label={`Remove tag ${tag}`}
                                onClick={() => {
                                    onChange({ ...filters, tags: filters.tags.filter(t => t !== tag) });
                                }}
                            >
                                {tag}
                                <X className="size-3" />
                            </button>
                        </Badge>
                    ))}
                </div>
            )}
        </div>
    );
}

interface Props {
    scope: BackupSnapshotScope;
    repos: BackupSnapshotRepoRef[];
    filters: BackupSnapshotFilterValues;
    onChange: (filters: BackupSnapshotFilterValues) => void;
}
