import { useParams } from "react-router";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";
import { ProjectAccessTokenTable } from "~/settings/module-shared/components";

import { ProjectProviderSettingsScopeHeader } from "@application/modules/projects/module-shared/components";

export function ProjectAccessTokensRoute() {
    const { id: projectId = "" } = useParams();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <div className="flex flex-col gap-4">
            <ProjectProviderSettingsScopeHeader projectId={projectId} />
            <ProjectAccessTokenTable
                key={scopedEnv ?? "all"}
                projectId={projectId}
                env={scopedEnv}
            />
        </div>
    );
}
