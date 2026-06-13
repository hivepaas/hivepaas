import { type KeyboardEvent, type ReactNode, useState } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { dashedBorderBox } from "@lib/styles";
import { format } from "date-fns";
import { ChevronDown } from "lucide-react";
import ReactTimeAgo from "react-time-ago";
import type { AppScheduledJobTask } from "~/projects/domain";
import { EAppScheduledJobTaskStatus } from "~/projects/module-shared/enums";

import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import { Button, Checkbox, Skeleton } from "@/components/ui";

const STATUS_LABELS: Record<EAppScheduledJobTaskStatus, string> = {
    [EAppScheduledJobTaskStatus.Done]: "Done",
    [EAppScheduledJobTaskStatus.Failed]: "Failed",
    [EAppScheduledJobTaskStatus.InProgress]: "In-Progress",
    [EAppScheduledJobTaskStatus.NotStarted]: "Not Started",
    [EAppScheduledJobTaskStatus.Canceled]: "Canceled",
};

const STATUS_CLASS_NAMES: Record<EAppScheduledJobTaskStatus, string> = {
    [EAppScheduledJobTaskStatus.Done]: "bg-green-500 text-white hover:bg-green-500/90",
    [EAppScheduledJobTaskStatus.Failed]: "bg-red-500 text-white hover:bg-red-500/90",
    [EAppScheduledJobTaskStatus.InProgress]: "bg-purple-400 text-white hover:bg-purple-400/90",
    [EAppScheduledJobTaskStatus.NotStarted]: "bg-blue-400 text-white hover:bg-blue-400/90",
    [EAppScheduledJobTaskStatus.Canceled]: "bg-zinc-500 text-white hover:bg-zinc-500/90",
};

const STATUS_BORDER_CLASS_NAMES: Record<EAppScheduledJobTaskStatus, string> = {
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

function canCancelTask(task: AppScheduledJobTask): boolean {
    return (
        task.status === EAppScheduledJobTaskStatus.NotStarted ||
        (task.status === EAppScheduledJobTaskStatus.InProgress && !task.config.controlDisabled)
    );
}

function shouldShowDuration(task: AppScheduledJobTask): task is AppScheduledJobTask & { startedAt: Date } {
    return task.status !== EAppScheduledJobTaskStatus.NotStarted && task.startedAt != null;
}

function StatusBadge({ status }: { status: EAppScheduledJobTaskStatus }) {
    return <Badge className={cn("h-7 px-3", STATUS_CLASS_NAMES[status])}>{STATUS_LABELS[status]}</Badge>;
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
    onCancel,
    onClick,
}: ScheduledJobTaskSummaryCardProps) {
    const [isDetailsContentOpen, setIsDetailsContentOpen] = useState(variant === "details");
    const isClickable = Boolean(onClick);
    const shouldShowDetailsContent = variant === "list" || isDetailsContentOpen;
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
                "rounded-lg border bg-background p-5 shadow-xs",
                variant === "list" ? [STATUS_BORDER_CLASS_NAMES[task.status]] : "border-0",
                isClickable &&
                    "cursor-pointer transition-colors hover:bg-muted/30 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
            )}
            style={{
                borderLeftWidth: variant === "list" ? 4 : 0,
            }}
            role={isClickable ? "button" : undefined}
            tabIndex={isClickable ? 0 : undefined}
            onClick={onClick}
            onKeyDown={handleKeyDown}
        >
            <dl className="grid grid-cols-1 items-center gap-y-2 sm:grid-cols-[150px_minmax(0,1fr)] sm:gap-x-8 sm:gap-y-4">
                <InfoRow label="Status">
                    <div className="flex flex-wrap items-center gap-3">
                        <StatusBadge status={task.status} />
                        {onCancel && canCancelTask(task) && (
                            <Button
                                type="button"
                                variant="link"
                                className="h-auto p-0 text-primary"
                                isLoading={isCancelling}
                                onClick={event => {
                                    event.stopPropagation();
                                    onCancel(task.id);
                                }}
                            >
                                Cancel Task
                            </Button>
                        )}
                        {variant === "details" && (
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
                        )}
                    </div>
                </InfoRow>

                {shouldShowDetailsContent && (
                    <InfoRow label="Schedule At">
                        <span>{formatDateTime(task.runAt)}</span>
                    </InfoRow>
                )}

                {variant === "details" && shouldShowDetailsContent && (
                    <>
                        <InfoRow label="Started At">
                            <span>{formatDateTime(task.startedAt)}</span>
                        </InfoRow>
                        <InfoRow label="Ended At">
                            <span>{formatDateTime(task.endedAt)}</span>
                        </InfoRow>
                    </>
                )}

                {shouldShowDetailsContent && shouldShowDuration(task) && (
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

                {shouldShowDetailsContent && (
                    <>
                        <InfoRow label="Priority">
                            <span>{priority}</span>
                        </InfoRow>
                        <InfoRow label="Control Enabled">
                            <Checkbox
                                checked={controlEnabled}
                                disabled
                            />
                        </InfoRow>
                    </>
                )}

                {variant === "details" && shouldShowDetailsContent && (
                    <>
                        <InfoRow label="Timeout">
                            <span>{formatValue(task.config.timeout)}</span>
                        </InfoRow>
                        <InfoRow label="Max Retry">
                            <span>{task.config.maxRetry}</span>
                        </InfoRow>
                    </>
                )}

                {shouldShowDetailsContent && (
                    <InfoRow label="Retries">
                        <span>{task.config.retry}</span>
                    </InfoRow>
                )}

                {variant === "details" && shouldShowDetailsContent && (
                    <>
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

            {children && <div className="mt-5 min-w-0">{children}</div>}
        </div>
    );
}

export function ScheduledJobTaskSummaryCardSkeleton({ variant = "list" }: { variant?: ScheduledJobTaskCardVariant }) {
    return (
        <div
            className={cn(
                "rounded-lg border border-border bg-background p-5",
                variant === "list" && "border-l-4 border-l-purple-200",
            )}
        >
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
    onCancel?: (taskID: string) => void;
    onClick?: () => void;
}
