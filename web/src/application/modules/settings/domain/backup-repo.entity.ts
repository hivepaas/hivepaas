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
    cloudStorage?: { id: string; name?: string } | null;
    volume?: { id: string; name?: string } | null;
    storagePrefix?: string;
    password?: string;
    compression?: string;
    packSize?: string;
    retention?: SettingBackupRepoRetention;
    secretMasked?: boolean;
}
