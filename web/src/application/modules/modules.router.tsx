import { Outlet, type RouteObject } from "react-router";

import { CommonDialogsContainer } from "@application/shared/dialogs-container";
import { ModuleLayout } from "@application/shared/layouts/module";
import { Page404NotFound } from "@application/shared/pages";

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
    children: [userManagementRouter],
} as const;
