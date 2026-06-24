import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppSecretFormRoute } from "../form-route";

export function AppConfigSecretEditRoute() {
    const { id: projectId, appId, secretId } = useParams<{ id: string; appId: string; secretId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");
    invariant(secretId, "secretId must be defined");

    return (
        <AppSecretFormRoute
            mode="edit"
            projectId={projectId}
            appId={appId}
            secretId={secretId}
        />
    );
}
