import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { SettingsBackupRepoTable } from "~/settings/module-shared/components";

export function SettingsBackupReposRoute() {
    return (
        <div className={cn(listBox)}>
            <SettingsBackupRepoTable />
        </div>
    );
}
