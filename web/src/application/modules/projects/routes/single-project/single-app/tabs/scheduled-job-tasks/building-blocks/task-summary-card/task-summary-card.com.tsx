import { type KeyboardEvent, type ReactNode, useState } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { dashedBorderBox } from "@lib/styles";
import { format } from "date-fns";
import { Calendar, ChevronDown, Clock, ShieldCheck } from "lucide-react";
import ReactTimeAgo from "react-time-ago";
import type { AppScheduledJobTask } from "~/projects/domain";
import { EAppScheduledJobTaskStatus } from "~/projects/module-shared/enums";

import { LogViewerActionButtons, PopConfirm } from "@application/shared/components";
import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import type { OpenApiConstant } from "@infrastructure/api";

import { Button, Checkbox, Skeleton } from "@/components/ui";

const STATUS_LABELS: Partial<Record<EAppScheduledJobTaskStatus, string>> = {
    [EAppScheduledJobTaskStatus.Done]: "Done",
    [EAppScheduledJobTaskStatus.Failed]: "Failed",
    [EAppScheduledJobTaskStatus.InProgress]: "In-Progress",
    [EAppScheduledJobTaskStatus.NotStarted]: "Not Started",
    [EAppScheduledJobTaskStatus.Canceled]: "Canceled",
};

const STATUS_CLASS_NAMES: Partial<Record<EAppScheduledJobTaskStatus, string>> = {
    [EAppScheduledJobTaskStatus.Done]: "bg-green-500 text-white hover:bg-green-500/90",
    [EAppScheduledJobTaskStatus.Failed]: "bg-red-500 text-white hover:bg-red-500/90",
    [EAppScheduledJobTaskStatus.InProgress]: "bg-purple-400 text-white hover:bg-purple-400/90",
    [EAppScheduledJobTaskStatus.NotStarted]: "bg-blue-400 text-white hover:bg-blue-400/90",
    [EAppScheduledJobTaskStatus.Canceled]: "bg-zinc-500 text-white hover:bg-zinc-500/90",
};

