import { Fragment, useEffect, useMemo, useState } from "react";

import { cn } from "@/lib/utils";
import { Badge } from "@components/ui/badge";
import { dashedBorderBox, listBox } from "@lib/styles";
import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { AppServiceTasksQueries } from "~/projects/data";
import type { AppServiceTask, AppServiceTaskNode } from "~/projects/domain";

import { AppLink, TableActions } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import { Button, Skeleton, Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui";

const APP_INSTANCES_REFETCH_INTERVAL_MS = 5_000;
const APP_INSTANCES_DURATION_TICK_MS = 5_000;
const TABLE_COLUMN_COUNT = 7;
const EMPTY_INSTANCES: AppServiceTask[] = [];

const TASK_STATE_CLASS_NAMES: Record<string, string> = {
    running: "bg-green-500 text-white hover:bg-green-500/90",
    shutdown: "bg-pink-400 text-white hover:bg-pink-400/90",
    complete: "bg-purple-500 text-white hover:bg-purple-500/90",
    failed: "bg-red-500 text-white hover:bg-red-500/90",
};

const NODE_ROLE_CLASS_NAMES: Record<string, string> = {
    manager: "bg-primary text-white",
    worker: "bg-blue-500 text-white",
};

function useCurrentTime(): Date {
    const [now, setNow] = useState(() => new Date());

    useEffect(() => {
        const interval = window.setInterval(() => {
            setNow(new Date());
        }, APP_INSTANCES_DURATION_TICK_MS);

        return () => {
            window.clearInterval(interval);
        };
    }, []);

    return now;
}

function normalize(value: string | number | null | undefined): string {
    return String(value ?? "")
        .trim()
        .toLowerCase();
}

function formatStateLabel(state: string | null | undefined): string {
    if (!state) {
        return "-";
    }

    return state
        .split("-")
        .filter(Boolean)
        .map(part => part.charAt(0).toUpperCase() + part.slice(1))
        .join(" ");
}

function formatDuration(timestamp: Date | null | undefined, now: Date): string {
    if (!timestamp) {
        return "-";
    }

    const durationMs = Math.max(0, now.getTime() - timestamp.getTime());

    if (durationMs < 1_000) {
        return "less than a minute";
    }

    return timeAgoFormatter
        .format(now.getTime() - durationMs, "round")
        .replace(/\s+ago$/, "")
        .replace(/^in\s+/, "");
}

function getNodeName(node: AppServiceTaskNode | null): string {
    if (node?.name) {
        return node.name;
    }

    if (node?.hostname) {
        return node.hostname;
    }

    return "-";
}

function getNodeAddr(node: AppServiceTaskNode | null): string {
    if (node?.addr) {
        return node.addr;
    }

    return "-";
}

function getNodeRoleLabel(node: AppServiceTaskNode | null): string {
    if (!node?.role) {
        return "-";
    }

    if (node.role === "manager" && node.isLeader) {
        return "Leader";
    }

    if (node.role === "manager") {
        return "Manager";
    }

    if (node.role === "worker") {
        return "Worker";
    }

    return formatStateLabel(node.role);
}

function matchesSearch(task: AppServiceTask, search: string): boolean {
    const query = normalize(search);

    if (!query) {
        return true;
    }

    const { node } = task;
    const searchableValues = [
        task.slot,
        task.status?.state,
        task.status?.err,
        node?.name,
        node?.hostname,
        node?.addr,
        node?.role,
        getNodeRoleLabel(node),
    ];

    return searchableValues.some(value => normalize(value).includes(query));
}

function StateBadge({ state }: { state: string | null | undefined }) {
    if (!state) {
        return <span>-</span>;
    }

    const normalizedState = normalize(state);
    const className = TASK_STATE_CLASS_NAMES[normalizedState] ?? "bg-cyan-500 text-white hover:bg-cyan-500/90";

    return <Badge className={cn("h-7 px-3", className)}>{formatStateLabel(state)}</Badge>;
}

function NodeRoleBadge({ node }: { node: AppServiceTaskNode | null }) {
    const normalizedRole = normalize(node?.role);

    if (!normalizedRole) {
        return <span>-</span>;
    }

    const className =
        normalizedRole === "manager" && node?.isLeader
            ? "bg-purple-500 text-white"
            : (NODE_ROLE_CLASS_NAMES[normalizedRole] ?? "bg-primary text-primary-foreground");

    return <Badge className={cn("h-7 px-3", className)}>{getNodeRoleLabel(node)}</Badge>;
}

function InstancesTableSkeleton() {
    return (
        <>
            {Array.from({ length: 3 }).map((_, rowIndex) => (
                <TableRow key={rowIndex}>
                    {Array.from({ length: TABLE_COLUMN_COUNT }).map((__, cellIndex) => (
                        <TableCell key={cellIndex}>
                            <Skeleton className="h-5 w-full" />
                        </TableCell>
                    ))}
                </TableRow>
            ))}
        </>
    );
}

export function AppInstancesRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();
    const [search, setSearch] = useState("");
    const [expandedErrorIds, setExpandedErrorIds] = useState<Set<string>>(() => new Set());
    const now = useCurrentTime();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    const { data, isFetching, isLoading } = AppServiceTasksQueries.useFindMany(
        {
            projectID: projectId,
            appID: appId,
        },
        {
            refetchInterval: APP_INSTANCES_REFETCH_INTERVAL_MS,
        },
    );
    const instances = data?.data ?? EMPTY_INSTANCES;

    const filteredInstances = useMemo(() => {
        return instances.filter(instance => matchesSearch(instance, search));
    }, [instances, search]);

    function toggleError(taskId: string) {
        setExpandedErrorIds(current => {
            const next = new Set(current);

            if (next.has(taskId)) {
                next.delete(taskId);
            } else {
                next.add(taskId);
            }

            return next;
        });
    }

    const isInitialLoading = isLoading || (isFetching && instances.length === 0);
    const emptyMessage = instances.length === 0 ? "No instances found." : "No matching instances.";

    return (
        <section className={cn(listBox)}>
            <div className="flex flex-col gap-5">
                <TableActions
                    search={{ value: search, onChange: setSearch }}
                    renderActions={
                        <AppLink.Basic
                            to={ROUTE.projects.single.apps.single.configuration.availabilityAndScaling.$route(
                                projectId,
                                appId,
                            )}
                            className="text-sm font-medium text-primary underline-offset-4 hover:underline"
                        >
                            Go to Availability &amp; Scaling
                        </AppLink.Basic>
                    }
                />

                <div className="relative rounded-md border overflow-hidden">
                    {isFetching && instances.length > 0 && (
                        <div className="absolute left-0 right-0 top-0 z-10 h-1 overflow-hidden bg-muted">
                            <div
                                className="h-full w-1/3 bg-primary"
                                style={{ animation: "progress 1.5s ease-in-out infinite" }}
                            />
                        </div>
                    )}

                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead className="w-[110px] text-center">Instance</TableHead>
                                <TableHead className="w-[150px] text-center">State</TableHead>
                                <TableHead className="w-[180px] text-center">Duration</TableHead>
                                <TableHead className="w-[100px] text-center">Error</TableHead>
                                <TableHead>Node</TableHead>
                                <TableHead className="w-[170px]">Node Addr</TableHead>
                                <TableHead className="w-[140px] text-center">Node Role</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {isInitialLoading ? (
                                <InstancesTableSkeleton />
                            ) : filteredInstances.length === 0 ? (
                                <TableRow>
                                    <TableCell
                                        colSpan={TABLE_COLUMN_COUNT}
                                        className="h-24 text-center text-muted-foreground"
                                    >
                                        {emptyMessage}
                                    </TableCell>
                                </TableRow>
                            ) : (
                                filteredInstances.map(instance => {
                                    const error = instance.status?.err.trim() ?? "";
                                    const isErrorExpanded = expandedErrorIds.has(instance.id);

                                    return (
                                        <Fragment key={instance.id}>
                                            <TableRow>
                                                <TableCell className="text-center font-medium">
                                                    {instance.slot}
                                                </TableCell>
                                                <TableCell className="text-center">
                                                    <StateBadge state={instance.status?.state} />
                                                </TableCell>
                                                <TableCell className="text-center">
                                                    {formatDuration(instance.status?.timestamp, now)}
                                                </TableCell>
                                                <TableCell className="text-center">
                                                    {error ? (
                                                        <Button
                                                            type="button"
                                                            variant="link"
                                                            className="h-auto p-0"
                                                            aria-expanded={isErrorExpanded}
                                                            onClick={() => {
                                                                toggleError(instance.id);
                                                            }}
                                                        >
                                                            {isErrorExpanded ? "Hide" : "Show"}
                                                        </Button>
                                                    ) : (
                                                        "-"
                                                    )}
                                                </TableCell>
                                                <TableCell>{getNodeName(instance.node)}</TableCell>
                                                <TableCell>{getNodeAddr(instance.node)}</TableCell>
                                                <TableCell className="text-center">
                                                    <NodeRoleBadge node={instance.node} />
                                                </TableCell>
                                            </TableRow>
                                            {error && isErrorExpanded && (
                                                <TableRow>
                                                    <TableCell
                                                        colSpan={TABLE_COLUMN_COUNT}
                                                        className="bg-muted/30 p-4"
                                                    >
                                                        <div className={cn(dashedBorderBox, "p-4")}>
                                                            <p className="mb-3 text-sm font-medium text-foreground">
                                                                Error message
                                                            </p>
                                                            <pre className="whitespace-pre-wrap break-words font-sans text-sm text-foreground">
                                                                {error}
                                                            </pre>
                                                        </div>
                                                    </TableCell>
                                                </TableRow>
                                            )}
                                        </Fragment>
                                    );
                                })
                            )}
                        </TableBody>
                    </Table>
                </div>
            </div>
        </section>
    );
}
