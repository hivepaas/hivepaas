import type { SettingsBaseEntity } from "~/settings/domain";

export interface SettingBackupRepoRetention {
    keepLast?: number;
    keepHourly?: number;
    keepDaily?: number;
    keepWeekly?: number;
    keepMonthly?: number;
}

export interface SettingBackupRepo extends SettingsBaseEntity {
    engine?: string;
    description?: string;
    storagePrefix?: string;
    compression?: string;
    packSize?: string;
    retention?: SettingBackupRepoRetention;
    secretMasked?: boolean;
}
