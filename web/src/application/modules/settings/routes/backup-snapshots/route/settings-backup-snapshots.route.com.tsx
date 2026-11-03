import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { BackupSnapshotTable } from "~/settings/module-shared/components";

export function SettingsBackupSnapshotsRoute() {
    return (
        <div className={cn(listBox)}>
            <BackupSnapshotTable scope={{ type: "settings" }} />
        </div>
    );
}
