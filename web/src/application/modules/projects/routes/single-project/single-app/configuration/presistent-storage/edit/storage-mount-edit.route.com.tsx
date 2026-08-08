import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { StorageMountFormRoute } from "../form-route";

export function StorageMountEditRoute() {
    const {
        id: projectId,
        env,
        appId,
        mountId,
    } = useParams<{
        id: string;
        env: string;
        appId: string;
        mountId: string;
    }>();

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");
    invariant(mountId, "mountId must be defined");

    return (
        <StorageMountFormRoute
            mode="edit"
            projectId={projectId}
            env={env}
            appId={appId}
            mountId={mountId}
        />
    );
}
