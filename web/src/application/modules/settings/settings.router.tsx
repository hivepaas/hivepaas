import type { ComponentType } from "react";

import { MODULE_IDS, ROUTE } from "@/application/shared/constants";
import { Navigate, Outlet, type RouteObject } from "react-router";

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
            path: "settings/github-apps",
            element: (
                <ConditionalModule id={MODULE_IDS.Settings}>
                    <Navigate
                        to={ROUTE.sources.githubApps.$route}
                        replace
                    />
                </ConditionalModule>
            ),
        },
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
    ],
} as const;
