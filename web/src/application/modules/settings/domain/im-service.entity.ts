import type { EImServiceKind } from "@application/shared/enums";

import type { SettingsBaseEntity } from "./settings.base.entity";

export interface ImServiceSlack {
    webhook: string;
}

export interface ImServiceDiscord {
    webhook: string;
}

export interface SettingImService extends SettingsBaseEntity {
    kind: EImServiceKind;
    slack?: ImServiceSlack | null;
    discord?: ImServiceDiscord | null;
    secretMasked?: boolean;
    inherited?: boolean;
}
