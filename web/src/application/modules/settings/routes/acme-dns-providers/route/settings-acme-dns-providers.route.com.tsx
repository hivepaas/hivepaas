import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SettingsAcmeDnsProviderTable } from "~/settings/module-shared/components";

export function SettingsAcmeDnsProvidersRoute() {
    return (
        <div className={cn(listBox)}>
            <SettingsAcmeDnsProviderTable />
        </div>
    );
}
