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

export type AppFeaturePreviewCommandRef = {
    id: string;
    name: string;
    type?: string;
};

export type AppFeaturePreviewSettings = {
    enabled: boolean;
    creationDelay: string;
    appsToClone: AppFeaturePreviewAppRef[];
    autoCloneApps: boolean;
    commands: AppFeaturePreviewCommandRef[];
};

export interface AppFeatureSettings extends SettingsBaseEntity {
    loggingSettings: AppFeatureToggleSettings;
    schedJobSettings: AppFeatureToggleSettings;
    terminalSettings: AppFeatureToggleSettings;
    previewSettings: AppFeaturePreviewSettings;
}
