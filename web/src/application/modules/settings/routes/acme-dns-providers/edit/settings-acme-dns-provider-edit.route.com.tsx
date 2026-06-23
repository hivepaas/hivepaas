import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { AcmeDnsProviderFormRoute } from "~/settings/module-shared/components/acme-dns-provider-form-route";

export function SettingsAcmeDnsProviderEditRoute() {
    const { acmeDnsProviderId = "" } = useParams();

    return (
        <div className={cn(listBox)}>
            <AcmeDnsProviderFormRoute
                mode="edit"
                scope={{ type: "settings" }}
                acmeDnsProviderId={acmeDnsProviderId}
            />
        </div>
    );
}
