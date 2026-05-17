import type { ComponentType } from "react";

import { ROUTE } from "@/application/shared/constants";
import { Outlet, type RouteObject } from "react-router";

import { ModuleTitle } from "@application/shared/components/module-title";

async function getLazyComponents() {
    return await import("./settings.module");
}

function createSettingsRoute(path: string, title: string, loadComponent: () => Promise<ComponentType>): RouteObject {
    return {
        path,
        element: (
            <ModuleTitle title={title}>
                <Outlet />
            </ModuleTitle>
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
    ],
} as const;
