import { useParams } from "react-router";
import { SslProviderFormRoute } from "~/settings/module-shared/components/ssl-provider-form-route";

export function ProjectSslProviderCreateRoute() {
    const { id: projectId = "" } = useParams();

    return (
        <SslProviderFormRoute
            mode="create"
            scope={{ type: "project", projectId }}
        />
    );
}
