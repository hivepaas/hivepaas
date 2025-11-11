import { Outlet, type RouteObject } from "react-router";

import { CommonDialogsContainer } from "@application/shared/dialogs-container";
import { ModuleLayout } from "@application/shared/layouts/module";
import { Page404NotFound } from "@application/shared/pages";

import { usersAndRolesRouter } from "./users-and-roles/users-and-roles.router";

export const modulesRouter: RouteObject = {
    errorElement: (
        <ModuleLayout>
            <Page404NotFound />
        </ModuleLayout>
    ),
    element: (
        <>
            <ModuleLayout>
                <Outlet />
            </ModuleLayout>
            <CommonDialogsContainer />
        </>
    ),
    children: [usersAndRolesRouter],
} as const;
