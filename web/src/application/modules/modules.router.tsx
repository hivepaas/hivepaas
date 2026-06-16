import { Outlet, type RouteObject } from "react-router";

import { CommonDialogsContainer } from "@application/shared/dialogs-container";
import { ModuleLayout } from "@application/shared/layouts/module";
import { Page404NotFound } from "@application/shared/pages";

import { clusterRouter } from "./cluster/cluster.router";
import { projectsRouter } from "./projects/projects.router";
import { settingsRouter } from "./settings/settings.router";
import { sourcesRouter } from "./sources/sources.router";
import { systemSettingsRouter } from "./system-settings";
import { systemStatusRouter } from "./system-status";
import { userManagementRouter } from "./user-management/user-management.router";

export const modulesRouter: RouteObject = {
    errorElement: (
        <ModuleLayout>
            <Page404NotFound />
        </ModuleLayout>
    ),
    element: (
        <ModuleLayout>
            <Outlet />
            <CommonDialogsContainer />
        </ModuleLayout>
    ),
    children: [
        userManagementRouter,
        clusterRouter,
        projectsRouter,
        sourcesRouter,
        settingsRouter,
        systemSettingsRouter,
        systemStatusRouter,
    ],
} as const;
