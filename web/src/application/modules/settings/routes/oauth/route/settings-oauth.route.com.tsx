import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SettingsOAuthTable } from "~/settings/module-shared/components";

export function SettingsOAuthRoute() {
    return (
        <div className={cn(listBox)}>
            <SettingsOAuthTable />
        </div>
    );
}
