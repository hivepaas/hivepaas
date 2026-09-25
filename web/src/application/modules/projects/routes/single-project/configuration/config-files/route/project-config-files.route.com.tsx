import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { ProjectConfigFilesTable } from "../building-blocks";

export function ProjectConfigFilesRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return <ProjectConfigFilesTable projectId={projectId} />;
}
