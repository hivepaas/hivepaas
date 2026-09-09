import { type KeyboardEvent, type ReactNode, useState } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { format } from "date-fns";
import { Calendar, ChevronDown, Clock, FolderTree, ShieldCheck } from "lucide-react";
import ReactTimeAgo from "react-time-ago";
import type { SystemTask } from "~/system-status/domain";
import { SystemTaskStatus } from "~/system-status/domain";

import { LogViewerActionButtons, PopConfirm } from "@application/shared/components";
import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import { Button, Skeleton } from "@/components/ui";
import { Avatar } from "@/components/ui/avatar";

const STATUS_LABELS: Record<SystemTaskStatus, string> = {
    [SystemTaskStatus.Done]: "Done",
    [SystemTaskStatus.Failed]: "Failed",
    [SystemTaskStatus.InProgress]: "In-Progress",
    [SystemTaskStatus.NotStarted]: "Not Started",
    [SystemTaskStatus.Canceled]: "Canceled",
};

const STATUS_CLASS_NAMES: Record<SystemTaskStatus, string> = {
    [SystemTaskStatus.Done]: "bg-green-500 text-white hover:bg-green-500/90",
    [SystemTaskStatus.Failed]: "bg-red-500 text-white hover:bg-red-500/90",
    [SystemTaskStatus.InProgress]: "bg-purple-400 text-white hover:bg-purple-400/90",
    [SystemTaskStatus.NotStarted]: "bg-blue-400 text-white hover:bg-blue-400/90",
    [SystemTaskStatus.Canceled]: "bg-amber-500 text-white hover:bg-amber-500/90",
};

const STATUS_BORDER_CLASS_NAMES: Record<SystemTaskStatus, string> = {
    [SystemTaskStatus.Done]: "border-l-green-500",
    [SystemTaskStatus.Failed]: "border-l-red-500",
    [SystemTaskStatus.InProgress]: "border-l-purple-400",
    [SystemTaskStatus.NotStarted]: "border-l-blue-400",
    [SystemTaskStatus.Canceled]: "border-l-amber-500",
};

function formatDuration(startedAt: Date, endedAt: Date): string {
    const durationMs = Math.max(0, endedAt.getTime() - startedAt.getTime());

    if (durationMs < 1_000) {
        return "less than a minute";
    }

    const anchor = Date.now();

    return timeAgoFormatter
        .format(anchor - durationMs, "round")
        .replace(/\s+ago$/, "")
        .replace(/^in\s+/, "");
}

function formatDateTime(date: Date | null): string {
    return date ? format(date, "yyyy-MM-dd HH:mm:ss") : "-";
}

function formatValue(value: string | number | null | undefined): string {
    if (value === null || value === undefined || value === "") {
        return "-";
    }

    return String(value);
}

function canCancelTask(task: SystemTask): boolean {
    return (
        !task.config.controlDisabled &&
        (task.status === SystemTaskStatus.NotStarted || task.status === SystemTaskStatus.InProgress)
    );
}

function shouldShowDuration(task: SystemTask): task is SystemTask & { startedAt: Date } {
    return task.status !== SystemTaskStatus.NotStarted && task.startedAt != null;
}

function StatusBadge({ status }: { status: SystemTaskStatus }) {
    return <Badge className={cn("h-7 px-3 text-sm", STATUS_CLASS_NAMES[status])}>{STATUS_LABELS[status]}</Badge>;
}

function JobBadge({ kind }: { kind: string }) {
    return <Badge className="h-7 bg-sky-400 px-3 text-sm text-white hover:bg-sky-400/90">{kind}</Badge>;
}

function TaskTypeBadge({ type }: { type: string }) {
    return (
        <Badge
            variant="outline"
            className="h-7 px-3 text-sm font-mono rounded-md border-border/70 bg-muted/50 text-foreground"
        >
            {type}
        </Badge>
    );
}

