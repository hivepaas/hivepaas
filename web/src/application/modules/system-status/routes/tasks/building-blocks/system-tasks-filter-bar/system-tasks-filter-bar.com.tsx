import { cn } from "@/lib/utils";
import { format } from "date-fns";
import { SystemTasksQueries } from "~/system-status/data";
import { type SystemTaskScope, SystemTaskStatus } from "~/system-status/domain";

import { AppsPublicQueries, ProjectsPublicQueries } from "@application/shared/data-public/queries";

import { Avatar } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { DateTimePicker } from "@/components/ui/date-time-picker";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

export interface SystemTaskFilterValues {
    projectId?: string;
    appId?: string;
    scopeOnly?: boolean;
    type?: string;
    status?: SystemTaskStatus;
    fromDate?: string;
    toDate?: string;
}

export interface SystemTasksFilterBarProps {
    scope?: SystemTaskScope;
    filters: SystemTaskFilterValues;
    onChange: (filters: SystemTaskFilterValues) => void;
    className?: string;
}

const ALL_TASK_STATUSES = [
    SystemTaskStatus.NotStarted,
    SystemTaskStatus.InProgress,
    SystemTaskStatus.Done,
    SystemTaskStatus.Failed,
    SystemTaskStatus.Canceled,
];

const STATUS_LABELS: Record<SystemTaskStatus, string> = {
    [SystemTaskStatus.NotStarted]: "Not Started",
    [SystemTaskStatus.InProgress]: "In-Progress",
    [SystemTaskStatus.Done]: "Done",
    [SystemTaskStatus.Failed]: "Failed",
    [SystemTaskStatus.Canceled]: "Canceled",
};

const STATUS_BADGE_CLASS_NAMES: Record<SystemTaskStatus, string> = {
    [SystemTaskStatus.Done]: "bg-green-500 text-white hover:bg-green-500/90",
    [SystemTaskStatus.Failed]: "bg-red-500 text-white hover:bg-red-500/90",
    [SystemTaskStatus.InProgress]: "bg-purple-400 text-white hover:bg-purple-400/90",
    [SystemTaskStatus.NotStarted]: "bg-blue-400 text-white hover:bg-blue-400/90",
    [SystemTaskStatus.Canceled]: "bg-amber-500 text-white hover:bg-amber-500/90",
};

const ALL_TASK_TYPES_FALLBACK = [
    "task:app-deploy",
    "task:app-clone",
    "task:app-preview",
    "task:sched-job-exec",
    "task:periodic-exec",
    "task:system-update",
    "task:workflow",
    "task:settings-revert",
    "task:app-labels-sweep",
];

