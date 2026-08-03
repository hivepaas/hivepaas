import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppHealthCheckFormRoute } from "../form-route";

export function AppPeriodicJobHealthCheckEditRoute() {
    const {
        id: projectId,
        env,
        appId,
        healthCheckId,
    } = useParams<{
        id: string;
        env: string;
        appId: string;
        healthCheckId: string;
    }>();

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");
    invariant(healthCheckId, "healthCheckId must be defined");

    return (
        <AppHealthCheckFormRoute
            mode="edit"
            projectId={projectId}
            env={env}
            appId={appId}
            healthCheckId={healthCheckId}
        />
    );
}
