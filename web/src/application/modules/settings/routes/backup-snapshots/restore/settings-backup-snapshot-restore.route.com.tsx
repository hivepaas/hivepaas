import { formBox } from "@lib/styles";
import { useParams } from "react-router";
import { BackupSnapshotRestoreRoute } from "~/settings/module-shared/components";

export function SettingsBackupSnapshotRestoreRoute() {
    const { snapshotId = "" } = useParams();

    return (
        <div className={formBox}>
            <BackupSnapshotRestoreRoute
                scope={{ type: "settings" }}
                snapshotId={snapshotId}
            />
        </div>
    );
}
