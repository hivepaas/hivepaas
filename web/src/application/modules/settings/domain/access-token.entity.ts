import type { SettingsBaseEntity } from "~/settings/domain";

import type { EAccessTokenKind } from "@application/shared/enums";

export interface SettingAccessToken extends SettingsBaseEntity {
    kind?: EAccessTokenKind;
    user: string;
    token: string;
    baseURL: string;
    secretMasked?: boolean;
    inherited?: boolean;
}