const STATUS_BORDER_CLASS_NAMES: Partial<Record<EAppScheduledJobTaskStatus, string>> = {
    [EAppScheduledJobTaskStatus.Done]: "border-l-green-500",
    [EAppScheduledJobTaskStatus.Failed]: "border-l-red-500",
    [EAppScheduledJobTaskStatus.InProgress]: "border-l-purple-400",
    [EAppScheduledJobTaskStatus.NotStarted]: "border-l-blue-400",
    [EAppScheduledJobTaskStatus.Canceled]: "border-l-zinc-500",
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

function formatStatusLabel(status: string) {
    if (!status.trim()) {
        return "-";
    }

    return status
        .replace(/[_-]+/g, " ")
        .replace(/\s+/g, " ")
        .trim()
        .replace(/\b\w/g, char => char.toUpperCase());
}

function canCancelTask(task: AppScheduledJobTask): boolean {
    return (
        !task.config.controlDisabled &&
        (task.status === EAppScheduledJobTaskStatus.NotStarted || task.status === EAppScheduledJobTaskStatus.InProgress)
    );
}

function shouldShowDuration(task: AppScheduledJobTask): task is AppScheduledJobTask & { startedAt: Date } {
    return task.status !== EAppScheduledJobTaskStatus.NotStarted && task.startedAt != null;
}

function StatusBadge({ status }: { status: OpenApiConstant<EAppScheduledJobTaskStatus> }) {
    return (
        <Badge className={cn("h-7 px-3 text-sm", STATUS_CLASS_NAMES[status as EAppScheduledJobTaskStatus])}>
            {STATUS_LABELS[status as EAppScheduledJobTaskStatus] ?? formatStatusLabel(status)}
        </Badge>
    );
}

function InfoRow({ label, children }: { label: string; children: ReactNode }) {
    return (
        <>
            <dt className="text-sm font-semibold text-foreground">{label}</dt>
            <dd className="min-w-0 text-sm text-foreground">{children}</dd>
        </>
    );
}

export function ScheduledJobTaskSummaryCard({
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
}: ScheduledJobTaskSummaryCardProps) {
    const [isDetailsContentOpen, setIsDetailsContentOpen] = useState(false);
    const isClickable = Boolean(onClick);
    const shouldShowDetailsContent = !isFullHeight && isDetailsContentOpen;
    const { priority } = task.config;
    const controlEnabled = !task.config.controlDisabled;

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
                    ? ["p-3 sm:p-4", STATUS_BORDER_CLASS_NAMES[task.status as EAppScheduledJobTaskStatus]]
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
                    {/* Header: Status, Cancel Link, and Schedule At */}
                    <div className="flex flex-wrap items-center justify-between gap-3">
                        <div className="flex flex-wrap items-center gap-2 sm:gap-2.5">
                            <StatusBadge status={task.status} />
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
                <dl className="grid grid-cols-1 items-center gap-y-2 sm:grid-cols-[150px_minmax(0,1fr)] sm:gap-x-8 sm:gap-y-4">
                    <InfoRow label="Status">
                        <div className="flex items-center justify-between gap-3 w-full">
                            <div className="flex flex-wrap items-center gap-3">
                                <StatusBadge status={task.status} />
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
                                    className="size-6 text-primary hover:text-primary"
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
                    </InfoRow>

                    {shouldShowDetailsContent && (
                        <>
                            <InfoRow label="Schedule At">
                                <span>{formatDateTime(task.runAt)}</span>
                            </InfoRow>
                            <InfoRow label="Started At">
                                <span>{formatDateTime(task.startedAt)}</span>
                            </InfoRow>
                            <InfoRow label="Ended At">
                                <span>{formatDateTime(task.endedAt)}</span>
                            </InfoRow>
                            {shouldShowDuration(task) && (
                                <InfoRow label="Duration">
                                    <span>
                                        {formatDuration(task.startedAt, task.endedAt ?? now)} from{" "}
                                        <ReactTimeAgo
                                            date={task.startedAt}
                                            locale="en-US"
                                        />
                                    </span>
                                </InfoRow>
                            )}
                            <InfoRow label="Priority">
                                <span>{priority}</span>
                            </InfoRow>
                            <InfoRow label="Control Enabled">
                                <Checkbox
                                    checked={controlEnabled}
                                    disabled
                                />
                            </InfoRow>
                            <InfoRow label="Timeout">
                                <span>{formatValue(task.config.timeout)}</span>
                            </InfoRow>
                            <InfoRow label="Max Retry">
                                <span>{task.config.maxRetry}</span>
                            </InfoRow>
                            <InfoRow label="Retries">
                                <span>{task.config.retry}</span>
                            </InfoRow>
                            <InfoRow label="Retry Delay">
                                <span>{formatValue(task.config.retryDelay)}</span>
                            </InfoRow>
                            {task.lastError.trim() ? (
                                <InfoRow label="Error">
                                    <div className={cn(dashedBorderBox, "break-words whitespace-pre-wrap")}>
                                        {task.lastError}
                                    </div>
                                </InfoRow>
                            ) : null}
                        </>
                    )}
                </dl>
            )}

            {children && (
                <div className={cn("mt-5 min-w-0", isFullscreen && "flex-1 min-h-0 flex flex-col")}>{children}</div>
            )}
        </div>
    );
}

export function ScheduledJobTaskSummaryCardSkeleton({ variant = "list" }: { variant?: ScheduledJobTaskCardVariant }) {
    if (variant === "list") {
        return (
            <div className="rounded-[8px] border border-border bg-background p-3 sm:p-4 border-l-4 border-l-muted">
                <div className="flex flex-col gap-2.5">
                    <div className="flex items-center justify-between gap-3">
                        <Skeleton className="h-7 w-20 rounded-md" />
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
        <div className="rounded-[8px] border border-border bg-background p-5">
            <div className="grid grid-cols-1 items-center gap-y-2 sm:grid-cols-[150px_minmax(0,1fr)] sm:gap-x-8 sm:gap-y-4">
                <Skeleton className="h-5 w-16" />
                <Skeleton className="h-7 w-56" />
                <Skeleton className="h-5 w-20" />
                <Skeleton className="h-5 w-72" />
                <Skeleton className="h-5 w-16" />
                <Skeleton className="h-5 w-32" />
            </div>
        </div>
    );
}

type ScheduledJobTaskCardVariant = "list" | "details";

interface ScheduledJobTaskSummaryCardProps {
    task: AppScheduledJobTask;
    now: Date;
    children?: ReactNode;
    variant?: ScheduledJobTaskCardVariant;
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
