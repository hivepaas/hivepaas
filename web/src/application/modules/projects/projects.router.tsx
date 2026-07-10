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
    githubAppCreate: "projects/:id/configuration/github-apps/create",
    githubAppEdit: "projects/:id/configuration/github-apps/:githubAppId/edit",
    webhooks: "projects/:id/configuration/webhooks",
    webhookCreate: "projects/:id/configuration/webhooks/create",
    webhookEdit: "projects/:id/configuration/webhooks/:repoWebhookId/edit",
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
    githubAppCreate: "projects/:id/provider-configuration/github-apps/create",
    githubAppEdit: "projects/:id/provider-configuration/github-apps/:githubAppId/edit",
    webhooks: "projects/:id/provider-configuration/webhooks",
    webhookCreate: "projects/:id/provider-configuration/webhooks/create",
    webhookEdit: "projects/:id/provider-configuration/webhooks/:repoWebhookId/edit",
    imPlatforms: "projects/:id/provider-configuration/im-platforms",
    notificationTargets: "projects/:id/provider-configuration/notification-targets",
    registryAuth: "projects/:id/provider-configuration/registry-auth",
    secrets: "projects/:id/provider-configuration/secrets",
    sshKeys: "projects/:id/provider-configuration/ssh-keys",
    sslProviders: "projects/:id/provider-configuration/ssl-providers",
    sslCertificates: "projects/:id/provider-configuration/ssl-certificates",
} as const;

