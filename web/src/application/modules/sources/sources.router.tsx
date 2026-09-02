import { Navigate, type RouteObject } from "react-router";

import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { ConditionalModule } from "@application/shared/permissions";

import { LegacyGithubAppEditRedirect, LegacyRepoWebhookEditRedirect } from "./sources-legacy-redirect.com";

export const sourcesRouter: RouteObject = {
    children: [
        {
            path: "sources",
            element: (
                <ConditionalModule id={MODULE_IDS.Settings}>
                    <Navigate
                        to={ROUTE.settings.githubApps.$route}
                        replace
                    />
                </ConditionalModule>
            ),
        },
        {
            path: "sources/github-apps",
            element: (
                <ConditionalModule id={MODULE_IDS.Settings}>
                    <Navigate
                        to={ROUTE.settings.githubApps.$route}
                        replace
                    />
                </ConditionalModule>
            ),
        },
        {
            path: "sources/github-apps/create",
            element: (
                <ConditionalModule id={MODULE_IDS.Settings}>
                    <Navigate
                        to={ROUTE.settings.githubApps.create.$route}
                        replace
                    />
                </ConditionalModule>
            ),
        },
        {
            path: "sources/github-apps/:githubAppId/edit",
            element: (
                <ConditionalModule id={MODULE_IDS.Settings}>
                    <LegacyGithubAppEditRedirect />
                </ConditionalModule>
            ),
        },
        {
            path: "sources/webhooks",
            element: (
                <ConditionalModule id={MODULE_IDS.Settings}>
                    <Navigate
                        to={ROUTE.settings.webhooks.$route}
                        replace
                    />
                </ConditionalModule>
            ),
        },
        {
            path: "sources/webhooks/create",
            element: (
                <ConditionalModule id={MODULE_IDS.Settings}>
                    <Navigate
                        to={ROUTE.settings.webhooks.create.$route}
                        replace
                    />
                </ConditionalModule>
            ),
        },
        {
            path: "sources/webhooks/:repoWebhookId/edit",
            element: (
                <ConditionalModule id={MODULE_IDS.Settings}>
                    <LegacyRepoWebhookEditRedirect />
                </ConditionalModule>
            ),
        },
    ],
} as const;
