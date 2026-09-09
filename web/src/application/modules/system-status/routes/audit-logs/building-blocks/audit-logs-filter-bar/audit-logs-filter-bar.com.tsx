import { cn } from "@/lib/utils";
import { format } from "date-fns";
import { AuditLogsQueries } from "~/system-status/data";
import { AuditLogResult, type AuditLogScope, AuditLogSource, AuditLogType } from "~/system-status/domain";

import { AppsPublicQueries, ProjectsPublicQueries, UsersPublicQueries } from "@application/shared/data-public/queries";

import { Avatar } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { DateTimePicker } from "@/components/ui/date-time-picker";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import { ResultBadge } from "../audit-log-summary-card/audit-log-summary-card.com";

export interface AuditLogFilterValues {
    type?: string;
    source?: string;
    result?: string;
    actorId?: string;
    fromDate?: string;
    toDate?: string;
    projectId?: string;
    appId?: string;
}

export interface AuditLogsFilterBarProps {
    scope?: AuditLogScope;
    filters: AuditLogFilterValues;
    onChange: (filters: AuditLogFilterValues) => void;
    className?: string;
}

const ALL_SOURCES = Object.values(AuditLogSource);
const ALL_TYPES_FALLBACK = Object.values(AuditLogType);

export function AuditLogsFilterBar({ scope, filters, onChange, className }: AuditLogsFilterBarProps) {
    const isGlobalScope = !scope || scope.type === "global";
    const isProjectOrEnvScope = scope?.type === "project" || scope?.type === "project-env";
    const projectID = isProjectOrEnvScope ? scope.projectID : "";

    // 1. Fetch available types from backend (with static fallback)
    const { data: typesResponse } = AuditLogsQueries.useFindTypes({ scope });
    const availableTypes =
        typesResponse?.data && typesResponse.data.length > 0 ? typesResponse.data : ALL_TYPES_FALLBACK;

    // 2. Fetch users for actor filter
    const { data: usersResponse } = UsersPublicQueries.useFindManyBase({});
    const users = usersResponse?.data ?? [];

    // 3. Fetch projects for project filter (global scope only)
    const { data: projectsResponse } = ProjectsPublicQueries.useFindManyPaginated({}, { enabled: isGlobalScope });
    const projects = projectsResponse?.data ?? [];

    // 4. Fetch apps for app filter (project / project-env scope only)
    const { data: appsResponse } = AppsPublicQueries.useFindManyBase(
        { projectID },
        { enabled: isProjectOrEnvScope && Boolean(projectID) },
    );
    const apps = appsResponse?.data ?? [];

    function handleProjectChange(val: string) {
        onChange({
            ...filters,
            projectId: val === "all" ? undefined : val,
        });
    }

    function handleAppChange(val: string) {
        onChange({
            ...filters,
            appId: val === "all" ? undefined : val,
        });
    }

    function handleTypeChange(val: string) {
        onChange({
            ...filters,
            type: val === "all" ? undefined : val,
        });
    }

    function handleSourceChange(val: string) {
        onChange({
            ...filters,
            source: val === "all" ? undefined : val,
        });
    }

    function handleResultChange(val: string) {
        onChange({
            ...filters,
            result: val === "all" ? undefined : val,
        });
    }

    function handleActorChange(val: string) {
        onChange({
            ...filters,
            actorId: val === "all" ? undefined : val,
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
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-7 gap-3">
                {/* Global Scope: Project Dropdown */}
                {isGlobalScope && (
                    <div className="flex flex-col gap-1.5 min-w-0">
                        <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                            Project
                        </span>
                        <Select
                            value={filters.projectId ?? "all"}
                            onValueChange={handleProjectChange}
                        >
                            <SelectTrigger className="h-9 w-full">
                                <SelectValue placeholder="All Projects" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="all">
                                    <span className="text-muted-foreground">All Projects</span>
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

                {/* Project / Env Scope: App Dropdown */}
                {isProjectOrEnvScope && (
                    <div className="flex flex-col gap-1.5 min-w-0">
                        <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                            App
                        </span>
                        <Select
                            value={filters.appId ?? "all"}
                            onValueChange={handleAppChange}
                        >
                            <SelectTrigger className="h-9 w-full">
                                <SelectValue placeholder="All Apps" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="all">
                                    <span className="text-muted-foreground">All Apps</span>
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

                {/* 1. Type Dropdown */}
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
                                <span className="text-muted-foreground">All Types</span>
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

                {/* 2. Source Dropdown */}
                <div className="flex flex-col gap-1.5 min-w-0">
                    <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                        Source
                    </span>
                    <Select
                        value={filters.source ?? "all"}
                        onValueChange={handleSourceChange}
                    >
                        <SelectTrigger className="h-9 w-full">
                            <SelectValue placeholder="All Sources" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">
                                <span className="text-muted-foreground">All Sources</span>
                            </SelectItem>
                            {ALL_SOURCES.map(sourceItem => (
                                <SelectItem
                                    key={sourceItem}
                                    value={sourceItem}
                                >
                                    <Badge
                                        variant="secondary"
                                        className="font-mono text-xs px-2 py-0.5 rounded-md text-muted-foreground bg-muted/40 border border-border/40"
                                    >
                                        {sourceItem}
                                    </Badge>
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </div>

                {/* 3. Result Dropdown */}
                <div className="flex flex-col gap-1.5 min-w-0">
                    <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                        Result
                    </span>
                    <Select
                        value={filters.result ?? "all"}
                        onValueChange={handleResultChange}
                    >
                        <SelectTrigger className="h-9 w-full">
                            <SelectValue placeholder="All Results" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">
                                <span className="text-muted-foreground">All Results</span>
                            </SelectItem>
                            <SelectItem value={AuditLogResult.Allowed}>
                                <ResultBadge result={AuditLogResult.Allowed} />
                            </SelectItem>
                            <SelectItem value={AuditLogResult.Denied}>
                                <ResultBadge result={AuditLogResult.Denied} />
                            </SelectItem>
                        </SelectContent>
                    </Select>
                </div>

                {/* 4. Actor Dropdown */}
                <div className="flex flex-col gap-1.5 min-w-0">
                    <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                        Actor
                    </span>
                    <Select
                        value={filters.actorId ?? "all"}
                        onValueChange={handleActorChange}
                    >
                        <SelectTrigger className="h-9 w-full">
                            <SelectValue placeholder="All Actors" />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="all">
                                <span className="text-muted-foreground">All Actors</span>
                            </SelectItem>
                            {users.map(user => {
                                const userName = user.fullName ? user.fullName : user.username;
                                return (
                                    <SelectItem
                                        key={user.id}
                                        value={user.id}
                                    >
                                        <div className="flex items-center gap-2 min-w-0">
                                            <Avatar
                                                name={userName}
                                                src={user.photo}
                                                className="size-4.5 text-[9px] shrink-0"
                                            />
                                            <span className="truncate">{userName}</span>
                                        </div>
                                    </SelectItem>
                                );
                            })}
                        </SelectContent>
                    </Select>
                </div>

                {/* 5. From Date */}
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

                {/* 6. To Date */}
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
