import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SettingsEmailAccountsTable } from "~/settings/module-shared/components";

export function SettingsEmailAccountsRoute() {
    return (
        <div className={cn(listBox)}>
            <SettingsEmailAccountsTable />
        </div>
    );
}
