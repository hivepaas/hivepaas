import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { StorageMountFormRoute } from "../form-route";

export function StorageMountCreateRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    return (
        <StorageMountFormRoute
            mode="create"
            projectId={projectId}
            appId={appId}
        />
    );
}
