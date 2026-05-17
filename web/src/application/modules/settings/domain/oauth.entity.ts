import type { SettingsBaseEntity } from "~/settings/domain";

import type { EOAuthKind } from "@application/shared/enums";

export interface SettingOAuth extends SettingsBaseEntity {
    kind?: EOAuthKind | (string & {});
    organization: string;
    clientId: string;
    clientSecret: string;
    callbackURL?: string;
    authURL?: string;
    tokenURL?: string;
    profileURL?: string;
    autoDiscoveryURL?: string;
    scopes?: string[];
    secretMasked?: boolean;
}
