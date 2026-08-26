import { useParams } from "react-router";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";
import { AcmeDnsProviderFormRoute } from "~/settings/module-shared/components/acme-dns-provider-form-route";

export function ProjectAcmeDnsProviderCreateRoute() {
    const { id: projectId = "" } = useParams();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <AcmeDnsProviderFormRoute
            mode="create"
            scope={{ type: "project", projectId, env: scopedEnv }}
        />
    );
}
