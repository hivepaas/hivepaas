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
        createSettingsRoute(ROUTE.settings.basicAuth.$pattern, "Basic auth", async () => {
            const { SettingsBasicAuthRoute } = await getLazyComponents();

            return SettingsBasicAuthRoute;
        }),
        createSettingsRoute(ROUTE.settings.registryAuth.$pattern, "Registry auth", async () => {
            const { SettingsRegistryAuthRoute } = await getLazyComponents();

            return SettingsRegistryAuthRoute;
        }),
        createSettingsRoute(ROUTE.settings.sslCertificates.$pattern, "SSL certificates", async () => {
            const { SettingsSslCertificatesRoute } = await getLazyComponents();

            return SettingsSslCertificatesRoute;
        }),
        createSettingsRoute(ROUTE.settings.emailAccounts.$pattern, "Email accounts", async () => {
            const { SettingsEmailAccountsRoute } = await getLazyComponents();

            return SettingsEmailAccountsRoute;
        }),
        createSettingsRoute(ROUTE.settings.imPlatforms.$pattern, "IM platforms", async () => {
            const { SettingsImPlatformsRoute } = await getLazyComponents();

            return SettingsImPlatformsRoute;
        }),
    ],
} as const;
