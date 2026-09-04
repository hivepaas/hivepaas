import { formBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { useParams } from "react-router";
import { BackupRepoFormRoute } from "~/settings/module-shared/components/backup-repo-form-route";

export function SettingsBackupRepoEditRoute() {
    const { backupRepoId = "" } = useParams();

    return (
        <div className={cn(formBox)}>
            <BackupRepoFormRoute
                mode="edit"
                scope={{ type: "settings" }}
                backupRepoId={backupRepoId}
            />
        </div>
    );
}
