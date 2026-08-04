import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { ProjectCommandPipeTable } from "../building-blocks";

export function ProjectCommandPipesRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return <ProjectCommandPipeTable projectId={projectId} />;
}
