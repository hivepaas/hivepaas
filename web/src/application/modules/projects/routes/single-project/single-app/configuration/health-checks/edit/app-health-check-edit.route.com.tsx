import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppHealthCheckFormRoute } from "../form-route";

export function AppHealthCheckEditRoute() {
    const {
        id: projectId,
        appId,
        healthCheckId,
    } = useParams<{
        id: string;
        appId: string;
        healthCheckId: string;
    }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");
    invariant(healthCheckId, "healthCheckId must be defined");

    return (
        <AppHealthCheckFormRoute
            mode="edit"
            projectId={projectId}
            appId={appId}
            healthCheckId={healthCheckId}
        />
    );
}
