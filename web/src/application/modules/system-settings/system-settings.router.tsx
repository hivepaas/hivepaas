import { MODULE_IDS, ROUTE } from "@/application/shared/constants";
import { Outlet, type RouteObject } from "react-router";

import { AppNavigate } from "@application/shared/components";
import { ModuleTitle } from "@application/shared/components/module-title";
import { ConditionalModule } from "@application/shared/permissions";

async function getLazyComponents() {
    return await import("./system-settings.module");
}

export const systemSettingsRouter: RouteObject = {
    children: [
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
    ],
} as const;
