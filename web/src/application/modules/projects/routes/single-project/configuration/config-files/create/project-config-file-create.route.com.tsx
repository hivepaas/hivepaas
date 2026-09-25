import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { ProjectConfigFileFormRoute } from "../form-route";

export function ProjectConfigFileCreateRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return (
        <ProjectConfigFileFormRoute
            mode="create"
            projectId={projectId}
        />
    );
}
