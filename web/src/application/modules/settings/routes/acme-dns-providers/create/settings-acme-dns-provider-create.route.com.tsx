import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { AcmeDnsProviderFormRoute } from "~/settings/module-shared/components/acme-dns-provider-form-route";

export function SettingsAcmeDnsProviderCreateRoute() {
    return (
        <div className={cn(listBox)}>
            <AcmeDnsProviderFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}
