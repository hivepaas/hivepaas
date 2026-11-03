import type { SettingsBaseEntity } from "~/settings/domain";

export interface SettingBasicAuth extends SettingsBaseEntity {
    username: string;
    password: string;
    secretMasked?: boolean;
}
