import { ROUTE } from "@/application/shared/constants";
import { Outlet, type RouteObject } from "react-router";

async function getLazyComponents() {
    return await import("./projects.module");
}

export const projectsRouter: RouteObject = {
    lazy: async () => {
        const { ProjectsDialogsContainer } = await getLazyComponents();

        return {
            element: (
                <>
                    <Outlet />
                    <ProjectsDialogsContainer />
                </>
            ),
        };
    },
    children: [
        {
            lazy: async () => {
                const { ProjectsLayout } = await getLazyComponents();

                return {
                    element: (
                        <ProjectsLayout>
                            <Outlet />
                        </ProjectsLayout>
                    ),
                };
            },
            children: [
                /**
                 * Projects
                 */
                {
                    path: ROUTE.projects.list.$pattern,
                    lazy: async () => {
                        const { ProjectsRoute } = await getLazyComponents();

                        return {
                            Component: ProjectsRoute,
                        };
                    },
                },
            ],
        },
        {
            lazy: async () => {
                const { SingleProjectLayout } = await getLazyComponents();

                return {
                    element: (
                        <SingleProjectLayout>
                            <Outlet />
                        </SingleProjectLayout>
                    ),
                };
            },
            children: [
                /**
                 * Single Project
                 */
                {
                    path: ROUTE.projects.single.configuration.general.$pattern,
                    lazy: async () => {
                        const { ProjectGeneralRoute } = await getLazyComponents();

                        return {
                            Component: ProjectGeneralRoute,
                        };
                    },
                },
                {
                    path: ROUTE.projects.single.configuration.envVariables.$pattern,
                    lazy: async () => {
                        const { ProjectEnvVariablesRoute } = await getLazyComponents();

                        return {
                            Component: ProjectEnvVariablesRoute,
                        };
                    },
                },
                {
                    path: ROUTE.projects.single.configuration.secrets.$pattern,
                    lazy: async () => {
                        const { ProjectSecretsRoute } = await getLazyComponents();

                        return {
                            Component: ProjectSecretsRoute,
                        };
                    },
                },
                {
                    path: ROUTE.projects.single.apps.$pattern,
                    lazy: async () => {
                        const { ProjectAppsRoute } = await getLazyComponents();

                        return {
                            Component: ProjectAppsRoute,
                        };
                    },
                },
            ],
        },
        {
            lazy: async () => {
                const { SingleAppLayout } = await getLazyComponents();

                return {
                    element: (
                        <SingleAppLayout>
                            <Outlet />
                        </SingleAppLayout>
                    ),
                };
            },
            children: [
                /**
                 * Single App – landing
                 */
                {
                    path: ROUTE.projects.single.apps.single.$pattern,
                    lazy: async () => {
                        const { SingleAppRoute } = await getLazyComponents();

                        return {
                            Component: SingleAppRoute,
                        };
                    },
                },
                /**
                 * Single App – Deployments
                 */
                {
                    path: ROUTE.projects.single.apps.single.deployments.$pattern,
                    lazy: async () => {
                        const { AppDeploymentsRoute } = await getLazyComponents();

                        return {
                            Component: AppDeploymentsRoute,
                        };
                    },
                },
                /**
                 * Single App – Logs
                 */
                {
                    path: ROUTE.projects.single.apps.single.logs.$pattern,
                    lazy: async () => {
                        const { AppLogsRoute } = await getLazyComponents();

                        return {
                            Component: AppLogsRoute,
                        };
                    },
                },
                /**
                 * Single App – Preview Deployments
                 */
                {
                    path: ROUTE.projects.single.apps.single.previewDeployments.$pattern,
                    lazy: async () => {
                        const { AppPreviewDeploymentsRoute } = await getLazyComponents();

                        return {
                            Component: AppPreviewDeploymentsRoute,
                        };
                    },
                },
                /**
                 * Single App – Configuration (with sidebar layout)
                 */
                {
                    lazy: async () => {
                        const { SingleAppConfigurationLayout } = await getLazyComponents();

                        return {
                            element: (
                                <SingleAppConfigurationLayout>
                                    <Outlet />
                                </SingleAppConfigurationLayout>
                            ),
                        };
                    },
                    children: [
                        {
                            path: ROUTE.projects.single.apps.single.configuration.general.$pattern,
                            lazy: async () => {
                                const { AppConfigGeneralRoute } = await getLazyComponents();

                                return { Component: AppConfigGeneralRoute };
                            },
                        },
                        {
                            path: ROUTE.projects.single.apps.single.configuration.deploymentSettings.$pattern,
                            lazy: async () => {
                                const { AppConfigDeploymentSettingsRoute } = await getLazyComponents();

                                return { Component: AppConfigDeploymentSettingsRoute };
                            },
                        },
                        {
                            path: ROUTE.projects.single.apps.single.configuration.containerSettings.$pattern,
                            lazy: async () => {
                                const { AppConfigContainerSettingsRoute } = await getLazyComponents();

                                return { Component: AppConfigContainerSettingsRoute };
                            },
                        },
                        {
                            path: ROUTE.projects.single.apps.single.configuration.httpSettings.$pattern,
                            lazy: async () => {
                                const { AppConfigHttpSettingsRoute } = await getLazyComponents();

                                return { Component: AppConfigHttpSettingsRoute };
                            },
                        },
                        {
                            path: ROUTE.projects.single.apps.single.configuration.envVariables.$pattern,
                            lazy: async () => {
                                const { AppConfigEnvVariablesRoute } = await getLazyComponents();

                                return { Component: AppConfigEnvVariablesRoute };
                            },
                        },
                        {
                            path: ROUTE.projects.single.apps.single.configuration.secrets.$pattern,
                            lazy: async () => {
                                const { AppConfigSecretsRoute } = await getLazyComponents();

                                return { Component: AppConfigSecretsRoute };
                            },
                        },
                        {
                            path: ROUTE.projects.single.apps.single.configuration.availabilityAndScaling.$pattern,
                            lazy: async () => {
                                const { AppConfigAvailabilityRoute } = await getLazyComponents();

                                return { Component: AppConfigAvailabilityRoute };
                            },
                        },
                        {
                            path: ROUTE.projects.single.apps.single.configuration.presistentStorage.$pattern,
                            lazy: async () => {
                                const { AppConfigStorageRoute } = await getLazyComponents();

                                return { Component: AppConfigStorageRoute };
                            },
                        },
                        {
                            path: ROUTE.projects.single.apps.single.configuration.networks.$pattern,
                            lazy: async () => {
                                const { AppConfigNetworksRoute } = await getLazyComponents();

                                return { Component: AppConfigNetworksRoute };
                            },
                        },
                        {
                            path: ROUTE.projects.single.apps.single.configuration.resources.$pattern,
                            lazy: async () => {
                                const { AppConfigResourcesRoute } = await getLazyComponents();

                                return { Component: AppConfigResourcesRoute };
                            },
                        },
                        {
                            path: ROUTE.projects.single.apps.single.configuration.dangerZone.$pattern,
                            lazy: async () => {
                                const { AppConfigDangerZoneRoute } = await getLazyComponents();

                                return { Component: AppConfigDangerZoneRoute };
                            },
                        },
                    ],
                },
            ],
        },
    ],
} as const;