const LEGACY_PROJECT_PROVIDER_SETTINGS_SOURCE_PATTERNS = {
    githubApps: "projects/:id/provider-settings/github-apps",
    githubAppCreate: "projects/:id/provider-settings/github-apps/create",
    githubAppEdit: "projects/:id/provider-settings/github-apps/:githubAppId/edit",
    webhooks: "projects/:id/provider-settings/webhooks",
    webhookCreate: "projects/:id/provider-settings/webhooks/create",
    webhookEdit: "projects/:id/provider-settings/webhooks/:repoWebhookId/edit",
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

// eslint-disable-next-line react-refresh/only-export-components
function ProjectGithubAppSourcesRouteRedirect({ target }: ProjectGithubAppSourcesRouteRedirectProps) {
    const { id, githubAppId } = useParams<{ id: string; githubAppId: string }>();

    if (!id) {
        return (
            <Navigate
                to={ROUTE.projects.list.$route}
                replace
            />
        );
    }

    const to =
        target === "create"
            ? ROUTE.projects.single.sources.githubApps.create.$route(id)
            : target === "edit" && githubAppId
              ? ROUTE.projects.single.sources.githubApps.edit.$route(id, githubAppId)
              : ROUTE.projects.single.sources.githubApps.$route(id);

    return (
        <Navigate
            to={to}
            replace
        />
    );
}

// eslint-disable-next-line react-refresh/only-export-components
function ProjectWebhookSourcesRouteRedirect({ target }: ProjectWebhookSourcesRouteRedirectProps) {
    const { id, repoWebhookId } = useParams<{ id: string; repoWebhookId: string }>();

    if (!id) {
        return (
            <Navigate
                to={ROUTE.projects.list.$route}
                replace
            />
        );
    }

    const to =
        target === "create"
            ? ROUTE.projects.single.sources.webhooks.create.$route(id)
            : target === "edit" && repoWebhookId
              ? ROUTE.projects.single.sources.webhooks.edit.$route(id, repoWebhookId)
              : ROUTE.projects.single.sources.webhooks.$route(id);

    return (
        <Navigate
            to={to}
            replace
        />
    );
}

interface ProjectGithubAppSourcesRouteRedirectProps {
    target: "list" | "create" | "edit";
}

interface ProjectWebhookSourcesRouteRedirectProps {
    target: "list" | "create" | "edit";
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
            element: <ProjectRouteRedirect to={ROUTE.projects.single.sources.githubApps.$route} />,
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
                    path: ROUTE.projects.single.sources.$pattern,
                    element: <ProjectRouteRedirect to={ROUTE.projects.single.sources.githubApps.$route} />,
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
                    element: <ProjectGithubAppSourcesRouteRedirect target="list" />,
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.githubAppCreate,
                    element: <ProjectGithubAppSourcesRouteRedirect target="create" />,
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.githubAppEdit,
                    element: <ProjectGithubAppSourcesRouteRedirect target="edit" />,
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.webhooks,
                    element: <ProjectWebhookSourcesRouteRedirect target="list" />,
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.webhookCreate,
                    element: <ProjectWebhookSourcesRouteRedirect target="create" />,
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_CONFIGURATION_PATTERNS.webhookEdit,
                    element: <ProjectWebhookSourcesRouteRedirect target="edit" />,
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_SETTINGS_SOURCE_PATTERNS.githubApps,
                    element: <ProjectGithubAppSourcesRouteRedirect target="list" />,
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_SETTINGS_SOURCE_PATTERNS.githubAppCreate,
                    element: <ProjectGithubAppSourcesRouteRedirect target="create" />,
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_SETTINGS_SOURCE_PATTERNS.githubAppEdit,
                    element: <ProjectGithubAppSourcesRouteRedirect target="edit" />,
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_SETTINGS_SOURCE_PATTERNS.webhooks,
                    element: <ProjectWebhookSourcesRouteRedirect target="list" />,
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_SETTINGS_SOURCE_PATTERNS.webhookCreate,
                    element: <ProjectWebhookSourcesRouteRedirect target="create" />,
                },
                {
                    path: LEGACY_PROJECT_PROVIDER_SETTINGS_SOURCE_PATTERNS.webhookEdit,
                    element: <ProjectWebhookSourcesRouteRedirect target="edit" />,
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
                    element: <ProjectGithubAppSourcesRouteRedirect target="list" />,
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.githubAppCreate,
                    element: <ProjectGithubAppSourcesRouteRedirect target="create" />,
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.githubAppEdit,
                    element: <ProjectGithubAppSourcesRouteRedirect target="edit" />,
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.webhooks,
                    element: <ProjectWebhookSourcesRouteRedirect target="list" />,
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.webhookCreate,
                    element: <ProjectWebhookSourcesRouteRedirect target="create" />,
                },
                {
                    path: LEGACY_PROJECT_CONFIGURATION_PATTERNS.webhookEdit,
                    element: <ProjectWebhookSourcesRouteRedirect target="edit" />,
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
                            path: ROUTE.projects.single.providerConfiguration.accessTokens.create.$pattern,
                            lazy: async () => {
                                const { ProjectAccessTokenCreateRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectAccessTokenCreateRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.accessTokens.edit.$pattern,
                            lazy: async () => {
                                const { ProjectAccessTokenEditRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectAccessTokenEditRoute,
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
                            path: ROUTE.projects.single.providerConfiguration.acmeDnsProviders.create.$pattern,
                            lazy: async () => {
                                const { ProjectAcmeDnsProviderCreateRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectAcmeDnsProviderCreateRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.acmeDnsProviders.edit.$pattern,
                            lazy: async () => {
                                const { ProjectAcmeDnsProviderEditRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectAcmeDnsProviderEditRoute,
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
                            path: ROUTE.projects.single.providerConfiguration.basicAuth.create.$pattern,
                            lazy: async () => {
                                const { ProjectBasicAuthCreateRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectBasicAuthCreateRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.basicAuth.edit.$pattern,
                            lazy: async () => {
                                const { ProjectBasicAuthEditRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectBasicAuthEditRoute,
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
                            path: ROUTE.projects.single.providerConfiguration.cloudStorages.create.$pattern,
                            lazy: async () => {
                                const { ProjectCloudStorageCreateRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectCloudStorageCreateRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.cloudStorages.edit.$pattern,
                            lazy: async () => {
                                const { ProjectCloudStorageEditRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectCloudStorageEditRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.commandTemplates.$pattern,
                            lazy: async () => {
                                const { ProjectCommandTemplatesRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectCommandTemplatesRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.commandTemplates.create.$pattern,
                            lazy: async () => {
                                const { ProjectCommandTemplateCreateRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectCommandTemplateCreateRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.commandTemplates.edit.$pattern,
                            lazy: async () => {
                                const { ProjectCommandTemplateEditRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectCommandTemplateEditRoute,
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
                            path: ROUTE.projects.single.providerConfiguration.emailAccounts.create.$pattern,
                            lazy: async () => {
                                const { ProjectEmailAccountCreateRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectEmailAccountCreateRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.emailAccounts.edit.$pattern,
                            lazy: async () => {
                                const { ProjectEmailAccountEditRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectEmailAccountEditRoute,
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
                            path: ROUTE.projects.single.providerConfiguration.imPlatforms.$pattern,
                            lazy: async () => {
                                const { ProjectImPlatformsRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectImPlatformsRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.imPlatforms.create.$pattern,
                            lazy: async () => {
                                const { ProjectImPlatformCreateRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectImPlatformCreateRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.imPlatforms.edit.$pattern,
                            lazy: async () => {
                                const { ProjectImPlatformEditRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectImPlatformEditRoute,
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
                            path: ROUTE.projects.single.providerConfiguration.notificationTargets.create.$pattern,
                            lazy: async () => {
                                const { ProjectNotificationTargetCreateRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectNotificationTargetCreateRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.notificationTargets.edit.$pattern,
                            lazy: async () => {
                                const { ProjectNotificationTargetEditRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectNotificationTargetEditRoute,
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
                            path: ROUTE.projects.single.providerConfiguration.registryAuth.create.$pattern,
                            lazy: async () => {
                                const { ProjectRegistryAuthCreateRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectRegistryAuthCreateRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.registryAuth.edit.$pattern,
                            lazy: async () => {
                                const { ProjectRegistryAuthEditRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectRegistryAuthEditRoute,
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
                            path: ROUTE.projects.single.providerConfiguration.secrets.create.$pattern,
                            lazy: async () => {
                                const { ProjectSecretCreateRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectSecretCreateRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.secrets.edit.$pattern,
                            lazy: async () => {
                                const { ProjectSecretEditRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectSecretEditRoute,
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
                            path: ROUTE.projects.single.providerConfiguration.sshKeys.create.$pattern,
                            lazy: async () => {
                                const { ProjectSSHKeyCreateRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectSSHKeyCreateRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.sshKeys.edit.$pattern,
                            lazy: async () => {
                                const { ProjectSSHKeyEditRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectSSHKeyEditRoute,
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
                            path: ROUTE.projects.single.providerConfiguration.sslProviders.create.$pattern,
                            lazy: async () => {
                                const { ProjectSslProviderCreateRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectSslProviderCreateRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.sslProviders.edit.$pattern,
                            lazy: async () => {
                                const { ProjectSslProviderEditRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectSslProviderEditRoute,
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
                        {
                            path: ROUTE.projects.single.providerConfiguration.sslCertificates.create.$pattern,
                            lazy: async () => {
                                const { ProjectSslCertCreateRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectSslCertCreateRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.providerConfiguration.sslCertificates.edit.$pattern,
                            lazy: async () => {
                                const { ProjectSslCertEditRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectSslCertEditRoute,
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
                                <ProjectWithSidebar section="sources">
                                    <Outlet />
                                </ProjectWithSidebar>
                            ),
                        };
                    },
                    children: [
                        {
                            path: ROUTE.projects.single.sources.githubApps.$pattern,
                            lazy: async () => {
                                const { ProjectGithubAppsRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectGithubAppsRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.sources.githubApps.create.$pattern,
                            lazy: async () => {
                                const { ProjectGithubAppCreateRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectGithubAppCreateRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.sources.githubApps.edit.$pattern,
                            lazy: async () => {
                                const { ProjectGithubAppEditRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectGithubAppEditRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.sources.webhooks.$pattern,
                            lazy: async () => {
                                const { ProjectWebhooksRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectWebhooksRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.sources.webhooks.create.$pattern,
                            lazy: async () => {
                                const { ProjectWebhookCreateRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectWebhookCreateRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.sources.webhooks.edit.$pattern,
                            lazy: async () => {
                                const { ProjectWebhookEditRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectWebhookEditRoute,
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
                        {
                            path: ROUTE.projects.single.clusterResources.networks.create.$pattern,
                            lazy: async () => {
                                const { ProjectNetworkCreateRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectNetworkCreateRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.clusterResources.networks.details.$pattern,
                            lazy: async () => {
                                const { ProjectNetworkDetailsRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectNetworkDetailsRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.clusterResources.volumes.$pattern,
                            lazy: async () => {
                                const { ProjectVolumesRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectVolumesRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.clusterResources.volumes.create.$pattern,
                            lazy: async () => {
                                const { ProjectVolumeCreateRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectVolumeCreateRoute,
                                };
                            },
                        },
                        {
                            path: ROUTE.projects.single.clusterResources.volumes.edit.$pattern,
                            lazy: async () => {
                                const { ProjectVolumeEditRoute } = await getLazyComponents();

                                return {
                                    Component: ProjectVolumeEditRoute,
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
                    path: ROUTE.projects.single.apps.single.previewDeployments.create.$pattern,
                    lazy: async () => {
                        const { AppPreviewDeploymentCreateRoute } = await getLazyComponents();

                        return {
                            Component: AppPreviewDeploymentCreateRoute,
                        };
                    },
                },
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
                            path: ROUTE.projects.single.apps.single.configuration.healthChecks.create.$pattern,
                            lazy: async () => {
                                const { AppHealthCheckCreateRoute } = await getLazyComponents();

                                return { Component: AppHealthCheckCreateRoute };
                            },
                        },
                        {
                            path: ROUTE.projects.single.apps.single.configuration.healthChecks.edit.$pattern,
                            lazy: async () => {
                                const { AppHealthCheckEditRoute } = await getLazyComponents();

                                return { Component: AppHealthCheckEditRoute };
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
                            path: ROUTE.projects.single.apps.single.configuration.scheduledJobs.create.$pattern,
                            lazy: async () => {
                                const { AppScheduledJobCreateRoute } = await getLazyComponents();

                                return { Component: AppScheduledJobCreateRoute };
                            },
                        },
                        {
                            path: ROUTE.projects.single.apps.single.configuration.scheduledJobs.edit.$pattern,
                            lazy: async () => {
                                const { AppScheduledJobEditRoute } = await getLazyComponents();

                                return { Component: AppScheduledJobEditRoute };
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
                            path: ROUTE.projects.single.apps.single.configuration.secrets.create.$pattern,
                            lazy: async () => {
                                const { AppConfigSecretCreateRoute } = await getLazyComponents();

                                return { Component: AppConfigSecretCreateRoute };
                            },
                        },
                        {
                            path: ROUTE.projects.single.apps.single.configuration.secrets.edit.$pattern,
                            lazy: async () => {
                                const { AppConfigSecretEditRoute } = await getLazyComponents();

                                return { Component: AppConfigSecretEditRoute };
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
                            path: ROUTE.projects.single.apps.single.configuration.configFiles.create.$pattern,
                            lazy: async () => {
                                const { AppConfigFileCreateRoute } = await getLazyComponents();

                                return { Component: AppConfigFileCreateRoute };
                            },
                        },
                        {
                            path: ROUTE.projects.single.apps.single.configuration.configFiles.edit.$pattern,
                            lazy: async () => {
                                const { AppConfigFileEditRoute } = await getLazyComponents();

                                return { Component: AppConfigFileEditRoute };
                            },
                        },
                        {
                            path: ROUTE.projects.single.apps.single.configuration.dataFiles.$pattern,
                            lazy: async () => {
                                const { AppDataFilesRoute } = await getLazyComponents();

                                return { Component: AppDataFilesRoute };
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
                            children: [
                                {
                                    index: true,
                                    lazy: async () => {
                                        const { AppConfigStorageListRoute } = await getLazyComponents();

                                        return { Component: AppConfigStorageListRoute };
                                    },
                                },
                                {
                                    path: "create",
                                    lazy: async () => {
                                        const { StorageMountCreateRoute } = await getLazyComponents();

                                        return { Component: StorageMountCreateRoute };
                                    },
                                },
                                {
                                    path: ":mountId/edit",
                                    lazy: async () => {
                                        const { StorageMountEditRoute } = await getLazyComponents();

                                        return { Component: StorageMountEditRoute };
                                    },
                                },
                            ],
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
