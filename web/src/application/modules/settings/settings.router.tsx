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
        createSettingsRoute(ROUTE.settings.registryAuth.$pattern, "Registry Auth", async () => {
            const { SettingsRegistryAuthRoute } = await getLazyComponents();

            return SettingsRegistryAuthRoute;
        }),
        createSettingsRoute(ROUTE.settings.sslCertificates.$pattern, "SSL Certificates", async () => {
            const { SettingsSslCertificatesRoute } = await getLazyComponents();

            return SettingsSslCertificatesRoute;
        }),
        createSettingsRoute(ROUTE.settings.emailAccounts.$pattern, "Email Accounts", async () => {
            const { SettingsEmailAccountsRoute } = await getLazyComponents();

            return SettingsEmailAccountsRoute;
        }),
        createSettingsRoute(ROUTE.settings.imPlatforms.$pattern, "IM Platforms", async () => {
            const { SettingsImPlatformsRoute } = await getLazyComponents();

            return SettingsImPlatformsRoute;
        }),
        createSettingsRoute(ROUTE.settings.sshKeys.$pattern, "SSH Keys", async () => {
            const { SettingsSSHKeysRoute } = await getLazyComponents();

            return SettingsSSHKeysRoute;
        }),
        createSettingsRoute(ROUTE.settings.accessTokens.$pattern, "Access Tokens", async () => {
            const { SettingsAccessTokensRoute } = await getLazyComponents();

            return SettingsAccessTokensRoute;
        }),
        createSettingsRoute(ROUTE.settings.cloudStorages.$pattern, "Cloud Storages", async () => {
            const { SettingsCloudStoragesRoute } = await getLazyComponents();

            return SettingsCloudStoragesRoute;
        }),
        createSettingsRoute(ROUTE.settings.oauth.$pattern, "OAuth", async () => {
            const { SettingsOAuthRoute } = await getLazyComponents();

            return SettingsOAuthRoute;
        }),
        createSettingsRoute(ROUTE.settings.notificationTargets.$pattern, "Notification Targets", async () => {
            const { SettingsNotificationTargetsRoute } = await getLazyComponents();

            return SettingsNotificationTargetsRoute;
        }),
    ],
} as const;
