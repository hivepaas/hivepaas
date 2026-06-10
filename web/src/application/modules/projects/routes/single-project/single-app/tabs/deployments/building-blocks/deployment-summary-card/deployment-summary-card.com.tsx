import { type KeyboardEvent, type MouseEvent, type ReactNode, useState } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { dashedBorderBox } from "@lib/styles";
import { format } from "date-fns";
import { ChevronDown, Info } from "lucide-react";
import ReactTimeAgo from "react-time-ago";
import type { AppDeployment, AppDeploymentOutput, AppDeploymentSourceUser } from "~/projects/domain";
import {
    EAppDeploymentMethod,
    EAppDeploymentStatus,
    EAppDeploymentTriggerSource,
} from "~/projects/module-shared/enums";

import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import { Avatar, Button, Skeleton } from "@/components/ui";

const STATUS_LABELS: Record<EAppDeploymentStatus, string> = {
    [EAppDeploymentStatus.Done]: "Done",
    [EAppDeploymentStatus.Failed]: "Failed",
    [EAppDeploymentStatus.InProgress]: "In-Progress",
    [EAppDeploymentStatus.NotStarted]: "Not Started",
    [EAppDeploymentStatus.Canceled]: "Canceled",
};

const STATUS_CLASS_NAMES: Record<EAppDeploymentStatus, string> = {
    [EAppDeploymentStatus.Done]: "bg-green-500 text-white hover:bg-green-500/90",
    [EAppDeploymentStatus.Failed]: "bg-red-500 text-white hover:bg-red-500/90",
    [EAppDeploymentStatus.InProgress]: "bg-purple-400 text-white hover:bg-purple-400/90",
    [EAppDeploymentStatus.NotStarted]: "bg-blue-400 text-white hover:bg-blue-400/90",
    [EAppDeploymentStatus.Canceled]: "bg-zinc-500 text-white hover:bg-zinc-500/90",
};

