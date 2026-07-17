import { MODULE_IDS, ROUTE } from "@/application/shared/constants";
import { Outlet, type RouteObject } from "react-router";

import { AppNavigate } from "@application/shared/components";
import { ModuleTitle } from "@application/shared/components/module-title";
import { ConditionalModule } from "@application/shared/permissions";

async function getLazyComponents() {
    return await import("./system-settings.module");
}

export const systemSettingsRouter: RouteObject = {
    lazy: async () => {
        const { SystemSettingsDialogsContainer } = await getLazyComponents();

        return {
            element: (
                <>
                    <Outlet />
                    <SystemSettingsDialogsContainer />
                </>
            ),
        };
    },
    children: [
        {
            lazy: async () => {
                const { HivePaaSLayout } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.System}>
                            <ModuleTitle title="HivePaaS">
                                <HivePaaSLayout>
                                    <Outlet />
                                </HivePaaSLayout>
                            </ModuleTitle>
                        </ConditionalModule>
                    ),
                };
            },
            path: ROUTE.systemSettings.hivepaas.$pattern,
            children: [
                {
                    index: true,
                    element: (
                        <AppNavigate.Basic
                            to={ROUTE.systemSettings.hivepaas.general.$route}
                            replace
                            ignorePrevPath
                        />
                    ),
                },
                {
                    path: "general",
                    lazy: async () => {
                        const { SystemSettingsHivePaaSGeneralRoute } = await getLazyComponents();

                        return { Component: SystemSettingsHivePaaSGeneralRoute };
                    },
                },
                {
                    path: "http-settings",
                    lazy: async () => {
                        const { SystemSettingsHivePaaSHttpSettingsRoute } = await getLazyComponents();

                        return { Component: SystemSettingsHivePaaSHttpSettingsRoute };
                    },
                },
            ],
        },
        {
            lazy: async () => {
                const { TraefikLayout } = await getLazyComponents();

                return {
                    element: (
                        <ConditionalModule id={MODULE_IDS.System}>
                            <ModuleTitle title="Traefik">
                                <TraefikLayout>
                                    <Outlet />
                                </TraefikLayout>
                            </ModuleTitle>
                        </ConditionalModule>
                    ),
                };
            },
            path: ROUTE.systemSettings.traefik.$pattern,
            children: [
                {
                    index: true,
                    element: (
                        <AppNavigate.Basic
                            to={ROUTE.systemSettings.traefik.general.$route}
                            replace
                            ignorePrevPath
                        />
                    ),
                },
                {
                    path: "general",
                    lazy: async () => {
                        const { SystemSettingsTraefikGeneralRoute } = await getLazyComponents();

                        return { Component: SystemSettingsTraefikGeneralRoute };
                    },
                },
            ],
        },
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
            path: ROUTE.systemSettings.dataBackup.$pattern,
            children: [
                {
                    index: true,
                    element: (
                        <AppNavigate.Basic
                            to={ROUTE.systemSettings.dataBackup.configuration.$route}
                            replace
                            ignorePrevPath
                        />
                    ),
                },
                {
                    path: "configuration",
                    lazy: async () => {
                        const { SystemSettingsDataBackupConfigurationRoute } = await getLazyComponents();

                        return { Component: SystemSettingsDataBackupConfigurationRoute };
                    },
                },
                {
                    path: "backup-files",
                    lazy: async () => {
                        const { SystemSettingsDataBackupBackupFilesRoute } = await getLazyComponents();

                        return { Component: SystemSettingsDataBackupBackupFilesRoute };
                    },
                },
                {
                    path: "actions",
                    lazy: async () => {
                        const { SystemSettingsDataBackupActionsRoute } = await getLazyComponents();

                        return { Component: SystemSettingsDataBackupActionsRoute };
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
            path: ROUTE.systemSettings.dataCleanup.$pattern,
            children: [
                {
                    index: true,
                    element: (
                        <AppNavigate.Basic
                            to={ROUTE.systemSettings.dataCleanup.configuration.$route}
                            replace
                            ignorePrevPath
                        />
                    ),
                },
                {
                    path: "configuration",
                    lazy: async () => {
                        const { SystemSettingsDataCleanupConfigurationRoute } = await getLazyComponents();

                        return { Component: SystemSettingsDataCleanupConfigurationRoute };
                    },
                },
                {
                    path: "actions",
                    lazy: async () => {
                        const { SystemSettingsDataCleanupActionsRoute } = await getLazyComponents();

                        return { Component: SystemSettingsDataCleanupActionsRoute };
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
            path: ROUTE.systemSettings.sslRenewal.$pattern,
            children: [
                {
                    index: true,
                    element: (
                        <AppNavigate.Basic
                            to={ROUTE.systemSettings.sslRenewal.configuration.$route}
                            replace
                            ignorePrevPath
                        />
                    ),
                },
                {
                    path: "configuration",
                    lazy: async () => {
                        const { SystemSettingsSslRenewalConfigurationRoute } = await getLazyComponents();

                        return { Component: SystemSettingsSslRenewalConfigurationRoute };
                    },
                },
                {
                    path: "actions",
                    lazy: async () => {
                        const { SystemSettingsSslRenewalActionsRoute } = await getLazyComponents();

                        return { Component: SystemSettingsSslRenewalActionsRoute };
                    },
                },
            ],
        },
    ],
} as const;
