import type { SettingsBaseEntity } from "~/settings/domain";

export interface SettingRegistryAuth extends SettingsBaseEntity {
    address: string;
    username: string;
    password: string;
    readonly: boolean;
    secretMasked?: boolean;
}
