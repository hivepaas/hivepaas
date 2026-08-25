import { useParams } from "react-router";
import { ProjectSSHKeyTable } from "~/settings/module-shared/components";

import { ProjectProviderSettingsScopeHeader } from "@application/modules/projects/module-shared/components";

export function ProjectSSHKeysRoute() {
    const { id: projectId = "" } = useParams();
    return (
        <div className="flex flex-col gap-4">
            <ProjectProviderSettingsScopeHeader projectId={projectId} />
            <ProjectSSHKeyTable projectId={projectId} />
        </div>
    );
}
