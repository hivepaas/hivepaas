import type { SettingsBaseEntity } from "~/settings/domain";

export type AppFeatureToggleSettings = {
    enabled: boolean;
};

export type AppFeaturePreviewAppRef = {
    id: string;
    name: string;
    photo?: string;
    key?: string;
    status?: string;
    env?: string;
};

export type AppFeaturePreviewSettings = {
    enabled: boolean;
    creationDelay: string;
    appsToClone: AppFeaturePreviewAppRef[];
    autoCloneApps: boolean;
};

export interface AppFeatureSettings extends SettingsBaseEntity {
    loggingSettings: AppFeatureToggleSettings;
    schedJobSettings: AppFeatureToggleSettings;
    terminalSettings: AppFeatureToggleSettings;
    previewSettings: AppFeaturePreviewSettings;
}
