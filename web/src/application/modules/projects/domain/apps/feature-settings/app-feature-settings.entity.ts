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
    /** Lets a pull request's comments run /hivepaas deploy and /hivepaas cancel. */
    allowPRComments: boolean;
    commands: AppFeaturePreviewCommandRef[];
};

export interface AppFeatureSettings extends SettingsBaseEntity {
    loggingSettings: AppFeatureToggleSettings;
    schedJobSettings: AppFeatureToggleSettings;
    terminalSettings: AppFeatureToggleSettings;
    previewSettings: AppFeaturePreviewSettings;
    /** The app's routes and calls, measured by OBI on the nodes that run it. Off by default. */
    performanceSettings: AppFeatureToggleSettings;
}
