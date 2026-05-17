import { useParams } from "react-router";
import { ProjectWithSidebar } from "~/projects/module-shared/components";
import { ProjectCloudStorageTable } from "~/settings/module-shared/components";

export function ProjectCloudStoragesRoute() {
    const { id: projectId = "" } = useParams();
    return (
        <ProjectWithSidebar projectId={projectId}>
            <ProjectCloudStorageTable projectId={projectId} />
        </ProjectWithSidebar>
    );
}
