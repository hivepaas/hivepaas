import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { ProjectCommandTemplateTable } from "../building-blocks";

export function ProjectCommandTemplatesRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return <ProjectCommandTemplateTable projectId={projectId} />;
}
