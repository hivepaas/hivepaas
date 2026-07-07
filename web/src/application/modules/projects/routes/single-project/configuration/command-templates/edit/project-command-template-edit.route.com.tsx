import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { ProjectCommandTemplateFormRoute } from "../form-route";

export function ProjectCommandTemplateEditRoute() {
    const { commandTemplateId, id: projectId } = useParams<{ commandTemplateId: string; id: string }>();

    invariant(projectId, "projectId must be defined");
    invariant(commandTemplateId, "commandTemplateId must be defined");

    return (
        <ProjectCommandTemplateFormRoute
            mode="edit"
            projectId={projectId}
            commandTemplateId={commandTemplateId}
        />
    );
}
