import { useParams } from "react-router";
import { SslCertFormRoute } from "~/settings/module-shared/components/ssl-cert-form-route";

export function ProjectSslCertEditRoute() {
    const { sslCertId = "", id: projectId = "" } = useParams();

    return (
        <SslCertFormRoute
            mode="edit"
            scope={{ type: "project", projectId }}
            sslCertId={sslCertId}
        />
    );
}
