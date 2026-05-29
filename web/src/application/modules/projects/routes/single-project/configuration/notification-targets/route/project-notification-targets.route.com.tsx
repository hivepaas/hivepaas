import { useParams } from "react-router";
import { ProjectNotificationTargetTable } from "~/settings/module-shared/components";

export function ProjectNotificationTargetsRoute() {
    const { id: projectId = "" } = useParams();
    return <ProjectNotificationTargetTable projectId={projectId} />;
}
