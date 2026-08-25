import { useParams } from "react-router";
import { ProjectNotificationTargetTable } from "~/settings/module-shared/components";

import { ProjectProviderSettingsScopeHeader } from "@application/modules/projects/module-shared/components";

export function ProjectNotificationTargetsRoute() {
    const { id: projectId = "" } = useParams();
    return (
        <div className="flex flex-col gap-4">
            <ProjectProviderSettingsScopeHeader projectId={projectId} />
            <ProjectNotificationTargetTable projectId={projectId} />
        </div>
    );
}
