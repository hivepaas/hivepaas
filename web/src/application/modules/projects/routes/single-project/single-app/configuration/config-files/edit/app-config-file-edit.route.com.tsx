import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { AppConfigFileFormRoute } from "../form-route";

export function AppConfigFileEditRoute() {
    const {
        id: projectId,
        appId,
        configFileId,
    } = useParams<{
        id: string;
        appId: string;
        configFileId: string;
    }>();

    invariant(projectId, "projectId must be defined");
    invariant(appId, "appId must be defined");
    invariant(configFileId, "configFileId must be defined");

    return (
        <AppConfigFileFormRoute
            mode="edit"
            projectId={projectId}
            appId={appId}
            configFileId={configFileId}
        />
    );
}
