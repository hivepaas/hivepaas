import { useMemo } from "react";

import { cn } from "@/lib/utils";
import { SearchableFilterSelect } from "~/operations/routes/tasks";
import { ProjectAppsQueries } from "~/projects/data/queries";
import { EAppScheduledJobType } from "~/projects/module-shared/enums";
import { SettingStatusBadge } from "~/settings/module-shared/components";

import { ESettingStatus } from "@application/shared/enums";

import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

const ALL = "all";
const LIST_ALL_PAGE = { page: 1, size: 1000 };

const JOB_TYPES = [
    EAppScheduledJobType.JobSequence,
    EAppScheduledJobType.ContainerCommand,
    EAppScheduledJobType.DataBackup,
    EAppScheduledJobType.FunctionInvoke,
];
const STATUSES = [ESettingStatus.Active, ESettingStatus.Disabled, ESettingStatus.Pending];

export interface EnvScheduledJobFilterValues {
    appId?: string;
    jobType?: EAppScheduledJobType;
    status?: ESettingStatus;
}

function FilterField({ label, children }: { label: string; children: React.ReactNode }) {
    return (
        <div className="flex flex-col gap-1.5 min-w-0">
            <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">{label}</span>
            {children}
        </div>
    );
}

/** The filters of an env's scheduled jobs, laid out as the task filters are. */
export function EnvScheduledJobsFilterBar({ projectId, env, filters, onChange }: Props) {
    const { data: appsData } = ProjectAppsQueries.useFindManyPaginated({
        projectID: projectId,
        env,
        pagination: LIST_ALL_PAGE,
    });

    const appItems = useMemo(
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

    return (
        <div
            className={cn(
                "rounded-lg border border-border/80 bg-card/60 p-3.5 sm:p-4 shadow-2xs backdrop-blur-xs flex flex-col gap-3.5 transition-all duration-200",
            )}
        >
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
                <FilterField label="App">
                    <SearchableFilterSelect
                        value={filters.appId ?? ALL}
                        onValueChange={value => {
                            onChange({ ...filters, appId: value === ALL ? undefined : value });
                        }}
                        placeholder="All Apps"
                        searchPlaceholder="Search apps..."
                        emptyText="No apps found."
                        items={appItems}
                    />
                </FilterField>

                <FilterField label="Type">
                    <Select
                        value={filters.jobType ?? ALL}
                        onValueChange={value => {
                            onChange({
                                ...filters,
                                jobType: value === ALL ? undefined : (value as EAppScheduledJobType),
                            });
                        }}
                    >
                        <SelectTrigger className="h-9 w-full">
                            <SelectValue placeholder="All Types" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value={ALL}>
                                <span>All Types</span>
                            </SelectItem>
                            {JOB_TYPES.map(jobType => (
                                <SelectItem
                                    key={jobType}
                                    value={jobType}
                                >
                                    <Badge
                                        variant="outline"
                                        className="font-mono text-xs px-2 py-0.5 rounded-md border-border/70 bg-muted/50 text-foreground"
                                    >
                                        {jobType}
                                    </Badge>
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </FilterField>

                <FilterField label="Status">
                    <Select
                        value={filters.status ?? ALL}
                        onValueChange={value => {
                            onChange({ ...filters, status: value === ALL ? undefined : (value as ESettingStatus) });
                        }}
                    >
                        <SelectTrigger className="h-9 w-full">
                            <SelectValue placeholder="All Statuses" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value={ALL}>
                                <span>All Statuses</span>
                            </SelectItem>
                            {STATUSES.map(status => (
                                <SelectItem
                                    key={status}
                                    value={status}
                                >
                                    <SettingStatusBadge status={status} />
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </FilterField>
            </div>
        </div>
    );
}

interface Props {
    projectId: string;
    env: string;
    filters: EnvScheduledJobFilterValues;
    onChange: (filters: EnvScheduledJobFilterValues) => void;
}
