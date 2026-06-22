import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SslCertFormRoute } from "~/settings/module-shared/components/ssl-cert-form-route";

export function SettingsSslCertCreateRoute() {
    return (
        <div className={cn(listBox, "p-0")}>
            <SslCertFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}
