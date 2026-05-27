import { Navigate, useParams } from "react-router";

import { ROUTE } from "@application/shared/constants";

export function ProjectGithubAppsLegacyRedirect() {
    const { id } = useParams<{ id: string }>();

    return (
        <Navigate
            to={id ? ROUTE.projects.single.configuration.githubApps.$route(id) : ROUTE.projects.list.$route}
            replace
        />
    );
}
