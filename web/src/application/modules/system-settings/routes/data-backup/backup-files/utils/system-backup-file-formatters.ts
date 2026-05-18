import type { SystemBackupFile } from "~/system-settings/domain";
import { ESystemBackupFileStorageType } from "~/system-settings/module-shared/enums";

export function formatBackupFileSize(size: number): string {
    if (!Number.isFinite(size) || size < 0) {
        return "-";
    }

    if (size === 0) {
        return "0 B";
    }

    const units = ["B", "KB", "MB", "GB", "TB"];
    const unitIndex = Math.min(Math.floor(Math.log(size) / Math.log(1024)), units.length - 1);
    const value = size / 1024 ** unitIndex;

    return `${value.toFixed(value >= 10 || unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`;
}

export function getBackupFileStorageLabel(file: Pick<SystemBackupFile, "storageType" | "storage">): string {
    if (file.storageType === ESystemBackupFileStorageType.Local) {
        return "local";
    }

    return file.storage?.name ?? "-";
}
