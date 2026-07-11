import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppDataFileFormRoute } from "../form-route";

export function AppDataFileCreateRoute() {
    const { id: projectId, appId } = useParams<{ id: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");

    return (
        <AppDataFileFormRoute
            projectId={projectId}
            appId={appId}
        />
    );
}
