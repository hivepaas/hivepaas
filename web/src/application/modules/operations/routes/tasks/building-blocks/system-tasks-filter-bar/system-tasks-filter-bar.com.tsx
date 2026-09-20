import { useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { format } from "date-fns";
import { Check, ChevronDown } from "lucide-react";
import { SystemTasksQueries } from "~/operations/data";
import { type SystemTaskScope, SystemTaskStatus } from "~/operations/domain";

import { AppsPublicQueries, ProjectsPublicQueries } from "@application/shared/data-public/queries";

import { Avatar } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command";
import { DateTimePicker } from "@/components/ui/date-time-picker";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

interface SearchableFilterItem {
    value: string;
    label: string;
    searchKey?: string;
    badge?: string;
    avatar?: {
        name: string;
        src?: string | null;
    };
}

interface SearchableFilterSelectProps {
    value: string;
    onValueChange: (value: string) => void;
    placeholder: string;
    searchPlaceholder: string;
    emptyText: string;
    items: SearchableFilterItem[];
}

function SearchableFilterSelect({
    value,
    onValueChange,
    placeholder,
    searchPlaceholder,
    emptyText,
    items,
}: SearchableFilterSelectProps) {
    const [open, setOpen] = useState(false);

    const selectedItem = useMemo(() => items.find(item => item.value === value), [items, value]);

    return (
        <Popover
            open={open}
            onOpenChange={setOpen}
        >
            <PopoverTrigger asChild>
                <Button
                    type="button"
                    variant="outline"
                    role="combobox"
                    aria-expanded={open}
                    className="h-9 w-full justify-between gap-2 px-3 py-2 text-sm font-normal border-input bg-transparent dark:bg-input/15 dark:hover:bg-input/25 overflow-hidden shadow-xs"
                >
                    <span className="flex items-center gap-2 min-w-0 flex-1 truncate text-left">
                        {selectedItem?.avatar && (
                            <Avatar
                                name={selectedItem.avatar.name}
                                src={selectedItem.avatar.src}
                                className="size-4.5 text-[9px] shrink-0"
                            />
                        )}
                        {selectedItem?.badge && (
                            <Badge
                                variant="outline"
                                className="font-mono text-[11px] px-1.5 py-0.5 rounded-md shrink-0 border-border/70 bg-muted/50 text-foreground"
                            >
                                {selectedItem.badge}
                            </Badge>
                        )}
                        <span className="truncate">{selectedItem ? selectedItem.label : placeholder}</span>
                    </span>
                    <ChevronDown className="size-4 shrink-0 opacity-50 text-muted-foreground" />
                </Button>
            </PopoverTrigger>
            <PopoverContent
                className="w-60 min-w-[var(--radix-popover-trigger-width)] p-0"
                align="start"
            >
                <Command>
                    <CommandInput placeholder={searchPlaceholder} />
                    <CommandList className="max-h-60">
                        <CommandEmpty className="p-2 text-xs text-muted-foreground text-center">
                            {emptyText}
                        </CommandEmpty>
                        <CommandGroup>
                            {items.map(item => (
                                <CommandItem
                                    key={item.value}
                                    value={`${item.searchKey ?? item.label} ${item.value}`}
                                    onSelect={() => {
                                        onValueChange(item.value);
                                        setOpen(false);
                                    }}
                                    className="flex items-center justify-between gap-2 py-1.5 cursor-pointer"
                                >
                                    <div className="flex items-center gap-2 min-w-0 flex-1">
                                        {item.avatar && (
                                            <Avatar
                                                name={item.avatar.name}
                                                src={item.avatar.src}
                                                className="size-4.5 text-[9px] shrink-0"
                                            />
                                        )}
                                        {item.badge && (
                                            <Badge
                                                variant="outline"
                                                className="font-mono text-[11px] px-1.5 py-0.5 rounded-md shrink-0 border-border/70 bg-muted/50 text-foreground"
                                            >
                                                {item.badge}
                                            </Badge>
                                        )}
                                        <span className="truncate">{item.label}</span>
                                    </div>
                                    {value === item.value && <Check className="size-4 shrink-0 text-primary" />}
                                </CommandItem>
                            ))}
                        </CommandGroup>
                    </CommandList>
                </Command>
            </PopoverContent>
        </Popover>
    );
}

export interface SystemTaskFilterValues {
    projectId?: string;
    appId?: string;
    scopeOnly?: boolean;
    type?: string;
    targetId?: string;
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
    const isAppScope = scope?.type === "app";
    const projectID = isProjectOrEnvScope ? scope.projectID : "";

    // 1. Fetch available types from backend (with static fallback)
    const { data: typesResponse } = SystemTasksQueries.useFindTypes({ scope });
    const availableTypes =
        typesResponse?.data && typesResponse.data.length > 0 ? typesResponse.data : ALL_TASK_TYPES_FALLBACK;

    // 2. Fetch projects for project filter (global scope only)
    const { data: projectsResponse } = ProjectsPublicQueries.useFindManyPaginated({}, { enabled: isGlobalScope });

    // 3. Fetch apps for app filter (project / project-env scope only)
    const { data: appsResponse } = AppsPublicQueries.useFindManyBase(
        { projectID },
        { enabled: isProjectOrEnvScope && Boolean(projectID) },
    );

    // 4. Fetch target objects for Target Job filter (app scope only)
    const { data: targetObjectsResponse } = SystemTasksQueries.useFindTargetObjects({ scope }, { enabled: isAppScope });

    const projectSelectValue = filters.scopeOnly ? "scope-only" : (filters.projectId ?? "all");
    const appSelectValue = filters.scopeOnly ? "scope-only" : (filters.appId ?? "all");

    const projectItems: SearchableFilterItem[] = useMemo(() => {
        const projects = projectsResponse?.data ?? [];
        return [
            { value: "all", label: "All Projects" },
            { value: "scope-only", label: "Global Only" },
            ...projects.map(project => ({
                value: project.id,
                label: project.name,
                searchKey: project.name,
                avatar: { name: project.name },
            })),
        ];
    }, [projectsResponse?.data]);

    const appItems: SearchableFilterItem[] = useMemo(() => {
        const apps = appsResponse?.data ?? [];
        return [
            { value: "all", label: "All Apps" },
            { value: "scope-only", label: "Project Only" },
            ...apps.map(app => ({
                value: app.id,
                label: app.name,
                searchKey: app.name,
                avatar: { name: app.name },
            })),
        ];
    }, [appsResponse?.data]);

    const targetJobItems: SearchableFilterItem[] = useMemo(() => {
        const targetObjects = targetObjectsResponse?.data ?? [];
        return [
            { value: "all", label: "All Jobs" },
            ...targetObjects.map(item => ({
                value: item.id,
                label: item.name,
                badge: item.type,
                searchKey: `${item.type} ${item.name}`,
            })),
        ];
    }, [targetObjectsResponse?.data]);

    function handleTargetJobChange(val: string) {
        onChange({
            ...filters,
            targetId: val === "all" ? undefined : val,
        });
    }

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
                        <SearchableFilterSelect
                            value={projectSelectValue}
                            onValueChange={handleProjectChange}
                            placeholder="All Projects"
                            searchPlaceholder="Search projects..."
                            emptyText="No projects found."
                            items={projectItems}
                        />
                    </div>
                )}

                {/* App Dropdown (project / project-env scope only) */}
                {isProjectOrEnvScope && (
                    <div className="flex flex-col gap-1.5 min-w-0">
                        <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                            App
                        </span>
                        <SearchableFilterSelect
                            value={appSelectValue}
                            onValueChange={handleAppChange}
                            placeholder="All Apps"
                            searchPlaceholder="Search apps..."
                            emptyText="No apps found."
                            items={appItems}
                        />
                    </div>
                )}

                {/* Target Job Dropdown (app scope only) */}
                {isAppScope && (
                    <div className="flex flex-col gap-1.5 min-w-0">
                        <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                            Target Job
                        </span>
                        <SearchableFilterSelect
                            value={filters.targetId ?? "all"}
                            onValueChange={handleTargetJobChange}
                            placeholder="All Jobs"
                            searchPlaceholder="Search jobs..."
                            emptyText="No jobs found."
                            items={targetJobItems}
                        />
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
