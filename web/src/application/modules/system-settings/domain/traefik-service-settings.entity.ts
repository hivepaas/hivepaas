import type { SettingsBaseEntity } from "~/settings/domain";

export interface TraefikAppSettings {
    replicas: number;
}

export interface TraefikServiceSettings extends SettingsBaseEntity {
    appSettings: TraefikAppSettings;
}
