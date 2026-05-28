import { MODULE_IDS, ROUTE } from "@/application/shared/constants";
import { Outlet, type RouteObject } from "react-router";
import { ProjectGithubAppsLegacyRedirect } from "~/projects/routes/single-project/configuration/github-apps/route/project-github-apps-legacy-redirect.com";

import { ConditionalModule } from "@application/shared/permissions";

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
            path: "projects/:id/github-apps",
            element: (
                <ConditionalModule id={MODULE_IDS.Project}>
                    <ProjectGithubAppsLegacyRedirect />
                </ConditionalModule>
            ),
        },
        {
            lazy: async () => {
                const { ProjectsLayout } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.Project}>
                            <ProjectsLayout>
                                <Outlet />
                            </ProjectsLayout>
                        </ConditionalModule>
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
                        <ConditionalModule id={MODULE_IDS.Project}>
                            <SingleProjectLayout>
                                <Outlet />
                            </SingleProjectLayout>
                        </ConditionalModule>
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
                    path: ROUTE.projects.single.configuration.basicAuth.$pattern,
                    lazy: async () => {
                        const { ProjectBasicAuthRoute } = await getLazyComponents();

                        return {
                            Component: ProjectBasicAuthRoute,
                        };
                    },
                },
                {
                    path: ROUTE.projects.single.configuration.githubApps.$pattern,
                    lazy: async () => {
                        const { ProjectGithubAppsRoute } = await getLazyComponents();

                        return {
                            Component: ProjectGithubAppsRoute,
                        };
                    },
                },
                {
                    path: ROUTE.projects.single.configuration.registryAuth.$pattern,
                    lazy: async () => {
                        const { ProjectRegistryAuthRoute } = await getLazyComponents();

                        return {
                            Component: ProjectRegistryAuthRoute,
                        };
                    },
                },
                {
                    path: ROUTE.projects.single.configuration.sslCertificates.$pattern,
                    lazy: async () => {
                        const { ProjectSslCertificatesRoute } = await getLazyComponents();

                        return {
                            Component: ProjectSslCertificatesRoute,
                        };
                    },
                },
                {
                    path: ROUTE.projects.single.configuration.emailAccounts.$pattern,
                    lazy: async () => {
                        const { ProjectEmailAccountsRoute } = await getLazyComponents();

                        return {
                            Component: ProjectEmailAccountsRoute,
                        };
                    },
                },
                {
                    path: ROUTE.projects.single.configuration.imPlatforms.$pattern,
                    lazy: async () => {
                        const { ProjectImPlatformsRoute } = await getLazyComponents();

                        return {
                            Component: ProjectImPlatformsRoute,
                        };
                    },
                },
                {
                    path: ROUTE.projects.single.configuration.sshKeys.$pattern,
                    lazy: async () => {
                        const { ProjectSSHKeysRoute } = await getLazyComponents();

                        return {
                            Component: ProjectSSHKeysRoute,
                        };
                    },
                },
                {
                    path: ROUTE.projects.single.configuration.accessTokens.$pattern,
                    lazy: async () => {
                        const { ProjectAccessTokensRoute } = await getLazyComponents();

                        return {
                            Component: ProjectAccessTokensRoute,
                        };
                    },
                },
                {
                    path: ROUTE.projects.single.configuration.cloudStorages.$pattern,
                    lazy: async () => {
                        const { ProjectCloudStoragesRoute } = await getLazyComponents();

                        return {
                            Component: ProjectCloudStoragesRoute,
                        };
                    },
                },
                {
                    path: ROUTE.projects.single.configuration.notificationTargets.$pattern,
                    lazy: async () => {
                        const { ProjectNotificationTargetsRoute } = await getLazyComponents();

                        return {
                            Component: ProjectNotificationTargetsRoute,
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
                        <ConditionalModule id={MODULE_IDS.Project}>
                            <SingleAppLayout>
                                <Outlet />
                            </SingleAppLayout>
                        </ConditionalModule>
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
                            path: ROUTE.projects.single.apps.single.configuration.configFiles.$pattern,
                            lazy: async () => {
                                const { AppConfigFilesRoute } = await getLazyComponents();

                                return { Component: AppConfigFilesRoute };
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
