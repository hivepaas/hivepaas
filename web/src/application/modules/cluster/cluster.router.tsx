import { ROUTE } from "@/application/shared/constants";
import { Outlet, type RouteObject } from "react-router";

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
                        <NodesLayout>
                            <Outlet />
                        </NodesLayout>
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
                        <SingleNodeLayout>
                            <Outlet />
                        </SingleNodeLayout>
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
