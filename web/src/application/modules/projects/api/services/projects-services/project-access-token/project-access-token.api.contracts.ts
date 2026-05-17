import type { PaginationState, SortingState } from "@infrastructure/data";
import type {
    AccessToken_CreateOne_Payload,
    AccessToken_UpdateMeta_Payload,
    AccessToken_UpdateOne_Payload,
} from "~/settings/api/services";
import type { SettingAccessToken } from "~/settings/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectAccessToken_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectAccessToken_FindManyPaginated_Res = ApiResponsePaginated<SettingAccessToken>;

export type ProjectAccessToken_FindOneById_Req = ApiRequestBase<{ projectID: string; id: string }>;

export type ProjectAccessToken_FindOneById_Res = ApiResponseBase<SettingAccessToken>;

export type ProjectAccessToken_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: AccessToken_CreateOne_Payload;
}>;

export type ProjectAccessToken_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type ProjectAccessToken_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: AccessToken_UpdateOne_Payload;
}>;

export type ProjectAccessToken_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectAccessToken_UpdateMeta_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: AccessToken_UpdateMeta_Payload;
}>;

export type ProjectAccessToken_UpdateMeta_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectAccessToken_DeleteOne_Req = ApiRequestBase<{ projectID: string; id: string }>;

export type ProjectAccessToken_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
