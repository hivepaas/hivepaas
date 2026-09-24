import { Link } from "react-router";
import { NodesQueries } from "~/cluster/data";
import { ENodeState } from "~/cluster/module-shared/enums";

import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

const NODE_LIMIT = 10;

/** The nodes, and whether each is up. Shown to whoever may read the cluster screens. */
export function NodesCard() {
    const { data: { data: nodes, meta } = DEFAULT_PAGINATED_DATA, isLoading } = NodesQueries.useFindManyPaginated(
        { pagination: { page: 1, size: NODE_LIMIT } },
        { refetchInterval: 30_000 },
    );

    return (
        <Card className="gap-0 py-0">
            <CardHeader className="flex flex-row items-center justify-between border-b px-5 py-4 [.border-b]:pb-4">
                <CardTitle className="text-[15px]">Nodes</CardTitle>
                <Link
                    to={ROUTE.cluster.nodes.$route}
                    className="text-[13px] font-medium underline-offset-4 hover:underline"
                >
                    {meta.page.total > nodes.length ? `All ${meta.page.total}` : "Manage"}
                </Link>
            </CardHeader>
            <CardContent className="px-0">
                {isLoading ? (
                    <div className="p-5">
                        <Skeleton className="h-6 w-full" />
                    </div>
                ) : (
                    <ul>
                        {nodes.map(node => {
                            const ready = node.state === ENodeState.Ready;
                            return (
                                <li
                                    key={node.id}
                                    className="border-b border-border/60 last:border-b-0"
                                >
                                    <Link
                                        to={ROUTE.cluster.nodes.single.$route(node.id)}
                                        className="flex items-center gap-3 px-5 py-3 text-[13px] hover:bg-muted/40"
                                    >
                                        <span
                                            className={
                                                ready
                                                    ? "size-2 rounded-full bg-green-600"
                                                    : "size-2 rounded-sm bg-destructive"
                                            }
                                            aria-hidden
                                        />
                                        <span className="grow truncate font-semibold">{node.hostname}</span>
                                        <Badge
                                            variant="secondary"
                                            className="capitalize"
                                        >
                                            {node.role}
                                            {node.isLeader ? " · leader" : ""}
                                        </Badge>
                                        {node.resources && (
                                            <span className="text-muted-foreground">
                                                {node.resources.cpus} CPU · {node.resources.memory}
                                            </span>
                                        )}
                                        <span
                                            className={
                                                ready ? "text-muted-foreground" : "font-semibold text-destructive"
                                            }
                                        >
                                            {ready ? "Ready" : node.state}
                                        </span>
                                    </Link>
                                </li>
                            );
                        })}
                    </ul>
                )}
            </CardContent>
        </Card>
    );
}
