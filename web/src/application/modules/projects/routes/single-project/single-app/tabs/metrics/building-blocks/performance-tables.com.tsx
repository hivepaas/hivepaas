import type { AppDependencyPeer, AppRouteMetricsRoute } from "~/projects/api/services";

import { AppLink } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

import { callKindLabel, formatMs, formatShare } from "./performance-format";

/** The routes the app served, the busiest first. */
export function RoutesTable({ routes }: { routes: AppRouteMetricsRoute[] }) {
    return (
        <div className="overflow-x-auto rounded-md border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead className="w-20">Method</TableHead>
                        <TableHead>Route</TableHead>
                        <TableHead className="text-right">Requests</TableHead>
                        <TableHead className="text-right">Failed</TableHead>
                        <TableHead className="text-right">p50</TableHead>
                        <TableHead className="text-right">p95</TableHead>
                        <TableHead className="text-right">p99</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {routes.map(item => (
                        <TableRow key={`${item.kind} ${item.method} ${item.route}`}>
                            <TableCell className="font-mono text-xs">
                                {item.method || "-"}
                                {item.kind !== "http" && (
                                    <Badge
                                        tone="neutral"
                                        className="ml-1"
                                    >
                                        {callKindLabel(item.kind)}
                                    </Badge>
                                )}
                            </TableCell>
                            <TableCell
                                className="max-w-[28rem] truncate font-mono text-xs"
                                title={item.route}
                            >
                                {item.route || <span className="text-muted-foreground">No route</span>}
                            </TableCell>
                            <TableCell className="text-right tabular-nums">{item.requests.toLocaleString()}</TableCell>
                            <TableCell className="text-right tabular-nums">
                                {formatShare(item.errors, item.requests)}
                            </TableCell>
                            <TableCell className="text-right tabular-nums">{formatMs(item.p50)}</TableCell>
                            <TableCell className="text-right tabular-nums">{formatMs(item.p95)}</TableCell>
                            <TableCell className="text-right tabular-nums">{formatMs(item.p99)}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}

/** A peer's operations, the busiest first: SELECT 1,204 · INSERT 37. */
function PeerOperations({ peer }: { peer: AppDependencyPeer }) {
    const operations = peer.operations
        .map(item => ({ name: [item.method, item.operation].filter(Boolean).join(" "), requests: item.requests }))
        .filter(item => item.name);
    if (operations.length === 0) {
        return null;
    }
    return (
        <span className="text-xs text-muted-foreground">
            {operations.map(item => `${item.name} ${item.requests.toLocaleString()}`).join(" · ")}
        </span>
    );
}

/** The peers the app called, the busiest first: the env's app behind one when it is known. */
export function PeersTable({ peers, projectID, env }: { peers: AppDependencyPeer[]; projectID: string; env: string }) {
    return (
        <div className="overflow-x-auto rounded-md border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead className="w-24">Kind</TableHead>
                        <TableHead>Peer</TableHead>
                        <TableHead className="text-right">Calls</TableHead>
                        <TableHead className="text-right">Failed</TableHead>
                        <TableHead className="text-right">p50</TableHead>
                        <TableHead className="text-right">p95</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {peers.map(item => (
                        <TableRow key={`${item.kind} ${item.peer}`}>
                            <TableCell>
                                <Badge tone={item.kind === "db" ? "violet" : "sky"}>{callKindLabel(item.kind)}</Badge>
                            </TableCell>
                            <TableCell className="max-w-[28rem]">
                                <span className="flex flex-col gap-0.5">
                                    {item.app ? (
                                        <span className="flex flex-wrap items-center gap-x-2">
                                            <AppLink.Basic
                                                to={ROUTE.projects.single.apps.single.metrics.$route(
                                                    projectID,
                                                    env,
                                                    item.app.id,
                                                )}
                                                className="text-link font-medium"
                                            >
                                                {item.app.name || item.app.key}
                                            </AppLink.Basic>
                                            <span
                                                className="truncate font-mono text-xs text-muted-foreground"
                                                title={item.peer}
                                            >
                                                {item.peer}
                                            </span>
                                        </span>
                                    ) : (
                                        <span
                                            className="truncate font-mono text-xs"
                                            title={item.peer}
                                        >
                                            {item.peer || <span className="text-muted-foreground">Unknown</span>}
                                        </span>
                                    )}
                                    <PeerOperations peer={item} />
                                </span>
                            </TableCell>
                            <TableCell className="text-right tabular-nums">{item.requests.toLocaleString()}</TableCell>
                            <TableCell className="text-right tabular-nums">
                                {formatShare(item.errors, item.requests)}
                            </TableCell>
                            <TableCell className="text-right tabular-nums">{formatMs(item.p50)}</TableCell>
                            <TableCell className="text-right tabular-nums">{formatMs(item.p95)}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}
