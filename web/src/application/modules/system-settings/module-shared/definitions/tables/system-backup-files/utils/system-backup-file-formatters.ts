import type { SystemBackupFile } from "~/system-settings/domain";
import { ESystemBackupFileStorageType } from "~/system-settings/module-shared/enums";

export function getBackupFileStorageLabel(file: Pick<SystemBackupFile, "storageType" | "storage">): string {
    if (file.storageType === ESystemBackupFileStorageType.Local) {
        return "local";
    }

    return file.storage?.name ?? "-";
}
