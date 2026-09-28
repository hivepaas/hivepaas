import type { ComponentType } from "react";

import { MODULE_IDS, ROUTE } from "@/application/shared/constants";
import { Navigate, Outlet, type RouteObject, useParams } from "react-router";

import { AppNavigate } from "@application/shared/components";
import { ModuleTitle } from "@application/shared/components/module-title";
import { ConditionalModule } from "@application/shared/permissions";

async function getLazyComponents() {
    return await import("./settings.module");
}

function createSettingsRoute(path: string, title: string, loadComponent: () => Promise<ComponentType>): RouteObject {
    return {
        path,
        element: (
            <ConditionalModule id={MODULE_IDS.Settings}>
                <ModuleTitle title={title}>
                    <Outlet />
                </ModuleTitle>
            </ConditionalModule>
        ),
        children: [
            {
                index: true,
                lazy: async () => {
                    const Component = await loadComponent();

                    return {
                        Component,
                    };
                },
            },
        ],
    };
}

function createSettingsModuleRoute(
    path: string,
    title: string,
    loadComponent: () => Promise<ComponentType>,
): RouteObject {
    return {
        path,
        element: (
            <ConditionalModule id={MODULE_IDS.Settings}>
                <ModuleTitle title={title}>
                    <Outlet />
                </ModuleTitle>
            </ConditionalModule>
        ),
        children: [
            {
                index: true,
                lazy: async () => {
                    const Component = await loadComponent();

                    return {
                        Component,
                    };
                },
            },
        ],
    };
}

// eslint-disable-next-line react-refresh/only-export-components
function LegacySettingsRouteRedirect() {
    const { "*": splat } = useParams<{ "*": string }>();

    if (!splat) {
        return (
            <Navigate
                to="/integrations/"
                replace
            />
        );
    }

    const target = `/integrations/${splat}`;
    const normalizedTarget = target.endsWith("/") ? target : `${target}/`;

    return (
        <Navigate
            to={normalizedTarget}
            replace
        />
    );
}

