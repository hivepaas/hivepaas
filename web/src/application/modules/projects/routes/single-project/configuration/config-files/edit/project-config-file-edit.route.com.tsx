import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { ProjectConfigFileFormRoute } from "../form-route";

export function ProjectConfigFileEditRoute() {
    const { id: projectId, configFileId } = useParams<{ id: string; configFileId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(configFileId, "configFileId must be defined");

    return (
        <ProjectConfigFileFormRoute
            mode="edit"
            projectId={projectId}
            configFileId={configFileId}
        />
    );
}
