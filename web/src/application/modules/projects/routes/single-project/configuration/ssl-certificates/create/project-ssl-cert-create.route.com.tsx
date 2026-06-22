import { useParams } from "react-router";
import { SslCertFormRoute } from "~/settings/module-shared/components/ssl-cert-form-route";

export function ProjectSslCertCreateRoute() {
    const { id: projectId = "" } = useParams();

    return (
        <SslCertFormRoute
            mode="create"
            scope={{ type: "project", projectId }}
        />
    );
}
