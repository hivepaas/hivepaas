import { NodesQueries } from "~/cluster/data";
import { ENodeState } from "~/cluster/module-shared/enums";
import { HomeAttentionQueries } from "~/home/data";
import { ProjectsQueries } from "~/projects/data/queries";
import { UsersQueries } from "~/user-management/data/queries";

import { DEFAULT_PAGINATED_DATA, MODULE_IDS, ROUTE } from "@application/shared/constants";
import { EUserRole } from "@application/shared/enums";
import { useConditionalModule, useProjectPermissionsStore } from "@application/shared/permissions";

import { NeedsAttentionCard, NodesCard, RecentTasksCard, SummaryTiles } from "../building-blocks";

function plural(count: number, noun: string) {
    return `${count} ${noun}${count === 1 ? "" : "s"}`;
}

/** The same request as NodesCard's, so the two share one answer. */
const NODES_REQUEST = { pagination: { page: 1, size: 10 } };

/**
 * Where a user lands after signing in: what is running, and what needs them.
 *
 * Every part is shown only to whoever may open the screens it leads to. What
 * needs attention is narrowed by the server; the rest asks the permissions the
 * dashboard already holds.
 */
export function HomeRoute() {
    const { canRead: canReadCluster } = useConditionalModule({ id: MODULE_IDS.Cluster });
    const { canRead: canReadSystem } = useConditionalModule({ id: MODULE_IDS.System });
    const { canRead: canReadProjectModule } = useConditionalModule({ id: MODULE_IDS.Project });
    const { canRead: canReadUsers } = useConditionalModule({ id: MODULE_IDS.User });
    const projectPermissions = useProjectPermissionsStore(state => state.projects);
    const canListProjects = canReadProjectModule || projectPermissions.some(project => project.actions.read);

    const { data: attention, isLoading: attentionLoading } = HomeAttentionQueries.useFindAll();
    const { data: projects = DEFAULT_PAGINATED_DATA } = ProjectsQueries.useFindManyPaginated(
        { pagination: { page: 1, size: 1 } },
        { enabled: canListProjects },
    );
    const { data: nodes = DEFAULT_PAGINATED_DATA } = NodesQueries.useFindManyPaginated(NODES_REQUEST, {
        enabled: canReadCluster,
        refetchInterval: 30_000,
    });

    // Counted, not listed: one of each is enough for the totals.
    const { data: admins = DEFAULT_PAGINATED_DATA } = UsersQueries.useFindManyPaginated(
        { pagination: { page: 1, size: 1 }, role: [EUserRole.Admin] },
        { enabled: canReadUsers },
    );
    const { data: members = DEFAULT_PAGINATED_DATA } = UsersQueries.useFindManyPaginated(
        { pagination: { page: 1, size: 1 }, role: [EUserRole.Member] },
        { enabled: canReadUsers },
    );

    const adminCount = admins.meta.page.total;
    const memberCount = members.meta.page.total;

    const items = attention?.data ?? [];
    const critical = items.filter(item => item.severity === "critical").length;
    const readyNodes = nodes.data.filter(node => node.state === ENodeState.Ready).length;

    const tiles = [
        {
            label: "Needs attention",
            value: attentionLoading ? "–" : String(items.length),
            note: critical > 0 ? `${critical} critical` : "All clear",
            alert: critical > 0,
        },
        ...(canListProjects
            ? [{ label: "Projects", value: String(projects.meta.page.total), to: ROUTE.projects.list.$route }]
            : []),
        ...(canReadCluster
            ? [
                  {
                      label: "Nodes",
                      value: `${readyNodes} / ${nodes.meta.page.total}`,
                      note: readyNodes < nodes.meta.page.total ? "Not all ready" : "All ready",
                      alert: readyNodes < nodes.meta.page.total,
                      to: ROUTE.cluster.nodes.$route,
                  },
              ]
            : []),
        ...(canReadUsers
            ? [
                  {
                      label: "Users",
                      value: String(adminCount + memberCount),
                      note: `${plural(adminCount, "admin")} · ${plural(memberCount, "member")}`,
                      to: ROUTE.userManagement.users.$route,
                  },
              ]
            : []),
    ];

    return (
        <div className="flex flex-col gap-6">
            <header className="flex flex-col gap-1.5">
                <h1 className="text-2xl font-bold tracking-tight">Welcome back</h1>
                <p className="text-sm text-muted-foreground">
                    What is running on your cluster, and what needs you today.
                </p>
            </header>

            <SummaryTiles tiles={tiles} />

            <div
                className={
                    canReadCluster
                        ? "grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,1.55fr)_minmax(0,1fr)]"
                        : "grid grid-cols-1 gap-4"
                }
            >
                <div className="flex min-w-0 flex-col gap-4">
                    <NeedsAttentionCard
                        items={items}
                        isLoading={attentionLoading}
                    />
                    {canReadSystem && <RecentTasksCard />}
                </div>
                {canReadCluster && (
                    <div className="flex min-w-0 flex-col gap-4">
                        <NodesCard />
                    </div>
                )}
            </div>
        </div>
    );
}
