import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { ProjectWithSidebar } from "~/projects/module-shared/components";
import { ProjectEmailAccountsTable } from "~/settings/module-shared/components";

export function ProjectEmailAccountsRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return (
        <ProjectWithSidebar projectId={projectId}>
            <ProjectEmailAccountsTable projectId={projectId} />
        </ProjectWithSidebar>
    );
}