export function SystemTaskSummaryCard({
    task,
    now,
    children,
    variant = "list",
    isCancelling = false,
    isFullscreen = false,
    isFullView = false,
    isFullHeight = false,
    fontSize,
    themeId,
    onCancel,
    onClick,
    onToggleFullscreen,
    onToggleFullView,
    onToggleFullHeight,
    onCycleFontSize,
    onSelectTheme,
}: SystemTaskSummaryCardProps) {
    const [isDetailsContentOpen, setIsDetailsContentOpen] = useState(true);
    const isClickable = Boolean(onClick);
    const shouldShowDetailsContent = !isFullHeight && isDetailsContentOpen;
    const { priority } = task.config;
    const controlEnabled = !task.config.controlDisabled;
    const jobKind = task.targetJob?.kind.trim();

    const hasScope = Boolean(task.scopeProject ?? task.scopeApp ?? task.scopeUser);

    function handleKeyDown(event: KeyboardEvent<HTMLElement>) {
        if (!onClick || (event.key !== "Enter" && event.key !== " ")) {
            return;
        }

        event.preventDefault();
        onClick();
    }

    return (
        <div
            className={cn(
                "rounded-[8px] border bg-background shadow-xs",
                variant === "list"
                    ? ["p-3 sm:p-4", STATUS_BORDER_CLASS_NAMES[task.status]]
                    : "border-0 p-0 shadow-none bg-transparent",
                isClickable &&
                    "cursor-pointer transition-colors hover:bg-muted/30 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
                isFullscreen && "flex-1 min-h-0 flex flex-col",
            )}
            style={{
                borderLeftWidth: variant === "list" ? 4 : 0,
            }}
            role={isClickable ? "button" : undefined}
            tabIndex={isClickable ? 0 : undefined}
            onClick={onClick}
            onKeyDown={handleKeyDown}
        >
            {variant === "list" ? (
                <div className="flex flex-col gap-2.5">
                    {/* Header: Status, Task Type, Job Kind, Cancel Link, and Schedule At */}
                    <div className="flex flex-wrap items-center justify-between gap-3">
                        <div className="flex flex-wrap items-center gap-2 sm:gap-2.5">
                            <StatusBadge status={task.status} />
                            {task.type && <TaskTypeBadge type={task.type} />}
                            {jobKind && <JobBadge kind={jobKind} />}
                            {onCancel && canCancelTask(task) && (
                                <PopConfirm
                                    title="Cancel task"
                                    description="Are you sure you want to cancel this task?"
                                    confirmText="Cancel task"
                                    cancelText="Cancel"
                                    variant="destructive"
                                    onConfirm={() => {
                                        onCancel(task.id);
                                    }}
                                >
                                    <Button
                                        type="button"
                                        variant="link"
                                        className="h-auto p-0 text-sm text-destructive hover:underline"
                                        isLoading={isCancelling}
                                        onClick={event => {
                                            event.stopPropagation();
                                        }}
                                    >
                                        Cancel
                                    </Button>
                                </PopConfirm>
                            )}
                        </div>

                        <div className="flex items-center gap-3">
                            <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
                                <Calendar className="size-4 shrink-0 text-muted-foreground/70" />
                                <span>Scheduled:</span>
                                <span className="font-medium text-foreground">{formatDateTime(task.runAt)}</span>
                            </div>
                        </div>
                    </div>

                    {/* Scope Row (14px font size) */}
                    {hasScope && (
                        <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
                            <div className="flex items-center gap-1.5">
                                <FolderTree className="size-4 shrink-0 text-muted-foreground/70" />
                                <span>Scope:</span>
                            </div>

                            <div className="flex flex-wrap items-center gap-2">
                                {task.scopeProject && (
                                    <div className="flex items-center gap-1.5">
                                        <Avatar
                                            name={task.scopeProject.name}
                                            src={task.scopeProject.photo}
                                            className="size-4.5 rounded-xs text-[9px] border border-border shrink-0"
                                        />
                                        <span className="font-medium text-foreground">{task.scopeProject.name}</span>
                                    </div>
                                )}

                                {task.scopeProject && task.scopeApp && (
                                    <span className="text-muted-foreground/40 select-none">•</span>
                                )}

                                {task.scopeApp && (
                                    <div className="flex items-center gap-1.5">
                                        <Avatar
                                            name={task.scopeApp.name}
                                            src={task.scopeApp.photo}
                                            className="size-4.5 rounded-xs text-[9px] border border-border shrink-0"
                                        />
                                        <span className="font-medium text-foreground">{task.scopeApp.name}</span>
                                    </div>
                                )}

                                {!task.scopeProject && !task.scopeApp && task.scopeUser && (
                                    <div className="flex items-center gap-1.5">
                                        <Avatar
                                            name={task.scopeUser.fullName ?? task.scopeUser.username}
                                            src={task.scopeUser.photo}
                                            className="size-4.5 rounded-full text-[9px] border border-border shrink-0"
                                        />
                                        <span className="font-medium text-foreground">
                                            {task.scopeUser.fullName ?? task.scopeUser.username}
                                        </span>
                                    </div>
                                )}
                            </div>
                        </div>
                    )}

                    {/* Meta Row / Chips */}
                    <div className="flex flex-wrap items-center gap-x-3.5 gap-y-1.5 text-sm text-muted-foreground pt-2 border-t border-border/50">
                        {shouldShowDuration(task) && (
                            <>
                                <div className="flex items-center gap-1.5">
                                    <Clock className="size-4 shrink-0 text-muted-foreground/70" />
                                    <span>
                                        {formatDuration(task.startedAt, task.endedAt ?? now)} (
                                        <ReactTimeAgo
                                            date={task.startedAt}
                                            locale="en-US"
                                        />
                                        )
                                    </span>
                                </div>
                                <span className="text-muted-foreground/40 select-none">•</span>
                            </>
                        )}

                        <div className="flex items-center gap-1">
                            <span>Priority:</span>
                            <span className="font-medium text-foreground capitalize">{priority}</span>
                        </div>

                        <span className="text-muted-foreground/40 select-none">•</span>

                        <div className="flex items-center gap-1">
                            <span>Retries:</span>
                            <span className="font-medium text-foreground">
                                {task.config.retry} / {task.config.maxRetry}
                            </span>
                        </div>

                        {task.config.timeout && task.config.timeout !== "" && task.config.timeout !== "-" && (
                            <>
                                <span className="text-muted-foreground/40 select-none">•</span>
                                <div className="flex items-center gap-1">
                                    <span>Timeout:</span>
                                    <span className="font-medium text-foreground">{task.config.timeout}</span>
                                </div>
                            </>
                        )}

                        <span className="text-muted-foreground/40 select-none">•</span>

                        <div className="flex items-center gap-1.5">
                            <ShieldCheck
                                className={cn(
                                    "size-4 shrink-0",
                                    controlEnabled ? "text-emerald-500" : "text-muted-foreground/50",
                                )}
                            />
                            <span>Control:</span>
                            <span
                                className={cn(
                                    "font-medium",
                                    controlEnabled ? "text-emerald-600 dark:text-emerald-400" : "text-muted-foreground",
                                )}
                            >
                                {controlEnabled ? "Enabled" : "Disabled"}
                            </span>
                        </div>
                    </div>
                </div>
            ) : (
                <div className="flex flex-col gap-3">
                    {/* Header Row: Badges, Cancel, Chevron on the left; Toolbar on the right */}
                    <div className="flex flex-wrap items-center justify-between gap-3">
                        <div className="flex flex-wrap items-center gap-2 sm:gap-2.5">
                            <StatusBadge status={task.status} />
                            {task.type && <TaskTypeBadge type={task.type} />}
                            {jobKind && <JobBadge kind={jobKind} />}
                            {onCancel && canCancelTask(task) && (
                                <PopConfirm
                                    title="Cancel task"
                                    description="Are you sure you want to cancel this task?"
                                    confirmText="Cancel task"
                                    cancelText="Cancel"
                                    variant="destructive"
                                    onConfirm={() => {
                                        onCancel(task.id);
                                    }}
                                >
                                    <Button
                                        type="button"
                                        variant="link"
                                        className="h-auto p-0 text-sm text-destructive hover:underline"
                                        isLoading={isCancelling}
                                        onClick={event => {
                                            event.stopPropagation();
                                        }}
                                    >
                                        Cancel
                                    </Button>
                                </PopConfirm>
                            )}
                            <Button
                                type="button"
                                variant="ghost"
                                size="icon-sm"
                                className="size-7 text-muted-foreground hover:text-foreground"
                                aria-label={isDetailsContentOpen ? "Hide task details" : "Show task details"}
                                title={isDetailsContentOpen ? "Hide task details" : "Show task details"}
                                aria-expanded={isDetailsContentOpen}
                                onClick={event => {
                                    event.stopPropagation();
                                    setIsDetailsContentOpen(current => !current);
                                }}
                            >
                                <ChevronDown
                                    className={cn(
                                        "size-4 transition-transform duration-200",
                                        isDetailsContentOpen && "rotate-180",
                                    )}
                                />
                            </Button>
                        </div>

                        <div className="flex items-center gap-3">
                            {onToggleFullscreen && (
                                <LogViewerActionButtons
                                    isFullscreen={isFullscreen}
                                    isFullView={isFullView}
                                    isFullHeight={isFullHeight}
                                    fontSize={fontSize}
                                    themeId={themeId}
                                    onToggleFullscreen={onToggleFullscreen}
                                    onToggleFullView={onToggleFullView}
                                    onToggleFullHeight={onToggleFullHeight}
                                    onCycleFontSize={onCycleFontSize}
                                    onSelectTheme={onSelectTheme}
                                />
                            )}
                        </div>
                    </div>

                    {/* Compact Scope Row when details are collapsed */}
                    {!shouldShowDetailsContent && hasScope && (
                        <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
                            <div className="flex items-center gap-1.5">
                                <FolderTree className="size-4 shrink-0 text-muted-foreground/70" />
                                <span>Scope:</span>
                            </div>

                            <div className="flex flex-wrap items-center gap-2">
                                {task.scopeProject && (
                                    <div className="flex items-center gap-1.5">
                                        <Avatar
                                            name={task.scopeProject.name}
                                            src={task.scopeProject.photo}
                                            className="size-4.5 rounded-xs text-[9px] border border-border shrink-0"
                                        />
                                        <span className="font-medium text-foreground">{task.scopeProject.name}</span>
                                    </div>
                                )}

                                {task.scopeProject && task.scopeApp && (
                                    <span className="text-muted-foreground/40 select-none">•</span>
                                )}

                                {task.scopeApp && (
                                    <div className="flex items-center gap-1.5">
                                        <Avatar
                                            name={task.scopeApp.name}
                                            src={task.scopeApp.photo}
                                            className="size-4.5 rounded-xs text-[9px] border border-border shrink-0"
                                        />
                                        <span className="font-medium text-foreground">{task.scopeApp.name}</span>
                                    </div>
                                )}

                                {!task.scopeProject && !task.scopeApp && task.scopeUser && (
                                    <div className="flex items-center gap-1.5">
                                        <Avatar
                                            name={task.scopeUser.fullName ?? task.scopeUser.username}
                                            src={task.scopeUser.photo}
                                            className="size-4.5 rounded-full text-[9px] border border-border shrink-0"
                                        />
                                        <span className="font-medium text-foreground">
                                            {task.scopeUser.fullName ?? task.scopeUser.username}
                                        </span>
                                    </div>
                                )}
                            </div>
                        </div>
                    )}

                    {/* Expanded Details Section */}
                    {shouldShowDetailsContent && (
                        <div className="flex flex-col gap-3.5 pt-3 mt-1 border-t border-border/60">
                            {/* Scope Hierarchy Banner */}
                            {hasScope && (
                                <div className="flex flex-col gap-1.5">
                                    <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                                        Scope Hierarchy
                                    </span>
                                    <div className="flex flex-wrap items-center gap-3 rounded-md border border-border bg-muted/30 px-3.5 py-2 text-sm">
                                        {task.scopeProject && (
                                            <div className="flex items-center gap-1.5">
                                                <span className="text-muted-foreground">Project:</span>
                                                <Avatar
                                                    name={task.scopeProject.name}
                                                    src={task.scopeProject.photo}
                                                    className="size-4.5 rounded-xs text-[9px] border border-border shrink-0"
                                                />
                                                <span className="font-medium text-foreground">
                                                    {task.scopeProject.name}
                                                </span>
                                                {task.scopeProject.key && (
                                                    <span className="font-mono text-muted-foreground">
                                                        ({task.scopeProject.key})
                                                    </span>
                                                )}
                                            </div>
                                        )}

                                        {task.scopeProject && task.scopeApp && (
                                            <span className="text-muted-foreground/40 select-none">•</span>
                                        )}

                                        {task.scopeApp && (
                                            <div className="flex items-center gap-1.5">
                                                <span className="text-muted-foreground">App:</span>
                                                <Avatar
                                                    name={task.scopeApp.name}
                                                    src={task.scopeApp.photo}
                                                    className="size-4.5 rounded-xs text-[9px] border border-border shrink-0"
                                                />
                                                <span className="font-medium text-foreground">
                                                    {task.scopeApp.name}
                                                </span>
                                                {task.scopeApp.key && (
                                                    <span className="font-mono text-muted-foreground">
                                                        ({task.scopeApp.key})
                                                    </span>
                                                )}
                                            </div>
                                        )}

                                        {task.scopeUser && (
                                            <>
                                                {Boolean(task.scopeProject ?? task.scopeApp) && (
                                                    <span className="text-muted-foreground/40 select-none">•</span>
                                                )}
                                                <div className="flex items-center gap-1.5">
                                                    <span className="text-muted-foreground">User:</span>
                                                    <Avatar
                                                        name={task.scopeUser.fullName ?? task.scopeUser.username}
                                                        src={task.scopeUser.photo}
                                                        className="size-4.5 rounded-full text-[9px] border border-border shrink-0"
                                                    />
                                                    <span className="font-medium text-foreground">
                                                        {task.scopeUser.fullName ?? task.scopeUser.username}
                                                    </span>
                                                    {task.scopeUser.email && (
                                                        <span className="text-muted-foreground">
                                                            ({task.scopeUser.email})
                                                        </span>
                                                    )}
                                                </div>
                                            </>
                                        )}
                                    </div>
                                </div>
                            )}

                            {/* Detailed Meta Grid */}
                            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-x-6 gap-y-2.5 text-sm">
                                <div>
                                    <span className="text-muted-foreground">Task ID:</span>{" "}
                                    <span className="font-mono text-foreground font-medium select-all">{task.id}</span>
                                </div>

                                <div>
                                    <span className="text-muted-foreground">Schedule At:</span>{" "}
                                    <span className="text-foreground font-medium">{formatDateTime(task.runAt)}</span>
                                </div>

                                <div>
                                    <span className="text-muted-foreground">Started At:</span>{" "}
                                    <span className="text-foreground font-medium">
                                        {formatDateTime(task.startedAt)}
                                    </span>
                                </div>

                                <div>
                                    <span className="text-muted-foreground">Ended At:</span>{" "}
                                    <span className="text-foreground font-medium">{formatDateTime(task.endedAt)}</span>
                                </div>

                                {shouldShowDuration(task) && (
                                    <div>
                                        <span className="text-muted-foreground">Duration:</span>{" "}
                                        <span className="text-foreground font-medium">
                                            {formatDuration(task.startedAt, task.endedAt ?? now)}
                                        </span>{" "}
                                        <span className="text-xs text-muted-foreground">
                                            (
                                            <ReactTimeAgo
                                                date={task.startedAt}
                                                locale="en-US"
                                            />
                                            )
                                        </span>
                                    </div>
                                )}

                                <div>
                                    <span className="text-muted-foreground">Priority:</span>{" "}
                                    <span className="text-foreground font-medium capitalize">{priority}</span>
                                </div>

                                <div className="flex items-center gap-1.5">
                                    <span className="text-muted-foreground">Control:</span>
                                    <div className="flex items-center gap-1">
                                        <ShieldCheck
                                            className={cn(
                                                "size-4 shrink-0",
                                                controlEnabled ? "text-emerald-500" : "text-muted-foreground/50",
                                            )}
                                        />
                                        <span
                                            className={cn(
                                                "font-medium",
                                                controlEnabled
                                                    ? "text-emerald-600 dark:text-emerald-400"
                                                    : "text-muted-foreground",
                                            )}
                                        >
                                            {controlEnabled ? "Enabled" : "Disabled"}
                                        </span>
                                    </div>
                                </div>

                                <div>
                                    <span className="text-muted-foreground">Timeout:</span>{" "}
                                    <span className="text-foreground font-medium">
                                        {formatValue(task.config.timeout)}
                                    </span>
                                </div>

                                <div>
                                    <span className="text-muted-foreground">Max Retry:</span>{" "}
                                    <span className="text-foreground font-medium">{task.config.maxRetry}</span>
                                </div>

                                <div>
                                    <span className="text-muted-foreground">Retries:</span>{" "}
                                    <span className="text-foreground font-medium">{task.config.retry}</span>
                                </div>

                                <div>
                                    <span className="text-muted-foreground">Retry Delay:</span>{" "}
                                    <span className="text-foreground font-medium">
                                        {formatValue(task.config.retryDelay)}
                                    </span>
                                </div>

                                {task.retryAt && (
                                    <div>
                                        <span className="text-muted-foreground">Retry At:</span>{" "}
                                        <span className="text-foreground font-medium">
                                            {formatDateTime(task.retryAt)}
                                        </span>
                                    </div>
                                )}

                                {task.targetJob?.id && (
                                    <div>
                                        <span className="text-muted-foreground">Job ID:</span>{" "}
                                        <span className="font-mono text-foreground select-all">
                                            {task.targetJob.id}
                                        </span>
                                    </div>
                                )}

                                {task.scopeProject?.id && (
                                    <div>
                                        <span className="text-muted-foreground">Scope Project ID:</span>{" "}
                                        <span className="font-mono text-foreground select-all">
                                            {task.scopeProject.id}
                                        </span>
                                    </div>
                                )}

                                {task.scopeApp?.id && (
                                    <div>
                                        <span className="text-muted-foreground">Scope App ID:</span>{" "}
                                        <span className="font-mono text-foreground select-all">{task.scopeApp.id}</span>
                                    </div>
                                )}

                                {task.scopeUser?.id && (
                                    <div>
                                        <span className="text-muted-foreground">Scope User ID:</span>{" "}
                                        <span className="font-mono text-foreground select-all">
                                            {task.scopeUser.id}
                                        </span>
                                    </div>
                                )}
                            </div>

                            {/* Error Banner */}
                            {task.lastError.trim() && (
                                <div className="flex flex-col gap-1.5 pt-1">
                                    <span className="text-[11px] font-semibold text-destructive uppercase tracking-wider">
                                        Error
                                    </span>
                                    <div className="rounded-md border border-destructive/30 bg-destructive/10 p-3 max-h-60 overflow-y-auto">
                                        <pre className="font-mono text-xs text-destructive whitespace-pre-wrap break-all select-all">
                                            {task.lastError}
                                        </pre>
                                    </div>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            )}

            {children && (
                <div className={cn("mt-5 min-w-0", isFullscreen && "flex-1 min-h-0 flex flex-col")}>{children}</div>
            )}
        </div>
    );
}

export function SystemTaskSummaryCardSkeleton({ variant = "list" }: { variant?: SystemTaskCardVariant }) {
    if (variant === "list") {
        return (
            <div className="rounded-[8px] border border-border bg-background p-3 sm:p-4 border-l-4 border-l-muted">
                <div className="flex flex-col gap-2.5">
                    <div className="flex items-center justify-between gap-3">
                        <div className="flex items-center gap-2.5">
                            <Skeleton className="h-7 w-20 rounded-md" />
                            <Skeleton className="h-7 w-28 rounded-md" />
                        </div>
                        <Skeleton className="h-5 w-44 rounded-md" />
                    </div>
                    <div className="flex items-center gap-3 pt-2 border-t border-border/40">
                        <Skeleton className="h-4 w-36" />
                        <Skeleton className="h-4 w-20" />
                        <Skeleton className="h-4 w-20" />
                        <Skeleton className="h-4 w-24" />
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="flex flex-col gap-3.5 p-1">
            <div className="flex items-center justify-between gap-3">
                <div className="flex items-center gap-2.5">
                    <Skeleton className="h-7 w-20 rounded-md" />
                    <Skeleton className="h-7 w-36 rounded-md" />
                    <Skeleton className="size-7 rounded-md" />
                </div>
                <div className="flex items-center gap-2">
                    <Skeleton className="size-8 rounded-md" />
                    <Skeleton className="size-8 rounded-md" />
                    <Skeleton className="size-8 rounded-md" />
                    <Skeleton className="size-8 rounded-md" />
                </div>
            </div>
            <div className="h-10 rounded-md bg-muted/40 border border-border" />
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4 pt-2">
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-4 w-40" />
                <Skeleton className="h-4 w-36" />
                <Skeleton className="h-4 w-36" />
                <Skeleton className="h-4 w-28" />
                <Skeleton className="h-4 w-24" />
                <Skeleton className="h-4 w-28" />
                <Skeleton className="h-4 w-24" />
            </div>
        </div>
    );
}

type SystemTaskCardVariant = "list" | "details";

interface SystemTaskSummaryCardProps {
    task: SystemTask;
    now: Date;
    children?: ReactNode;
    variant?: SystemTaskCardVariant;
    isCancelling?: boolean;
    isFullscreen?: boolean;
    isFullView?: boolean;
    isFullHeight?: boolean;
    fontSize?: number;
    themeId?: string;
    onCancel?: (taskID: string) => void;
    onClick?: () => void;
    onToggleFullscreen?: () => void;
    onToggleFullView?: () => void;
    onToggleFullHeight?: () => void;
    onCycleFontSize?: () => void;
    onSelectTheme?: (themeId: string) => void;
}
