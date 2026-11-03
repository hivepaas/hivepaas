import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { ProjectSecretFormRoute } from "../form-route";

export function ProjectSecretEditRoute() {
    const { id: projectId, secretId } = useParams<{ id: string; secretId: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(secretId, "secretId must be defined");

    return (
        <ProjectSecretFormRoute
            mode="edit"
            projectId={projectId}
            secretId={secretId}
        />
    );
}
