import { useParams } from "react-router";
import { getProjectEnvFilterParam, useSelectedProjectEnv } from "~/projects/module-shared/hooks";
import { SslCertFormRoute } from "~/settings/module-shared/components/ssl-cert-form-route";

export function ProjectSslCertEditRoute() {
    const { id: projectId = "", sslCertId = "" } = useParams();
    const selectedEnv = useSelectedProjectEnv(projectId);
    const scopedEnv = getProjectEnvFilterParam(selectedEnv);

    return (
        <SslCertFormRoute
            mode="edit"
            scope={{ type: "project", projectId, env: scopedEnv }}
            sslCertId={sslCertId}
        />
    );
}
