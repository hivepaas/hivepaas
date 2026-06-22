import { useParams } from "react-router";
import { SslProviderFormRoute } from "~/settings/module-shared/components/ssl-provider-form-route";

export function ProjectSslProviderEditRoute() {
    const { sslProviderId = "", id: projectId = "" } = useParams();

    return (
        <SslProviderFormRoute
            mode="edit"
            scope={{ type: "project", projectId }}
            sslProviderId={sslProviderId}
        />
    );
}
