import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { StorageMountFormRoute } from "../form-route";

export function StorageMountEditRoute() {
    const { id: projectId, appId, mountId } = useParams<{ id: string; appId: string; mountId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");
    invariant(mountId, "mountId must be defined");

    return (
        <StorageMountFormRoute
            mode="edit"
            projectId={projectId}
            appId={appId}
            mountId={mountId}
        />
    );
}