const STATUS_BORDER_CLASS_NAMES: Record<EAppDeploymentStatus, string> = {
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

function getUserDisplayName(user: AppDeploymentSourceUser): string {
    return user.fullName || user.email || user.username;
}

function canCancelDeployment(status: EAppDeploymentStatus): boolean {
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

function shouldShowCommit(deployment: AppDeployment): deployment is RepoDeployment & { output: AppDeploymentOutput } {
    return isRepoDeployment(deployment) && deployment.output != null;
}

function stopCardClick(event: MouseEvent<HTMLElement>) {
    event.stopPropagation();
}

function StatusBadge({ status }: { status: EAppDeploymentStatus }) {
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

function RepositoryValue({ repoUrl, repoRef }: { repoUrl: string; repoRef: string }) {
    return (
        <div className="flex min-w-0 flex-wrap items-center gap-x-1 gap-y-1">
            {repoUrl ? (
                <a
                    href={repoUrl}
                    target="_blank"
                    rel="noreferrer"
                    className="min-w-0 max-w-full break-all underline underline-offset-4 hover:text-primary"
                    onClick={stopCardClick}
                >
                    {repoUrl}
                </a>
            ) : (
                <span />
            )}
            <span>-</span>
            <span className="break-words">{repoRef}</span>
        </div>
    );
}

function CommitValue({
    output,
    isCommitMessageOpen,
    onToggleCommitMessage,
}: {
    output: AppDeploymentOutput;
    isCommitMessageOpen: boolean;
    onToggleCommitMessage: () => void;
}) {
    const commitURL = output.commitURL || undefined;
    const hasCommitMessage = output.commitMessage.trim().length > 0;

    return (
        <div className="flex min-w-0 flex-wrap items-center gap-x-1 gap-y-1">
            {output.commitAuthor && <span>{output.commitAuthor}</span>}
            {output.commitHashShort && (
                <>
                    {output.commitAuthor && <span>-</span>}
                    {commitURL ? (
                        <a
                            href={commitURL}
                            target="_blank"
                            rel="noreferrer"
                            className="underline underline-offset-4 hover:text-primary"
                            onClick={stopCardClick}
                        >
                            {output.commitHashShort}
                        </a>
                    ) : (
                        <span>{output.commitHashShort}</span>
                    )}
                </>
            )}
            {output.commitTitle && (
                <>
                    {(output.commitAuthor || output.commitHashShort) && <span className="shrink-0">-</span>}
                    <span className="flex min-w-0 flex-1 items-center gap-x-1">
                        {commitURL ? (
                            <a
                                href={commitURL}
                                target="_blank"
                                rel="noreferrer"
                                className="min-w-0 flex-1 truncate underline underline-offset-4 hover:text-primary"
                                title={output.commitTitle}
                                onClick={stopCardClick}
                            >
                                {output.commitTitle}
                            </a>
                        ) : (
                            <span
                                className="min-w-0 flex-1 truncate"
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
                                className="size-6 shrink-0 text-primary hover:text-primary"
                                aria-label="Toggle commit message"
                                title="Toggle commit message"
                                aria-pressed={isCommitMessageOpen}
                                onClick={event => {
                                    event.stopPropagation();
                                    onToggleCommitMessage();
                                }}
                            >
                                <Info className="size-4" />
                            </Button>
                        )}
                    </span>
                </>
            )}
        </div>
    );
}

export function DeploymentSummaryCard({
    deployment,
    now,
    children,
    variant = "list",
    isCancelling = false,
    onCancel,
    onClick,
}: DeploymentSummaryCardProps) {
    const { output } = deployment;
    const sourceUser = deployment.trigger?.sourceUser;
    const [isCommitMessageOpen, setIsCommitMessageOpen] = useState(false);
    const [isDetailsContentOpen, setIsDetailsContentOpen] = useState(false);
    const isRepo = isRepoDeployment(deployment);
    const isClickable = Boolean(onClick);
    const shouldShowDetailsContent = variant === "list" || isDetailsContentOpen;

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
                variant === "list" ? [STATUS_BORDER_CLASS_NAMES[deployment.status]] : "border-0",
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
            <dl className="grid grid-cols-1 items-center gap-y-2 sm:grid-cols-[130px_minmax(0,1fr)] sm:gap-x-8 sm:gap-y-4">
                <InfoRow label="Status">
                    <div className="flex flex-wrap items-center gap-3">
                        <StatusBadge status={deployment.status} />
                        {onCancel && canCancelDeployment(deployment.status) && (
                            <Button
                                type="button"
                                variant="link"
                                className="h-auto p-0 text-primary"
                                isLoading={isCancelling}
                                onClick={event => {
                                    event.stopPropagation();
                                    onCancel(deployment.id);
                                }}
                            >
                                Cancel Deployment
                            </Button>
                        )}
                        {variant === "details" && (
                            <Button
                                type="button"
                                variant="ghost"
                                size="icon-sm"
                                className="size-6 text-primary hover:text-primary"
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
                        )}
                    </div>
                </InfoRow>

                {variant === "details" && shouldShowDetailsContent && (
                    <>
                        <InfoRow label="Started At">
                            <span>{formatDateTime(deployment.startedAt)}</span>
                        </InfoRow>
                        <InfoRow label="Ended At">
                            <span>{formatDateTime(deployment.endedAt)}</span>
                        </InfoRow>
                    </>
                )}

                {shouldShowDetailsContent && shouldShowDuration(deployment) && (
                    <InfoRow label="Duration">
                        <span>
                            {formatDuration(deployment.startedAt, deployment.endedAt ?? now)} from{" "}
                            <ReactTimeAgo
                                date={deployment.startedAt}
                                locale="en-US"
                            />
                        </span>
                    </InfoRow>
                )}

                {shouldShowDetailsContent && deployment.trigger?.source === EAppDeploymentTriggerSource.RepoWebhook && (
                    <InfoRow label="Trigger">
                        <Badge className="h-7 bg-emerald-500 px-4 text-white hover:bg-emerald-500/90">Webhook</Badge>
                    </InfoRow>
                )}

                {shouldShowDetailsContent && sourceUser && (
                    <InfoRow label="Trigger">
                        <div className="flex min-w-0 items-center gap-2">
                            <Avatar
                                name={getUserDisplayName(sourceUser)}
                                src={sourceUser.photo}
                                className="size-7 text-xs"
                            />
                            <span className="truncate">{getUserDisplayName(sourceUser)}</span>
                        </div>
                    </InfoRow>
                )}

                {shouldShowDetailsContent && isRepo && (
                    <InfoRow label="Repository">
                        <RepositoryValue
                            repoUrl={deployment.settings.repoSource.repoUrl}
                            repoRef={deployment.settings.repoSource.repoRef}
                        />
                    </InfoRow>
                )}

                {shouldShowDetailsContent &&
                    shouldShowCommit(deployment) &&
                    (deployment.output.commitAuthor ||
                        deployment.output.commitHashShort ||
                        deployment.output.commitTitle) && (
                        <InfoRow label="Commit">
                            <CommitValue
                                output={deployment.output}
                                isCommitMessageOpen={isCommitMessageOpen}
                                onToggleCommitMessage={() => {
                                    setIsCommitMessageOpen(current => !current);
                                }}
                            />
                        </InfoRow>
                    )}

                {shouldShowDetailsContent && output?.commitMessage && isCommitMessageOpen && (
                    <div className={cn(dashedBorderBox, "sm:col-span-2 break-words whitespace-pre-wrap")}>
                        {output.commitMessage}
                    </div>
                )}

                {shouldShowDetailsContent && output && output.imageTags.length > 0 && (
                    <InfoRow label="Docker Image">
                        <span className="break-words">{output.imageTags.join(", ")}</span>
                    </InfoRow>
                )}
            </dl>

            {shouldShowDetailsContent && children && <div className="mt-5 min-w-0">{children}</div>}
        </div>
    );
}

export function DeploymentSummaryCardSkeleton({ variant = "list" }: { variant?: DeploymentSummaryCardVariant }) {
    return (
        <div
            className={cn(
                "rounded-lg border border-border bg-background p-5",
                variant === "list" && "border-l-4 border-l-purple-200",
            )}
        >
            <div className="grid grid-cols-1 items-center gap-y-2 sm:grid-cols-[130px_minmax(0,1fr)] sm:gap-x-8 sm:gap-y-4">
                <Skeleton className="h-5 w-16" />
                <Skeleton className="h-7 w-56" />
                <Skeleton className="h-5 w-20" />
                <Skeleton className="h-5 w-72" />
                <Skeleton className="h-5 w-16" />
                <Skeleton className="h-7 w-32" />
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
    onCancel?: (deploymentID: string) => void;
    onClick?: () => void;
}
