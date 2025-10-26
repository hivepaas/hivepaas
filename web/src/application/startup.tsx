import { Outlet, RouterProvider, createBrowserRouter } from "react-router-dom";

import { ApplicationProfileInit, AuthRouteProtection } from "@application/root";

import { AppLoader } from "@application/shared/components/loaders";
import { Page404NotFound } from "@application/shared/pages";

import { authenticationRouter } from "@application/authentication";

import { modulesRouter } from "@application/modules";

const router = createBrowserRouter([
    {
        element: (
            <AuthRouteProtection>
                <Outlet />
            </AuthRouteProtection>
        ),
        HydrateFallback: AppLoader,
        children: [
            {
                children: [authenticationRouter],
            },

            {
                element: <Outlet />,
                children: [modulesRouter],
            },

            {
                path: "*",
                Component: Page404NotFound,
            },
        ],
    },
]);

export function Startup() {
    return (
        <ApplicationProfileInit>
            <RouterProvider router={router} />
        </ApplicationProfileInit>
    );
}
