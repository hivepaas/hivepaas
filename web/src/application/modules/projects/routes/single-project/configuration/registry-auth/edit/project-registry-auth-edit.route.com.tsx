import { useParams } from "react-router";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";
import { RegistryAuthFormRoute } from "~/settings/module-shared/components/registry-auth-form-route";

export function ProjectRegistryAuthEditRoute() {
    const { id: projectId = "", registryAuthId = "" } = useParams();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <RegistryAuthFormRoute
            mode="edit"
            scope={{ type: "project", projectId, env: scopedEnv }}
            registryAuthId={registryAuthId}
        />
    );
}
