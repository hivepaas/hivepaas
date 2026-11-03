import { useParams } from "react-router";
import invariant from "tiny-invariant";
import { NetworkManagementTable } from "~/cluster/module-shared/components";
import { ProjectProviderSettingsScopeHeader } from "~/projects/module-shared/components";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";

export function ProjectNetworksRoute() {
    const { id: projectId = "" } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <div className="flex flex-col gap-4">
            <ProjectProviderSettingsScopeHeader projectId={projectId} />
            <NetworkManagementTable
                key={scopedEnv ?? "all"}
                scope={{
                    type: "project",
                    projectId,
                    env: scopedEnv,
                }}
            />
        </div>
    );
}