export const settingsRouter: RouteObject = {
    lazy: async () => {
        const { SettingsDialogsContainer } = await getLazyComponents();

        return {
            element: (
                <>
                    <Outlet />
                    <SettingsDialogsContainer />
                </>
            ),
        };
    },
    children: [
        {
            path: ROUTE.settings.$pattern,
            element: (
                <ConditionalModule id={MODULE_IDS.Settings}>
                    <Navigate
                        to={ROUTE.settings.githubApps.$route}
                        replace
                    />
                </ConditionalModule>
            ),
        },
        {
            path: "settings/github-apps",
            element: (
                <ConditionalModule id={MODULE_IDS.Settings}>
                    <Navigate
                        to={ROUTE.settings.githubApps.$route}
                        replace
                    />
                </ConditionalModule>
            ),
        },
        createSettingsRoute(ROUTE.settings.githubApps.$pattern, "Github Apps", async () => {
            const { SettingsGithubAppsRoute } = await getLazyComponents();

            return SettingsGithubAppsRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.githubApps.create.$pattern, "Github Apps", async () => {
            const { SettingsGithubAppCreateRoute } = await getLazyComponents();

            return SettingsGithubAppCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.githubApps.edit.$pattern, "Github Apps", async () => {
            const { SettingsGithubAppEditRoute } = await getLazyComponents();

            return SettingsGithubAppEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.webhooks.$pattern, "Webhooks", async () => {
            const { SettingsWebhooksRoute } = await getLazyComponents();

            return SettingsWebhooksRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.webhooks.create.$pattern, "Webhooks", async () => {
            const { SettingsWebhookCreateRoute } = await getLazyComponents();

            return SettingsWebhookCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.webhooks.edit.$pattern, "Webhooks", async () => {
            const { SettingsWebhookEditRoute } = await getLazyComponents();

            return SettingsWebhookEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.basicAuth.$pattern, "Basic Auth", async () => {
            const { SettingsBasicAuthRoute } = await getLazyComponents();

            return SettingsBasicAuthRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.basicAuth.create.$pattern, "Basic Auth", async () => {
            const { SettingsBasicAuthCreateRoute } = await getLazyComponents();

            return SettingsBasicAuthCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.basicAuth.edit.$pattern, "Basic Auth", async () => {
            const { SettingsBasicAuthEditRoute } = await getLazyComponents();

            return SettingsBasicAuthEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.keyAuth.$pattern, "Key Auth", async () => {
            const { SettingsKeyAuthRoute } = await getLazyComponents();

            return SettingsKeyAuthRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.keyAuth.create.$pattern, "Key Auth", async () => {
            const { SettingsKeyAuthCreateRoute } = await getLazyComponents();

            return SettingsKeyAuthCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.keyAuth.edit.$pattern, "Key Auth", async () => {
            const { SettingsKeyAuthEditRoute } = await getLazyComponents();

            return SettingsKeyAuthEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.registryAuth.$pattern, "Registry Auth", async () => {
            const { SettingsRegistryAuthRoute } = await getLazyComponents();

            return SettingsRegistryAuthRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.registryAuth.create.$pattern, "Registry Auth", async () => {
            const { SettingsRegistryAuthCreateRoute } = await getLazyComponents();

            return SettingsRegistryAuthCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.registryAuth.edit.$pattern, "Registry Auth", async () => {
            const { SettingsRegistryAuthEditRoute } = await getLazyComponents();

            return SettingsRegistryAuthEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.sslProviders.$pattern, "SSL Providers", async () => {
            const { SettingsSslProvidersRoute } = await getLazyComponents();

            return SettingsSslProvidersRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.sslProviders.create.$pattern, "SSL Providers", async () => {
            const { SettingsSslProviderCreateRoute } = await getLazyComponents();

            return SettingsSslProviderCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.sslProviders.edit.$pattern, "SSL Providers", async () => {
            const { SettingsSslProviderEditRoute } = await getLazyComponents();

            return SettingsSslProviderEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.sslCertificates.$pattern, "SSL Certificates", async () => {
            const { SettingsSslCertificatesRoute } = await getLazyComponents();

            return SettingsSslCertificatesRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.sslCertificates.create.$pattern, "SSL Certificates", async () => {
            const { SettingsSslCertCreateRoute } = await getLazyComponents();

            return SettingsSslCertCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.sslCertificates.edit.$pattern, "SSL Certificates", async () => {
            const { SettingsSslCertEditRoute } = await getLazyComponents();

            return SettingsSslCertEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.emailAccounts.$pattern, "Email Accounts", async () => {
            const { SettingsEmailAccountsRoute } = await getLazyComponents();

            return SettingsEmailAccountsRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.emailAccounts.create.$pattern, "Email Accounts", async () => {
            const { SettingsEmailAccountCreateRoute } = await getLazyComponents();

            return SettingsEmailAccountCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.emailAccounts.edit.$pattern, "Email Accounts", async () => {
            const { SettingsEmailAccountEditRoute } = await getLazyComponents();

            return SettingsEmailAccountEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.imPlatforms.$pattern, "IM Platforms", async () => {
            const { SettingsImPlatformsRoute } = await getLazyComponents();

            return SettingsImPlatformsRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.imPlatforms.create.$pattern, "IM Platforms", async () => {
            const { SettingsImPlatformCreateRoute } = await getLazyComponents();

            return SettingsImPlatformCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.imPlatforms.edit.$pattern, "IM Platforms", async () => {
            const { SettingsImPlatformEditRoute } = await getLazyComponents();

            return SettingsImPlatformEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.sshKeys.$pattern, "SSH Keys", async () => {
            const { SettingsSSHKeysRoute } = await getLazyComponents();

            return SettingsSSHKeysRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.sshKeys.create.$pattern, "SSH Keys", async () => {
            const { SettingsSSHKeyCreateRoute } = await getLazyComponents();

            return SettingsSSHKeyCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.sshKeys.edit.$pattern, "SSH Keys", async () => {
            const { SettingsSSHKeyEditRoute } = await getLazyComponents();

            return SettingsSSHKeyEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.accessTokens.$pattern, "Access Tokens", async () => {
            const { SettingsAccessTokensRoute } = await getLazyComponents();

            return SettingsAccessTokensRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.accessTokens.create.$pattern, "Access Tokens", async () => {
            const { SettingsAccessTokenCreateRoute } = await getLazyComponents();

            return SettingsAccessTokenCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.accessTokens.edit.$pattern, "Access Tokens", async () => {
            const { SettingsAccessTokenEditRoute } = await getLazyComponents();

            return SettingsAccessTokenEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.acmeDnsProviders.$pattern, "ACME DNS Providers", async () => {
            const { SettingsAcmeDnsProvidersRoute } = await getLazyComponents();

            return SettingsAcmeDnsProvidersRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.acmeDnsProviders.create.$pattern, "ACME DNS Providers", async () => {
            const { SettingsAcmeDnsProviderCreateRoute } = await getLazyComponents();

            return SettingsAcmeDnsProviderCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.acmeDnsProviders.edit.$pattern, "ACME DNS Providers", async () => {
            const { SettingsAcmeDnsProviderEditRoute } = await getLazyComponents();

            return SettingsAcmeDnsProviderEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.cloudStorages.$pattern, "Cloud Storages", async () => {
            const { SettingsCloudStoragesRoute } = await getLazyComponents();

            return SettingsCloudStoragesRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.cloudStorages.create.$pattern, "Cloud Storages", async () => {
            const { SettingsCloudStorageCreateRoute } = await getLazyComponents();

            return SettingsCloudStorageCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.cloudStorages.edit.$pattern, "Cloud Storages", async () => {
            const { SettingsCloudStorageEditRoute } = await getLazyComponents();

            return SettingsCloudStorageEditRoute;
        }),
        createSettingsRoute(ROUTE.settings.oauth.$pattern, "OAuth", async () => {
            const { SettingsOAuthRoute } = await getLazyComponents();

            return SettingsOAuthRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.oauth.create.$pattern, "OAuth", async () => {
            const { SettingsOAuthCreateRoute } = await getLazyComponents();

            return SettingsOAuthCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.oauth.edit.$pattern, "OAuth", async () => {
            const { SettingsOAuthEditRoute } = await getLazyComponents();

            return SettingsOAuthEditRoute;
        }),
        createSettingsRoute(ROUTE.appSettings.imageBuild.$pattern, "Image Build", async () => {
            const { SettingsImageBuildRoute } = await getLazyComponents();

            return SettingsImageBuildRoute;
        }),
        createSettingsRoute(ROUTE.appSettings.appPlacement.$pattern, "App Placement", async () => {
            const { SettingsAppPlacementRoute } = await getLazyComponents();

            return SettingsAppPlacementRoute;
        }),
        createSettingsRoute(ROUTE.settings.notificationTargets.$pattern, "Notification Targets", async () => {
            const { SettingsNotificationTargetsRoute } = await getLazyComponents();

            return SettingsNotificationTargetsRoute;
        }),
        createSettingsModuleRoute(
            ROUTE.settings.notificationTargets.create.$pattern,
            "Notification Targets",
            async () => {
                const { SettingsNotificationTargetCreateRoute } = await getLazyComponents();

                return SettingsNotificationTargetCreateRoute;
            },
        ),
        createSettingsModuleRoute(
            ROUTE.settings.notificationTargets.edit.$pattern,
            "Notification Targets",
            async () => {
                const { SettingsNotificationTargetEditRoute } = await getLazyComponents();

                return SettingsNotificationTargetEditRoute;
            },
        ),
        createSettingsRoute(ROUTE.settings.backupSnapshots.$pattern, "Backup Snapshots", async () => {
            const { SettingsBackupSnapshotsRoute } = await getLazyComponents();

            return SettingsBackupSnapshotsRoute;
        }),
        createSettingsRoute(ROUTE.settings.backupRepos.$pattern, "Backup Repos", async () => {
            const { SettingsBackupReposRoute } = await getLazyComponents();

            return SettingsBackupReposRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.backupRepos.create.$pattern, "Backup Repos", async () => {
            const { SettingsBackupRepoCreateRoute } = await getLazyComponents();

            return SettingsBackupRepoCreateRoute;
        }),
        createSettingsModuleRoute(ROUTE.settings.backupRepos.edit.$pattern, "Backup Repos", async () => {
            const { SettingsBackupRepoEditRoute } = await getLazyComponents();

            return SettingsBackupRepoEditRoute;
        }),
        {
            lazy: async () => {
                const { DataBackupLayout } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.System}>
                            <ModuleTitle title="Data Backup">
                                <DataBackupLayout>
                                    <Outlet />
                                </DataBackupLayout>
                            </ModuleTitle>
                        </ConditionalModule>
                    ),
                };
            },
            path: ROUTE.appSettings.dataBackup.$pattern,
            children: [
                {
                    index: true,
                    element: (
                        <AppNavigate.Basic
                            to={ROUTE.appSettings.dataBackup.configuration.$route}
                            replace
                            ignorePrevPath
                        />
                    ),
                },
                {
                    path: "configuration",
                    lazy: async () => {
                        const { SettingsDataBackupConfigurationRoute } = await getLazyComponents();

                        return { Component: SettingsDataBackupConfigurationRoute };
                    },
                },
                {
                    path: "backup-files",
                    lazy: async () => {
                        const { SettingsDataBackupBackupFilesRoute } = await getLazyComponents();

                        return { Component: SettingsDataBackupBackupFilesRoute };
                    },
                },
                {
                    path: "actions",
                    lazy: async () => {
                        const { SettingsDataBackupActionsRoute } = await getLazyComponents();

                        return { Component: SettingsDataBackupActionsRoute };
                    },
                },
            ],
        },
        {
            lazy: async () => {
                const { DataCleanupLayout } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.System}>
                            <ModuleTitle title="Data Cleanup">
                                <DataCleanupLayout>
                                    <Outlet />
                                </DataCleanupLayout>
                            </ModuleTitle>
                        </ConditionalModule>
                    ),
                };
            },
            path: ROUTE.appSettings.dataCleanup.$pattern,
            children: [
                {
                    index: true,
                    element: (
                        <AppNavigate.Basic
                            to={ROUTE.appSettings.dataCleanup.configuration.$route}
                            replace
                            ignorePrevPath
                        />
                    ),
                },
                {
                    path: "configuration",
                    lazy: async () => {
                        const { SettingsDataCleanupConfigurationRoute } = await getLazyComponents();

                        return { Component: SettingsDataCleanupConfigurationRoute };
                    },
                },
                {
                    path: "actions",
                    lazy: async () => {
                        const { SettingsDataCleanupActionsRoute } = await getLazyComponents();

                        return { Component: SettingsDataCleanupActionsRoute };
                    },
                },
            ],
        },
        {
            lazy: async () => {
                const { SslRenewalLayout } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.System}>
                            <ModuleTitle title="SSL Renewal">
                                <SslRenewalLayout>
                                    <Outlet />
                                </SslRenewalLayout>
                            </ModuleTitle>
                        </ConditionalModule>
                    ),
                };
            },
            path: ROUTE.appSettings.sslRenewal.$pattern,
            children: [
                {
                    index: true,
                    element: (
                        <AppNavigate.Basic
                            to={ROUTE.appSettings.sslRenewal.configuration.$route}
                            replace
                            ignorePrevPath
                        />
                    ),
                },
                {
                    path: "configuration",
                    lazy: async () => {
                        const { SettingsSslRenewalConfigurationRoute } = await getLazyComponents();

                        return { Component: SettingsSslRenewalConfigurationRoute };
                    },
                },
                {
                    path: "actions",
                    lazy: async () => {
                        const { SettingsSslRenewalActionsRoute } = await getLazyComponents();

                        return { Component: SettingsSslRenewalActionsRoute };
                    },
                },
            ],
        },
        {
            lazy: async () => {
                const { BackupRepoCleanupLayout } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.System}>
                            <ModuleTitle title="Backup Repo Cleanup">
                                <BackupRepoCleanupLayout>
                                    <Outlet />
                                </BackupRepoCleanupLayout>
                            </ModuleTitle>
                        </ConditionalModule>
                    ),
                };
            },
            path: ROUTE.appSettings.backupRepoCleanup.$pattern,
            children: [
                {
                    index: true,
                    element: (
                        <AppNavigate.Basic
                            to={ROUTE.appSettings.backupRepoCleanup.configuration.$route}
                            replace
                            ignorePrevPath
                        />
                    ),
                },
                {
                    path: "configuration",
                    lazy: async () => {
                        const { SettingsBackupRepoCleanupConfigurationRoute } = await getLazyComponents();

                        return { Component: SettingsBackupRepoCleanupConfigurationRoute };
                    },
                },
                {
                    path: "actions",
                    lazy: async () => {
                        const { SettingsBackupRepoCleanupActionsRoute } = await getLazyComponents();

                        return { Component: SettingsBackupRepoCleanupActionsRoute };
                    },
                },
            ],
        },
        {
            path: "settings",
            element: (
                <ConditionalModule id={MODULE_IDS.Settings}>
                    <Navigate
                        to={ROUTE.appSettings.imageBuild.$route}
                        replace
                    />
                </ConditionalModule>
            ),
        },
        {
            path: "settings/*",
            element: <LegacySettingsRouteRedirect />,
        },
        {
            path: "providers-and-keys",
            element: (
                <ConditionalModule id={MODULE_IDS.Settings}>
                    <Navigate
                        to={ROUTE.settings.githubApps.$route}
                        replace
                    />
                </ConditionalModule>
            ),
        },
        {
            path: "providers-and-keys/*",
            element: <LegacySettingsRouteRedirect />,
        },
    ],
} as const;
