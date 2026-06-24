import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppConfigFileFormRoute } from "../form-route";

export function AppConfigFileCreateRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    return (
        <AppConfigFileFormRoute
            mode="create"
            projectId={projectId}
            appId={appId}
        />
    );
}
