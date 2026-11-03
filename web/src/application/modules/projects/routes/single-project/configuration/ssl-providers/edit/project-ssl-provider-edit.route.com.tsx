import { useParams } from "react-router";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";
import { SslProviderFormRoute } from "~/settings/module-shared/components/ssl-provider-form-route";

export function ProjectSslProviderEditRoute() {
    const { id: projectId = "", sslProviderId = "" } = useParams();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <SslProviderFormRoute
            mode="edit"
            scope={{ type: "project", projectId, env: scopedEnv }}
            sslProviderId={sslProviderId}
        />
    );
}
