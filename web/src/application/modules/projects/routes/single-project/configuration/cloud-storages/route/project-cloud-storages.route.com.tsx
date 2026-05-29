import { useParams } from "react-router";
import { ProjectCloudStorageTable } from "~/settings/module-shared/components";

export function ProjectCloudStoragesRoute() {
    const { id: projectId = "" } = useParams();
    return <ProjectCloudStorageTable projectId={projectId} />;
}
