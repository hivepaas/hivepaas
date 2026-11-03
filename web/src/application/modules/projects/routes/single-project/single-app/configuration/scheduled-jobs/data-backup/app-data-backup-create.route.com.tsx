import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppDataBackupFormRoute } from "./app-data-backup-form-route.com";

export function AppDataBackupCreateRoute() {
    const { id: projectId, env, appId } = useParams<{ id: string; env: string; appId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");

    return (
        <AppDataBackupFormRoute
            mode="create"
            projectId={projectId}
            env={env}
            appId={appId}
        />
    );
}
