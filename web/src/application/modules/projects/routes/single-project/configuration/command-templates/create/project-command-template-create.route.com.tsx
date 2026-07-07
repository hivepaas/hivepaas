import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { ProjectCommandTemplateFormRoute } from "../form-route";

export function ProjectCommandTemplateCreateRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return (
        <ProjectCommandTemplateFormRoute
            mode="create"
            projectId={projectId}
        />
    );
}
