import type { SettingsBaseEntity } from "~/settings/domain";

export interface SettingKeyAuth extends SettingsBaseEntity {
    keyId: string;
    secretKey: string;
    secretMasked?: boolean;
}
