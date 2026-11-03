import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppJobSequenceFormRoute } from "./app-job-sequence-form-route.com";

export function AppJobSequenceCreateRoute() {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    return (
        <AppJobSequenceFormRoute
            mode="create"
            projectId={projectId}
            env={env}
            appId={appId}
        />
    );
}
