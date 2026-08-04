import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { ProjectCommandPipeFormRoute } from "../form-route";

export function ProjectCommandPipeEditRoute() {
    const { commandPipeId, id: projectId } = useParams<{ commandPipeId: string; id: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(commandPipeId, "commandPipeId must be defined");

    return (
        <ProjectCommandPipeFormRoute
            mode="edit"
            projectId={projectId}
            commandPipeId={commandPipeId}
        />
    );
}
