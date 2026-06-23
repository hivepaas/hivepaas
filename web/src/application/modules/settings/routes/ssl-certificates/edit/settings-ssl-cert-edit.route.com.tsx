import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { SslCertFormRoute } from "~/settings/module-shared/components/ssl-cert-form-route";

export function SettingsSslCertEditRoute() {
    const { sslCertId = "" } = useParams();

    return (
        <div className={cn(listBox)}>
            <SslCertFormRoute
                mode="edit"
                scope={{ type: "settings" }}
                sslCertId={sslCertId}
            />
        </div>
    );
}
