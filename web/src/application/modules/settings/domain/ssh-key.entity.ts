import type { SettingsBaseEntity } from "~/settings/domain";

import type { ESSHKeyType } from "@application/shared/enums";

export interface SettingSSHKey extends SettingsBaseEntity {
    keyType?: ESSHKeyType | "";
    publicKey?: string;
    privateKey: string;
    passphrase?: string;
    secretMasked?: boolean;
    inherited?: boolean;
}
