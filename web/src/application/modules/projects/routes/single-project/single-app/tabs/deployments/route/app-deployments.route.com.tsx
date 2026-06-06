import { type ReactNode, useEffect, useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { listBox } from "@lib/styles";
import { useParams } from "react-router";
import ReactTimeAgo from "react-time-ago";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { AppDeploymentsCommands, AppDeploymentsQueries } from "~/projects/data";
import type { AppDeployment, AppDeploymentOutput, AppDeploymentSourceUser } from "~/projects/domain";
import {
    EAppDeploymentMethod,
    EAppDeploymentStatus,
    EAppDeploymentTriggerSource,
} from "~/projects/module-shared/enums";

import { TableActions } from "@application/shared/components";
import { DEFAULT_PAGINATED_DATA } from "@application/shared/constants";
import { useTableState } from "@application/shared/hooks/table";
import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import { Avatar, Button, Skeleton, TablePagination } from "@/components/ui";

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

function getUserDisplayName(user: AppDeploymentSourceUser): string {
    return user.fullName || user.email || user.username;
}

function canCancelDeployment(status: EAppDeploymentStatus): boolean {
    return status === EAppDeploymentStatus.InProgress || status === EAppDeploymentStatus.NotStarted;
}

function shouldShowDuration(deployment: AppDeployment): deployment is AppDeployment & { startedAt: Date } {
    return deployment.status !== EAppDeploymentStatus.NotStarted && deployment.startedAt != null;
}

function shouldShowCommit(deployment: AppDeployment): deployment is AppDeployment & { output: AppDeploymentOutput } {
    return deployment.settings.activeMethod === EAppDeploymentMethod.Repo && deployment.output != null;
}

function useCurrentTime(enabled: boolean): Date {
    const [now, setNow] = useState(() => new Date());

    useEffect(() => {
        if (!enabled) {
            return;
        }

        const interval = window.setInterval(() => {
            setNow(new Date());
        }, 60_000);

        return () => {
            window.clearInterval(interval);
        };
    }, [enabled]);

    return now;
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

function CommitValue({ output }: { output: AppDeploymentOutput }) {
    const commitURL = output.commitURL || undefined;

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
                    {(output.commitAuthor || output.commitHashShort) && <span>-</span>}
                    {commitURL ? (
                        <a
                            href={commitURL}
                            target="_blank"
                            rel="noreferrer"
                            className="min-w-0 max-w-full truncate underline underline-offset-4 hover:text-primary"
                            title={output.commitTitle}
                        >
                            {output.commitTitle}
                        </a>
                    ) : (
                        <span className="min-w-0 max-w-full truncate">{output.commitTitle}</span>
                    )}
                </>
            )}
        </div>
    );
}

function DeploymentCard({
    deployment,
    now,
    isCancelling,
    onCancel,
}: {
    deployment: AppDeployment;
    now: Date;
    isCancelling: boolean;
    onCancel: (deploymentID: string) => void;
}) {
    const { output } = deployment;
    const sourceUser = deployment.trigger?.sourceUser;

    return (
        <article className="rounded-lg border border-border border-l-4 border-l-purple-300 bg-background p-5 shadow-xs">
            <dl className="grid grid-cols-1 items-center gap-y-2 sm:grid-cols-[130px_minmax(0,1fr)] sm:gap-x-8 sm:gap-y-4">
                <InfoRow label="Status">
                    <div className="flex flex-wrap items-center gap-3">
                        <StatusBadge status={deployment.status} />
                        {canCancelDeployment(deployment.status) && (
                            <Button
                                type="button"
                                variant="link"
                                className="h-auto p-0 text-primary"
                                isLoading={isCancelling}
                                onClick={() => {
                                    onCancel(deployment.id);
                                }}
                            >
                                Cancel Deployment
                            </Button>
                        )}
                    </div>
                </InfoRow>

                {shouldShowDuration(deployment) && (
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

                {deployment.trigger?.source === EAppDeploymentTriggerSource.RepoWebhook && (
                    <InfoRow label="Trigger">
                        <Badge className="h-7 bg-emerald-500 px-4 text-white hover:bg-emerald-500/90">Webhook</Badge>
                    </InfoRow>
                )}

                {sourceUser && (
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

                {shouldShowCommit(deployment) &&
                    (deployment.output.commitAuthor ||
                        deployment.output.commitHashShort ||
                        deployment.output.commitTitle) && (
                        <InfoRow label="Commit">
                            <CommitValue output={deployment.output} />
                        </InfoRow>
                    )}

                {output && output.imageTags.length > 0 && (
                    <InfoRow label="Docker Image">
                        <span className="break-words">{output.imageTags.join(", ")}</span>
                    </InfoRow>
                )}
            </dl>
        </article>
    );
}

function DeploymentCardSkeleton() {
    return (
        <div className="rounded-lg border border-border border-l-4 border-l-purple-200 bg-background p-5">
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

export function AppDeploymentsRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const { pagination, setPagination, sorting, search, setSearch } = useTableState();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data: { data: deployments, meta } = DEFAULT_PAGINATED_DATA, isFetching } =
        AppDeploymentsQueries.useFindManyPaginated({
            projectID: projectId,
            appID: appId,
            pagination,
            sorting,
            search,
        });

    const hasActiveDeployment = useMemo(
        () => deployments.some(deployment => deployment.status === EAppDeploymentStatus.InProgress),
        [deployments],
    );
    const now = useCurrentTime(hasActiveDeployment);
    const pageCount = Math.max(1, Math.ceil(meta.page.total / pagination.size));

    const { mutate: cancelDeployment, isPending: isCancelling } = AppDeploymentsCommands.useCancel({
        onSuccess: () => {
            toast.success("Deployment canceled");
        },
    });

    return (
        <section className={cn(listBox)}>
            <div className="flex flex-col gap-5">
                <TableActions search={{ value: search, onChange: setSearch }} />

                <div className="flex flex-col gap-4">
                    {isFetching && deployments.length === 0 ? (
                        <>
                            <DeploymentCardSkeleton />
                            <DeploymentCardSkeleton />
                            <DeploymentCardSkeleton />
                        </>
                    ) : deployments.length === 0 ? (
                        <div className="rounded-lg border border-dashed border-border p-8 text-center text-sm text-muted-foreground">
                            No deployments found.
                        </div>
                    ) : (
                        deployments.map(deployment => (
                            <DeploymentCard
                                key={deployment.id}
                                deployment={deployment}
                                now={now}
                                isCancelling={isCancelling}
                                onCancel={deploymentID => {
                                    cancelDeployment({
                                        projectID: projectId,
                                        appID: appId,
                                        deploymentID,
                                    });
                                }}
                            />
                        ))
                    )}
                </div>

                {meta.page.total > 0 && (
                    <TablePagination
                        pageIndex={pagination.page - 1}
                        pageSize={pagination.size}
                        pageCount={pageCount}
                        totalCount={meta.page.total}
                        showPageSizeSelector={false}
                        onPageChange={pageIndex => {
                            setPagination(prev => ({ ...prev, page: pageIndex + 1 }));
                        }}
                    />
                )}
            </div>
        </section>
    );
}
