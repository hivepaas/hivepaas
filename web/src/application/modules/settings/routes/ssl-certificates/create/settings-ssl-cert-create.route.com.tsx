import { formBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SslCertFormRoute } from "~/settings/module-shared/components/ssl-cert-form-route";

export function SettingsSslCertCreateRoute() {
    return (
        <div className={cn(formBox)}>
            <SslCertFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}
