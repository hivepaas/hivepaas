import type { SettingStorageSettings } from "~/settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type StorageSettings_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type StorageSettings_FindOne_Res = ApiResponseBase<SettingStorageSettings>;

export type StorageSettings_UpdateOne_Payload = {
    updateVer: number;
    bindSettings?: SettingStorageSettings["bindSettings"];
    volumeSettings?: SettingStorageSettings["volumeSettings"];
    clusterVolumeSettings?: SettingStorageSettings["clusterVolumeSettings"];
    tmpfsSettings?: SettingStorageSettings["tmpfsSettings"];
};

export type StorageSettings_UpdateOne_Req = ApiRequestBase<{
    payload: StorageSettings_UpdateOne_Payload;
}>;
export type StorageSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type StorageSettings_UpdateStatus_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    availableInProjects?: boolean;
    default?: boolean;
};

export type StorageSettings_UpdateStatus_Req = ApiRequestBase<{
    payload: StorageSettings_UpdateStatus_Payload;
}>;
export type StorageSettings_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type StorageSettings_DeleteOne_Req = ApiRequestBase<Record<string, never>>;
export type StorageSettings_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
