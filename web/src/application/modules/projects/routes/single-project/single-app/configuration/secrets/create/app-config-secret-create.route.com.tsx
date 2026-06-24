import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppSecretFormRoute } from "../form-route";

export function AppConfigSecretCreateRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    return (
        <AppSecretFormRoute
            mode="create"
            projectId={projectId}
            appId={appId}
        />
    );
}
