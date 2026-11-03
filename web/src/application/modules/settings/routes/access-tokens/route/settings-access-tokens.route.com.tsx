import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SettingsAccessTokenTable } from "~/settings/module-shared/components";

export function SettingsAccessTokensRoute() {
    return (
        <div className={cn(listBox)}>
            <SettingsAccessTokenTable />
        </div>
    );
}
