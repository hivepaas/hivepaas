import { Navigate, useParams } from "react-router";

import { ROUTE } from "@application/shared/constants";

interface ProjectRouteRedirectProps {
    to: (id: string) => string;
}

export function ProjectRouteRedirect({ to }: ProjectRouteRedirectProps) {
    const { id } = useParams<{ id: string }>();

    return (
        <Navigate
            to={id ? to(id) : ROUTE.projects.list.$route}
            replace
        />
    );
}
