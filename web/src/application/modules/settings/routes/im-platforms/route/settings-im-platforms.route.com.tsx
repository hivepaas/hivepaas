import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SettingsImPlatformTable } from "~/settings/module-shared/components";

export function SettingsImPlatformsRoute() {
    return (
        <div className={cn(listBox)}>
            <SettingsImPlatformTable />
        </div>
    );
}
