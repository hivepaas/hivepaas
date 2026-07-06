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
                const { NetworksLayout } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.Cluster}>
                            <NetworksLayout>
                                <Outlet />
                            </NetworksLayout>
                        </ConditionalModule>
                    ),
                };
            },
            children: [
                /**
                 * Networks
                 */
                {
                    path: ROUTE.cluster.networks.$pattern,
                    lazy: async () => {
                        const { NetworksRoute } = await getLazyComponents();

                        return {
                            Component: NetworksRoute,
                        };
                    },
                },
                {
                    path: ROUTE.cluster.networks.create.$pattern,
                    lazy: async () => {
                        const { NetworkCreateRoute } = await getLazyComponents();

                        return {
                            Component: NetworkCreateRoute,
                        };
                    },
                },
                {
                    path: ROUTE.cluster.networks.details.$pattern,
                    lazy: async () => {
                        const { NetworkDetailsRoute } = await getLazyComponents();

                        return {
                            Component: NetworkDetailsRoute,
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
        {
            lazy: async () => {
                const { VolumesLayout } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.Cluster}>
                            <VolumesLayout>
                                <Outlet />
                            </VolumesLayout>
                        </ConditionalModule>
                    ),
                };
            },
            children: [
                {
                    path: ROUTE.cluster.volumes.$pattern,
                    lazy: async () => {
                        const { VolumesRoute } = await getLazyComponents();

                        return {
                            Component: VolumesRoute,
                        };
                    },
                },
                {
                    path: ROUTE.cluster.volumes.create.$pattern,
                    lazy: async () => {
                        const { VolumeCreateRoute } = await getLazyComponents();

                        return {
                            Component: VolumeCreateRoute,
                        };
                    },
                },
                {
                    path: ROUTE.cluster.volumes.edit.$pattern,
                    lazy: async () => {
                        const { VolumeEditRoute } = await getLazyComponents();

                        return {
                            Component: VolumeEditRoute,
                        };
                    },
                },
            ],
        },
    ],
} as const;
