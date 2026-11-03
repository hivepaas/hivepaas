import { BackupSnapshotTable, SYSTEM_BACKUP_SOURCE_TAG } from "~/settings/module-shared/components";

const SYSTEM_BACKUP_TAGS = [SYSTEM_BACKUP_SOURCE_TAG];

/** The system backup's snapshots: those of every repository, taken by a system backup. */
export function SettingsDataBackupSnapshotsRoute() {
    return (
        <BackupSnapshotTable
            scope={{ type: "settings" }}
            fixedTags={SYSTEM_BACKUP_TAGS}
        />
    );
}
