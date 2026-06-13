import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SettingsSslProviderTable } from "~/settings/module-shared/components";

export function SettingsSslProvidersRoute() {
    return (
        <div className={cn(listBox)}>
            <SettingsSslProviderTable />
        </div>
    );
}
