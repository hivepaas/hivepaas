import { useParams } from "react-router";
import invariant from "tiny-invariant";

import { ProjectSecretsTable } from "../building-blocks";

export function ProjectSecretsRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return <ProjectSecretsTable projectId={projectId} />;
}
