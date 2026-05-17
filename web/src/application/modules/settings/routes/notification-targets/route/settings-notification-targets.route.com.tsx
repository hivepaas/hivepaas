import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SettingsNotificationTargetTable } from "~/settings/module-shared/components";

export function SettingsNotificationTargetsRoute() {
    return (
        <div className={cn(listBox)}>
            <SettingsNotificationTargetTable />
        </div>
    );
}
