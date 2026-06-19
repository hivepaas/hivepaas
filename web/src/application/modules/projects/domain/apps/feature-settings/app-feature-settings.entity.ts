import type { SettingsBaseEntity } from "~/settings/domain";

export type AppFeatureToggleSettings = {
    enabled: boolean;
};

export interface AppFeatureSettings extends SettingsBaseEntity {
    loggingSettings: AppFeatureToggleSettings;
    schedJobSettings: AppFeatureToggleSettings;
    terminalSettings: AppFeatureToggleSettings;
}