export function SystemTasksFilterBar({ scope, filters, onChange, className }: SystemTasksFilterBarProps) {
    const isGlobalScope = !scope || scope.type === "global";
    const isProjectOrEnvScope = scope?.type === "project" || scope?.type === "project-env";
    const projectID = isProjectOrEnvScope ? scope.projectID : "";

    // 1. Fetch available types from backend (with static fallback)
    const { data: typesResponse } = SystemTasksQueries.useFindTypes({ scope });
    const availableTypes =
        typesResponse?.data && typesResponse.data.length > 0 ? typesResponse.data : ALL_TASK_TYPES_FALLBACK;

    // 2. Fetch projects for project filter (global scope only)
    const { data: projectsResponse } = ProjectsPublicQueries.useFindManyPaginated({}, { enabled: isGlobalScope });
    const projects = projectsResponse?.data ?? [];

    // 3. Fetch apps for app filter (project / project-env scope only)
    const { data: appsResponse } = AppsPublicQueries.useFindManyBase(
        { projectID },
        { enabled: isProjectOrEnvScope && Boolean(projectID) },
    );
    const apps = appsResponse?.data ?? [];

    const projectSelectValue = filters.scopeOnly ? "scope-only" : (filters.projectId ?? "all");
    const appSelectValue = filters.scopeOnly ? "scope-only" : (filters.appId ?? "all");

    function handleProjectChange(val: string) {
        if (val === "scope-only") {
            onChange({
                ...filters,
                projectId: undefined,
                scopeOnly: true,
            });
        } else if (val === "all") {
            onChange({
                ...filters,
                projectId: undefined,
                scopeOnly: undefined,
            });
        } else {
            onChange({
                ...filters,
                projectId: val,
                scopeOnly: undefined,
            });
        }
    }

    function handleAppChange(val: string) {
        if (val === "scope-only") {
            onChange({
                ...filters,
                appId: undefined,
                scopeOnly: true,
            });
        } else if (val === "all") {
            onChange({
                ...filters,
                appId: undefined,
                scopeOnly: undefined,
            });
        } else {
            onChange({
                ...filters,
                appId: val,
                scopeOnly: undefined,
            });
        }
    }

    function handleTypeChange(val: string) {
        onChange({
            ...filters,
            type: val === "all" ? undefined : val,
        });
    }

    function handleStatusChange(val: string) {
        onChange({
            ...filters,
            status: val === "all" ? undefined : (val as SystemTaskStatus),
        });
    }

    function handleFromDateChange(date: Date | undefined) {
        onChange({
            ...filters,
            fromDate: date ? format(date, "yyyy-MM-dd") : undefined,
        });
    }

    function handleToDateChange(date: Date | undefined) {
        onChange({
            ...filters,
            toDate: date ? format(date, "yyyy-MM-dd") : undefined,
        });
    }

    const fromDateValue = filters.fromDate ? new Date(`${filters.fromDate}T00:00:00`) : undefined;
    const toDateValue = filters.toDate ? new Date(`${filters.toDate}T00:00:00`) : undefined;

    return (
        <div
            className={cn(
                "rounded-lg border border-border/80 bg-card/60 p-3.5 sm:p-4 shadow-2xs backdrop-blur-xs flex flex-col gap-3.5 transition-all duration-200",
                className,
            )}
        >
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
                {/* 1. Project Dropdown (Global scope only) */}
                {isGlobalScope && (
                    <div className="flex flex-col gap-1.5 min-w-0">
                        <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                            Project
                        </span>
                        <Select
                            value={projectSelectValue}
                            onValueChange={handleProjectChange}
                        >
                            <SelectTrigger className="h-9 w-full">
                                <SelectValue placeholder="All Projects" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="all">
                                    <span>All Projects</span>
                                </SelectItem>
                                <SelectItem value="scope-only">
                                    <span>Global Only</span>
                                </SelectItem>
                                {projects.map(project => (
                                    <SelectItem
                                        key={project.id}
                                        value={project.id}
                                    >
                                        <div className="flex items-center gap-2 min-w-0">
                                            <Avatar
                                                name={project.name}
                                                className="size-4.5 text-[9px] shrink-0"
                                            />
                                            <span className="truncate">{project.name}</span>
                                        </div>
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>
                )}

                {/* App Dropdown (project / project-env scope only) */}
                {isProjectOrEnvScope && (
                    <div className="flex flex-col gap-1.5 min-w-0">
                        <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                            App
                        </span>
                        <Select
                            value={appSelectValue}
                            onValueChange={handleAppChange}
                        >
                            <SelectTrigger className="h-9 w-full">
                                <SelectValue placeholder="All Apps" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="all">
                                    <span>All Apps</span>
                                </SelectItem>
                                <SelectItem value="scope-only">
                                    <span>Project Only</span>
                                </SelectItem>
                                {apps.map(app => (
                                    <SelectItem
                                        key={app.id}
                                        value={app.id}
                                    >
                                        <div className="flex items-center gap-2 min-w-0">
                                            <Avatar
                                                name={app.name}
                                                className="size-4.5 text-[9px] shrink-0"
                                            />
                                            <span className="truncate">{app.name}</span>
                                        </div>
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        </Select>
                    </div>
                )}

                {/* 2. Type Dropdown */}
                <div className="flex flex-col gap-1.5 min-w-0">
                    <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                        Type
                    </span>
                    <Select
                        value={filters.type ?? "all"}
                        onValueChange={handleTypeChange}
                    >
                        <SelectTrigger className="h-9 w-full">
                            <SelectValue placeholder="All Types" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">
                                <span>All Types</span>
                            </SelectItem>
                            {availableTypes.map(typeItem => (
                                <SelectItem
                                    key={typeItem}
                                    value={typeItem}
                                >
                                    <Badge
                                        variant="outline"
                                        className="font-mono text-xs px-2 py-0.5 rounded-md border-border/70 bg-muted/50 text-foreground"
                                    >
                                        {typeItem}
                                    </Badge>
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </div>

                {/* 3. Status Dropdown */}
                <div className="flex flex-col gap-1.5 min-w-0">
                    <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                        Status
                    </span>
                    <Select
                        value={filters.status ?? "all"}
                        onValueChange={handleStatusChange}
                    >
                        <SelectTrigger className="h-9 w-full">
                            <SelectValue placeholder="All Statuses" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">
                                <span>All Statuses</span>
                            </SelectItem>
                            {ALL_TASK_STATUSES.map(statusItem => (
                                <SelectItem
                                    key={statusItem}
                                    value={statusItem}
                                >
                                    <Badge
                                        className={cn(
                                            "h-6 px-2.5 text-xs font-semibold shadow-2xs",
                                            STATUS_BADGE_CLASS_NAMES[statusItem],
                                        )}
                                    >
                                        {STATUS_LABELS[statusItem]}
                                    </Badge>
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </div>

                {/* 4. From Date */}
                <div className="flex flex-col gap-1.5 min-w-0">
                    <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                        From Date
                    </span>
                    <DateTimePicker
                        value={fromDateValue}
                        onChange={handleFromDateChange}
                        placeholder="From date"
                        granularity="day"
                        showClearButton
                        className="h-9 text-xs sm:text-sm"
                    />
                </div>

                {/* 5. To Date */}
                <div className="flex flex-col gap-1.5 min-w-0">
                    <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                        To Date
                    </span>
                    <DateTimePicker
                        value={toDateValue}
                        onChange={handleToDateChange}
                        placeholder="To date"
                        granularity="day"
                        showClearButton
                        className="h-9 text-xs sm:text-sm"
                    />
                </div>
            </div>
        </div>
    );
}
