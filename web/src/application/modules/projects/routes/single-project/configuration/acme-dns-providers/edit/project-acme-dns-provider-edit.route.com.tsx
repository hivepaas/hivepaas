import { useParams } from "react-router";
import { AcmeDnsProviderFormRoute } from "~/settings/module-shared/components/acme-dns-provider-form-route";

export function ProjectAcmeDnsProviderEditRoute() {
    const { acmeDnsProviderId = "", id: projectId = "" } = useParams();

    return (
        <AcmeDnsProviderFormRoute
            mode="edit"
            scope={{ type: "project", projectId }}
            acmeDnsProviderId={acmeDnsProviderId}
        />
    );
}
