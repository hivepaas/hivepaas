import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SslProviderFormRoute } from "~/settings/module-shared/components/ssl-provider-form-route";

export function SettingsSslProviderCreateRoute() {
    return (
        <div className={cn(listBox)}>
            <SslProviderFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}
