import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";
import { ProjectSslProviderTable } from "~/settings/module-shared/components";

import { ProjectProviderSettingsScopeHeader } from "@application/modules/projects/module-shared/components";

export function ProjectSslProvidersRoute() {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <div className="flex flex-col gap-4">
            <ProjectProviderSettingsScopeHeader projectId={projectId} />
            <ProjectSslProviderTable
                key={scopedEnv ?? "all"}
                projectId={projectId}
                env={scopedEnv}
            />
        </div>
    );
}
