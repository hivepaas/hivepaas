import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SettingsSSHKeyTable } from "~/settings/module-shared/components";

export function SettingsSSHKeysRoute() {
    return (
        <div className={cn(listBox)}>
            <SettingsSSHKeyTable />
        </div>
    );
}
