import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SettingsKeyAuthTable } from "~/settings/module-shared/components";

export function SettingsKeyAuthRoute() {
    return (
        <div className={cn(listBox)}>
            <SettingsKeyAuthTable />
        </div>
    );
}
