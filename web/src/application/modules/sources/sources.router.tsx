import type { ComponentType } from "react";

import { Navigate, Outlet, type RouteObject } from "react-router";

import { ModuleTitle } from "@application/shared/components/module-title";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { ConditionalModule } from "@application/shared/permissions";

async function getLazyComponents() {
    return await import("./sources.module");
}

function createSourcesRoute(
    path: string,
    title: string,
    loadComponent: () => Promise<ComponentType>,
    children: RouteObject[] = [],
): RouteObject {
    return {
        path,
        element: (
            <ConditionalModule id={MODULE_IDS.Settings}>
                <ModuleTitle title={title}>
                    <Outlet />
                </ModuleTitle>
            </ConditionalModule>
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
            ...children,
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
                <ConditionalModule id={MODULE_IDS.Settings}>
                    <Navigate
                        to={ROUTE.sources.githubApps.$route}
                        replace
                    />
                </ConditionalModule>
            ),
        },
        createSourcesRoute(
            ROUTE.sources.githubApps.$pattern,
            "Github Apps",
            async () => {
                const { SourcesGithubAppsRoute } = await getLazyComponents();

                return SourcesGithubAppsRoute;
            },
            [
                {
                    path: "create",
                    lazy: async () => {
                        const { SourcesGithubAppCreateRoute } = await getLazyComponents();

                        return { Component: SourcesGithubAppCreateRoute };
                    },
                },
                {
                    path: ":githubAppId/edit",
                    lazy: async () => {
                        const { SourcesGithubAppEditRoute } = await getLazyComponents();

                        return { Component: SourcesGithubAppEditRoute };
                    },
                },
            ],
        ),
        createSourcesRoute(
            ROUTE.sources.webhooks.$pattern,
            "Webhooks",
            async () => {
                const { SourcesWebhooksRoute } = await getLazyComponents();

                return SourcesWebhooksRoute;
            },
            [
                {
                    path: "create",
                    lazy: async () => {
                        const { SourcesWebhookCreateRoute } = await getLazyComponents();

                        return { Component: SourcesWebhookCreateRoute };
                    },
                },
                {
                    path: ":repoWebhookId/edit",
                    lazy: async () => {
                        const { SourcesWebhookEditRoute } = await getLazyComponents();

                        return { Component: SourcesWebhookEditRoute };
                    },
                },
            ],
        ),
    ],
} as const;
