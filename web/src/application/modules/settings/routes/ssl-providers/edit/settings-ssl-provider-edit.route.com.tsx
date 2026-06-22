import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { SslProviderFormRoute } from "~/settings/module-shared/components/ssl-provider-form-route";

export function SettingsSslProviderEditRoute() {
    const { sslProviderId = "" } = useParams();

    return (
        <div className={cn(listBox, "p-0")}>
            <SslProviderFormRoute
                mode="edit"
                scope={{ type: "settings" }}
                sslProviderId={sslProviderId}
            />
        </div>
    );
}
