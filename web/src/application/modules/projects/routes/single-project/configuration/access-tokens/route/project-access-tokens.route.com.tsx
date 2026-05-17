import { useParams } from "react-router";
import { ProjectWithSidebar } from "~/projects/module-shared/components";
import { ProjectAccessTokenTable } from "~/settings/module-shared/components";

export function ProjectAccessTokensRoute() {
    const { id: projectId = "" } = useParams();
    return (
        <ProjectWithSidebar projectId={projectId}>
            <ProjectAccessTokenTable projectId={projectId} />
        </ProjectWithSidebar>
    );
}
