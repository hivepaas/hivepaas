import { useParams } from "react-router";
import { ProjectSecretsTable } from "../building-blocks";
import invariant from "tiny-invariant";

import { ProjectWithSidebar } from "~/projects/module-shared/components";

export function ProjectSecretsRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return (
        <ProjectWithSidebar projectId={projectId}>
            <div className="bg-background rounded-lg p-4 w-full mx-auto">
                <ProjectSecretsTable projectId={projectId} />
            </div>
        </ProjectWithSidebar>
    );
}
