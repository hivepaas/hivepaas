import type { PaginationState, SortingState } from "@infrastructure/data";
import type {
    CloudStorage_CreateOne_Payload,
    CloudStorage_UpdateMeta_Payload,
    CloudStorage_UpdateOne_Payload,
} from "~/settings/api/services";
import type { SettingCloudStorage } from "~/settings/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectCloudStorage_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectCloudStorage_FindManyPaginated_Res = ApiResponsePaginated<SettingCloudStorage>;

export type ProjectCloudStorage_FindOneById_Req = ApiRequestBase<{ projectID: string; id: string }>;

export type ProjectCloudStorage_FindOneById_Res = ApiResponseBase<SettingCloudStorage>;

export type ProjectCloudStorage_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: CloudStorage_CreateOne_Payload;
}>;

export type ProjectCloudStorage_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type ProjectCloudStorage_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: CloudStorage_UpdateOne_Payload;
}>;

export type ProjectCloudStorage_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectCloudStorage_UpdateMeta_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: CloudStorage_UpdateMeta_Payload;
}>;

export type ProjectCloudStorage_UpdateMeta_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectCloudStorage_DeleteOne_Req = ApiRequestBase<{ projectID: string; id: string }>;

export type ProjectCloudStorage_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
