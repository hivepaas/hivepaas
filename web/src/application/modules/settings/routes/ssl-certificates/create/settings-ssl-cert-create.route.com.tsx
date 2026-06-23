import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SslCertFormRoute } from "~/settings/module-shared/components/ssl-cert-form-route";

export function SettingsSslCertCreateRoute() {
    return (
        <div className={cn(listBox)}>
            <SslCertFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}
