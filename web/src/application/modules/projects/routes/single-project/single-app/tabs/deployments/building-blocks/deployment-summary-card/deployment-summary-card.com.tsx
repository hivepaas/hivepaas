import { type KeyboardEvent, type MouseEvent, type ReactNode, useState } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { dashedBorderBox } from "@lib/styles";
import { format } from "date-fns";
import { Box, ChevronDown, Clock, FileText, FolderGit2, GitBranch, GitCommit, Info } from "lucide-react";
import ReactTimeAgo from "react-time-ago";
import type { AppDeployment, AppDeploymentSourceUser } from "~/projects/domain";
import {
    EAppDeploymentMethod,
    EAppDeploymentStatus,
    EAppDeploymentTriggerSource,
} from "~/projects/module-shared/enums";

import { LogViewerActionButtons, PopConfirm } from "@application/shared/components";
import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import type { OpenApiConstant } from "@infrastructure/api";

import { Avatar, Button, Skeleton } from "@/components/ui";

const STATUS_LABELS: Partial<Record<EAppDeploymentStatus, string>> = {
    [EAppDeploymentStatus.Done]: "Done",
    [EAppDeploymentStatus.Failed]: "Failed",
    [EAppDeploymentStatus.InProgress]: "In-Progress",
    [EAppDeploymentStatus.NotStarted]: "Not Started",
    [EAppDeploymentStatus.Canceled]: "Canceled",
};

const STATUS_CLASS_NAMES: Partial<Record<EAppDeploymentStatus, string>> = {
    [EAppDeploymentStatus.Done]: "bg-green-500 text-white hover:bg-green-500/90",
    [EAppDeploymentStatus.Failed]: "bg-red-500 text-white hover:bg-red-500/90",
    [EAppDeploymentStatus.InProgress]: "bg-purple-400 text-white hover:bg-purple-400/90",
    [EAppDeploymentStatus.NotStarted]: "bg-blue-400 text-white hover:bg-blue-400/90",
    [EAppDeploymentStatus.Canceled]: "bg-zinc-500 text-white hover:bg-zinc-500/90",
};

