import { ROUTE } from "@/application/shared/constants";
import { Outlet, type RouteObject } from "react-router";

async function getLazyComponents() {
    return await import("./settings.module");
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
            path: ROUTE.settings.basicAuth.$pattern,
            lazy: async () => {
                const { SettingsBasicAuthRoute } = await getLazyComponents();

                return {
                    Component: SettingsBasicAuthRoute,
                };
            },
        },
        {
            path: ROUTE.settings.registryAuth.$pattern,
            lazy: async () => {
                const { SettingsRegistryAuthRoute } = await getLazyComponents();

                return {
                    Component: SettingsRegistryAuthRoute,
                };
            },
        },
        {
            path: ROUTE.settings.sslCertificates.$pattern,
            lazy: async () => {
                const { SettingsSslCertificatesRoute } = await getLazyComponents();

                return {
                    Component: SettingsSslCertificatesRoute,
                };
            },
        },
        {
            path: ROUTE.settings.emailAccounts.$pattern,
            lazy: async () => {
                const { SettingsEmailAccountsRoute } = await getLazyComponents();

                return {
                    Component: SettingsEmailAccountsRoute,
                };
            },
        },
        {
            path: ROUTE.settings.imPlatforms.$pattern,
            lazy: async () => {
                const { SettingsImPlatformsRoute } = await getLazyComponents();

                return {
                    Component: SettingsImPlatformsRoute,
                };
            },
        },
    ],
} as const;
