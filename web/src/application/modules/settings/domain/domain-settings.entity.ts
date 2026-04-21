import type { ESslCertType, ESslKeyType } from "@application/shared/enums";

import type { SettingsBaseEntity } from "./settings.base.entity";

export interface DomainCertSettings {
    certType: ESslCertType;
    keyType: ESslKeyType;
    /** Duration string from API (e.g. `720h`, `30d`). */
    validPeriod: string;
    email: string;
    autoRenew: boolean;
}

export interface SettingDomainSettings extends SettingsBaseEntity {
    rootDomain: string;
    certSettings?: DomainCertSettings | null;
}
