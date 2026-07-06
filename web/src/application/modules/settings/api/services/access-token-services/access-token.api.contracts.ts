import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingAccessToken } from "~/settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type AccessToken_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type AccessToken_FindManyPaginated_Res = ApiResponsePaginated<SettingAccessToken>;

export type AccessToken_FindOneById_Req = ApiRequestBase<{ id: string }>;

export type AccessToken_FindOneById_Res = ApiResponseBase<SettingAccessToken>;

export type AccessToken_CreateOne_Payload = {
    availableInProjects: boolean;
    default: boolean;
    expireAt: Date | null;
    kind: string;
    name: string;
    user: string;
    token: string;
    baseURL: string;
};

export type AccessToken_CreateOne_Req = ApiRequestBase<{ payload: AccessToken_CreateOne_Payload }>;

export type AccessToken_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type AccessToken_UpdateOne_Payload = AccessToken_CreateOne_Payload & {
    updateVer: number;
};

export type AccessToken_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: AccessToken_UpdateOne_Payload;
}>;

export type AccessToken_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type AccessToken_UpdateMeta_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    availableInProjects?: boolean;
    default?: boolean;
};

export type AccessToken_UpdateMeta_Req = ApiRequestBase<{
    id: string;
    payload: AccessToken_UpdateMeta_Payload;
}>;

export type AccessToken_UpdateMeta_Res = ApiResponseBase<{ type: "success" }>;

export type AccessToken_DeleteOne_Req = ApiRequestBase<{ id: string }>;

export type AccessToken_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;

export type AccessToken_TestConn_Payload = {
    expireAt: Date | null;
    kind: string;
    name: string;
    user: string;
    token: string;
    baseURL: string;
};

export type AccessToken_TestConn_Req = ApiRequestBase<{ payload: AccessToken_TestConn_Payload }>;

export type AccessToken_TestConn_Res = ApiResponseBase<{ type: "success" }>;
