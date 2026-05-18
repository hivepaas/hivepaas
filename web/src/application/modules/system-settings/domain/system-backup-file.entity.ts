import type { SettingsBaseEntity } from "~/settings/domain";

import type { ESettingStatus, ESettingType } from "@application/shared/enums";

import type { ESystemBackupFileStorageType } from "../module-shared/enums";

export interface SystemBackupFileStorage extends SettingsBaseEntity {
    status: ESettingStatus;
}

export interface SystemBackupFile extends SettingsBaseEntity {
    type: typeof ESettingType.File;
    status: ESettingStatus;
    storageType: ESystemBackupFileStorageType;
    storage?: SystemBackupFileStorage | null;
    bucket?: string;
    mimetype: string;
    size: number;
    path: string;
}
