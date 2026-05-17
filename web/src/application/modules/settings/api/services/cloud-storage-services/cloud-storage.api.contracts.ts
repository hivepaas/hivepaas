import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingCloudStorage, SettingCloudStorageS3 } from "~/settings/domain";

import type { ECloudStorageKind, ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type CloudStorage_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type CloudStorage_FindManyPaginated_Res = ApiResponsePaginated<SettingCloudStorage>;

export type CloudStorage_FindOneById_Req = ApiRequestBase<{ id: string }>;

export type CloudStorage_FindOneById_Res = ApiResponseBase<SettingCloudStorage>;

export type CloudStorage_CreateOne_Payload = {
    availableInProjects: boolean;
    default: boolean;
    kind: ECloudStorageKind;
    name: string;
    s3: SettingCloudStorageS3;
};

export type CloudStorage_CreateOne_Req = ApiRequestBase<{ payload: CloudStorage_CreateOne_Payload }>;

export type CloudStorage_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type CloudStorage_UpdateOne_Payload = CloudStorage_CreateOne_Payload & {
    updateVer: number;
};

export type CloudStorage_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: CloudStorage_UpdateOne_Payload;
}>;

export type CloudStorage_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type CloudStorage_UpdateMeta_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    availableInProjects?: boolean;
    default?: boolean;
};

export type CloudStorage_UpdateMeta_Req = ApiRequestBase<{
    id: string;
    payload: CloudStorage_UpdateMeta_Payload;
}>;

export type CloudStorage_UpdateMeta_Res = ApiResponseBase<{ type: "success" }>;

export type CloudStorage_DeleteOne_Req = ApiRequestBase<{ id: string }>;

export type CloudStorage_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;

export type CloudStorage_TestConn_Payload = {
    kind: ECloudStorageKind;
    name: string;
    s3: SettingCloudStorageS3;
};

export type CloudStorage_TestConn_Req = ApiRequestBase<{ payload: CloudStorage_TestConn_Payload }>;

export type CloudStorage_TestConn_Res = ApiResponseBase<{ type: "success" }>;
