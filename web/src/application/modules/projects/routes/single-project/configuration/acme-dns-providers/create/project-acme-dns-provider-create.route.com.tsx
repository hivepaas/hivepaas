import { useParams } from "react-router";
import { AcmeDnsProviderFormRoute } from "~/settings/module-shared/components/acme-dns-provider-form-route";

export function ProjectAcmeDnsProviderCreateRoute() {
    const { id: projectId = "" } = useParams();

    return (
        <AcmeDnsProviderFormRoute
            mode="create"
            scope={{ type: "project", projectId }}
        />
    );
}
