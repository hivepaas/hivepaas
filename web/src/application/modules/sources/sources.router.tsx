import type { ComponentType } from "react";

import { Navigate, Outlet, type RouteObject } from "react-router";

import { ModuleTitle } from "@application/shared/components/module-title";
import { ROUTE } from "@application/shared/constants";

async function getLazyComponents() {
    return await import("./sources.module");
}

function createSourcesRoute(path: string, title: string, loadComponent: () => Promise<ComponentType>): RouteObject {
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

export const sourcesRouter: RouteObject = {
    lazy: async () => {
        const { SourcesDialogsContainer } = await getLazyComponents();

        return {
            element: (
                <>
                    <Outlet />
                    <SourcesDialogsContainer />
                </>
            ),
        };
    },
    children: [
        {
            path: ROUTE.sources.$pattern,
            element: (
                <Navigate
                    to={ROUTE.sources.githubApps.$route}
                    replace
                />
            ),
        },
        createSourcesRoute(ROUTE.sources.githubApps.$pattern, "Github Apps", async () => {
            const { SourcesGithubAppsRoute } = await getLazyComponents();

            return SourcesGithubAppsRoute;
        }),
        createSourcesRoute(ROUTE.sources.webhooks.$pattern, "Webhooks", async () => {
            const { SourcesWebhooksRoute } = await getLazyComponents();

            return SourcesWebhooksRoute;
        }),
    ],
} as const;
