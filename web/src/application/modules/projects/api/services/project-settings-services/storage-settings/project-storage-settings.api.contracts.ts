import type { ProjectStorageSettings } from "~/projects/domain";
import type {
    StorageSettings_UpdateOne_Payload,
    StorageSettings_UpdateStatus_Payload,
} from "~/settings/api/services/storage-settings-services";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type ProjectStorageSettings_FindOne_Req = ApiRequestBase<{
    projectID: string;
}>;
export type ProjectStorageSettings_FindOne_Res = ApiResponseBase<ProjectStorageSettings>;

export type ProjectStorageSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: StorageSettings_UpdateOne_Payload;
}>;
export type ProjectStorageSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectStorageSettings_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    payload: StorageSettings_UpdateStatus_Payload;
}>;
export type ProjectStorageSettings_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectStorageSettings_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
}>;
export type ProjectStorageSettings_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
