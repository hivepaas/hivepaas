import { useParams } from "react-router";
import { ProjectWithSidebar } from "~/projects/module-shared/components";
import { ProjectNotificationTargetTable } from "~/settings/module-shared/components";

export function ProjectNotificationTargetsRoute() {
    const { id: projectId = "" } = useParams();
    return (
        <ProjectWithSidebar projectId={projectId}>
            <ProjectNotificationTargetTable projectId={projectId} />
        </ProjectWithSidebar>
    );
}
