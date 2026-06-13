import type { ESslKeyType, ESslProviderKind } from "@application/shared/enums";

import type { SettingsBaseEntity } from "./settings.base.entity";

export type SettingSslProviderLetsEncrypt = Record<string, never>;

export interface SettingSslProviderEab {
    eabKid: string;
    eabHmacKey: string;
}

export interface SettingSslProvider extends SettingsBaseEntity {
    kind: ESslProviderKind;
    email: string;
    defaultKeyType: ESslKeyType | "";
    letsEncrypt?: SettingSslProviderLetsEncrypt | null;
    zeroSSL?: SettingSslProviderEab | null;
    googleTS?: SettingSslProviderEab | null;
    secretMasked?: boolean;
}
