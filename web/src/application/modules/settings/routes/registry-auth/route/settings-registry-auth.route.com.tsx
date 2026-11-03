import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SettingsRegistryAuthTable } from "~/settings/module-shared/components";

export function SettingsRegistryAuthRoute() {
    return (
        <div className={cn(listBox)}>
            <SettingsRegistryAuthTable />
        </div>
    );
}
