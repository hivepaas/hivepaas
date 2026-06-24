import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { ProjectSecretFormRoute } from "../form-route";

export function ProjectSecretCreateRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return (
        <ProjectSecretFormRoute
            mode="create"
            projectId={projectId}
        />
    );
}
