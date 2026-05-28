import { MODULE_IDS, ROUTE } from "@/application/shared/constants";
import { Outlet, type RouteObject } from "react-router";

import { ConditionalModule } from "@application/shared/permissions";

async function getLazyComponents() {
    return await import("./cluster.module");
}

export const clusterRouter: RouteObject = {
    lazy: async () => {
        const { ClusterDialogsContainer } = await getLazyComponents();

        return {
            element: (
                <>
                    <Outlet />
                    <ClusterDialogsContainer />
                </>
            ),
        };
    },
    children: [
        {
            lazy: async () => {
                const { NodesLayout } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.Cluster}>
                            <NodesLayout>
                                <Outlet />
                            </NodesLayout>
                        </ConditionalModule>
                    ),
                };
            },
            children: [
                /**
                 * Nodes
                 */
                {
                    path: ROUTE.cluster.nodes.$pattern,
                    lazy: async () => {
                        const { NodesRoute } = await getLazyComponents();

                        return {
                            Component: NodesRoute,
                        };
                    },
                },
            ],
        },
        {
            lazy: async () => {
                const { SingleNodeLayout } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.Cluster}>
                            <SingleNodeLayout>
                                <Outlet />
                            </SingleNodeLayout>
                        </ConditionalModule>
                    ),
                };
            },
            children: [
                /**
                 * Single Node
                 */
                {
                    path: ROUTE.cluster.nodes.single.$pattern,
                    lazy: async () => {
                        const { SingleNodeRoute } = await getLazyComponents();

                        return {
                            Component: SingleNodeRoute,
                        };
                    },
                },
            ],
        },
    ],
} as const;
