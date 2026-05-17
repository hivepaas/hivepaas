import type { SettingsBaseEntity } from "~/settings/domain";

export interface SettingSSHKey extends SettingsBaseEntity {
    privateKey: string;
    passphrase?: string;
    targets?: string[];
    secretMasked?: boolean;
    inherited?: boolean;
}
