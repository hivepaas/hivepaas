import { useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { format } from "date-fns";
import { Check, ChevronDown } from "lucide-react";
import { AuditLogsQueries } from "~/operations/data";
import { AuditLogResult, type AuditLogScope, AuditLogSource, AuditLogType } from "~/operations/domain";

import { AppsPublicQueries, ProjectsPublicQueries, UsersPublicQueries } from "@application/shared/data-public/queries";

import { Avatar } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command";
import { DateTimePicker } from "@/components/ui/date-time-picker";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import { ResultBadge } from "../audit-log-summary-card/audit-log-summary-card.com";

interface SearchableFilterItem {
    value: string;
    label: string;
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
                                    value={`${item.label} ${item.value}`}
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

export interface AuditLogFilterValues {
    type?: string;
    source?: string;
    result?: string;
    actorId?: string;
    fromDate?: string;
    toDate?: string;
    projectId?: string;
    appId?: string;
    scopeOnly?: boolean;
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

    // 3. Fetch projects for project filter (global scope only)
    const { data: projectsResponse } = ProjectsPublicQueries.useFindManyPaginated({}, { enabled: isGlobalScope });

    // 4. Fetch apps for app filter (project / project-env scope only)
    const { data: appsResponse } = AppsPublicQueries.useFindManyBase(
        { projectID },
        { enabled: isProjectOrEnvScope && Boolean(projectID) },
    );

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

    const actorItems: SearchableFilterItem[] = useMemo(() => {
        const users = usersResponse?.data ?? [];
        return [
            { value: "all", label: "All Actors" },
            ...users.map(user => {
                const userName = user.fullName ? user.fullName : user.username;
                return {
                    value: user.id,
                    label: userName,
                    searchKey: `${userName} ${user.username} ${user.email}`,
                    avatar: {
                        name: userName,
                        src: user.photo,
                    },
                };
            }),
        ];
    }, [usersResponse?.data]);

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

                {/* Project / Env Scope: App Dropdown */}
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
                                <span>All Sources</span>
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
                                <span>All Results</span>
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
                    <SearchableFilterSelect
                        value={filters.actorId ?? "all"}
                        onValueChange={handleActorChange}
                        placeholder="All Actors"
                        searchPlaceholder="Search actors..."
                        emptyText="No actors found."
                        items={actorItems}
                    />
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
