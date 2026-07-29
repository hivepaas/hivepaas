import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppConfigFileFormRoute } from "../form-route";

export function AppConfigFileEditRoute() {
    const {
        id: projectId,
        env,
        appId,
        configFileId,
    } = useParams<{
        id: string;
        env: string;
        appId: string;
        configFileId: string;
    }>();

    invariant(projectId, "projectId must be defined");
    invariant(env, "env must be defined");
    invariant(appId, "appId must be defined");
    invariant(configFileId, "configFileId must be defined");

    return (
        <AppConfigFileFormRoute
            mode="edit"
            projectId={projectId}
            env={env}
            appId={appId}
            configFileId={configFileId}
        />
    );
}
