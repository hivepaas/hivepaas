import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export interface ProjectSettingsImportSettingPayload {
    id: string;
}

export interface ProjectSettingsImport_Import_Payload {
    settings: ProjectSettingsImportSettingPayload[];
    dataViewAllowed: boolean;
}

export type ProjectSettingsImport_Import_Req = ApiRequestBase<{
    projectID: string;
    payload: ProjectSettingsImport_Import_Payload;
}>;

export type ProjectSettingsImport_Import_Res = ApiResponseBase<{ type: "success" }>;
