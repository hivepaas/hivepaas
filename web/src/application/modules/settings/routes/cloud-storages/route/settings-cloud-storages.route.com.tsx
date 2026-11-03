import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SettingsCloudStorageTable } from "~/settings/module-shared/components";

export function SettingsCloudStoragesRoute() {
    return (
        <div className={cn(listBox)}>
            <SettingsCloudStorageTable />
        </div>
    );
}
