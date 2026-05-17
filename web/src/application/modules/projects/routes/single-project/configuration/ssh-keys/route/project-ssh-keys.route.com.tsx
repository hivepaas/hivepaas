import { useParams } from "react-router";
import { ProjectWithSidebar } from "~/projects/module-shared/components";
import { ProjectSSHKeyTable } from "~/settings/module-shared/components";

export function ProjectSSHKeysRoute() {
    const { id: projectId = "" } = useParams();
    return (
        <ProjectWithSidebar projectId={projectId}>
            <ProjectSSHKeyTable projectId={projectId} />
        </ProjectWithSidebar>
    );
}
