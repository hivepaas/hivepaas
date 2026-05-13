import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SettingsSslCertTable } from "~/settings/module-shared/components";

export function SettingsSslCertificatesRoute() {
    return (
        <div className={cn(listBox)}>
            <SettingsSslCertTable />
        </div>
    );
}
