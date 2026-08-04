import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { ProjectCommandPipeFormRoute } from "../form-route";

export function ProjectCommandPipeCreateRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return (
        <ProjectCommandPipeFormRoute
            mode="create"
            projectId={projectId}
        />
    );
}
