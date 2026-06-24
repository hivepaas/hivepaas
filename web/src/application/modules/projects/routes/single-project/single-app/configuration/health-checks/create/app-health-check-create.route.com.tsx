import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppHealthCheckFormRoute } from "../form-route";

export function AppHealthCheckCreateRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    return (
        <AppHealthCheckFormRoute
            mode="create"
            projectId={projectId}
            appId={appId}
        />
    );
}
