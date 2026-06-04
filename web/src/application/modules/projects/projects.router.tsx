import { ROUTE } from "@/application/shared/constants";
import { Navigate, Outlet, type RouteObject, useParams } from "react-router";
import { ConditionalProjectsAccess } from "~/projects/routes/conditional-projects-access.com";
import { ProjectRouteRedirect } from "~/projects/routes/single-project/project-route-redirect.com";

async function getLazyComponents() {
    return await import("./projects.module");
}

const LEGACY_PROJECT_CONFIGURATION_PATTERNS = {
    accessTokens: "projects/:id/configuration/access-tokens",
    basicAuth: "projects/:id/configuration/basic-auth",
    cloudStorages: "projects/:id/configuration/cloud-storages",
    emailAccounts: "projects/:id/configuration/email-accounts",
    envVariables: "projects/:id/configuration/env-variables",
    githubApps: "projects/:id/configuration/github-apps",
    webhooks: "projects/:id/configuration/webhooks",
    imPlatforms: "projects/:id/configuration/im-platforms",
    notificationTargets: "projects/:id/configuration/notification-targets",
    registryAuth: "projects/:id/configuration/registry-auth",
    secrets: "projects/:id/configuration/secrets",
    sshKeys: "projects/:id/configuration/ssh-keys",
    sslCertificates: "projects/:id/configuration/ssl-certificates",
} as const;

// eslint-disable-next-line react-refresh/only-export-components
function SingleAppRouteRedirect() {
    const { id, appId } = useParams<{ id: string; appId: string }>();

    return (
        <Navigate
            to={
                id && appId
                    ? ROUTE.projects.single.apps.single.configuration.general.$route(id, appId)
                    : ROUTE.projects.list.$route
            }
            replace
        />
    );
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
            element: <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.githubApps.$route} />,
        },
        {
            lazy: async () => {
                const { ProjectsLayout } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalProjectsAccess>
                            <ProjectsLayout>
                                <Outlet />
                            </ProjectsLayout>
                        </ConditionalProjectsAccess>
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
                    path: ROUTE.projects.single.configuration.$pattern,
                    element: <ProjectRouteRedirect to={ROUTE.projects.single.configuration.general.$route} />,
                },
                {
                    path: ROUTE.projects.single.providerConfiguration.$pattern,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.accessTokens.$route} />
                    ),
                },
                {
                    path: ROUTE.projects.single.clusterResources.$pattern,
                    element: <ProjectRouteRedirect to={ROUTE.projects.single.clusterResources.networks.$route} />,
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.accessTokens,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.accessTokens.$route} />
                    ),
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.basicAuth,
                    element: <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.basicAuth.$route} />,
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.cloudStorages,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.cloudStorages.$route} />
                    ),
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.emailAccounts,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.emailAccounts.$route} />
                    ),
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.envVariables,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.envVariables.$route} />
                    ),
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.githubApps,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.githubApps.$route} />
                    ),
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.webhooks,
                    element: <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.webhooks.$route} />,
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.imPlatforms,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.imPlatforms.$route} />
                    ),
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.notificationTargets,
                    element: (
                        <ProjectRouteRedirect
                            to={ROUTE.projects.single.providerConfiguration.notificationTargets.$route}
                        />
                    ),
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.registryAuth,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.registryAuth.$route} />
                    ),
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.secrets,
                    element: <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.secrets.$route} />,
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.sshKeys,
                    element: <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.sshKeys.$route} />,
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.sslCertificates,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.sslCertificates.$route} />
                    ),
                },
                {
                    lazy: async () => {
                        const { ProjectWithSidebar } = await getLazyComponents();

                        return {
                            element: (
                                <ProjectWithSidebar section="configuration">
                                    <Outlet />
                                </ProjectWithSidebar>
                            ),
                        };
                    },
                    children: [
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
                            path: ROUTE.projects.single.configuration.buildSettings.$pattern,
                            lazy: async () => {
                                const { ProjectImageBuildSettingsRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectImageBuildSettingsRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.configuration.storageSettings.$pattern,
                            lazy: async () => {
                                const { ProjectStorageSettingsRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectStorageSettingsRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.configuration.domainSettings.$pattern,
                            lazy: async () => {
                                const { ProjectDomainSettingsRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectDomainSettingsRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.configuration.dangerZone.$pattern,
                            lazy: async () => {
                                const { ProjectDangerZoneRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectDangerZoneRoute,
                                };
                            },
                        },
                    ],
                },
                {
                    lazy: async () => {
                        const { ProjectWithSidebar } = await getLazyComponents();

                        return {
                            element: (
                                <ProjectWithSidebar section="providerConfiguration">
                                    <Outlet />
                                </ProjectWithSidebar>
                            ),
                        };
                    },
                    children: [
                        {
                            path: ROUTE.projects.single.providerConfiguration.accessTokens.$pattern,
                            lazy: async () => {
                                const { ProjectAccessTokensRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectAccessTokensRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.basicAuth.$pattern,
                            lazy: async () => {
                                const { ProjectBasicAuthRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectBasicAuthRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.cloudStorages.$pattern,
                            lazy: async () => {
                                const { ProjectCloudStoragesRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectCloudStoragesRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.emailAccounts.$pattern,
                            lazy: async () => {
                                const { ProjectEmailAccountsRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectEmailAccountsRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.envVariables.$pattern,
                            lazy: async () => {
                                const { ProjectEnvVariablesRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectEnvVariablesRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.githubApps.$pattern,
                            lazy: async () => {
                                const { ProjectGithubAppsRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectGithubAppsRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.webhooks.$pattern,
                            lazy: async () => {
                                const { ProjectWebhooksRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectWebhooksRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.imPlatforms.$pattern,
                            lazy: async () => {
                                const { ProjectImPlatformsRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectImPlatformsRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.notificationTargets.$pattern,
                            lazy: async () => {
                                const { ProjectNotificationTargetsRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectNotificationTargetsRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.registryAuth.$pattern,
                            lazy: async () => {
                                const { ProjectRegistryAuthRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectRegistryAuthRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.secrets.$pattern,
                            lazy: async () => {
                                const { ProjectSecretsRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectSecretsRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.sshKeys.$pattern,
                            lazy: async () => {
                                const { ProjectSSHKeysRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectSSHKeysRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.sslCertificates.$pattern,
                            lazy: async () => {
                                const { ProjectSslCertificatesRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectSslCertificatesRoute,
                                };
                            },
                        },
                    ],
                },
                {
                    lazy: async () => {
                        const { ProjectWithSidebar } = await getLazyComponents();

                        return {
                            element: (
                                <ProjectWithSidebar section="clusterResources">
                                    <Outlet />
                                </ProjectWithSidebar>
                            ),
                        };
                    },
                    children: [
                        {
                            path: ROUTE.projects.single.clusterResources.networks.$pattern,
                            lazy: async () => {
                                const { ProjectNetworksRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectNetworksRoute,
                                };
                            },
                        },
                    ],
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
                    element: <SingleAppRouteRedirect />,
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
                            path: ROUTE.projects.single.apps.single.configuration.healthChecks.$pattern,
                            lazy: async () => {
                                const { AppHealthChecksRoute } = await getLazyComponents();

                                return { Component: AppHealthChecksRoute };
                            },
                        },
                        {
                            path: ROUTE.projects.single.apps.single.configuration.scheduledJobs.$pattern,
                            lazy: async () => {
                                const { AppScheduledJobsRoute } = await getLazyComponents();

                                return { Component: AppScheduledJobsRoute };
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
