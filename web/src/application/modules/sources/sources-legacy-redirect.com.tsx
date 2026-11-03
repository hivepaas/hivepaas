import { Navigate, useParams } from "react-router";

import { ROUTE } from "@application/shared/constants";

export function LegacyGithubAppEditRedirect() {
    const { githubAppId = "" } = useParams();

    return (
        <Navigate
            to={ROUTE.settings.githubApps.edit.$route(githubAppId)}
            replace
        />
    );
}

export function LegacyRepoWebhookEditRedirect() {
    const { repoWebhookId = "" } = useParams();

    return (
        <Navigate
            to={ROUTE.settings.webhooks.edit.$route(repoWebhookId)}
            replace
        />
    );
}
