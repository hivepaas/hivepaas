import type { ESslCertType, ESslKeyType } from "@application/shared/enums";

import type { SettingsBaseEntity } from "./settings.base.entity";

export interface SslCertEventNotification {
    success?: { id: string; name: string } | null;
    successUseDefault: boolean;
    failure?: { id: string; name: string } | null;
    failureUseDefault: boolean;
}

export interface SettingSslCert extends SettingsBaseEntity {
    certType: ESslCertType;
    domain: string;
    certificate: string;
    privateKey: string;
    keyType: ESslKeyType;
    /** Duration string from API (e.g. `720h`, `30d`). */
    validPeriod: string;
    email: string;
    autoRenew: boolean;
    renewableFrom?: Date | null;
    notifyFrom?: Date | null;
    notification?: SslCertEventNotification | null;
    secretMasked?: boolean;
}