const STATUS_BORDER_CLASS_NAMES: Partial<Record<EAppDeploymentStatus, string>> = {
    [EAppDeploymentStatus.Done]: "border-l-green-500",
    [EAppDeploymentStatus.Failed]: "border-l-red-500",
    [EAppDeploymentStatus.InProgress]: "border-l-purple-400",
    [EAppDeploymentStatus.NotStarted]: "border-l-blue-400",
    [EAppDeploymentStatus.Canceled]: "border-l-zinc-500",
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

function getUserDisplayName(user: AppDeploymentSourceUser): string {
    return user.fullName || user.email || user.username;
}

function canCancelDeployment(status: OpenApiConstant<EAppDeploymentStatus>): boolean {
    return status === EAppDeploymentStatus.InProgress || status === EAppDeploymentStatus.NotStarted;
}

function shouldShowDuration(deployment: AppDeployment): deployment is AppDeployment & { startedAt: Date } {
    return deployment.status !== EAppDeploymentStatus.NotStarted && deployment.startedAt != null;
}

type RepoDeployment = AppDeployment & {
    settings: Extract<AppDeployment["settings"], { activeMethod: typeof EAppDeploymentMethod.Repo }>;
};

function isRepoDeployment(deployment: AppDeployment): deployment is RepoDeployment {
    return deployment.settings.activeMethod === EAppDeploymentMethod.Repo;
}

function stopCardClick(event: MouseEvent<HTMLElement>) {
    event.stopPropagation();
}

function formatRepoRef(ref: string): string {
    return ref.replace(/^refs\/heads\//, "");
}

function getRepoShortName(url: string): string {
    return (
        url
            .replace(/^https?:\/\/[^/]+\//, "")
            .replace(/^git@[^:]+:/, "")
            .replace(/\.git$/, "") || url
    );
}

function StatusBadge({ status }: { status: OpenApiConstant<EAppDeploymentStatus> }) {
    return (
        <Badge className={cn("h-7 px-3 text-sm", STATUS_CLASS_NAMES[status as EAppDeploymentStatus])}>
            {STATUS_LABELS[status as EAppDeploymentStatus] ?? formatStatusLabel(status)}
        </Badge>
    );
}

export function DeploymentSummaryCard({
    deployment,
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
}: DeploymentSummaryCardProps) {
    const { output } = deployment;
    const sourceUser = deployment.trigger?.sourceUser;
    const [isCommitMessageOpen, setIsCommitMessageOpen] = useState(false);
    const [isDetailsContentOpen, setIsDetailsContentOpen] = useState(true);
    const isRepo = isRepoDeployment(deployment);
    const isClickable = Boolean(onClick);
    const shouldShowDetailsContent = !isFullHeight && isDetailsContentOpen;
    const hasCommitMessage = Boolean(output?.commitMessage && output.commitMessage.trim().length > 0);

    function handleClick(event: MouseEvent<HTMLElement>) {
        if (!onClick) {
            return;
        }

        const target = event.target as HTMLElement | null;
        if (target?.closest("a, button, [data-prevent-card-click]")) {
            return;
        }

        onClick();
    }

    function handleKeyDown(event: KeyboardEvent<HTMLElement>) {
        if (!onClick || (event.key !== "Enter" && event.key !== " ")) {
            return;
        }

        const target = event.target as HTMLElement | null;
        if (target?.closest("a, button, [data-prevent-card-click]")) {
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
                    ? ["p-3 sm:p-4", STATUS_BORDER_CLASS_NAMES[deployment.status as EAppDeploymentStatus]]
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
            onClick={handleClick}
            onKeyDown={handleKeyDown}
        >
            {variant === "list" ? (
                <div className="flex flex-col gap-2.5">
                    {/* Row 1: Header - Status, Cancel Link, Commit Title / Image / Fallback Title, Duration */}
                    <div className="flex flex-wrap items-center justify-between gap-3">
                        <div className="flex min-w-0 flex-1 items-center gap-2 sm:gap-2.5">
                            <StatusBadge status={deployment.status} />
                            {onCancel && canCancelDeployment(deployment.status) && (
                                <PopConfirm
                                    title="Cancel deployment"
                                    description="Are you sure you want to cancel this deployment?"
                                    confirmText="Cancel deployment"
                                    cancelText="Cancel"
                                    variant="destructive"
                                    onConfirm={() => {
                                        onCancel(deployment.id);
                                    }}
                                >
                                    <Button
                                        type="button"
                                        variant="link"
                                        className="h-auto p-0 text-sm text-destructive hover:underline shrink-0"
                                        isLoading={isCancelling}
                                        onClick={event => {
                                            event.stopPropagation();
                                        }}
                                    >
                                        Cancel
                                    </Button>
                                </PopConfirm>
                            )}

                            {/* Main Headline / Title */}
                            {output?.commitTitle ? (
                                <div className="flex min-w-0 flex-1 items-center gap-1.5">
                                    {output.commitURL ? (
                                        <a
                                            href={output.commitURL}
                                            target="_blank"
                                            rel="noreferrer"
                                            className="truncate text-sm font-semibold text-foreground hover:text-link hover:underline"
                                            title={output.commitTitle}
                                            onClick={stopCardClick}
                                        >
                                            {output.commitTitle}
                                        </a>
                                    ) : (
                                        <span
                                            className="truncate text-sm font-semibold text-foreground"
                                            title={output.commitTitle}
                                        >
                                            {output.commitTitle}
                                        </span>
                                    )}

                                    {hasCommitMessage && (
                                        <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon-sm"
                                            className={cn(
                                                "size-6 shrink-0 rounded transition-colors",
                                                isCommitMessageOpen
                                                    ? "bg-primary/10 text-primary"
                                                    : "text-muted-foreground hover:text-primary",
                                            )}
                                            aria-label={
                                                isCommitMessageOpen ? "Hide commit message" : "Show commit message"
                                            }
                                            title={isCommitMessageOpen ? "Hide commit message" : "Show commit message"}
                                            aria-expanded={isCommitMessageOpen}
                                            onClick={event => {
                                                event.stopPropagation();
                                                setIsCommitMessageOpen(current => !current);
                                            }}
                                        >
                                            <Info className="size-4" />
                                        </Button>
                                    )}
                                </div>
                            ) : isRepo ? (
                                <span className="truncate text-sm font-semibold text-foreground">
                                    {formatRepoRef(deployment.settings.repoSource.repoRef)}
                                </span>
                            ) : output?.imageTags && output.imageTags.length > 0 ? (
                                <span className="truncate font-mono text-sm font-semibold text-foreground">
                                    {output.imageTags.join(", ")}
                                </span>
                            ) : (
                                <span className="text-sm font-semibold text-foreground">
                                    Deployment #{deployment.id.slice(0, 8)}
                                </span>
                            )}
                        </div>

                        {/* Right side: Duration */}
                        {shouldShowDuration(deployment) && (
                            <div className="flex items-center gap-1.5 text-sm text-muted-foreground shrink-0">
                                <Clock className="size-4 shrink-0 text-muted-foreground/70" />
                                <span>
                                    {formatDuration(deployment.startedAt, deployment.endedAt ?? now)} (
                                    <ReactTimeAgo
                                        date={deployment.startedAt}
                                        locale="en-US"
                                    />
                                    )
                                </span>
                            </div>
                        )}
                    </div>

                    {/* Row 2: Meta Row / Chips */}
                    <div className="flex flex-wrap items-center gap-x-3.5 gap-y-1.5 text-sm text-muted-foreground pt-2 border-t border-border/50">
                        {isRepo && (
                            <div className="flex items-center gap-1 font-medium text-foreground">
                                <GitBranch className="size-4 shrink-0 text-muted-foreground/70" />
                                <span title={deployment.settings.repoSource.repoRef}>
                                    {formatRepoRef(deployment.settings.repoSource.repoRef)}
                                </span>
                            </div>
                        )}

                        {output?.commitHashShort && (
                            <>
                                {isRepo && <span className="text-muted-foreground/40 select-none">•</span>}
                                <div className="flex items-center gap-1 font-mono">
                                    <GitCommit className="size-4 shrink-0 text-muted-foreground/70" />
                                    {output.commitURL ? (
                                        <a
                                            href={output.commitURL}
                                            target="_blank"
                                            rel="noreferrer"
                                            className="font-medium text-foreground underline underline-offset-4 hover:text-link"
                                            onClick={stopCardClick}
                                        >
                                            {output.commitHashShort}
                                        </a>
                                    ) : (
                                        <span className="font-medium text-foreground">{output.commitHashShort}</span>
                                    )}
                                </div>
                            </>
                        )}

                        {isRepo && deployment.settings.repoSource.repoUrl && (
                            <>
                                <span className="text-muted-foreground/40 select-none">•</span>
                                <div className="flex items-center gap-1 max-w-[260px] sm:max-w-[340px] truncate">
                                    <FolderGit2 className="size-4 shrink-0 text-muted-foreground/70" />
                                    <a
                                        href={deployment.settings.repoSource.repoUrl}
                                        target="_blank"
                                        rel="noreferrer"
                                        className="truncate underline underline-offset-4 hover:text-link"
                                        title={deployment.settings.repoSource.repoUrl}
                                        onClick={stopCardClick}
                                    >
                                        {getRepoShortName(deployment.settings.repoSource.repoUrl)}
                                    </a>
                                </div>
                            </>
                        )}

                        {deployment.trigger?.source === EAppDeploymentTriggerSource.RepoWebhook && (
                            <>
                                <span className="text-muted-foreground/40 select-none">•</span>
                                <Badge className="h-5 px-1.5 text-xs bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 hover:bg-emerald-500/20 border-emerald-500/30">
                                    Webhook
                                </Badge>
                            </>
                        )}

                        {sourceUser && (
                            <>
                                <span className="text-muted-foreground/40 select-none">•</span>
                                <div className="flex items-center gap-1.5">
                                    <Avatar
                                        name={getUserDisplayName(sourceUser)}
                                        src={sourceUser.photo}
                                        className="size-5 text-[10px]"
                                    />
                                    <span className="truncate max-w-[150px]">{getUserDisplayName(sourceUser)}</span>
                                </div>
                            </>
                        )}

                        {!sourceUser && output?.commitAuthor && (
                            <>
                                <span className="text-muted-foreground/40 select-none">•</span>
                                <div className="flex items-center gap-1">
                                    <span className="text-muted-foreground">Author:</span>
                                    <span className="font-medium text-foreground">{output.commitAuthor}</span>
                                </div>
                            </>
                        )}

                        {output?.imageTags && output.imageTags.length > 0 && (
                            <>
                                <span className="text-muted-foreground/40 select-none">•</span>
                                <div
                                    className="flex items-center gap-1 font-mono text-xs max-w-[220px] truncate"
                                    title={output.imageTags.join(", ")}
                                >
                                    <Box className="size-4 shrink-0 text-muted-foreground/70" />
                                    <span className="truncate">{output.imageTags.join(", ")}</span>
                                </div>
                            </>
                        )}

                        {hasCommitMessage && (
                            <>
                                <span className="text-muted-foreground/40 select-none">•</span>
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="sm"
                                    className={cn(
                                        "h-5 px-1.5 text-xs gap-1 transition-colors rounded text-muted-foreground hover:text-primary",
                                        isCommitMessageOpen && "bg-primary/10 text-primary font-medium",
                                    )}
                                    aria-label={isCommitMessageOpen ? "Hide commit message" : "Show commit message"}
                                    title={isCommitMessageOpen ? "Hide commit message" : "Show commit message"}
                                    aria-expanded={isCommitMessageOpen}
                                    onClick={event => {
                                        event.stopPropagation();
                                        setIsCommitMessageOpen(current => !current);
                                    }}
                                >
                                    <FileText className="size-3" />
                                    <span>Message</span>
                                    <ChevronDown
                                        className={cn(
                                            "size-3 transition-transform duration-200",
                                            isCommitMessageOpen && "rotate-180",
                                        )}
                                    />
                                </Button>
                            </>
                        )}
                    </div>

                    {/* Row 3: Expanded Commit Message (Collapsible) */}
                    {hasCommitMessage && isCommitMessageOpen && output?.commitMessage && (
                        <div
                            data-prevent-card-click
                            className={cn(
                                dashedBorderBox,
                                "mt-1 max-h-72 overflow-y-auto break-words whitespace-pre-wrap rounded-md bg-muted/40 p-3 font-mono text-xs sm:text-sm text-foreground/90",
                            )}
                        >
                            {output.commitMessage}
                        </div>
                    )}

                    {/* Optional: Error Banner */}
                    {output?.error && output.error.trim().length > 0 && (
                        <div
                            data-prevent-card-click
                            className={cn(
                                dashedBorderBox,
                                "mt-1 break-words whitespace-pre-wrap rounded-md border-destructive/40 bg-destructive/5 p-3 text-xs sm:text-sm text-destructive",
                            )}
                        >
                            {output.error}
                        </div>
                    )}
                </div>
            ) : (
                <div className="flex flex-col gap-3.5">
                    {/* Header Row: Status Badge, Method Badge, Webhook Badge, Cancel Button, Details Toggle Chevron, and Toolbar Action Buttons */}
                    <div className="flex flex-wrap items-center justify-between gap-3">
                        <div className="flex flex-wrap items-center gap-2 sm:gap-2.5">
                            <StatusBadge status={deployment.status} />

                            {isRepo ? (
                                <Badge
                                    variant="outline"
                                    className="h-7 px-3 text-sm font-medium rounded-md border-border/70 bg-muted/50 text-foreground flex items-center gap-1.5"
                                >
                                    <GitBranch className="size-3.5 shrink-0 text-muted-foreground" />
                                    <span>{formatRepoRef(deployment.settings.repoSource.repoRef)}</span>
                                </Badge>
                            ) : output?.imageTags && output.imageTags.length > 0 ? (
                                <Badge
                                    variant="outline"
                                    className="h-7 px-3 text-sm font-mono rounded-md border-border/70 bg-muted/50 text-foreground flex items-center gap-1.5"
                                >
                                    <Box className="size-3.5 shrink-0 text-muted-foreground" />
                                    <span>{output.imageTags[0]}</span>
                                </Badge>
                            ) : (
                                <Badge
                                    variant="outline"
                                    className="h-7 px-3 text-sm font-medium rounded-md border-border/70 bg-muted/50 text-foreground"
                                >
                                    Deployment
                                </Badge>
                            )}

                            {deployment.trigger?.source === EAppDeploymentTriggerSource.RepoWebhook && (
                                <Badge className="h-7 px-2.5 text-xs bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 hover:bg-emerald-500/20 border border-emerald-500/30">
                                    Webhook
                                </Badge>
                            )}

                            {onCancel && canCancelDeployment(deployment.status) && (
                                <PopConfirm
                                    title="Cancel deployment"
                                    description="Are you sure you want to cancel this deployment?"
                                    confirmText="Cancel deployment"
                                    cancelText="Cancel"
                                    variant="destructive"
                                    onConfirm={() => {
                                        onCancel(deployment.id);
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
                                aria-label={
                                    isDetailsContentOpen ? "Hide deployment details" : "Show deployment details"
                                }
                                title={isDetailsContentOpen ? "Hide deployment details" : "Show deployment details"}
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

                    {/* Compact Summary when details are collapsed */}
                    {!shouldShowDetailsContent && (
                        <div className="flex flex-wrap items-center gap-x-3.5 gap-y-1.5 text-sm text-muted-foreground pt-2 border-t border-border/50">
                            {isRepo && (
                                <div className="flex items-center gap-1 font-medium text-foreground">
                                    <GitBranch className="size-4 shrink-0 text-muted-foreground/70" />
                                    <span title={deployment.settings.repoSource.repoRef}>
                                        {formatRepoRef(deployment.settings.repoSource.repoRef)}
                                    </span>
                                </div>
                            )}

                            {output?.commitHashShort && (
                                <>
                                    {isRepo && <span className="text-muted-foreground/40 select-none">•</span>}
                                    <div className="flex items-center gap-1 font-mono">
                                        <GitCommit className="size-4 shrink-0 text-muted-foreground/70" />
                                        {output.commitURL ? (
                                            <a
                                                href={output.commitURL}
                                                target="_blank"
                                                rel="noreferrer"
                                                className="font-medium text-foreground underline underline-offset-4 hover:text-link"
                                                onClick={stopCardClick}
                                            >
                                                {output.commitHashShort}
                                            </a>
                                        ) : (
                                            <span className="font-medium text-foreground">
                                                {output.commitHashShort}
                                            </span>
                                        )}
                                    </div>
                                </>
                            )}

                            {sourceUser && (
                                <>
                                    <span className="text-muted-foreground/40 select-none">•</span>
                                    <div className="flex items-center gap-1.5">
                                        <Avatar
                                            name={getUserDisplayName(sourceUser)}
                                            src={sourceUser.photo}
                                            className="size-4.5 rounded-full text-[9px] border border-border shrink-0"
                                        />
                                        <span className="font-medium text-foreground">
                                            {getUserDisplayName(sourceUser)}
                                        </span>
                                    </div>
                                </>
                            )}

                            {shouldShowDuration(deployment) && (
                                <>
                                    <span className="text-muted-foreground/40 select-none">•</span>
                                    <div className="flex items-center gap-1.5 text-muted-foreground">
                                        <Clock className="size-4 shrink-0 text-muted-foreground/70" />
                                        <span>
                                            {formatDuration(deployment.startedAt, deployment.endedAt ?? now)} (
                                            <ReactTimeAgo
                                                date={deployment.startedAt}
                                                locale="en-US"
                                            />
                                            )
                                        </span>
                                    </div>
                                </>
                            )}
                        </div>
                    )}

                    {/* Expanded Details Section */}
                    {shouldShowDetailsContent && (
                        <div className="flex flex-col gap-3.5 pt-3 border-t border-border/60">
                            {/* Source & Trigger Banner */}
                            <div className="flex flex-col gap-1.5">
                                <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                                    Source & Trigger
                                </span>
                                <div className="flex flex-wrap items-center justify-between gap-3 rounded-md border border-border bg-muted/30 px-3.5 py-2.5 text-sm">
                                    {isRepo ? (
                                        <div className="flex flex-wrap items-center gap-3">
                                            <div className="flex items-center gap-1.5 font-medium text-foreground">
                                                <FolderGit2 className="size-4 shrink-0 text-muted-foreground/70" />
                                                {deployment.settings.repoSource.repoUrl ? (
                                                    <a
                                                        href={deployment.settings.repoSource.repoUrl}
                                                        target="_blank"
                                                        rel="noreferrer"
                                                        className="underline underline-offset-4 hover:text-link"
                                                        title={deployment.settings.repoSource.repoUrl}
                                                        onClick={stopCardClick}
                                                    >
                                                        {getRepoShortName(deployment.settings.repoSource.repoUrl)}
                                                    </a>
                                                ) : (
                                                    <span>Repository</span>
                                                )}
                                            </div>

                                            <span className="text-muted-foreground/40 select-none">•</span>

                                            <div className="flex items-center gap-1 text-foreground">
                                                <GitBranch className="size-4 shrink-0 text-muted-foreground/70" />
                                                <span className="font-mono text-xs font-semibold bg-muted/70 px-1.5 py-0.5 rounded border border-border/60">
                                                    {formatRepoRef(deployment.settings.repoSource.repoRef)}
                                                </span>
                                            </div>

                                            {output?.commitHashShort && (
                                                <>
                                                    <span className="text-muted-foreground/40 select-none">•</span>
                                                    <div className="flex items-center gap-1.5 font-mono text-xs">
                                                        <GitCommit className="size-4 shrink-0 text-muted-foreground/70" />
                                                        {output.commitURL ? (
                                                            <a
                                                                href={output.commitURL}
                                                                target="_blank"
                                                                rel="noreferrer"
                                                                className="font-medium text-foreground underline underline-offset-4 hover:text-link"
                                                                onClick={stopCardClick}
                                                            >
                                                                {output.commitHashShort}
                                                            </a>
                                                        ) : (
                                                            <span className="font-medium text-foreground">
                                                                {output.commitHashShort}
                                                            </span>
                                                        )}
                                                    </div>
                                                </>
                                            )}

                                            {output?.commitTitle && (
                                                <>
                                                    <span className="text-muted-foreground/40 select-none">•</span>
                                                    <div className="flex items-center gap-1.5 max-w-[340px] sm:max-w-[420px] truncate">
                                                        <span
                                                            className="truncate text-foreground font-medium"
                                                            title={output.commitTitle}
                                                        >
                                                            {output.commitTitle}
                                                        </span>
                                                        {hasCommitMessage && (
                                                            <Button
                                                                type="button"
                                                                variant="ghost"
                                                                size="icon-sm"
                                                                className={cn(
                                                                    "size-6 shrink-0 rounded transition-colors",
                                                                    isCommitMessageOpen
                                                                        ? "bg-primary/10 text-primary"
                                                                        : "text-muted-foreground hover:text-primary",
                                                                )}
                                                                aria-label={
                                                                    isCommitMessageOpen
                                                                        ? "Hide commit message"
                                                                        : "Show commit message"
                                                                }
                                                                title={
                                                                    isCommitMessageOpen
                                                                        ? "Hide commit message"
                                                                        : "Show commit message"
                                                                }
                                                                aria-expanded={isCommitMessageOpen}
                                                                onClick={event => {
                                                                    event.stopPropagation();
                                                                    setIsCommitMessageOpen(current => !current);
                                                                }}
                                                            >
                                                                <Info className="size-3.5" />
                                                            </Button>
                                                        )}
                                                    </div>
                                                </>
                                            )}
                                        </div>
                                    ) : (
                                        <div className="flex items-center gap-2">
                                            <Box className="size-4 shrink-0 text-muted-foreground/70" />
                                            <span className="text-muted-foreground">Docker Image:</span>
                                            <span className="font-mono text-foreground font-medium">
                                                {output?.imageTags.length ? output.imageTags.join(", ") : "-"}
                                            </span>
                                        </div>
                                    )}

                                    {/* Trigger on right of banner */}
                                    {sourceUser && (
                                        <div className="flex items-center gap-1.5 shrink-0">
                                            <span className="text-muted-foreground">Trigger:</span>
                                            <Avatar
                                                name={getUserDisplayName(sourceUser)}
                                                src={sourceUser.photo}
                                                className="size-5 rounded-full text-[10px] border border-border shrink-0"
                                            />
                                            <span className="font-medium text-foreground">
                                                {getUserDisplayName(sourceUser)}
                                            </span>
                                            {sourceUser.email && (
                                                <span className="text-xs text-muted-foreground hidden sm:inline">
                                                    ({sourceUser.email})
                                                </span>
                                            )}
                                        </div>
                                    )}

                                    {deployment.trigger?.source === EAppDeploymentTriggerSource.RepoWebhook && (
                                        <div className="flex items-center gap-1.5 shrink-0">
                                            <span className="text-muted-foreground">Trigger:</span>
                                            <Badge className="h-5 px-1.5 text-xs bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border border-emerald-500/30">
                                                Webhook
                                            </Badge>
                                        </div>
                                    )}
                                </div>
                            </div>

                            {/* Detailed 4-Column Metadata Grid */}
                            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-x-6 gap-y-2.5 text-sm">
                                <div>
                                    <span className="text-muted-foreground">Deployment ID:</span>{" "}
                                    <span className="font-mono text-foreground font-medium select-all">
                                        {deployment.id}
                                    </span>
                                </div>

                                <div>
                                    <span className="text-muted-foreground">Started At:</span>{" "}
                                    <span className="text-foreground font-medium">
                                        {formatDateTime(deployment.startedAt)}
                                    </span>
                                </div>

                                <div>
                                    <span className="text-muted-foreground">Ended At:</span>{" "}
                                    <span className="text-foreground font-medium">
                                        {formatDateTime(deployment.endedAt)}
                                    </span>
                                </div>

                                {shouldShowDuration(deployment) && (
                                    <div>
                                        <span className="text-muted-foreground">Duration:</span>{" "}
                                        <span className="text-foreground font-medium">
                                            {formatDuration(deployment.startedAt, deployment.endedAt ?? now)}
                                        </span>{" "}
                                        <span className="text-xs text-muted-foreground">
                                            (
                                            <ReactTimeAgo
                                                date={deployment.startedAt}
                                                locale="en-US"
                                            />
                                            )
                                        </span>
                                    </div>
                                )}

                                <div>
                                    <span className="text-muted-foreground">Method:</span>{" "}
                                    <span className="text-foreground font-medium capitalize">
                                        {deployment.settings.activeMethod === EAppDeploymentMethod.Repo
                                            ? "Git Repository"
                                            : "Docker Image"}
                                    </span>
                                </div>

                                {isRepo && (
                                    <div>
                                        <span className="text-muted-foreground">Branch / Ref:</span>{" "}
                                        <span className="font-mono text-foreground font-medium">
                                            {formatRepoRef(deployment.settings.repoSource.repoRef)}
                                        </span>
                                    </div>
                                )}

                                {output?.commitAuthor && (
                                    <div>
                                        <span className="text-muted-foreground">Commit Author:</span>{" "}
                                        <span className="text-foreground font-medium">{output.commitAuthor}</span>
                                    </div>
                                )}

                                {output?.commitHashShort && (
                                    <div>
                                        <span className="text-muted-foreground">Commit Hash:</span>{" "}
                                        {output.commitURL ? (
                                            <a
                                                href={output.commitURL}
                                                target="_blank"
                                                rel="noreferrer"
                                                className="font-mono text-foreground font-medium underline underline-offset-4 hover:text-link select-all"
                                                onClick={stopCardClick}
                                            >
                                                {output.commitHashShort}
                                            </a>
                                        ) : (
                                            <span className="font-mono text-foreground font-medium select-all">
                                                {output.commitHashShort}
                                            </span>
                                        )}
                                    </div>
                                )}

                                {output?.imageTags && output.imageTags.length > 0 && (
                                    <div className="sm:col-span-2">
                                        <span className="text-muted-foreground">Docker Image:</span>{" "}
                                        <span className="font-mono text-foreground font-medium select-all">
                                            {output.imageTags.join(", ")}
                                        </span>
                                    </div>
                                )}

                                <div>
                                    <span className="text-muted-foreground">Created At:</span>{" "}
                                    <span className="text-foreground font-medium">
                                        {formatDateTime(deployment.createdAt)}
                                    </span>
                                </div>

                                {deployment.updatedAt && (
                                    <div>
                                        <span className="text-muted-foreground">Updated At:</span>{" "}
                                        <span className="text-foreground font-medium">
                                            {formatDateTime(deployment.updatedAt)}
                                        </span>
                                    </div>
                                )}
                            </div>

                            {/* Commit Message Box (when toggled open) */}
                            {hasCommitMessage && isCommitMessageOpen && output?.commitMessage && (
                                <div className="flex flex-col gap-1.5 pt-1">
                                    <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                                        Commit Message
                                    </span>
                                    <div className="rounded-md border border-border bg-muted/40 p-3 max-h-60 overflow-y-auto">
                                        <pre className="font-mono text-xs text-foreground/90 whitespace-pre-wrap break-words">
                                            {output.commitMessage}
                                        </pre>
                                    </div>
                                </div>
                            )}

                            {/* Error Banner */}
                            {output?.error.trim() ? (
                                <div className="flex flex-col gap-1.5 pt-1">
                                    <span className="text-[11px] font-semibold text-destructive uppercase tracking-wider">
                                        Error
                                    </span>
                                    <div className="rounded-md border border-destructive/30 bg-destructive/10 p-3 max-h-60 overflow-y-auto">
                                        <pre className="font-mono text-xs text-destructive whitespace-pre-wrap break-all select-all">
                                            {output.error}
                                        </pre>
                                    </div>
                                </div>
                            ) : null}
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

export function DeploymentSummaryCardSkeleton({ variant = "list" }: { variant?: DeploymentSummaryCardVariant }) {
    if (variant === "list") {
        return (
            <div className="rounded-[8px] border border-border bg-background p-3 sm:p-4 border-l-4 border-l-muted">
                <div className="flex flex-col gap-2.5">
                    <div className="flex items-center justify-between gap-3">
                        <div className="flex items-center gap-2.5 flex-1">
                            <Skeleton className="h-7 w-20 rounded-md" />
                            <Skeleton className="h-5 w-64 max-w-[50%] rounded-md" />
                        </div>
                        <Skeleton className="h-5 w-40 rounded-md" />
                    </div>
                    <div className="flex items-center gap-3 pt-2 border-t border-border/40">
                        <Skeleton className="h-4.5 w-24" />
                        <Skeleton className="h-4.5 w-20" />
                        <Skeleton className="h-4.5 w-36" />
                        <Skeleton className="h-4.5 w-28" />
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
                    <Skeleton className="h-7 w-28 rounded-md" />
                    <Skeleton className="size-7 rounded-md" />
                </div>
                <div className="flex items-center gap-2">
                    <Skeleton className="size-8 rounded-md" />
                    <Skeleton className="size-8 rounded-md" />
                    <Skeleton className="size-8 rounded-md" />
                    <Skeleton className="size-8 rounded-md" />
                </div>
            </div>
            <div className="h-11 rounded-md bg-muted/40 border border-border" />
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

type DeploymentSummaryCardVariant = "list" | "details";

interface DeploymentSummaryCardProps {
    deployment: AppDeployment;
    now: Date;
    children?: ReactNode;
    variant?: DeploymentSummaryCardVariant;
    isCancelling?: boolean;
    isFullscreen?: boolean;
    isFullView?: boolean;
    isFullHeight?: boolean;
    fontSize?: number;
    themeId?: string;
    onCancel?: (deploymentID: string) => void;
    onClick?: () => void;
    onToggleFullscreen?: () => void;
    onToggleFullView?: () => void;
    onToggleFullHeight?: () => void;
    onCycleFontSize?: () => void;
    onSelectTheme?: (themeId: string) => void;
}
