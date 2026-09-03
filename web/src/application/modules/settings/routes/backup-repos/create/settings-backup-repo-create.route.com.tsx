import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { BackupRepoFormRoute } from "~/settings/module-shared/components/backup-repo-form-route";

export function SettingsBackupRepoCreateRoute() {
    return (
        <div className={cn(listBox)}>
            <BackupRepoFormRoute
                mode="create"
                scope={{ type: "settings" }}
            />
        </div>
    );
}
