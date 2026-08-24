import type { AppFeatureSettings, AppFeatureToggleSettings } from "~/projects/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type AppFeatureSettings_FindOne_Req = ApiRequestBase<{ projectID: string; env: string; appID: string }>;
export type AppFeatureSettings_FindOne_Res = ApiResponseBase<AppFeatureSettings>;

export type AppFeaturePreviewSettingsUpdatePayload = {
    enabled: boolean;
    creationDelay: string;
    appsToClone: { id: string }[];
    autoCloneApps: boolean;
    commands: { id: string }[];
};

export type AppFeatureSettings_UpdatePayload = {
    inheritable: boolean;
    default: boolean;
    updateVer: number;
    loggingSettings: AppFeatureToggleSettings;
    schedJobSettings: AppFeatureToggleSettings;
    terminalSettings: AppFeatureToggleSettings;
    previewSettings: AppFeaturePreviewSettingsUpdatePayload;
};

export type AppFeatureSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    payload: AppFeatureSettings_UpdatePayload;
}>;
export type AppFeatureSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
