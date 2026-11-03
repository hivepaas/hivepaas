import { useParams } from "react-router";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";
import { AcmeDnsProviderFormRoute } from "~/settings/module-shared/components/acme-dns-provider-form-route";

export function ProjectAcmeDnsProviderEditRoute() {
    const { acmeDnsProviderId = "", id: projectId = "" } = useParams();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <AcmeDnsProviderFormRoute
            mode="edit"
            scope={{ type: "project", projectId, env: scopedEnv }}
            acmeDnsProviderId={acmeDnsProviderId}
        />
    );
}
