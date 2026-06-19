import { ROUTE } from "@/application/shared/constants";
import { Navigate, Outlet, type RouteObject, useParams } from "react-router";
import { ConditionalProjectsAccess } from "~/projects/routes/conditional-projects-access.com";
import { ProjectRouteRedirect } from "~/projects/routes/single-project/project-route-redirect.com";

async function getLazyComponents() {
    return await import("./projects.module");
}

const LEGACY_PROJECT_SETTINGS_PATTERNS = {
    root: "projects/:id/configuration",
    general: "projects/:id/configuration/general",
    buildSettings: "projects/:id/configuration/build-settings",
    storageSettings: "projects/:id/configuration/storage-settings",
    domainSettings: "projects/:id/configuration/domain-settings",
    dangerZone: "projects/:id/configuration/danger-zone",
} as const;

const LEGACY_PROJECT_CONFIGURATION_PATTERNS = {
    accessTokens: "projects/:id/configuration/access-tokens",
    acmeDnsProviders: "projects/:id/configuration/acme-dns-providers",
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
    sslProviders: "projects/:id/configuration/ssl-providers",
    sslCertificates: "projects/:id/configuration/ssl-certificates",
} as const;

const LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS = {
    root: "projects/:id/provider-configuration",
    accessTokens: "projects/:id/provider-configuration/access-tokens",
    acmeDnsProviders: "projects/:id/provider-configuration/acme-dns-providers",
    basicAuth: "projects/:id/provider-configuration/basic-auth",
    cloudStorages: "projects/:id/provider-configuration/cloud-storages",
    emailAccounts: "projects/:id/provider-configuration/email-accounts",
    envVariables: "projects/:id/provider-configuration/env-variables",
    githubApps: "projects/:id/provider-configuration/github-apps",
    webhooks: "projects/:id/provider-configuration/webhooks",
    imPlatforms: "projects/:id/provider-configuration/im-platforms",
    notificationTargets: "projects/:id/provider-configuration/notification-targets",
    registryAuth: "projects/:id/provider-configuration/registry-auth",
    secrets: "projects/:id/provider-configuration/secrets",
    sshKeys: "projects/:id/provider-configuration/ssh-keys",
    sslProviders: "projects/:id/provider-configuration/ssl-providers",
    sslCertificates: "projects/:id/provider-configuration/ssl-certificates",
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
                    path: LEGACY_PROJECT_SETTINGS_PATTERNS.root,
                    element: <ProjectRouteRedirect to={ROUTE.projects.single.configuration.general.$route} />,
                },
                {
                    path: LEGACY_PROJECT_SETTINGS_PATTERNS.general,
                    element: <ProjectRouteRedirect to={ROUTE.projects.single.configuration.general.$route} />,
                },
                {
                    path: LEGACY_PROJECT_SETTINGS_PATTERNS.buildSettings,
                    element: <ProjectRouteRedirect to={ROUTE.projects.single.configuration.buildSettings.$route} />,
                },
                {
                    path: LEGACY_PROJECT_SETTINGS_PATTERNS.storageSettings,
                    element: <ProjectRouteRedirect to={ROUTE.projects.single.configuration.storageSettings.$route} />,
                },
                {
                    path: LEGACY_PROJECT_SETTINGS_PATTERNS.domainSettings,
                    element: <ProjectRouteRedirect to={ROUTE.projects.single.configuration.domainSettings.$route} />,
                },
                {
                    path: LEGACY_PROJECT_SETTINGS_PATTERNS.dangerZone,
                    element: <ProjectRouteRedirect to={ROUTE.projects.single.configuration.dangerZone.$route} />,
                },
                {
                    path: ROUTE.projects.single.providerConfiguration.$pattern,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.accessTokens.$route} />
                    ),
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.root,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.accessTokens.$route} />
                    ),
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.accessTokens,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.accessTokens.$route} />
                    ),
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.acmeDnsProviders,
                    element: (
                        <ProjectRouteRedirect
                            to={ROUTE.projects.single.providerConfiguration.acmeDnsProviders.$route}
                        />
                    ),
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.basicAuth,
                    element: <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.basicAuth.$route} />,
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.cloudStorages,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.cloudStorages.$route} />
                    ),
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.emailAccounts,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.emailAccounts.$route} />
                    ),
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.envVariables,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.envVariables.$route} />
                    ),
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.githubApps,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.githubApps.$route} />
                    ),
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.webhooks,
                    element: <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.webhooks.$route} />,
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.imPlatforms,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.imPlatforms.$route} />
                    ),
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.notificationTargets,
                    element: (
                        <ProjectRouteRedirect
                            to={ROUTE.projects.single.providerConfiguration.notificationTargets.$route}
                        />
                    ),
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.registryAuth,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.registryAuth.$route} />
                    ),
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.secrets,
                    element: <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.secrets.$route} />,
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.sshKeys,
                    element: <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.sshKeys.$route} />,
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.sslProviders,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.sslProviders.$route} />
                    ),
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.sslCertificates,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.sslCertificates.$route} />
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
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.acmeDnsProviders,
                    element: (
                        <ProjectRouteRedirect
                            to={ROUTE.projects.single.providerConfiguration.acmeDnsProviders.$route}
                        />
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
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.sslProviders,
                    element: (
                        <ProjectRouteRedirect to={ROUTE.projects.single.providerConfiguration.sslProviders.$route} />
                    ),
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
                            path: ROUTE.projects.single.providerConfiguration.acmeDnsProviders.$pattern,
                            lazy: async () => {
                                const { ProjectAcmeDnsProvidersRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectAcmeDnsProvidersRoute,
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
                            path: ROUTE.projects.single.providerConfiguration.sslProviders.$pattern,
                            lazy: async () => {
                                const { ProjectSslProvidersRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectSslProvidersRoute,
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
                 * Single App – Instances
                 */
                {
                    path: ROUTE.projects.single.apps.single.instances.$pattern,
                    lazy: async () => {
                        const { AppInstancesRoute } = await getLazyComponents();

                        return {
                            Component: AppInstancesRoute,
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
                {
                    path: ROUTE.projects.single.apps.single.deployments.details.$pattern,
                    lazy: async () => {
                        const { AppDeploymentDetailsRoute } = await getLazyComponents();

                        return {
                            Component: AppDeploymentDetailsRoute,
                        };
                    },
                },
                /**
                 * Single App – Scheduled Job Tasks
                 */
                {
                    path: ROUTE.projects.single.apps.single.scheduledJobTasks.$pattern,
                    lazy: async () => {
                        const { AppScheduledJobTasksRoute } = await getLazyComponents();

                        return {
                            Component: AppScheduledJobTasksRoute,
                        };
                    },
                },
                {
                    path: ROUTE.projects.single.apps.single.scheduledJobTasks.details.$pattern,
                    lazy: async () => {
                        const { AppScheduledJobTaskDetailsRoute } = await getLazyComponents();

                        return {
                            Component: AppScheduledJobTaskDetailsRoute,
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
                 * Single App – Terminal
                 */
                {
                    path: ROUTE.projects.single.apps.single.terminal.$pattern,
                    lazy: async () => {
                        const { AppTerminalRoute } = await getLazyComponents();

                        return {
                            Component: AppTerminalRoute,
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
                            path: ROUTE.projects.single.apps.single.configuration.featureSettings.$pattern,
                            lazy: async () => {
                                const { AppFeatureSettingsRoute } = await getLazyComponents();

                                return { Component: AppFeatureSettingsRoute };
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
