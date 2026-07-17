import type { SettingsBaseEntity } from "~/settings/domain";

import type { ESettingStatus, ESettingType } from "@application/shared/enums";

import type { ESystemBackupCompressionFormat, ESystemBackupEncryptionFormat } from "../module-shared/enums";

export interface SystemBackupSchedule {
    interval: string;
    cronExpr: string;
    initialTime?: Date | null;
}

export interface SystemBackupCompression {
    format: ESystemBackupCompressionFormat;
}

export interface SystemBackupEncryption {
    format: ESystemBackupEncryptionFormat;
    secret: string;
}

export interface SystemBackupCloudStorage extends SettingsBaseEntity {
    bucket?: string;
    destinationDir: string;
}

export interface SystemBackupDBConfig {
    backupDeletedObjects: boolean;
}

export interface SystemBackupNotification {
    success?: {
        id: string;
        name: string;
    };
    successUseDefault: boolean;
    failure?: {
        id: string;
        name: string;
    };
    failureUseDefault: boolean;
}

export interface SystemBackupSettings extends SettingsBaseEntity {
    type: typeof ESettingType.SystemBackup;
    status: ESettingStatus;
    schedule: SystemBackupSchedule;
    compression: SystemBackupCompression;
    encryption: SystemBackupEncryption;
    cloudStorage?: SystemBackupCloudStorage | null;
    dbBackupConfig: SystemBackupDBConfig;
    notification?: SystemBackupNotification | null;
    nextRuns: Date[];
}
