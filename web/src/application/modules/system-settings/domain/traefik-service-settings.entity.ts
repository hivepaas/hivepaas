import type { SettingsBaseEntity } from "~/settings/domain";

import type { ESettingType } from "@application/shared/enums";

export interface TraefikAppSettings {
    replicas: number;
}

export interface TraefikServiceSettings extends SettingsBaseEntity {
    type: typeof ESettingType.TraefikService;
    appSettings: TraefikAppSettings;
}
