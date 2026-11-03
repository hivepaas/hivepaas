import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SettingsBasicAuthTable } from "~/settings/module-shared/components";

export function SettingsBasicAuthRoute() {
    return (
        <div className={cn(listBox)}>
            <SettingsBasicAuthTable />
        </div>
    );
}
