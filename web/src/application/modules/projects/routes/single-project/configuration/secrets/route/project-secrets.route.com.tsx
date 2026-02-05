import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectWithSidebar } from "~/projects/module-shared/components";

import { ProjectSecretsTable } from "../building-blocks";

export function ProjectSecretsRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return (
        <ProjectWithSidebar projectId={projectId}>
            <ProjectSecretsTable projectId={projectId} />
        </ProjectWithSidebar>
    );
}
