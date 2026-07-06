import type { SettingsBaseEntity } from "~/settings/domain";

export interface SettingAccessToken extends SettingsBaseEntity {
    kind?: string;
    user: string;
    token: string;
    baseURL: string;
    secretMasked?: boolean;
    inherited?: boolean;
}
