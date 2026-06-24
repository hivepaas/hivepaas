import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppScheduledJobFormRoute } from "../form-route";

export function AppScheduledJobCreateRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    return (
        <AppScheduledJobFormRoute
            mode="create"
            projectId={projectId}
            appId={appId}
        />
    );
}
