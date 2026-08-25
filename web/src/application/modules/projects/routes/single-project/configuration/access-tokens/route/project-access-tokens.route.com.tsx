import { useParams } from "react-router";
import { ProjectAccessTokenTable } from "~/settings/module-shared/components";

import { ProjectProviderSettingsScopeHeader } from "@application/modules/projects/module-shared/components";

export function ProjectAccessTokensRoute() {
    const { id: projectId = "" } = useParams();
    return (
        <div className="flex flex-col gap-4">
            <ProjectProviderSettingsScopeHeader projectId={projectId} />
            <ProjectAccessTokenTable projectId={projectId} />
        </div>
    );
}
